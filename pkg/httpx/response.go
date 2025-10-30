package httpx

type ErrorResponse struct {
	Error string `json:"error"`
}

func OK[T any](data T) map[string]any { return map[string]any{"data": data} }
