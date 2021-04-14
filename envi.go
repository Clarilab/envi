package envi

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Envi interface {
	// FromMap loads the given key-value pairs and loads them into the local map.
	FromMap(map[string]string)

	// LoadEnv loads the given keys from environment.
	LoadEnv(vars ...string)

	// LoadStringFromFile loads a string value under given key from a file.
	LoadStringFromFile(key, filePath string) error

	// LoadJSON loads key-value pairs from one or many json blobs.
	LoadJSON(...[]byte) error

	// LoadJSONFiles loads key-value pairs from one or more json files.
	LoadJSONFiles(...string) error

	// LoadYAML loads key-value pairs from one or many yaml blobs.
	LoadYAML(...[]byte) error

	// LoadYAMLFiles loads key-value pairs from one or more yaml files.
	LoadYAMLFiles(...string) error

	// EnsureVars checks, if all given keys have a non-empty value.
	EnsureVars(...string) error

	// ToEnv writes all key-value pairs to the environment.
	ToEnv()

	// ToMap returns a map, containing all key-value pairs.
	ToMap() map[string]string
}

type envi struct {
	loadedVars map[string]string
}

func NewEnvi() Envi {
	return &envi{
		loadedVars: make(map[string]string),
	}
}

func (envi *envi) FromMap(vars map[string]string) {
	for key := range vars {
		envi.loadedVars[key] = vars[key]
	}
}

func (envi *envi) LoadEnv(vars ...string) {
	for _, key := range vars {
		envi.loadedVars[key] = os.Getenv(key)
	}
}

func (envi *envi) LoadStringFromFile(key, filePath string) error {
	blob, err := ioutil.ReadFile(filePath)
	if err != nil {
		return errors.Wrapf(err, "failed to read file '%s'", filePath)
	}

	envi.loadedVars[key] = string(blob)

	return nil
}

func (envi *envi) LoadJSONFiles(paths ...string) error {
	for i := range paths {
		blob, err := ioutil.ReadFile(paths[i])
		if err != nil {
			return errors.Wrapf(err, "failed to read json file '%s'", paths[i])
		}

		err = envi.LoadJSON(blob)
		if err != nil {
			return errors.Wrapf(err, "failed to load json file '%s'", paths[i])
		}
	}

	return nil
}

func (envi *envi) LoadJSON(blobs ...[]byte) error {
	for i := range blobs {
		var decoded map[string]string

		err := json.Unmarshal(blobs[i], &decoded)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal json")
		}

		for key := range decoded {
			envi.loadedVars[key] = decoded[key]
		}
	}

	return nil
}

func (envi *envi) LoadYAMLFiles(paths ...string) error {
	for i := range paths {
		blob, err := ioutil.ReadFile(paths[i])
		if err != nil {
			return errors.Wrapf(err, "failed to read yaml file '%s'", paths[i])
		}

		err = envi.LoadYAML(blob)
		if err != nil {
			return errors.Wrapf(err, "failed to load yaml file '%s'", paths[i])
		}
	}

	return nil
}

func (envi *envi) LoadYAML(blobs ...[]byte) error {
	for i := range blobs {
		var decoded map[string]string

		err := yaml.Unmarshal(blobs[i], &decoded)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal yaml")
		}

		for key := range decoded {
			envi.loadedVars[key] = decoded[key]
		}
	}

	return nil
}

func (envi *envi) EnsureVars(requiredVars ...string) error {
	var missingVars []string

	for _, key := range requiredVars {
		if envi.loadedVars[key] == "" {
			missingVars = append(missingVars, key)
		}
	}

	if len(missingVars) > 0 {
		return &RequiredEnvVarsMissing{MissingVars: missingVars}
	}

	return nil
}

func (envi *envi) ToEnv() {
	for key := range envi.loadedVars {
		os.Setenv(key, envi.loadedVars[key])
	}
}

func (envi *envi) ToMap() map[string]string {
	return envi.loadedVars
}
