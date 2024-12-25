package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	StaticDir     string `yaml:"static_directory"`
	Url           string `yaml:"url"`
	Port          string `yaml:"port"`
	StoragePath   string `yaml:"storage_path"`
	WorksheetName string `yaml:"worksheet_name"`
}

const (
	defaultConfigPath = "config/config.yaml"
)

func MustLoad() *Config {
	path := fetchConfigPath()

	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config path does not exist" + path)
	}

	var config Config

	if err := cleanenv.ReadConfig(path, &config); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &config
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "config file path")
	flag.Parse()

	if res == "" {
		res = defaultConfigPath
	}
	return res
}
