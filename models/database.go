package models

import (
	"alphabot/config"
	"alphabot/utils"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() error {
	cfg := config.AppConfig
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		utils.Logger.Error().Err(err).Msg("failed to connect database")
		os.Exit(1)
	}
	utils.Logger.Info().Msg("connected to database")
	DB = db
	return nil
}

func CloseDatabase() {
	sqlDB, err := DB.DB()
	if err != nil {
		utils.Logger.Error().Err(err).Msg("failed to get database")
		os.Exit(1)
	}
	sqlDB.Close()
	utils.Logger.Info().Msg("closed database")
}

func AutoMigrate() error {
	utils.Logger.Info().Msg("migrating database")
	err := DB.AutoMigrate(
		&AlphaSymbol{},
		&AutoVolumeRecord{},
		&NotificationLog{},
		&HolderHistory{},
		&DataUpdate{},
	)
	if err != nil {
		utils.Logger.Error().Msg("failed to migrate database")
		os.Exit(1)
	}
	utils.Logger.Info().Msg("migrated database")
	return nil
}
