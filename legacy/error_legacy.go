package envi_legacy

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

// EnvVarNotFoundError says, that a given Environment Variable is not found.
type EnvVarNotFoundError struct {
	variable string
}

// Error implements the error interface.
func (e *EnvVarNotFoundError) Error() string {
	return fmt.Sprintf("environment variable '%s' not found", e.variable)
}

// FailedToReadFileError says, that a given file could not be read.
type FailedToReadFileError struct {
	path string
}

// Error implements the error interface.
func (e *FailedToReadFileError) Error() string {
	return fmt.Sprintf("failed to read file '%s'", e.path)
}
