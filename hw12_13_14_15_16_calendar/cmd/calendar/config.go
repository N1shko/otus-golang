package main

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Logger  LoggerConf  `mapstructure:"logger"`
	Server  ServerConf  `mapstructure:"server"`
	Storage StorageConf `mapstructure:"storage"`
}

type LoggerConf struct {
	Level string `mapstructure:"level"`
}

type ServerConf struct {
	Port        int    `mapstructure:"port"`
	StorageType string `mapstructure:"storageType"`
}

type StorageConf struct {
	Type  string        `mapstructure:"type"`
	DB    DBConf        `mapstructure:"db"`
	Local LocalPathConf `mapstructure:"local"`
}

type DBConf struct {
	User             string `mapstructure:"user"`
	Password         string `mapstructure:"password"`
	Host             string `mapstructure:"host"`
	Port             int    `mapstructure:"port"`
	Database         string `mapstructure:"database"`
	ConnectionString string `mapstructure:"connectionString"`
}

type LocalPathConf struct {
	Path string `mapstructure:"path"`
}

func NewConfig(path string) (Config, error) {
	var cfg Config

	viper.SetConfigFile(path)

	viper.AutomaticEnv()
	viper.SetEnvPrefix("CALENDAR") //it is gonna be k8s secret one day
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return cfg, err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
