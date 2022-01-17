package db

import (
	"go.uber.org/zap"

	"github.com/benshields/messagebox/internal/pkg/config"
)

func Setup(cfg config.DatabaseConfiguration, log *zap.Logger) error {
	sugar := log.Sugar()
	defer sugar.Sync()
	sugar.Debugw("db.Setup", "config", cfg)
	return nil
}
