package main

import (
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
)

type Config struct {
	Path string `koanf:"path"`
}

func ReadConfig(configFile string) (conf *Config, err error) {
	f := file.Provider(configFile)
	k := koanf.New("/")

	// load default values
	k.Load(confmap.Provider(map[string]interface{}{
		"goshot/path": ".",
	}, "/"), nil)

	// load YAML config
	if configFile != "" {
		if err = k.Load(f, yaml.Parser()); err != nil {
			return
		}
	}

	// unmarshal the data
	err = k.Unmarshal("goshot", &conf)
	if err != nil {
		return
	}

	// remove any trailing slashes
	conf.Path = strings.TrimRight(conf.Path, "/")

	return
}
