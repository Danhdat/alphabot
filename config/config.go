package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken  string
	TelegramChatID    string
	TelegramChannelID string
	ServerPort        string
	DBHost            string
	DBPort            string
	DBName            string
	DBUser            string
	DBPassword        string
}

var AppConfig *Config

func LoadConfig() {
	_ = godotenv.Load()

	AppConfig = &Config{
		TelegramBotToken:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:    os.Getenv("TELEGRAM_CHAT_ID"),
		TelegramChannelID: os.Getenv("TELEGRAM_CHANNEL_ID"),
		ServerPort:        os.Getenv("SERVER_PORT"),
		DBHost:            os.Getenv("DB_HOST"),
		DBPort:            os.Getenv("DB_PORT"),
		DBName:            os.Getenv("DB_NAME"),
		DBUser:            os.Getenv("DB_USER"),
		DBPassword:        os.Getenv("DB_PASSWORD"),
	}
}
