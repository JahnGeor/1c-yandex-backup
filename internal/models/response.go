package models

type ResponseError struct {
	Message     string `json:"message"`
	Description string `json:"description"`
	ErrorType   string `json:"error"`
}

func (r ResponseError) Error() string {
	return r.Message + " " + r.Description + " " + r.ErrorType
}
