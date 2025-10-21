package response

type Resp struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}

func Error(data interface{}) Resp {
	return Resp{
		Status: "error",
		Data:   data,
	}
}

func Success(data interface{}) Resp {
	return Resp{
		Status: "ok",
		Data:   data,
	}
}
