package dto

type Response[T any] struct {
	Code string `json:"code"`
	Message string `json:"message"`
	Data T `json:"data"`
}

func CreateResponseError(message string) Response[string] {
	return Response[string]{
		Code: "error",
		Message: message,
		Data: "",
	}
}

func CreateResponseSuccess[T any](data T) Response[T] {
	return Response[T]{
		Code: "200",
		Message: "success",
		Data: data,
	}
}

func CreateResponseErrorData(message string, data map[string]string) Response[map[string]string] {
	return Response[map[string]string]{
		Code: "401",
		Message: message,
		Data: data,
	}
}
