package logger

import (
	"github.com/benshields/messagebox/internal/pkg/config"
	"go.uber.org/zap"
)

const (
	development = "development"
	production  = "production"
)

type PresetError struct {
	Preset string
}

func (e PresetError) Error() string {
	return "unknown LoggerConfiguration.Preset value: " + e.Preset
}

func Setup(cfg config.LoggerConfiguration) (*zap.Logger, error) {
	var zlog *zap.Logger
	var err error
	switch cfg.Preset {
	case development:
		zlog, err = zap.NewDevelopment()
	case production:
		zlog, err = zap.NewProduction()
	default:
		return nil, PresetError{Preset: cfg.Preset}
	}
	return zlog, err
}
