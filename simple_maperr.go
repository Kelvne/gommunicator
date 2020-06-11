package gommunicator

// SimpleErrorCode Simple Error code type
type SimpleErrorCode string

const (
	// Basic error
	Basic SimpleErrorCode = "BASIC"
)

// SimpleError structure
type SimpleError struct {
	Code    SimpleErrorCode
	Type    ErrType
	Message string
}

// NewSimpleError returns a new SimpleError
func NewSimpleError(code SimpleErrorCode, message string) *SimpleError {
	return &SimpleError{
		Code:    code,
		Type:    SimpleErrorType,
		Message: message,
	}
}

// GetMessage the simple error message
func (err *SimpleError) GetMessage() string {
	code := SimpleErrorCode(err.GetCode())
	switch code {
	case Basic:
		{
			return err.Message
		}
	default:
		break
	}
	return ""
}

// GetCode returns the simple error code
func (err *SimpleError) GetCode() string {
	return string(err.Code)
}

// GetType returns the simple error ErrType
func (err *SimpleError) GetType() ErrType {
	return err.Type
}

// GetContext returns the simple error context
func (err *SimpleError) GetContext() map[string]interface{} {
	return map[string]interface{}{}
}

func (err *SimpleError) Error() string {
	return err.GetMessage()
}
