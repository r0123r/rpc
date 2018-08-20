// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// r0123r - update for Extjs Direct rpc
package json

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/rpc"
)

var null = json.RawMessage([]byte("null"))

// ----------------------------------------------------------------------------
// Request and Response
// ----------------------------------------------------------------------------

// serverRequest represents a JSON-RPC request received by the server.
type serverRequest struct {
	// A String containing the name of the method to be invoked.
	Method string `json:"method"`
	// An Array of objects to pass as arguments to the method.
	Params *json.RawMessage `json:"data"`
	// The request id. This can be of any type. It is used to match the
	// response with the request that it is replying to.
	Id     *json.RawMessage `json:"tid"`
	Type   string           `json:"type"`
	Action string           `json:"action"`
}

// serverResponse represents a JSON-RPC response returned by the server.
type serverResponse struct {
	// The Object that was returned by the invoked method. This must be null
	// in case there was an error invoking the method.
	Result interface{} `json:"result"`
	// This must be the same id as the request it is responding to.
	Id     *json.RawMessage `json:"tid"`
	Type   string           `json:"type"`
	Action string           `json:"action"`
	Method string           `json:"method"`
}
type serverErrorResponse struct {
	// An Error object if there was an error invoking the method. It must be
	// null if there was no error.
	Error interface{} `json:"message"`
	// This must be the same id as the request it is responding to.
	Id     *json.RawMessage `json:"tid"`
	Type   string           `json:"type"`
	Action string           `json:"action"`
	Method string           `json:"method"`
}

// ----------------------------------------------------------------------------
// Codec
// ----------------------------------------------------------------------------

// NewCodec returns a new JSON Codec.
func NewCodec() *Codec {
	return &Codec{}
}

// Codec creates a CodecRequest to process each request.
type Codec struct {
}

// NewRequest returns a CodecRequest.
func (c *Codec) NewRequest(r *http.Request) rpc.CodecRequest {
	return newCodecRequest(r)
}

// ----------------------------------------------------------------------------
// CodecRequest
// ----------------------------------------------------------------------------

// newCodecRequest returns a new CodecRequest.
func newCodecRequest(r *http.Request) rpc.CodecRequest {
	// Decode the request body and check if RPC method is valid.
	req := new(serverRequest)
	err := json.NewDecoder(r.Body).Decode(req)
	r.Body.Close()
	return &CodecRequest{request: req, err: err}
}

// CodecRequest decodes and encodes a single request.
type CodecRequest struct {
	request *serverRequest
	err     error
}

// Method returns the RPC method for the current request.
//
// The method uses a dotted notation as in "Service.Method".
func (c *CodecRequest) Method() (string, error) {
	if c.err == nil {
		return c.request.Action + "." + c.request.Method, nil
	}
	return "", c.err
}

// ReadRequest fills the request object for the RPC method.
func (c *CodecRequest) ReadRequest(args interface{}) error {
	if c.err == nil {
		params := [1]interface{}{args}
		if c.request.Params == nil { //ExtDirect data=null
			c.request.Params = &null
		} else {

			c.err = errors.New("rpc: method request ill-formed: missing params field")
		}

		c.err = json.Unmarshal(*c.request.Params, &params)
	}
	return c.err
}

// WriteResponse encodes the response and writes it to the ResponseWriter.
//
// The err parameter is the error resulted from calling the RPC method,
// or nil if there was no error.
func (c *CodecRequest) WriteResponse(w http.ResponseWriter, reply interface{}, methodErr error) error {
	if c.err != nil {
		return c.err
	}
	if methodErr != nil {
		res := &serverErrorResponse{
			Error:  methodErr.Error(),
			Id:     c.request.Id,
			Action: c.request.Action,
			Type:   "exception",
			Method: c.request.Method,
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(res)
	} else {
		res := &serverResponse{
			Result: reply,
			Id:     c.request.Id,
			Action: c.request.Action,
			Type:   c.request.Type,
			Method: c.request.Method,
		}
		if c.request.Id == nil {
			// Id is null for notifications and they don't have a response.
			res.Id = &null
		} else {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			encoder := json.NewEncoder(w)
			encoder.Encode(res)
		}
	}
	return nil
}
