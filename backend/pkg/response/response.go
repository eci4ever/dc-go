package response

type API struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Errors  []string    `json:"errors,omitempty"`
}

func OK(data interface{}) API {
	return API{Success: true, Data: data}
}

func Created(data interface{}) API {
	return API{Success: true, Data: data}
}

func Error(message string) API {
	return API{Success: false, Message: message}
}

func ValidationError(errors []string) API {
	return API{Success: false, Message: "validation failed", Errors: errors}
}

func NotFound() API {
	return API{Success: false, Message: "not found"}
}
