package errors

type ErrorModel struct {
	Message   string      `json:"message"`
	IsSuccess bool        `json:"IsSucess"`
	Error     interface{} `json:"error"`
}
