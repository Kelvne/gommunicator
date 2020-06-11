package gommunicator

// ErrType describes the type of an error
type ErrType string

// Error types
const (
	InternalErrorType ErrType = "Aconteceu um erro interno!"
	SimpleErrorType           = "Aconteceu um erro!"
)

// MapErr interface
type MapErr interface {
	GetMessage() string
	GetCode() string
	GetType() ErrType
	GetContext() map[string]interface{}
	Error() string
}
