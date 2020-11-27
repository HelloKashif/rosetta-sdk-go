// Copyright 2020 Coinbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reconciler

import (
	"context"
	"sync"
	"time"

	"github.com/HelloKashif/rosetta-sdk-go/parser"
	"github.com/HelloKashif/rosetta-sdk-go/storage"
	"github.com/HelloKashif/rosetta-sdk-go/types"
	"github.com/HelloKashif/rosetta-sdk-go/utils"
)

const (
	// ActiveReconciliation is included in the reconciliation
	// error message if reconciliation failed during active
	// reconciliation.
	ActiveReconciliation = "ACTIVE"

	// InactiveReconciliation is included in the reconciliation
	// error message if reconciliation failed during inactive
	// reconciliation.
	InactiveReconciliation = "INACTIVE"
)

const (
	// BlockGone is when the block where a reconciliation
	// is supposed to happen is orphaned.
	BlockGone = "BLOCK_GONE"

	// HeadBehind is when the synced tip (where balances
	// were last computed) is behind the *types.BlockIdentifier
	// returned by the call to /account/balance.
	HeadBehind = "HEAD_BEHIND"

	// BacklogFull is when the reconciliation backlog is full.
	BacklogFull = "BACKLOG_FULL"

	// TipFailure is returned when looking up the live
	// balance fails but we are at tip. This usually occurs
	// when the node processes an re-org that we have yet
	// to process (so the index we are querying at may be
	// ahead of the nodes tip).
	TipFailure = "TIP_FAILURE"
)

const (
	// pruneActiveReconciliation indicates if historical balances
	// should be pruned during active reconciliation.
	pruneActiveReconciliation = true

	// pruneInactiveReconciliation indicates if historical balances
	// should be pruned during inactive reconciliation.
	pruneInactiveReconciliation = false
)

const (
	// defaultBacklogSize is the limit of account lookups
	// that can be enqueued to reconcile before new
	// requests are dropped.
	defaultBacklogSize = 250000

	// waitToCheckDiff is the syncing difference (live-head)
	// to retry instead of exiting. In other words, if the
	// processed head is behind the live head by <
	// waitToCheckDiff we should try again after sleeping.
	// TODO: Make configurable
	waitToCheckDiff = 10

	// waitToCheckDiffSleep is the amount of time to wait
	// to check a balance difference if the syncer is within
	// waitToCheckDiff from the block a balance was queried at.
	waitToCheckDiffSleep = 5 * time.Second

	// zeroString is a string of value 0.
	zeroString = "0"

	// inactiveReconciliationSleep is used as the time.Duration
	// to sleep when there are no seen accounts to reconcile.
	inactiveReconciliationSleep = 1 * time.Second

	// defaultInactiveFrequency is the minimum
	// number of blocks the reconciler should wait between
	// inactive reconciliations for each account.
	defaultInactiveFrequency = 200

	// defaultReconcilerConcurrency is the number of goroutines
	// to start for reconciliation. Half of the goroutines are assigned
	// to inactive reconciliation and half are assigned to active
	// reconciliation.
	defaultReconcilerConcurrency = 8

	// safeBalancePruneDepth is the depth from the last balance
	// change that we consider safe to prune. We are very conservative
	// here to prevent removing balances we may need in a reorg.
	safeBalancePruneDepth = int64(500) // nolint:gomnd
)

