package spz

// SpzError represents an error related to SPZ file processing
type SpzError struct {
	Message string
}

func (e *SpzError) Error() string {
	return e.Message
}
