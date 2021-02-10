package envi

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

// LoadConfig parses a file to load all Mappings and returns the data as map[string]string
// fileName is optional
// default value is ./secretFile
func LoadConfig(fileName ...string) (map[string]string, error) {
	path := "./secretFile"

	if fileName != nil {
		path = fileName[0]
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "can not read file")
	}

	var mappings map[string]string
	err = json.Unmarshal(data, &mappings)
	if err != nil {
		return nil, errors.Wrap(err, "can not unmarshal data")
	}
	return mappings, nil
}

// LoadFromSecretFile parses a json file to load all mappings
// fileName is optional
// default value is ./secretFile
func LoadFromSecretFile(fileName ...string) error {
	mappings, err := LoadConfig(fileName...)
	if err != nil {
		return errors.Wrap(err, "can not load config")
	}

	for key, value := range mappings {
		os.Setenv(key, value)
	}

	return nil
}

// LoadEnvVars loads the given Environment Variables.
// All Vars are required.
func LoadEnvVars(required []string) (loadedVars map[string]string, err error) {
	loadedVars = make(map[string]string)

	for _, key := range required {
		loadedVars[key] = os.Getenv(key)
	}

	missingVars := listMissing(loadedVars)
	if len(missingVars) > 0 {
		err = &RequiredEnvVarsMissing{MissingVars: missingVars}

		return
	}

	return
}

// LoadEnvVarsWithOptional loads the given Environment Variables.
// These are seperated into required and optional Vars.
func LoadEnvVarsWithOptional(required, optional []string) (loadedVars map[string]string, err error) {
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
