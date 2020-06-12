package gommunicator

// ActionHandler is handler callback of a action
type ActionHandler func(*DataTransactionRequest) error

// RegisterAction registers a callback for a new handler
func (gom *Gommunicator) RegisterAction(action string, handler ActionHandler) *Gommunicator {
	gom.actions[action] = handler
	return gom
}

// CallAction calls a registered callback
func (gom *Gommunicator) CallAction(request *DataTransactionRequest) error {
	if callback, ok := gom.actions[request.Action]; ok == true {
		err := callback(request)
		if err != nil {
			gom.errorHandler(err)
			return err
		}
	}

	return nil
}
