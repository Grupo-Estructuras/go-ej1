package common

import "fmt"

type ParseError struct {
	parseobject string
}
type StatusCodeError struct {
	code int
}

func (err *ParseError) Error() string {
	return "Could not parse " + err.parseobject
}

func (err *StatusCodeError) Error() string {
	return fmt.Sprintf("Last status code was %d", err.code)
}

func NewParseError(parseobject string) *ParseError {
	err := ParseError{parseobject: parseobject}
	return &err
}

func NewStatusCodeError(code int) *StatusCodeError {
	err := StatusCodeError{code: code}
	return &err
}
