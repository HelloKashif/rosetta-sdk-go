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

// Generated by: OpenAPI Generator (https://openapi-generator.tech)

package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/HelloKashif/rosetta-sdk-go/asserter"
	"github.com/HelloKashif/rosetta-sdk-go/types"
)

// A CallAPIController binds http requests to an api service and writes the service results to the
// http response
type CallAPIController struct {
	service  CallAPIServicer
	asserter *asserter.Asserter
}

// NewCallAPIController creates a default api controller
func NewCallAPIController(
	s CallAPIServicer,
	asserter *asserter.Asserter,
) Router {
	return &CallAPIController{
		service:  s,
		asserter: asserter,
	}
}

// Routes returns all of the api route for the CallAPIController
func (c *CallAPIController) Routes() Routes {
	return Routes{
		{
			"Call",
			strings.ToUpper("Post"),
			"/call",
			c.Call,
		},
	}
}

// Call - Make a Network-Specific Procedure Call
func (c *CallAPIController) Call(w http.ResponseWriter, r *http.Request) {
	callRequest := &types.CallRequest{}
	if err := json.NewDecoder(r.Body).Decode(&callRequest); err != nil {
		EncodeJSONResponse(&types.Error{
			Message: err.Error(),
		}, http.StatusInternalServerError, w)

		return
	}

	// Assert that CallRequest is correct
	if err := c.asserter.CallRequest(callRequest); err != nil {
		EncodeJSONResponse(&types.Error{
			Message: err.Error(),
		}, http.StatusInternalServerError, w)

		return
	}

	result, serviceErr := c.service.Call(r.Context(), callRequest)
	if serviceErr != nil {
		EncodeJSONResponse(serviceErr, http.StatusInternalServerError, w)

		return
	}

	EncodeJSONResponse(result, http.StatusOK, w)
}
