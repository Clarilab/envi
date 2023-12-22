package envi

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

const (
	fileTypeJSON uint8 = iota + 1
	fileTypeYAML
)

// Envi is a config loader to load all sorts of configuration files.
type Envi struct {
	mu         sync.Mutex
	loadedVars map[string]string
}

// NewEnvi creates a new Envi instance.
func NewEnvi() *Envi {
	return &Envi{
		mu:         sync.Mutex{},
		loadedVars: make(map[string]string),
	}
}

// FromMap loads the given key-value pairs and loads them into the local map.
func (envi *Envi) FromMap(vars map[string]string) {
	envi.mu.Lock()

	for key := range vars {
		envi.loadedVars[key] = vars[key]
	}

	envi.mu.Unlock()
}

// LoadEnv loads the given keys from the environment variables.
func (envi *Envi) LoadEnv(vars ...string) {
	envi.mu.Lock()

	for _, key := range vars {
		envi.loadedVars[key] = os.Getenv(key)
	}

	envi.mu.Unlock()
}

// LoadYAMLFilesFromEnvPaths loads yaml files from the paths in the given environment variables.
func (envi *Envi) LoadYAMLFilesFromEnvPaths(vars ...string) error {
	const errMessage = "failed to load yaml files from env paths: %w"
	for _, key := range vars {
		path := os.Getenv(key)

		if path == "" {
			return fmt.Errorf(errMessage, &EnvVarNotFoundError{key})
		}

		if err := envi.LoadYAMLFile(path); err != nil {
			return fmt.Errorf(errMessage, err)
		}
	}

	return nil
}

// LoadYAMLFilesFromEnvPaths loads json files from the paths in the given environment variables.
func (envi *Envi) LoadJSONFilesFromEnvPaths(vars ...string) error {
	const errMessage = "failed to load json files from env paths: %w"

	for _, key := range vars {
		path := os.Getenv(key)

		if path == "" {
			return fmt.Errorf(errMessage, &EnvVarNotFoundError{key})
		}

		if err := envi.LoadJSONFile(path); err != nil {
			return fmt.Errorf(errMessage, err)
		}
	}

	return nil
}

// LoadYAMLFilesFromEnvPaths loads the file content from the path in the given environment variable
// to the value of the given key.
func (envi *Envi) LoadFileFromEnvPath(key string, envPath string) error {
	const errMessage = "failed to load file from env paths: %w"

	filePath := os.Getenv(envPath)

	if filePath == "" {
		return fmt.Errorf(errMessage, &EnvVarNotFoundError{envPath})
	}

	if err := envi.LoadFile(key, filePath); err != nil {
		return fmt.Errorf(errMessage, err)
	}

	return nil
}

// LoadFile loads a string value under given key from a file.
func (envi *Envi) LoadFile(key, filePath string) error {
	const errMessage = "failed to load file: %w"

	blob, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf(errMessage, &FailedToReadFileError{filePath})
	}

	envi.mu.Lock()

	envi.loadedVars[key] = string(blob)

	envi.mu.Unlock()

	return nil
}

// LoadJSONFiles loads key-value pairs from one or more json files.
func (envi *Envi) LoadJSONFiles(paths ...string) error {
	const errMessage = "failed to load json files: %w"

	for i := range paths {
		if err := envi.LoadJSONFile(paths[i]); err != nil {
			return fmt.Errorf(errMessage, err)
		}
	}

	return nil
}

// LoadJSONFile loads key-value pairs from a json file.
func (envi *Envi) LoadJSONFile(path string) error {
	const errMessage = "failed to load json file: %w"

	blob, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf(errMessage, &FailedToReadFileError{path})
	}

	err = envi.LoadJSON(blob)
	if err != nil {
		return fmt.Errorf(errMessage, err)
	}

	return nil
}

// LoadJSONFilePrefixed loads key-value pairs from a json file
// and prefixes the keys from the file with the given string.
func (envi *Envi) LoadJSONFilePrefixed(prefix, path string) error {
	const errMessage = "failed to load json file: %w"

	blob, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf(errMessage, &FailedToReadFileError{path})
	}

	err = envi.LoadJSONPrefixed(prefix, blob)
	if err != nil {
		return fmt.Errorf(errMessage, err)
	}

	return nil
}

