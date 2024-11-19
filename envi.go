package envi

import (
	"cmp"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

const (
	tagDefault  = "default"
	tagEnv      = "env"
	tagType     = "type"
	tagRequired = "required"
	tagWatch    = "watch"
)

// unmarshalFunc describes how to unmarshal a file.
type unmarshalFunc func([]byte, any) error

// FileWatcher is an interface for watching file changes.
type FileWatcher interface {
	OnChange()
	OnError(error)
}

type fileWatcherInstance struct {
	watcher *fsnotify.Watcher
	ctx     context.Context
	cancel  context.CancelFunc
}

// Envi holds references to all active file watchers.
type Envi struct {
	errorChan    chan error
	fileWatchers map[string]fileWatcherInstance
	fileHashes   map[string]string
}

// Errors returns an error channel where filewatcher errors are sent to.
func (e *Envi) Errors() <-chan error {
	return e.errorChan
}

// Close closes all file watchers attached to the Envi instance.
func (e *Envi) Close() error {
	var errs []error

	close(e.errorChan)

	for filePath, instance := range e.fileWatchers {
		instance.cancel()

		if err := instance.watcher.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close watcher for file %s with error: %w", filePath, err))
		}
	}

	if len(errs) > 0 {
		return &CloseError{Errors: errs}
	}

	return nil
}

// New creates a new Envi instance.
func New() *Envi {
	return &Envi{
		errorChan:    make(chan error, 100),
		fileWatchers: make(map[string]fileWatcherInstance, 0),
		fileHashes:   make(map[string]string),
	}
}

/*
Load loads all config files and environment variables into the input struct.
Supported types are JSON, YAML and text files, as well as strings.

If you want to watch a file for changes, the "watch" tag has to be set to true and the underlying struct
has to implement the envi.FileWatcher interface.

While using the "default" tag, the "env" tag can be omitted. If not omitted, the value from the
environment variable will be used.

When using the text file type, envi will try to load the file content into the first string field of that struct.

Example config:

	type Config struct {
		Environment string   `env:"ENVIRONMENT" required:"true"`
		YAMLConfig  YAMLFile `type:"yaml" watch:"true" default:"./config.yaml"`
		TextConfig  TextFile `env:"MY_TEXT_CONFIG_FILE" type:"text"`
	}

	type YAMLFile struct {
		Key1 string `yaml:"key1" required:"true"`
		Key2 string `yaml:"key2"`
	}

	func (y *YAMLFile) OnChange() {
		fmt.Println("YAML file changed")
	}

	func (y *YAMLFile) OnError(err error) {
		fmt.Println("error while reloading YAML file:", err)
	}

	type TextFile struct {
		Value string `default:"bar"`
	}

Available tags are:
  - default: default value (supports file paths for files and standard data types bool, float32, float64, int32, int64, string)
  - env: environment variable name
  - type: describes the file type (json, yaml, text), defaults to yaml if omitted
  - required: indicates that the field is required, "Load()" will return an error in this case
  - watch: indicates that the file should be watched for changes
*/
func (e *Envi) Load(config any) error {
	const errMsg = "error while getting config: %w"

	err := e.loadConfig(config)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	errs := validate(config)
	if len(errs) > 0 {
		return fmt.Errorf(errMsg, &ValidationError{Errors: errs})
	}

	return nil
}

