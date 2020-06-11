package gommunicator

import (
	"fmt"
	"strings"
)

// DataTransactionRequest is the request object to the services cluster
type DataTransactionRequest struct {
	// DataTransaction ID
	ID string `json:"id"`
	// Payload to be read
	Data interface{} `json:"data"`
	// Service name
	Service string `json:"service"`
	// Action name
	Action string `json:"action"`
}

// DataTransactionResponse is the response object to the services cluster
type DataTransactionResponse struct {
	// DataTransaction ID
	ID string `json:"id"`
	// True if the data transaction was successful
	Success bool `json:"success"`
	// Message to be read
	Message string `json:"message"`
	// Title of the response
	Title string `json:"title"`
	// Payload to be read
	Data interface{} `json:"data"`
	// Flag that is false if this should not be handled as a data transaction
	// default is `true`
	ShouldHandleAsDT bool `json:"handleAsDataTransaction"`
	// Flag that is false if this is an action response
	// default is `false`
	ActionResponse bool `json:"actionResponse"`
	// Action name if this is an action response
	// default is "" (an empty string)
	ActionName string `json:"actionName,omitempty"`
	// DataTransaction context (metadata)
	Context map[string]interface{} `json:"context,omitempty"`
}

// DataTransaction holder for handling data transactions
type DataTransaction struct {
	id       string
	data     interface{}
	messages []string
	context  map[string]interface{}
	action   *string
}

// NewDataTransaction returns a new DataTransaction object
func NewDataTransaction(id string) *DataTransaction {
	return &DataTransaction{
		id:       id,
		data:     nil,
		messages: make([]string, 0),
		context:  make(map[string]interface{}),
	}
}

// FromResponse returns a new DataTransaction object from a DataTransactionResponse
func FromResponse(response *DataTransactionResponse) *DataTransaction {
	return &DataTransaction{
		id:       response.ID,
		action:   &response.ActionName,
		context:  response.Context,
		data:     response.Data,
		messages: strings.Split(response.Message, "\n"),
	}
}

// FromRequest returns a new DataTransaction object from a DataTransactionResponse
func FromRequest(request *DataTransactionRequest) *DataTransaction {
	return &DataTransaction{
		id:       request.ID,
		action:   &request.Action,
		context:  make(map[string]interface{}),
		data:     request.Data,
		messages: make([]string, 0),
	}
}

// SetAction sets the action of this request is representing
func (transaction *DataTransaction) SetAction(action string) *DataTransaction {
	transaction.action = &action
	return transaction
}

// AddMessage adds a new message to the DataTransaction
func (transaction *DataTransaction) AddMessage(message string) *DataTransaction {
	transaction.messages = append(transaction.messages, message)
	return transaction
}

// AddToContext adds a new context attribute to the DataTransaction
func (transaction *DataTransaction) AddToContext(key string, value interface{}) *DataTransaction {
	transaction.context[key] = value
	return transaction
}

// GetContext returns the context of the DataTransaction
func (transaction *DataTransaction) GetContext() map[string]interface{} {
	return transaction.context
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
		Success:          true,
		ShouldHandleAsDT: true,
		Data:             transaction.data,
		Message:          transaction.GetMessage(),
		ID:               transaction.id,
		Title:            title,
		ActionName:       "",
		ActionResponse:   false,
	}
}

// Fail return a valid failed DataTransactionResponse
func (transaction *DataTransaction) Fail(title string) *DataTransactionResponse {
	return &DataTransactionResponse{
		Success:          false,
		ShouldHandleAsDT: true,
		Data:             transaction.data,
		Message:          transaction.GetMessage(),
		ID:               transaction.id,
		Title:            title,
		ActionName:       "",
		ActionResponse:   false,
	}
}

// Returns a new DataTransactionResponse with context appended
// Useful on internal services communication
func (transaction *DataTransaction) ContextResponse(title string) *DataTransactionResponse {
	return &DataTransactionResponse{
		Success:          true,
		ShouldHandleAsDT: true,
		Data:             transaction.data,
		Message:          transaction.GetMessage(),
		ID:               transaction.id,
		Title:            title,
		ActionName:       "",
		ActionResponse:   false,
		Context:          *transaction.context,
	}
}

// FailFromMapErr return a valid failed DataTransactionResponse
func (transaction *DataTransaction) FailFromMapErr(err MapErr) *DataTransactionResponse {
	return &DataTransactionResponse{
		Success:          false,
		ShouldHandleAsDT: true,
		Data:             transaction.data,
		Message:          err.GetMessage(),
		ID:               transaction.id,
		Title:            string(err.GetType()),
		ActionName:       "",
		ActionResponse:   false,
	}
}

// Request return a valid DataTransactionRequest
func (transaction *DataTransaction) Request(action, service string, payload interface{}) *DataTransactionRequest {
	return &DataTransactionRequest{
		ID:      transaction.id,
		Data:    payload,
		Action:  action,
		Service: service,
	}
}
