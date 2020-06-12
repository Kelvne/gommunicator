package gommunicator

// ActionHandler is handler callback of a action
type ActionHandler func(*DataTransactionRequest) error

// MiddlewareFunc is the middle func for actions
type MiddlewareFunc func(*Context) error

// Apply applies middlewares and returns an ActionHandler
func (gom *Gommunicator) Apply(handler MiddlewareFunc, middlewares ...MiddlewareFunc) ActionHandler {
	return func(dt *DataTransactionRequest) error {
		err := dt.Decode(&req)
		if err != nil {
			return err
		}

		c := new(Context)
		c.Request = dt

		c.Set("payload", req)

		for _, middleware := range middlewares {
			if err := middleware(c); err != nil {
				return err
			}
		}

		return handler(c)
	}
}

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
			gom.onErr(err)
			return err
		}
	}

	return nil
}
