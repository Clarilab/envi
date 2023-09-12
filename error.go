package envi

import (
	"fmt"
	"strings"
)

// RequiredEnvVarsMissing says, that a required Environment Variable is not given.
type RequiredEnvVarsMissing struct {
	MissingVars []string
}

func (e *RequiredEnvVarsMissing) Error() string {
	return fmt.Sprintf("One or more required environment variables are missing\nThe missing variables are: %s", e.printMissingVars())
}

func (e *RequiredEnvVarsMissing) printMissingVars() string {
	return strings.Join(e.MissingVars, ", ")
}

// ErrEnvVarNotFound says, that a given Environment Variable is not found.
type ErrEnvVarNotFound struct {
	variable string
}

// Error implements the error interface.
func (e *ErrEnvVarNotFound) Error() string {
	return fmt.Sprintf("environment variable %s not found", e.variable)
}