// LoadAndWatchJSONFile loads key-value pairs from a json file,
// then watches that file and reloads it when it changes.
// Accepts optional callback functions that are executed
// after the file was reloaded. Returns and error when something
// goes wrong. When no error is returned, returns a close function
// that should be deferred in the calling function, and an error
// channel where errors that occur during the file watching get sent.
func (envi *Envi) LoadAndWatchJSONFile(path string, callback ...func() error) (error, func() error, <-chan error) {
	const errMessage = "failed to load and watch json file: %w"

	loadFunc := envi.getLoadFunc(fileTypeJSON, path)

	err, closeFunc, watchErrChan := envi.loadAndWatchFile(path, loadFunc, callback...)
	if err != nil {
		return fmt.Errorf(errMessage, err), nil, nil
	}

	return nil, closeFunc, watchErrChan
}

// LoadAndWatchJSONFilePrefixed works exactly as LoadAndWatchJSONFile,
// except it prefixes the keys of the loaded variables with the given
// string.
func (envi *Envi) LoadAndWatchJSONFilePrefixed(prefix, path string, callback ...func() error) (error, func() error, <-chan error) {
	const errMessage = "failed to load and watch json file: %w"

	loadFunc := envi.getLoadFunc(fileTypeJSON, path, prefix)

	err, closeFunc, watchErrChan := envi.loadAndWatchFile(path, loadFunc, callback...)
	if err != nil {
		return fmt.Errorf(errMessage, err), nil, nil
	}

	return nil, closeFunc, watchErrChan
}

// LoadJSON loads key-value pairs from one or many json blobs.
func (envi *Envi) LoadJSON(blobs ...[]byte) error {
	const errMessage = "failed to load json: %w"

	for i := range blobs {
		var decoded map[string]string

		err := json.Unmarshal(blobs[i], &decoded)
		if err != nil {
			return fmt.Errorf(errMessage, err)
		}

		envi.mu.Lock()

		for key := range decoded {
			envi.loadedVars[key] = decoded[key]
		}

		envi.mu.Unlock()
	}

	return nil
}

// LoadJSONPrefixed loads key-value pairs from one or many json blobs
// and prefixes the keys from the blobs with the given string.
func (envi *Envi) LoadJSONPrefixed(prefix string, blobs ...[]byte) error {
	const errMessage = "failed to load json: %w"

	for i := range blobs {
		var decoded map[string]string

		err := json.Unmarshal(blobs[i], &decoded)
		if err != nil {
			return fmt.Errorf(errMessage, err)
		}

		envi.mu.Lock()

		for key := range decoded {
			envi.loadedVars[prefix+key] = decoded[key]
		}

		envi.mu.Unlock()
	}

	return nil
}

// LoadYAMLFiles loads key-value pairs from one or more yaml files.
func (envi *Envi) LoadYAMLFiles(paths ...string) error {
	const errMessage = "failed to load yaml files: %w"

	for i := range paths {
		if err := envi.LoadYAMLFile(paths[i]); err != nil {
			return fmt.Errorf(errMessage, err)
		}
	}

	return nil
}

// LoadYAMLFile loads key-value pairs from a yaml file.
func (envi *Envi) LoadYAMLFile(path string) error {
	const errMessage = "failed to load yaml file: %w"

	blob, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf(errMessage, &FailedToReadFileError{path})
	}

	err = envi.LoadYAML(blob)
	if err != nil {
		return fmt.Errorf(errMessage, err)
	}

	return nil
}

// LoadYAMLFilePrefixed loads key-value pairs from a yaml file
// and prefixes the keys from the file with the given string.
func (envi *Envi) LoadYAMLFilePrefixed(prefix, path string) error {
	const errMessage = "failed to load yaml file prefixed: %w"

	blob, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf(errMessage, &FailedToReadFileError{path})
	}

	err = envi.LoadYAMLPrefixed(prefix, blob)
	if err != nil {
		return fmt.Errorf(errMessage, err)
	}

	return nil
}

// LoadAndWatchYAMLFile loads key-value pairs from a yaml file,
// then watches that file and reloads it when it changes.
// Accepts optional callback functions that are executed
// after the file was reloaded. Returns and error when something
// goes wrong. When no error is returned, returns a close function
// that should be deferred in the calling function, and an error
// channel where errors that occur during the file watching get sent.
func (envi *Envi) LoadAndWatchYAMLFile(path string, callbacks ...func() error) (error, func() error, <-chan error) {
	const errMessage = "failed to load and watch yaml file: %w"

	loadFunc := envi.getLoadFunc(fileTypeYAML, path)

	err, closeFunc, watchErrChan := envi.loadAndWatchFile(path, loadFunc, callbacks...)
	if err != nil {
		return fmt.Errorf(errMessage, err), nil, nil
	}

	return nil, closeFunc, watchErrChan
}

