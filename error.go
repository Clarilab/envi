package envi

// RequiredEnvVarsMissing says, that a required Environment Variable is not given.
type RequiredEnvVarsMissing struct {
	MissingVars []string
}

func (e *RequiredEnvVarsMissing) Error() string {
	return "One or more required environment variables are missing"
}
