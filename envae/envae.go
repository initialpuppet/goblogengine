// Package envae takes a suitably tagged configuration struct and populates it
// with values from the current App Engine configuration.
//
// Rob Pike says (https://www.youtube.com/watch?v=PAAkCSZUG1c) that two things
// one should avoid doing when programming in Go are accepting interface{} and
// using reflection. This package does both. It's written as a learning
// exercise. Do not use it - there are better ways of managing your App Engine
// application's configuration.
package envae

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// ErrorMissingConfValue is returned if a configuration value defined in the
// configuration struct has not been set in the App Engine environemnt.
var ErrorMissingConfValue = errors.New("envae: missing configuration value")

// ErrorInvalidConfValue is returned if a configuration value defined in the
// configuration struct exists in the App Engine environment but is invalid.
var ErrorInvalidConfValue = errors.New("envae: invalid configuration value")

const tagPrefix = "envae"

// Populate accepts a pointer to a suitably tagged struct and fills it with
// matching values from the App Engine configuration as specified in app.yaml.
//
// Structs should be tagged in the format: envae:config_value. There is no
// hierarchical structure within the configuration file but structs nested by
// value at any depth can be used provided they are tagged correctly.
func Populate(config interface{}) error {
	rv := reflect.ValueOf(config)
	if rv.Kind() != reflect.Ptr {
		return errors.New("envae: interface must be a pointer to struct")
	}
	confval := rv.Elem()
	return fillFields(confval)
}

func fillFields(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		// Configuration structs can be nested
		fval := v.Field(i)
		if fval.Kind() == reflect.Struct {
			fillFields(fval)
			continue
		}

		field := t.Field(i)
		tag := field.Tag.Get(tagPrefix)
		if tag != "" {
			err := setFromEnv(&fval, tag)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

var lookupEnv = os.LookupEnv // allows mocking

func setFromEnv(v *reflect.Value, key string) error {
	val, set := lookupEnv(key)
	if set == false {
		return ErrorMissingConfValue
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(val)
	case reflect.Int:
		converted, err := strconv.ParseInt(val, 10, 0)
		if err != nil {
			return ErrorInvalidConfValue
		}
		v.SetInt(converted)
	case reflect.Bool:
		converted, err := strconv.ParseBool(val)
		if err != nil {
			return ErrorInvalidConfValue
		}
		v.SetBool(converted)
	case reflect.Slice:
		// TODO: check this is a slice of strings
		strs := strings.Split(val, ",")
		v.Set(reflect.ValueOf(strs))
	}

	return nil
}
