package envi

import "os"

// LoadEnvVars loads the given Environment Variables.
// These are seperated into required and optional Vars.
func LoadEnvVars(required, optional []string) (loadedVars map[string]string, err error) {
	loadedVars = make(map[string]string)

	for _, key := range required {
		loadedVars[key] = os.Getenv(key)
	}

	missingVars := listMissing(loadedVars)
	if len(missingVars) > 0 {
		err = &RequiredEnvVarsMissing{MissingVars: missingVars}

		return
	}

	for _, key := range optional {
		loadedVars[key] = os.Getenv(key)
	}

	return
}

func listMissing(vars map[string]string) (missing []string) {
	for key, value := range vars {
		if value == "" {
			missing = append(missing, key)
		}
	}

	return
}
