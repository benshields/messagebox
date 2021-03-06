package db

import (
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/benshields/messagebox/internal/pkg/config"
)

const driver = "postgres"

var globalDB *gorm.DB // FIXME use DI instead of global

func Setup(cfg config.DatabaseConfiguration, log *zap.Logger) (*gorm.DB, error) {
	if log != nil {
		sugar := log.Sugar()
		defer sugar.Sync()
		sugar.Debugw("db.Setup", "config", cfg) // TODO this logs the password which is a vulnerability
	}

	dsn := "host=" + cfg.Host + " port=" + cfg.Port + " user=" + cfg.User + " dbname=" + cfg.DatabaseName + "  sslmode=disable password=" + cfg.Password
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	gdb, err := db.DB()
	if err != nil {
		return nil, err
	}

	gdb.SetMaxIdleConns(cfg.MaxIdleConns)
	gdb.SetMaxOpenConns(cfg.MaxOpenConns)

	globalDB = db

	return db, nil
}

func Get() *gorm.DB {
	return globalDB
}
