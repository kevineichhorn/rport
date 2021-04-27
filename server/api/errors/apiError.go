package errors

// APIError wraps error which is interpreted as in http error
type APIError struct {
	Message string
	Err error
	Code int
}

// Error interface implementation
func (ae APIError) Error() string {
	if ae.Err != nil {
		return ae.Err.Error()
	}

	return ae.Message
}
