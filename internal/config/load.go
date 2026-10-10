package config

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
)

func init() {
	config.AddDriver(yaml.Driver)

	config.WithOptions(config.ParseEnv, config.ParseDefault, config.ParseTime)

	config.WithOptions(func(opt *config.Options) {
		opt.DecoderConfig.TagName = "key"
	})
}

func Load(file string) *Config {
	err := config.LoadFiles(file)
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	c := &Config{}

	err = config.Decode(c)
	if err != nil {
		log.Fatal("failed to parse config: ", err)
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(c)
	if err != nil {
		log.Fatal("config validation failed: ", err)
	}

	return c
}
