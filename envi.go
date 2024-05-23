package envi

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

const tagDefault = "default"
const tagEnv = "env"
const tagType = "type"
const tagRequired = "required"
const tagWatch = "watch"

type unmarshalFunc func([]byte, any) error

type FileWatcher interface {
	OnChange()
	OnError(error)
}

type Envi struct {
	fileWatchers []*fsnotify.Watcher
}

func (e *Envi) Close() error {
	for _, watcher := range e.fileWatchers {
		if err := watcher.Close(); err != nil {
			return err
		}
	}

	return nil
}

func New() *Envi {
	return &Envi{
		fileWatchers: make([]*fsnotify.Watcher, 0),
	}
}

// here could be a large description
func (e *Envi) GetConfig(config any) error {
	v := reflect.ValueOf(config)
	t := reflect.TypeOf(config)

	v = resolveValuePointer(v)
	t = resolveTypePointer(t)

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("%T is not a struct", config)
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		field = resolveValuePointer(field)

		defaultTag := getStructTag(t.Field(i), tagDefault)
		envTag := getStructTag(t.Field(i), tagEnv)

		split := strings.Split(envTag, ",")
		if len(split) == 0 {
			return fmt.Errorf("either env or default tag need to be filled for field: %s", t.Field(i).Name)
		}

		if envTag == "" && defaultTag == "" {
			return fmt.Errorf("either env or default tag need to be filled for field: %s", t.Field(i).Name)
		}

		switch field.Kind() {
		case reflect.Struct:
			typeTag := getStructTag(t.Field(i), tagType)
			watchTag := getStructTag(t.Field(i), tagWatch)

			path := cmp.Or(os.Getenv(envTag), defaultTag)
			typeVal := cmp.Or(typeTag, "yaml")

			unmarshalMap := map[string]unmarshalFunc{
				"yaml": yaml.Unmarshal,
				"yml":  yaml.Unmarshal,
				"json": json.Unmarshal,
				"text": unmarshalText,
			}

			unmarshalFunc, ok := unmarshalMap[typeVal]
			if !ok {
				return fmt.Errorf("invalid type %s", typeVal)
			}

			err := loadFile(field, path, unmarshalFunc)
			if err != nil {
				return fmt.Errorf("error unmarshaling %s file: %w", typeVal, err)
			}

			if watchTag == "true" {
				err = e.watchFile(field, path, unmarshalFunc)
				if err != nil {
					return fmt.Errorf("error watching %s file: %w", typeVal, err)
				}
			}
		case reflect.String:
			tagVal := getStructTag(t.Field(i), tagEnv)

			if tagVal == "" {
				return fmt.Errorf("tag env not set for field %s", t.Field(i).Name)
			}

			field.SetString(cmp.Or(os.Getenv(tagVal), defaultTag))
		default:
			return fmt.Errorf("invalid field type %s", field.Kind().String())
		}
	}

	err := validate(config)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
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

func loadFile(field reflect.Value, path string, unmarshal unmarshalFunc) error {
	err := handleDefaults(field)
	if err != nil {
		return fmt.Errorf("error handling defaults: %w", err)
	}

	blob, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	err = unmarshal(blob, field.Addr().Interface())
	if err != nil {
		return fmt.Errorf("could not unmarshal: %w", err)
	}

	return nil
}

func handleDefaults(field reflect.Value) error {
	for i := range field.NumField() {
		defaultTag := getStructTag(field.Type().Field(i), tagDefault)

		if defaultTag != "" {
			switch field.Field(i).Kind() {
			case reflect.Int32:
				fallthrough
			case reflect.Int64:
				parsedInt, err := strconv.ParseInt(defaultTag, 10, 64)
				if err != nil {
					return fmt.Errorf("could not parse int: %w", err)
				}

				field.Field(i).SetInt(parsedInt)
			case reflect.Float32:
				fallthrough
			case reflect.Float64:
				parsedFloat, err := strconv.ParseFloat(defaultTag, 64)
				if err != nil {
					return fmt.Errorf("could not parse float: %w", err)
				}

				field.Field(i).SetFloat(parsedFloat)
			case reflect.String:
				field.Field(i).SetString(defaultTag)
			case reflect.Bool:
				b, err := strconv.ParseBool(defaultTag)
				if err != nil {
					return fmt.Errorf("could not parse bool: %w", err)
				}

				field.Field(i).SetBool(b)
			default:
				return fmt.Errorf("invalid field type %s", field.Kind().String())
			}
		}
	}

	return nil
}

func (e *Envi) watchFile(field reflect.Value, path string, unmarshal unmarshalFunc) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to init watcher: %w", err)
	}

	e.fileWatchers = append(e.fileWatchers, watcher)

	go fileWatcher(watcher, field, path, unmarshal)

	err = watcher.Add(path)
	if err != nil {
		watcher.Close()

		return fmt.Errorf("failed to add path to watcher: %w", err)
	}

	return nil
}
func validate(e any) error {
	v := reflect.ValueOf(e)
	t := reflect.TypeOf(e)

	v = resolveValuePointer(v)
	t = resolveTypePointer(t)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		if field.Kind() == reflect.Struct {
			err := validate(field.Interface())
			if err != nil {
				return err
			}
		}

		required := getStructTag(t.Field(i), tagRequired)

		if required == "true" && field.IsZero() {
			return fmt.Errorf("field %s is required", t.Field(i).Name)
		}
	}

	return nil
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

func fileWatcher(
	watcher *fsnotify.Watcher,
	field reflect.Value,
	filePath string,
	unmarshal func([]byte, any) error,
) {
	callback, ok := field.Addr().Interface().(FileWatcher)
	if !ok {
		return
	}

	mutex := new(sync.Mutex)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Has(fsnotify.Write) {
				mutex.Lock()

				err := loadFile(field, filePath, unmarshal)
				if err != nil {
					callback.OnError(fmt.Errorf("error reloading watched file: %w", err))

					continue
				}

				mutex.Unlock()

				callback.OnChange()
			} else if event.Has(fsnotify.Remove) {
				err := watcher.Add(filePath)
				if err != nil {
					callback.OnError(fmt.Errorf("error reenabling watcher for file: %w", err))
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			callback.OnError(fmt.Errorf("error while watching file: %w", err))
		}
	}
}
