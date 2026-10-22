package config

import (
	"reflect"
	"time"
)

const koanfTag = "koanf"

var timeType = reflect.TypeOf(time.Time{})

// KoanfKeys returns the flattened koanf keys declared by the struct tags of v.
//
// Nested structs are walked recursively, so a field tagged `koanf:"port"` inside
// a struct tagged `koanf:"server"` is reported as "server.port".
//
// It is used to verify that a config file covers every declared field, and that
// it does not carry keys the structs do not know about.
func KoanfKeys(v any) []string {
	t := reflect.TypeOf(v)
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == nil {
		return nil
	}

	var keys []string
	collectKoanfKeys(t, "", &keys)
	return keys
}

func collectKoanfKeys(t reflect.Type, prefix string, keys *[]string) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue // unexported
		}

		tag := field.Tag.Get(koanfTag)
		if tag == "" || tag == "-" {
			continue
		}

		key := tag
		if prefix != "" {
			key = prefix + "." + tag
		}

		fieldType := field.Type
		for fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct && fieldType != timeType {
			collectKoanfKeys(fieldType, key, keys)
			continue
		}

		*keys = append(*keys, key)
	}
}
