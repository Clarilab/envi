package envi

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

const tagDefault = "default"
const tagEnv = "env"
const tagType = "type"
const tagRequired = "required"

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

		defaultTag := getStructTag(t.Field(i), tagDefault)
		envTag := getStructTag(t.Field(i), tagEnv)

		if envTag == "" && defaultTag == "" {
			return fmt.Errorf("either env or default tag need to be filled for field: %s", t.Field(i).Name)
		}

		switch field.Kind() {
		case reflect.Struct:
			typeTag := getStructTag(t.Field(i), tagType)

			path := cmp.Or(os.Getenv(envTag), defaultTag)
			typeVal := cmp.Or(typeTag, "yaml")

			blob, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading file: %w", err)
			}

			switch typeVal {
			case "yaml", "yml":
				err := loadFile(field, blob, yaml.Unmarshal)
				if err != nil {
					return fmt.Errorf("error unmarshaling yaml: %w", err)
				}
			case "json":
				err := loadFile(field, blob, json.Unmarshal)
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

				err := loadFile(field, blob, unmarshal)
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

func loadFile(field reflect.Value, blob []byte, unmarshal func([]byte, any) error) error {
	castedVal := reflect.New(field.Type())

	err := unmarshal(blob, castedVal.Interface())
	if err != nil {
		return fmt.Errorf("could not unmarshal: %w", err)
	}

	castedVal = resolveValuePointer(castedVal)
	field.Set(castedVal)

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
