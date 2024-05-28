package envi

import (
	"fmt"
	"strings"
)

// InvalidKindError is returned when a field is not of the expected kind.
type InvalidKindError struct {
	FieldName string
	Expected  string
	Got       string
}

func (e *InvalidKindError) Error() string {
	return fmt.Sprintf("expected field %s to be kind %s got %s", e.FieldName, e.Expected, e.Got)
}

// UnmarshalError is returned when an error occurs while unmarshalling.
type UnmarshalError struct {
	Type string
	Err  error
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("could not unmarshal %s: %s", e.Type, e.Err.Error())
}

// ValidationError is returned when one or multiple errors occured while validating the config.
type ValidationError struct {
	Errors []error
}

func (e *ValidationError) Error() string {
	sb := strings.Builder{}

	for _, err := range e.Errors {
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	}

	return sb.String()
}

// FieldRequiredError is returned when a required field is not set.
type FieldRequiredError struct {
	FieldName string
}

func (e *FieldRequiredError) Error() string {
	return fmt.Sprintf("field %s is required", e.FieldName)
}

// MissingTagError is returned when a required tag is not set.
type MissingTagError struct {
	Tag string
}

func (e *MissingTagError) Error() string {
	return fmt.Sprintf("tag %s not set", e.Tag)
}

type InvalidTagError struct {
	Tag string
}

func (e *InvalidTagError) Error() string {
	return fmt.Sprintf("invalid tag %s", e.Tag)
}

// ParsingError is returned when an error occurs while parsing a value into a specific datatype.
type ParsingError struct {
	Type string
	Err  error
}

func (e *ParsingError) Error() string {
	return fmt.Sprintf("could not parse %s: %s", e.Type, e.Err.Error())
}

// CloseError is returned when one or multiple errors occured while closing the file watchers.
type CloseError struct {
	Errors []error
}

func (e *CloseError) Error() string {
	sb := strings.Builder{}

	for _, err := range e.Errors {
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	}

	return sb.String()
}
