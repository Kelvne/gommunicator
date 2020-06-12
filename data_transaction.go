package gommunicator

import (
	"fmt"
	"strings"

	"github.com/kelvne/gommunicator/deco"
)

// DataTransactionRequest is the request object to the services cluster
type DataTransactionRequest struct {
	DedupID string `json:"dedupId"` // Prevent duplication ID

	ID      string      `json:"id"`      // DataTransaction ID
	Service string      `json:"service"` // Request to service name
	Action  string      `json:"action"`  // Action name
	Data    interface{} `json:"data"`    // Payload to be read
	Timeout int         `json:"timeout"` // Timeout policy in seconds

	IncomingService string  `json:"incomingService"` // The name of the service requesting
	ActionID        *string `json:"actionId"`        // ActionID represents the internal id for atomic internal request/response
}

// Decode is a helper method for transforming incoming data
// ALWAYS SEND A INITIALIZED POINTER!
// Ex.
// 	type Base struct {
//		Message string `json:"msg"`
// 	}
//
// 	dto := <- dataTransactionObject
// 	base := new(Base)
// 	dto.Decode(base)
func (dt *DataTransactionRequest) Decode(incoming interface{}) error {
	return deco.DecodeRequest(dt, incoming)
}

// DataTransactionResponse is the response object to the services cluster
type DataTransactionResponse struct {
	DedupID string `json:"dedupId"` // Prevent duplication ID

	ID      string      `json:"id"`      // DataTransaction ID
	Action  string      `json:"action"`  // Action name
	Success bool        `json:"success"` // True if the data transaction was successful
	Title   string      `json:"title"`   // Title of the response
	Message string      `json:"message"` // Message to be read
	Data    interface{} `json:"data"`    // Payload to be read

	ActionID *string `json:"actionId"` // ActionID represents the internal id for atomic internal request/response
}

// Decode is a helper method for transforming incoming data
// ALWAYS SEND A INITIALIZED POINTER!
// Ex.
// 	type Base struct {
//		Message string `json:"msg"`
// 	}
//
// 	dto := <- dataTransactionObject
// 	base := new(Base)
// 	dto.Decode(base)
func (dt *DataTransactionResponse) Decode(incoming interface{}) error {
	return deco.DecodeResponse(dt, incoming)
}

// DataTransaction holder for handling data transactions
type DataTransaction struct {
	id       string
	action   string
	messages []string
	data     interface{}

	actionID *string
}

// NewDataTransaction returns a new DataTransaction object
func NewDataTransaction(id, action string) *DataTransaction {
	return &DataTransaction{
		id:       id,
		action:   action,
		messages: make([]string, 0),
		data:     nil,

		actionID: nil,
	}
}

// FromResponse returns a new DataTransaction object from a DataTransactionResponse
func FromResponse(response *DataTransactionResponse) *DataTransaction {
	return &DataTransaction{
		id:       response.ID,
		action:   response.Action,
		messages: strings.Split(response.Message, "\n"),
		data:     response.Data,

		actionID: response.ActionID,
	}
}

// FromRequest returns a new DataTransaction object from a DataTransactionResponse
func FromRequest(request *DataTransactionRequest) *DataTransaction {
	return &DataTransaction{
		id:       request.ID,
		actionID: request.ActionID,
		action:   request.Action,
		data:     request.Data,
		messages: make([]string, 0),
	}
}

// SetAction sets the action of this request is representing
func (transaction *DataTransaction) SetAction(action string) *DataTransaction {
	transaction.action = action
	return transaction
}

// AddMessage adds a new message to the DataTransaction
func (transaction *DataTransaction) AddMessage(message string) *DataTransaction {
	transaction.messages = append(transaction.messages, message)
	return transaction
}

// GetMessage returns the main message of the DataTransaction
func (transaction *DataTransaction) GetMessage() string {
	message := ""

	for _, msg := range transaction.messages {
		message = fmt.Sprintf("%s\n%s", message, msg)
	}

	return message
}

// Success return a valid successful DataTransactionResponse
func (transaction *DataTransaction) Success(title string) *DataTransactionResponse {
	return &DataTransactionResponse{
		Success:  true,
		Data:     transaction.data,
		Message:  transaction.GetMessage(),
		ID:       transaction.id,
		ActionID: transaction.actionID,
		Title:    title,
		Action:   transaction.action,
	}
}

// Fail return a valid failed DataTransactionResponse
func (transaction *DataTransaction) Fail(title string) *DataTransactionResponse {
	return &DataTransactionResponse{
		Success:  false,
		Data:     transaction.data,
		Message:  transaction.GetMessage(),
		ID:       transaction.id,
		ActionID: transaction.actionID,
		Title:    title,
		Action:   transaction.action,
	}
}

// FailFromMapErr return a valid failed DataTransactionResponse
func (transaction *DataTransaction) FailFromMapErr(err MapErr) *DataTransactionResponse {
	return &DataTransactionResponse{
		Success:  false,
		Data:     transaction.data,
		Message:  err.GetMessage(),
		ID:       transaction.id,
		ActionID: transaction.actionID,
		Title:    string(err.GetType()),
		Action:   transaction.action,
	}
}
