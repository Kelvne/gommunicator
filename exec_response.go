package gommunicator

import (
	"errors"
)

type responseCallback func(*DataTransactionResponse) error

var callbacks map[string]responseCallback = make(map[string]responseCallback)

func registerCallback(actionID string, callback responseCallback) {
	callbacks[actionID] = callback
}

func callCallback(response *DataTransactionResponse) error {
	if response.ActionID != nil {
		if callback, ok := callbacks[*response.ActionID]; ok == true {
			delete(callbacks, *response.ActionID)
			return callback(response)
		}
	}

	return errors.New("callback not found")
}
