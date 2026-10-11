package contracts

type ErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