func (e *Envi) loadConfig(config any) error {
	const errMsg = "error while loading config: %w"

	v := reflect.ValueOf(config)
	t := reflect.TypeOf(config)

	if v.Kind() != reflect.Pointer {
		return fmt.Errorf(errMsg, &InvalidKindError{
			FieldName: t.Name(),
			Expected:  "pointer",
			Got:       v.Kind().String(),
		})
	}

	v = resolveValuePointer(v)
	t = resolveTypePointer(t)

	if v.Kind() != reflect.Struct {
		return fmt.Errorf(errMsg, &InvalidKindError{
			FieldName: t.Name(),
			Expected:  "struct",
			Got:       v.Kind().String(),
		})
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		// filter out unexported fields (CanSet() is false for unexported fields)
		if !field.CanSet() {
			continue
		}

		field = resolveValuePointer(field)

		defaultTag := getStructTag(t.Field(i), tagDefault)
		envTag := getStructTag(t.Field(i), tagEnv)

		if envTag == "" && defaultTag == "" {
			return fmt.Errorf(errMsg, &MissingTagError{Tag: "env or default"})
		}

		switch field.Kind() {
		case reflect.Struct:
			typeTag := getStructTag(t.Field(i), tagType)
			watchTag := getStructTag(t.Field(i), tagWatch)

			path := cmp.Or(os.Getenv(envTag), defaultTag)

			var err error
			path, err = filepath.Abs(path)
			if err != nil {
				return fmt.Errorf(errMsg, err)
			}

			typeVal := cmp.Or(typeTag, "yaml")

			unmarshalMap := map[string]unmarshalFunc{
				"yaml": yaml.Unmarshal,
				"yml":  yaml.Unmarshal,
				"json": json.Unmarshal,
				"text": unmarshalText,
			}

			unmarshalFunc, ok := unmarshalMap[typeVal]
			if !ok {
				return fmt.Errorf(errMsg, &InvalidTagError{Tag: "type"})
			}

			_, err = e.loadFile(field, path, unmarshalFunc)
			if err != nil {
				return fmt.Errorf(errMsg, err)
			}

			if watchTag == "true" {
				err = e.watchFile(field, path, unmarshalFunc)
				if err != nil {
					return fmt.Errorf(errMsg, err)
				}
			}
		case reflect.String:
			tagVal := getStructTag(t.Field(i), tagEnv)

			if tagVal == "" && defaultTag == "" {
				return fmt.Errorf(errMsg, &MissingTagError{Tag: "env or default"})
			}

			field.SetString(cmp.Or(os.Getenv(tagVal), defaultTag))
		default:
			return fmt.Errorf(errMsg, &InvalidKindError{
				FieldName: field.Type().Name(),
				Expected:  "string, struct",
				Got:       field.Kind().String(),
			})
		}
	}

	return nil
}

func unmarshalText(data []byte, v any) error {
	val := strings.Trim(string(data), "\n")

	rv := reflect.ValueOf(v)
	rv = resolveValuePointer(rv)

	var valueSet bool

	for i := range rv.NumField() {
		if rv.Field(i).Kind() == reflect.String {
			rv.Field(i).SetString(val)
			valueSet = true
			break
		}
	}

	if !valueSet {
		return fmt.Errorf("failed to find target value for text file")
	}

	return nil
}

// loadFile loads the file at path, checks if it is different from the already loaded file if exists, and unmarshals into the config value.
func (e *Envi) loadFile(field reflect.Value, path string, unmarshal unmarshalFunc) (bool, error) {
	const errMsg = "error while loading file: %w"

	err := handleDefaults(field)
	if err != nil {
		return false, fmt.Errorf(errMsg, err)
	}

	blob, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf(errMsg, err)
	}

	newHash := fmt.Sprintf("%x", md5.Sum(blob))
	if oldHash, ok := e.fileHashes[path]; ok && newHash == oldHash {
		return false, nil // The file has not changed, do not run trigger
	} else {
		e.fileHashes[path] = newHash
	}

	err = unmarshal(blob, field.Addr().Interface())
	if err != nil {
		return false, fmt.Errorf(errMsg, err)
	}

	return true, nil
}

