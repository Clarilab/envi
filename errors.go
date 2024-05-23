package envi

import (
	"fmt"
	"strings"
)

type InvalidKindError struct {
	FieldName string
	Expected  string
	Got       string
}

func (e *InvalidKindError) Error() string {
	return fmt.Sprintf("expected field %s to be kind %s got %s", e.FieldName, e.Expected, e.Got)
}

type InvalidTypeError struct {
	FieldName string
	Expected  string
	Got       string
}

func (e *InvalidTypeError) Error() string {
	return fmt.Sprintf("expected type %s, got %s", e.Expected, e.Got)
}

type UnmarshalError struct {
	Type string
	Err  error
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("could not unmarshal %s: %s", e.Type, e.Err.Error())
}

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

type FieldRequiredError struct {
	FieldName string
}

func (e *FieldRequiredError) Error() string {
	return fmt.Sprintf("field %s is required", e.FieldName)
}

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

type ParsingError struct {
	Type string
	Err  error
}

func (e *ParsingError) Error() string {
	return fmt.Sprintf("could not parse %s: %s", e.Type, e.Err.Error())
}

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
