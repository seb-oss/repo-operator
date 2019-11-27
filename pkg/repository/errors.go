package repository

// ErrorsJSON : Struct
type ErrorsJSON struct {
	Errors []ErrorJSON `json:"errors,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// ErrorJSON : Struct
type ErrorJSON struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
