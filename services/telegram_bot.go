package services

import (
	"alphabot/config"
	"alphabot/utils"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBotService struct {
	bot       *tgbotapi.BotAPI
	chatID    int64
	channelID string
}

func NewTelegramBotService() (*TelegramBotService, error) {
	bot, err := tgbotapi.NewBotAPI(config.AppConfig.TelegramBotToken)
	if err != nil {
		utils.Logger.Error().Err(err).Msg("failed to create telegram bot")
		return nil, err
	}
	var chatID int64
	if config.AppConfig.TelegramChatID != "" {
		if parsedChatID, err := strconv.ParseInt(config.AppConfig.TelegramChatID, 10, 64); err == nil {
			chatID = parsedChatID
		}
	}
	channelID := config.AppConfig.TelegramChannelID
	return &TelegramBotService{
		bot:       bot,
		chatID:    chatID,
		channelID: channelID,
	}, nil
}

func (s *TelegramBotService) StartBot() {
	utils.Logger.Info().Msg("🚀 Starting bot...")
	utils.Logger.Info().Msg("✅ Bot is ready to receive messages...")
	utils.Logger.Info().Msgf("Bot started: %s", s.bot.Self.UserName)
}

func (s *TelegramBotService) Stop() {
	utils.Logger.Info().Msg("🛑 Stopping bot...")
}

func (s *TelegramBotService) SendTelegramToChannel(channelID string, message string) {
	utils.Logger.Info().
		Str("channel_id", channelID).
		Msg("Sending message to channel")

	msg := tgbotapi.NewMessageToChannel(channelID, message)
	//msg.ParseMode = "MarkdownV2"

	_, err := s.bot.Send(msg)
	if err != nil {
		utils.Logger.Error().
			Err(err).
			Str("channel_id", channelID).
			Msg("Failed to send message to channel")
		return
	}

	utils.Logger.Info().
		Str("channel_id", channelID).
		Msg("Message sent successfully")
}

func (s *TelegramBotService) GetChannelID() string {
	return s.channelID
}
