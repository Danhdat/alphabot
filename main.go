package main

import (
	"alphabot/config"
	"alphabot/models"
	"alphabot/services"
	"alphabot/utils"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	config.LoadConfig()

	if config.AppConfig == nil {
		utils.Logger.Fatal().Msg("Config is nil")
	}

	utils.Logger.Info().Msg("🚀 Starting server...")

	// 1️⃣ Init DB trước
	if err := models.InitDatabase(); err != nil {
		utils.Logger.Fatal().Err(err).Msg("failed to initialize database")
	}

	// 2️⃣ Migrate trước
	if err := models.AutoMigrate(); err != nil {
		utils.Logger.Fatal().Err(err).Msg("failed to migrate database")
	}

	defer models.CloseDatabase()

	// 3️⃣ Setup route
	http.HandleFunc("/health", healthCheck)

	utils.Logger.Info().Msg("Server running on port " + config.AppConfig.ServerPort)

	// 4️⃣ Start server cuối cùng
	go func() {
		utils.Logger.Info().Msg("Server running on port " + config.AppConfig.ServerPort)
		if err := http.ListenAndServe(":"+config.AppConfig.ServerPort, nil); err != nil && err != http.ErrServerClosed {
			utils.Logger.Fatal().Err(err).Msg("server failed")
		}
	}()
	//
	botService, err := services.NewTelegramBotService()
	if err != nil {
		utils.Logger.Error().Msgf("Error initializing bot: %v", err)
	}
	// 5️⃣ Start scheduler
	fetchService := services.NewFetcherService()
	scheduler := services.NewScheduler1(fetchService)
	scheduler.Start()

	autoVolumeService := services.NewAutoVolumeService()
	scheduler2 := services.NewScheduler2(autoVolumeService)
	scheduler2.Start()

	analyzeService := services.NewAnalyzeService(botService)
	scheduler3 := services.NewScheduler3(analyzeService, botService.GetChannelID())
	scheduler3.Start()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan
	utils.Logger.Info().Msg("Shutting down server...")
	scheduler.Stop()
	scheduler2.Stop()
	scheduler3.Stop()
	time.Sleep(2 * time.Second)
	utils.Logger.Info().Msg("Server stopped")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
