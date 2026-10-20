package config

import (
	"fmt"
	"strings"

	"github.com/flashlabs/rootpath/location"
	"github.com/knadh/koanf/parsers/yaml"
	env2 "github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var k = koanf.New(".")

func getConfigFile(env []string) (*koanf.Koanf, error) {
	k = koanf.New(".")
	confile := "config/config.yaml"
	if len(env) > 0 && env[0] != "" {
		confile = fmt.Sprintf("config/config.%s.yaml", env[0])
		if err := location.Chdir(); err != nil {
			return k, nil
		}
	}

	if err := k.Load(file.Provider(confile), yaml.Parser()); err != nil {
		return nil, err
	}

	return k, nil
}

func NewConfig[T any](env ...string) (*T, error) {
	k, err := getConfigFile(env)
	if err != nil {
		return nil, err
	}

	err = k.Load(env2.Provider("", ".", func(s string) string {
		val := strings.ToLower(strings.ReplaceAll(s, "_", "."))
		return val
	}), nil)
	if err != nil {
		return nil, err
	}
	conf := new(T)
	err = k.Unmarshal("", conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
