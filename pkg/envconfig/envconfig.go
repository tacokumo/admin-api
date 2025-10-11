package envconfig

import (
	"os"
	"reflect"
	"strconv"

	"github.com/cockroachdb/errors"
)

// LoadFromEnv populates a struct from environment variables based on `env` struct tags.
// Accepts any struct type and returns an error if the type is not supported.
func LoadFromEnv[T any]() (T, error) {
	var cfg T
	var zero T

	value := reflect.ValueOf(&cfg).Elem()
	if value.Kind() != reflect.Struct {
		return zero, errors.New("envconfig: LoadFromEnv requires struct type")
	}

	if err := loadFromEnvRecursive(value); err != nil {
		return zero, errors.WithStack(err)
	}

	return cfg, nil
}

func loadFromEnvRecursive(v reflect.Value) error {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		if tag := fieldType.Tag.Get("env"); tag != "" {
			if envValue := os.Getenv(tag); envValue != "" {
				if !field.CanSet() {
					continue
				}

				switch field.Kind() {
				case reflect.String:
					field.SetString(envValue)
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					intVal, err := strconv.ParseInt(envValue, 10, field.Type().Bits())
					if err != nil {
						return errors.Wrapf(err, "envconfig: parse int for field %s", fieldType.Name)
					}
					field.SetInt(intVal)
				case reflect.Bool:
					boolVal, err := strconv.ParseBool(envValue)
					if err != nil {
						return errors.Wrapf(err, "envconfig: parse bool for field %s", fieldType.Name)
					}
					field.SetBool(boolVal)
				}
			}
		}

		if field.Kind() == reflect.Struct {
			if err := loadFromEnvRecursive(field); err != nil {
				return err
			}
		}
	}

	return nil
}
