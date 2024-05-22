package envi

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
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

type FileWatcher interface {
	Notify()
}

// here could be a large description
func GetEnvs(e any) error {
	v := reflect.ValueOf(e)
	t := reflect.TypeOf(e)

	v = resolveValuePointer(v)
	t = resolveTypePointer(t)

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("%T is not a struct", e)
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		field = resolveValuePointer(field)

		defaultTag := getStructTag(t.Field(i), tagDefault)
		envTag := getStructTag(t.Field(i), tagEnv)

		if envTag == "" && defaultTag == "" {
			return fmt.Errorf("either env or default tag need to be filled for field: %s", t.Field(i).Name)
		}

		switch field.Kind() {
		case reflect.Struct:
			typeTag := getStructTag(t.Field(i), tagType)
			watchTag := getStructTag(t.Field(i), tagWatch)

			path := cmp.Or(os.Getenv(envTag), defaultTag)
			typeVal := cmp.Or(typeTag, "yaml")

			switch typeVal {
			case "yaml", "yml":
				err := loadFile(field, path, yaml.Unmarshal)
				if err != nil {
					return fmt.Errorf("error unmarshaling yaml: %w", err)
				}

				if watchTag == "true" {
					watcher, err := fsnotify.NewWatcher()
					if err != nil {
						return fmt.Errorf("failed to init watcher: %w", err)
					}

					go fileWatcher(watcher, field, path, yaml.Unmarshal)

					err = watcher.Add(path)
					if err != nil {
						watcher.Close()
						return fmt.Errorf("failed to add path to watcher: %w", err)
					}
				}
			case "json":
				err := loadFile(field, path, json.Unmarshal)
				if err != nil {
					return fmt.Errorf("error unmarshaling json: %w", err)
				}
			case "text":
				unmarshal := func(data []byte, v any) error {
					val := strings.Trim(string(data), "\n")

					rv := reflect.ValueOf(v)
					rv = resolveValuePointer(rv)
					rv.Field(0).SetString(val)

					return nil
				}

				err := loadFile(field, path, unmarshal)
				if err != nil {
					return fmt.Errorf("error unmarshaling text: %w", err)
				}
			default:
				return fmt.Errorf("invalid type %s", typeVal)
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

	err := validate(e)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

func loadFile(field reflect.Value, path string, unmarshal func([]byte, any) error) error {
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
		fmt.Println("callback is not of type FileWatcher")
		return
	}

	mutex := new(sync.Mutex)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Has(fsnotify.Chmod) || event.Has(fsnotify.Write) {
				mutex.Lock()
				err := loadFile(field, filePath, unmarshal)
				if err != nil {
					// watchErrChan <- fmt.Errorf("error reloading watched file: %w", err)
					fmt.Println("error reloading watched file: %w", err)
					continue
				}
				mutex.Unlock()

				callback.Notify()
			} else if event.Has(fsnotify.Remove) {
				err := watcher.Add(filePath)
				if err != nil {
					// watchErrChan <- fmt.Errorf("error reenabling watcher for file: %w", err)
					fmt.Println("error reenabling watcher for file: %w", err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			// watchErrChan <- fmt.Errorf("error while watching file: %w", err)
			fmt.Println("error while watching file: %w", err)
		}
	}
}
