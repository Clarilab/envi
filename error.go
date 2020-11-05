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
