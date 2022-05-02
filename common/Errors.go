package common

type ParseError struct {
	parseobject string
}

func (pe *ParseError) Error() string {
	return "Could not parse " + pe.parseobject
}

func NewParseError(parseobject string) *ParseError {
	err := ParseError{parseobject: parseobject}
	return &err
}
