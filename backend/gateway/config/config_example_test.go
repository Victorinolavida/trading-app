package config

import (
	"reflect"
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	pkgconfig "github.com/victorinolavida/pkg/config"
)

// The example file is what `just setup` copies into config.yaml. A missing or
// mistyped key there silently produces an empty config value, so it is checked
// against the struct definitions on every test run.
func TestExampleConfigCoversEveryField(t *testing.T) {
	k := koanf.New(".")
	if err := k.Load(file.Provider("config.example.yaml"), yaml.Parser()); err != nil {
		t.Fatalf("load example config: %v", err)
	}

	declared := pkgconfig.KoanfKeys(Config{})
	present := make(map[string]bool, len(declared))
	for _, key := range declared {
		present[key] = true
	}

	// Every declared field must exist in the example file.
	for _, key := range declared {
		if !k.Exists(key) {
			t.Errorf("config.example.yaml is missing key %q", key)
		}
	}

	// The example file must not carry keys the structs do not know about.
	for key := range k.All() {
		if !present[key] {
			t.Errorf("config.example.yaml has unknown key %q (check the koanf tags)", key)
		}
	}
}

// The example config must actually produce a usable configuration.
func TestExampleConfigUnmarshals(t *testing.T) {
	k := koanf.New(".")
	if err := k.Load(file.Provider("config.example.yaml"), yaml.Parser()); err != nil {
		t.Fatalf("load example config: %v", err)
	}

	var conf Config
	if err := k.Unmarshal("", &conf); err != nil {
		t.Fatalf("unmarshal example config: %v", err)
	}

	// Fields that are legitimately empty in the example.
	allowedEmpty := map[string]bool{
		"cookie.domain": true, // empty domain means host-only cookie
		"cookie.secure": true, // false until https is in place
	}

	for _, key := range pkgconfig.KoanfKeys(Config{}) {
		value := k.Get(key)
		if isEmpty(value) && !allowedEmpty[key] {
			t.Errorf("config.example.yaml leaves %q empty", key)
		}
	}

	if conf.Cookie.Name == conf.Cookie.CSRFName {
		t.Errorf("cookie name and csrf name must differ, both are %q", conf.Cookie.Name)
	}
	if conf.Token.TTL <= 0 || conf.Cookie.TTL <= 0 {
		t.Error("token and cookie TTLs must be positive")
	}
	if conf.Cookie.TTL > conf.Token.TTL {
		t.Errorf("cookie ttl %s outlives token ttl %s", conf.Cookie.TTL, conf.Token.TTL)
	}
	if conf.Server.AllowOrigins == nil {
		t.Error("server.allow_origins must be set")
	}
}

func isEmpty(value any) bool {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}