func handleDefaults(field reflect.Value) error {
	const errMsg = "error while handling defaults: %w"

	for i := range field.NumField() {
		defaultTag := getStructTag(field.Type().Field(i), tagDefault)

		if defaultTag != "" {
			switch field.Field(i).Kind() {
			case reflect.Int32:
				fallthrough
			case reflect.Int64:
				parsedInt, err := strconv.ParseInt(defaultTag, 10, 64)
				if err != nil {
					return fmt.Errorf(errMsg, &ParsingError{Type: "int", Err: err})
				}

				field.Field(i).SetInt(parsedInt)
			case reflect.Float32:
				fallthrough
			case reflect.Float64:
				parsedFloat, err := strconv.ParseFloat(defaultTag, 64)
				if err != nil {
					return fmt.Errorf(errMsg, &ParsingError{Type: "float", Err: err})
				}

				field.Field(i).SetFloat(parsedFloat)
			case reflect.String:
				field.Field(i).SetString(defaultTag)
			case reflect.Bool:
				b, err := strconv.ParseBool(defaultTag)
				if err != nil {
					return fmt.Errorf(errMsg, &ParsingError{Type: "bool", Err: err})
				}

				field.Field(i).SetBool(b)
			default:
				return fmt.Errorf(errMsg, &InvalidKindError{
					FieldName: field.Type().Field(i).Name,
					Expected:  "string, int, float, bool",
					Got:       field.Field(i).Kind().String(),
				})
			}
		}
	}

	return nil
}

func (e *Envi) watchFile(field reflect.Value, path string, unmarshal unmarshalFunc) error {
	const errMsg = "error while watching file: %w"

	dirPath := filepath.Dir(path)
	if _, ok := e.fileWatchers[dirPath]; !ok {
		ctx, cancel := context.WithCancel(context.Background())

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf(errMsg, err)
		}

		e.fileWatchers[dirPath] = fileWatcherInstance{
			watcher: watcher,
			ctx:     ctx,
			cancel:  cancel,
		}

		err = watcher.Add(dirPath) // needs to be the directory of the file to ensure working on linux systems
		if err != nil {
			watcher.Close()

			return fmt.Errorf(errMsg, err)
		}
	}

	fileWatcher := e.fileWatchers[dirPath]

	go e.fileWatcher(fileWatcher.ctx, fileWatcher.watcher, field, path, unmarshal)

	return nil
}

func validate(e any) []error {
	v := reflect.ValueOf(e)
	t := reflect.TypeOf(e)

	v = resolveValuePointer(v)
	t = resolveTypePointer(t)

	errors := make([]error, 0)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		if field.Kind() == reflect.Struct {
			errs := validate(field.Interface())
			if len(errs) > 0 {
				errors = append(errors, errs...)
			}
		}

		required := getStructTag(t.Field(i), tagRequired)

		if required == "true" && field.IsZero() {
			errors = append(errors, &FieldRequiredError{FieldName: t.Field(i).Name})
		}
	}

	return errors
}

func resolveValuePointer(rv reflect.Value) reflect.Value {
	if rv.Kind() == reflect.Pointer {
		rv = resolveValuePointer(rv.Elem())
	}

	return rv
}

func resolveTypePointer(rt reflect.Type) reflect.Type {
	if rt.Kind() == reflect.Ptr {
		rt = resolveTypePointer(rt.Elem())
	}

	return rt
}

func getStructTag(f reflect.StructField, tagName string) string {
	return f.Tag.Get(tagName)
}

func (e *Envi) fileWatcher(
	ctx context.Context,
	watcher *fsnotify.Watcher,
	field reflect.Value,
	filePath string,
	unmarshal func([]byte, any) error,
) {
	const errMsg = "error reloading watched file: %w"

	callback, ok := field.Addr().Interface().(FileWatcher)
	if !ok {
		return
	}

	mutex := new(sync.Mutex)

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// ensure we're only watching the file we're interested in
			if filepath.Base(event.Name) != filepath.Base(filePath) {
				continue
			}

			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
				mutex.Lock()

				callOnChange, err := e.loadFile(field, filePath, unmarshal)
				if err != nil {
					wrappedErr := fmt.Errorf(errMsg, err)
					callback.OnError(wrappedErr)

					select {
					case e.errorChan <- wrappedErr: // send the error to the channel if there's space
					default:
						// drop the error if the channel is full
					}

					continue
				}

				mutex.Unlock()

				if callOnChange {
					callback.OnChange()
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			wrappedErr := fmt.Errorf(errMsg, err)
			callback.OnError(wrappedErr)

			select {
			case e.errorChan <- wrappedErr: // send the error to the channel if there's space
			default:
				// drop the error if the channel is full
			}
		}
	}
}
