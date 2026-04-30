package error

import "strings"

type AppError struct {
	error
	httpCode int
	Code     string
	Message  string
}

func (ae AppError) HTTPCode() int {
	return ae.httpCode
}

func (ae AppError) UnWrap() error {
	return ae.error
}

func (ae AppError) Error() string {
	if ae.Message != "" {
		return ae.Message
	}
	if ae.error != nil {
		return ae.error.Error()
	}
	return ""
}
func (ae AppError) isEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
