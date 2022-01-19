package config

import (
	"github.com/spf13/viper"
)

const defaultConfigPath = "config/default.yaml"

type ConfigError struct {
	ConfigPath string
	Err        error
}

func (e ConfigError) Error() string {
	return "config.New(" + e.ConfigPath + ") failed with: " + e.Err.Error()
}

type Configuration struct {
	Logger   LoggerConfiguration
	Server   ServerConfiguration
	Database DatabaseConfiguration
}

type LoggerConfiguration struct {
	Preset string
}

type ServerConfiguration struct {
	Host string
	Port string
	Mode string
}

type DatabaseConfiguration struct {
	DatabaseName string
	User         string
	Password     string
	Host         string
	Port         string
	MaxOpenConns int
	MaxIdleConns int
}

func New(configPath string) (*Configuration, error) {
	if configPath == "" {
		configPath = defaultConfigPath
	}
	return setup(configPath)
}

func setup(configPath string) (*Configuration, error) {
	var cfg *Configuration

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		setupErr := ConfigError{
			ConfigPath: configPath,
			Err:        err,
		}
		return nil, setupErr
	}

	err := viper.Unmarshal(&cfg)
	if err != nil {
		setupErr := ConfigError{
			ConfigPath: configPath,
			Err:        err,
		}
		return nil, setupErr
	}

	return cfg, nil
}