// Helper functions are used by Reconciler to compare
// computed balances from a block with the balance calculated
// by the node. Defining an interface allows the client to determine
// what sort of storage layer they want to use to provide the required
// information.
type Helper interface {
	DatabaseTransaction(ctx context.Context) storage.DatabaseTransaction

	CurrentBlock(
		ctx context.Context,
		dbTx storage.DatabaseTransaction,
	) (*types.BlockIdentifier, error)

	IndexAtTip(
		ctx context.Context,
		index int64,
	) (bool, error)

	CanonicalBlock(
		ctx context.Context,
		dbTx storage.DatabaseTransaction,
		block *types.BlockIdentifier,
	) (bool, error)

	ComputedBalance(
		ctx context.Context,
		dbTx storage.DatabaseTransaction,
		account *types.AccountIdentifier,
		currency *types.Currency,
		index int64,
	) (*types.Amount, error)

	LiveBalance(
		ctx context.Context,
		account *types.AccountIdentifier,
		currency *types.Currency,
		index int64,
	) (*types.Amount, *types.BlockIdentifier, error)

	// PruneBalances is invoked by the reconciler
	// to indicate that all historical balance states
	// <= to index can be removed.
	PruneBalances(
		ctx context.Context,
		account *types.AccountIdentifier,
		currency *types.Currency,
		index int64,
	) error
}

// Handler is called by Reconciler after a reconciliation
// is performed. When a reconciliation failure is observed,
// it is up to the client to trigger a halt (by returning
// an error) or to continue (by returning nil).
type Handler interface {
	ReconciliationFailed(
		ctx context.Context,
		reconciliationType string,
		account *types.AccountIdentifier,
		currency *types.Currency,
		computedBalance string,
		liveBalance string,
		block *types.BlockIdentifier,
	) error

	ReconciliationSucceeded(
		ctx context.Context,
		reconciliationType string,
		account *types.AccountIdentifier,
		currency *types.Currency,
		balance string,
		block *types.BlockIdentifier,
	) error

	ReconciliationExempt(
		ctx context.Context,
		reconciliationType string,
		account *types.AccountIdentifier,
		currency *types.Currency,
		computedBalance string,
		liveBalance string,
		block *types.BlockIdentifier,
		exemption *types.BalanceExemption,
	) error

	ReconciliationSkipped(
		ctx context.Context,
		reconciliationType string,
		account *types.AccountIdentifier,
		currency *types.Currency,
		cause string,
	) error
}

// InactiveEntry is used to track the last
// time that an *types.AccountCurrency was reconciled.
type InactiveEntry struct {
	Entry     *types.AccountCurrency
	LastCheck *types.BlockIdentifier
}

// Reconciler contains all logic to reconcile balances of
// types.AccountIdentifiers returned in types.Operations
// by a Rosetta Server.
type Reconciler struct {
	helper  Helper
	handler Handler
	parser  *parser.Parser

	lookupBalanceByBlock bool
	interestingAccounts  []*types.AccountCurrency
	backlogSize          int
	changeQueue          chan *parser.BalanceChange
	inactiveFrequency    int64
	debugLogging         bool
	balancePruning       bool

	// Reconciler concurrency is separated between
	// active and inactive concurrency to allow for
	// fine-grained tuning of reconciler behavior.
	// When there are many transactions in a block
	// on a resource-constrained machine (laptop),
	// it is useful to allocate more resources to
	// active reconciliation as it is synchronous
	// (when lookupBalanceByBlock is enabled).
	ActiveConcurrency   int
	InactiveConcurrency int

	// highWaterMark is used to skip requests when
	// we are very far behind the live head.
	highWaterMark int64

	// seenAccounts are stored for inactive account
	// reconciliation. seenAccounts must be stored
	// separately from inactiveQueue to prevent duplicate
	// accounts from being added to the inactive reconciliation
	// queue. If this is not done, it is possible a goroutine
	// could be processing an account (not in the queue) when
	// we do a lookup to determine if we should add to the queue.
	seenAccounts  map[string]struct{}
	inactiveQueue []*InactiveEntry

	// inactiveQueueMutex needed because we can't peek at the tip
	// of a channel to determine when it is ready to look at.
	inactiveQueueMutex sync.Mutex

	// lastIndexChecked is the last block index reconciled actively.
	lastIndexMutex   sync.Mutex
	lastIndexChecked int64

	// queueMap tracks the *types.AccountCurrency items
	// in the active reconciliation queue and being actively
	// reconciled. It ensures we don't accidentally attempt to prune
	// computed balances being used by other goroutines.
	queueMap      map[string]*utils.BST
	queueMapMutex sync.Mutex
}