// LoadAndWatchYAMLFilePrefixed works exactly as LoadAndWatchYAMLFile,
// except it prefixes the keys of the loaded variables with the given
// string.
func (envi *Envi) LoadAndWatchYAMLFilePrefixed(prefix, path string, callbacks ...func() error) (error, func() error, <-chan error) {
	const errMessage = "failed to load and watch yaml file: %w"

	loadFunc := envi.getLoadFunc(fileTypeYAML, path, prefix)

	err, closeFunc, watchErrChan := envi.loadAndWatchFile(path, loadFunc, callbacks...)
	if err != nil {
		return fmt.Errorf(errMessage, err), nil, nil
	}

	return nil, closeFunc, watchErrChan
}

// LoadYAML loads key-value pairs from one or many yaml blobs.
func (envi *Envi) LoadYAML(blobs ...[]byte) error {
	const errMessage = "failed to load yaml file: %w"

	for i := range blobs {
		var decoded map[string]string

		err := yaml.Unmarshal(blobs[i], &decoded)
		if err != nil {
			return fmt.Errorf(errMessage, err)
		}

		envi.mu.Lock()

		for key := range decoded {
			envi.loadedVars[key] = decoded[key]
		}

		envi.mu.Unlock()
	}

	return nil
}

// LoadYAMLPrefixed loads key-value pairs from one or many yaml blobs
// and prefixes the keys from the blobs with the given string.
func (envi *Envi) LoadYAMLPrefixed(prefix string, blobs ...[]byte) error {
	const errMessage = "failed to load yaml file: %w"

	for i := range blobs {
		var decoded map[string]string

		err := yaml.Unmarshal(blobs[i], &decoded)
		if err != nil {
			return fmt.Errorf(errMessage, err)
		}

		envi.mu.Lock()

		for key := range decoded {
			envi.loadedVars[prefix+key] = decoded[key]
		}

		envi.mu.Unlock()
	}

	return nil
}

// EnsureVars checks, if all given keys have a non-empty value.
func (envi *Envi) EnsureVars(requiredVars ...string) error {
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

// ToEnv writes all key-value pairs to the environment.
func (envi *Envi) ToEnv() {
	for key := range envi.loadedVars {
		os.Setenv(key, envi.loadedVars[key])
	}
}

// ToMap returns a map, containing all key-value pairs.
func (envi *Envi) ToMap() map[string]string {
	return envi.loadedVars
}

func (envi *Envi) loadAndWatchFile(
	path string,
	loadFunc func() error,
	callbacks ...func() error,
) (error, func() error, <-chan error) {
	const (
		errMessage        = "failed to load and watch file: %w"
		errChanBufferSize = 10
	)

	err := loadFunc()
	if err != nil {
		return fmt.Errorf(errMessage, err), nil, nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf(errMessage, err), nil, nil
	}

	watchErrChan := make(chan error, errChanBufferSize)

	go envi.fileWatcher(watcher, path, loadFunc, watchErrChan, callbacks...)

	err = watcher.Add(path)
	if err != nil {
		watcher.Close()
		return fmt.Errorf(errMessage, err), nil, nil
	}

	return nil, watcher.Close, watchErrChan
}

func (envi *Envi) fileWatcher(
	watcher *fsnotify.Watcher,
	filePath string,
	loadFunc func() error,
	watchErrChan chan<- error,
	callbacks ...func() error,
) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Has(fsnotify.Write) {
				err := loadFunc()
				if err != nil {
					watchErrChan <- fmt.Errorf("error reloading watched file: %w", err)
					continue
				}

				for i := range callbacks {
					err = callbacks[i]()
					if err != nil {
						watchErrChan <- fmt.Errorf("error executing callback for watched file: %w", err)
					}
				}
			} else if event.Has(fsnotify.Remove) {
				err := watcher.Add(filePath)
				if err != nil {
					watchErrChan <- fmt.Errorf("error reenabling watcher for file: %w", err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			watchErrChan <- fmt.Errorf("error while watching file: %w", err)
		}
	}
}

func (envi *Envi) getLoadFunc(filetype uint8, path string, prefix ...string) func() error {
	switch filetype {
	case fileTypeJSON:
		if len(prefix) > 0 {
			return func() error {
				err := envi.LoadJSONFilePrefixed(prefix[0], path)

				return err
			}
		}

		return func() error {
			err := envi.LoadJSONFile(path)

			return err
		}
	case fileTypeYAML:
		if len(prefix) > 0 {
			return func() error {
				err := envi.LoadYAMLFilePrefixed(prefix[0], path)

				return err
			}
		}

		return func() error {
			err := envi.LoadYAMLFile(path)

			return err
		}
	default:
		return nil
	}
}
