package models

type ResponseError struct {
	Description string `json:"description"`
	ErrorType   string `json:"error"`
	StatusCode  int
}

func (e *ResponseError) Error() string {
	return e.ErrorType
}
