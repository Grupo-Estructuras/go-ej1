package common

import "fmt"

type ParseError struct {
	parseobject string
}
type StatusCodeError struct {
	code int
}

func (err *ParseError) Error() string {
	return "No se pudo leer " + err.parseobject
}

func (err *StatusCodeError) Error() string {
	return fmt.Sprintf("El último código error fue %d", err.code)
}

func NewParseError(parseobject string) *ParseError {
	err := ParseError{parseobject: parseobject}
	return &err
}

func NewStatusCodeError(code int) *StatusCodeError {
	err := StatusCodeError{code: code}
	return &err
}
