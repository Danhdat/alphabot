package services

import (
	"alphabot/analysis"
	"alphabot/models"
	"alphabot/utils"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type AnalyzeService struct {
	alphaRepo           *models.AlphaSymbolRepository
	volumeRepo          *models.AutoVolumeRecordRepository
	notificationLogRepo *models.NotificationLogRepository
	telegramBotService  *TelegramBotService
}

func NewAnalyzeService(telegramBotService *TelegramBotService) *AnalyzeService {
	return &AnalyzeService{
		alphaRepo:           models.NewAlphaSymbolRepository(),
		volumeRepo:          models.NewAutoVolumeRecordRepository(),
		notificationLogRepo: models.NewNotificationLogRepository(),
		telegramBotService:  telegramBotService,
	}
}

func (s *AnalyzeService) AnalyzeVolume(volumes []float64) models.VolumeAnalysis {
	for i, j := 0, len(volumes)-1; i < j; i, j = i+1, j-1 {
		volumes[i], volumes[j] = volumes[j], volumes[i]
	}
	if len(volumes) < models.VOLUME_SMA_PERIOD+1 {
		return models.VolumeAnalysis{}
	}
	currentVolume := decimal.NewFromFloat(volumes[len(volumes)-1])
	var sum float64
	for i := len(volumes) - models.VOLUME_SMA_PERIOD; i < len(volumes); i++ {
		sum += volumes[i]
	}
	volumeSMA := sum / float64(models.VOLUME_SMA_PERIOD)
	var volumeSignal, volumeStrength, confirmation string
	confirmation = "null"
	var volumeRatio decimal.Decimal
	if volumeSMA > 0 {
		volumeRatio = currentVolume.Div(decimal.NewFromFloat(volumeSMA))
	} else {
		volumeRatio = decimal.Zero
	}
	if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(models.VOLUME_SPIKE_3X)) {
		volumeSignal = "🔥 VOLUME EXPLOSION"
		volumeStrength = "EXTREME"
	} else if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(models.VOLUME_SPIKE_2X)) {
		volumeSignal = "🚀 HIGH VOLUME"
		volumeStrength = "STRONG"
	} else if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(models.VOLUME_SPIKE_1_5X)) {
		volumeSignal = "📈 ABOVE AVERAGE VOLUME"
		volumeStrength = "MODERATE"
		confirmation = "Tín hiệu TRUNG BÌNH - Có sự quan tâm tăng lên"
	} else if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(1.0)) {
		volumeSignal = "🟡 NORMAL VOLUME"
		volumeStrength = "NORMAL"
	} else {
		volumeSignal = "📉 LOW VOLUME"
		volumeStrength = "WEAK"
	}
	return models.VolumeAnalysis{
		CurrentVolume:  currentVolume,
		VolumeSMA21:    decimal.NewFromFloat(volumeSMA),
		VolumeRatio:    volumeRatio,
		VolumeSignal:   volumeSignal,
		VolumeStrength: volumeStrength,
		Confirmation:   confirmation,
	}

}

func (s *AnalyzeService) AnalyzeAndNotifyVolumes(channelID string) error {
	alphaSymbols, err := s.alphaRepo.GetAllAlphaName()
	if err != nil {
		return err
	}
	utils.Logger.Info().Msgf("Analyzing volumes for %d alpha symbols", len(alphaSymbols))
	processedSymbols := make(map[string]bool)
	loc := time.FixedZone("UTC+7", 7*60*60)
	for _, symbol := range alphaSymbols {
		if processedSymbols[symbol] {
			continue
		}
		records22, _ := s.volumeRepo.GetLastNBySymbol(symbol, 22)
		if len(records22) == 0 {
			continue
		}
		var volumes []float64
		for _, r := range records22 {
			volumes = append(volumes, r.QuoteAssetVolume)
		}
		//chỉ tới cây nến 21
		var totalCandlestickLength float64 = 0
		var totalCandlestickBody float64 = 0
		for _, r := range records22[1:] { // Bỏ qua records22[0] (nến 22 - mới nhất đã đóng)
			totalCandlestickLength += r.CandlestickLength()
			totalCandlestickBody += r.CandlestickBody()
		}
		averageCandlestickBody := totalCandlestickBody / float64(len(records22)-1)
		volumeAnalysis := s.AnalyzeVolume(volumes)
		// Lấy bản ghi MỚI NHẤT (records22[0]) - nến 22 (đã đóng gần nhất)
		latestRecord := records22[0]
		if (volumeAnalysis.VolumeStrength == "EXTREME" || volumeAnalysis.VolumeStrength == "STRONG") && latestRecord.QuoteAssetVolume > 1000 && !checkAllRedCandles(records22) {
			currentTime := time.Now().In(loc)
			formattedTime := currentTime.Format("2006-01-02 15:04:05")

			// Phân tích mô hình trên nến ĐÃ ĐÓNG (records22[1] và records22[2])
			engulfingResult := analysis.DetectEngulfing(records22)
			confirmation1 := engulfingResult.Confirmation
			pattern1 := engulfingResult.Pattern
			breakoutResult := analysis.DetectBreakout(records22, averageCandlestickBody)
			confirmation2 := breakoutResult.Confirmation
			pattern2 := breakoutResult.Pattern
			hammerResult := analysis.DetectHammer(records22)
			confirmation3 := hammerResult.Confirmation
			pattern3 := hammerResult.Pattern
			dojiResult := analysis.DetectDojiSpecial(records22)
			confirmation4 := dojiResult.Confirmation
			pattern4 := dojiResult.Pattern
			patternString := utils.FormatElements(pattern1, pattern2, pattern3, pattern4)
			confirmationString := utils.FormatElements(confirmation1, confirmation2, confirmation3, confirmation4)

			// Ưu tiên: Breakout > Engulfing > Hammer > Doji
			direction := 0
			if breakoutResult.IsDetected {
				direction = breakoutResult.Direction
			} else if engulfingResult.IsDetected {
				direction = engulfingResult.Direction
			} else if hammerResult.IsDetected {
				direction = hammerResult.Direction
			} else if dojiResult.IsDetected {
				direction = dojiResult.Direction
			}
			// Tạo chuỗi hiển thị các nến từ records22[4] đến records22[0]
			var candlestickPattern strings.Builder
			candlestickPattern.WriteString("💡 ")
			for i := 4; i >= 0; i-- {
				if records22[i].Candlestick() == 1 {
					candlestickPattern.WriteString("🟢")
				} else {
					candlestickPattern.WriteString("🔴")
				}
			}
			count, _ := s.notificationLogRepo.CountBySymbolToday(symbol)
			countofWeek, _ := s.notificationLogRepo.CountBySymbolThisWeek(symbol)
			message := fmt.Sprintf("💰 Symbol: %s\n"+
				"📅 Time: %s\n"+
				"🚀 Volume: %s (SMA21: %s)\n"+
				"💵 Price: %s\n"+
				"🎯 Strength: %s\n"+
				"🔥 Signal: %s\n"+
				"🔖 Daily Occurrences: %d\n"+
				"✨ Pattern: %s\n"+
				"📊 Confirmation: %s\n"+
				"💎 Weekly Occurrences: %d\n"+
				"%s\n",
				strings.TrimSuffix(latestRecord.Symbol, "USDT"),
				formattedTime,
				utils.FormatVolume(decimal.NewFromFloat(latestRecord.QuoteAssetVolume)),
				utils.FormatVolume(volumeAnalysis.VolumeSMA21),
				utils.FormatPrice(decimal.NewFromFloat(latestRecord.ClosePrice)),
				volumeAnalysis.VolumeStrength,
				volumeAnalysis.VolumeSignal,
				count+1,
				patternString,
				confirmationString,
				countofWeek+1,
				candlestickPattern.String(),
			)
			s.telegramBotService.SendTelegramToChannel(channelID, message)
			notificationLog := &models.NotificationLog{
				Symbol:    symbol,
				CreatedAt: time.Now(),
				Direction: direction,
			}
			s.notificationLogRepo.Create(notificationLog)
			// Đánh dấu symbol đã được xử lý
			processedSymbols[symbol] = true
			time.Sleep(1 * time.Second)
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func checkAllRedCandles(records []models.AutoVolumeRecord) bool {
	for i := 4; i >= 0; i-- {
		if records[i].Candlestick() == 1 {
			return false
		}
	}
	return true
}

type Scheduler3 struct {
	analyzeService *AnalyzeService
	channelID      string
	stopChan       chan struct{}
}

func NewScheduler3(analyzeService *AnalyzeService, channelID string) *Scheduler3 {
	return &Scheduler3{
		analyzeService: analyzeService,
		channelID:      channelID,
		stopChan:       make(chan struct{}),
	}
}
func (s *Scheduler3) Start() {
	// Hàm helper để tính thời gian đến giờ:02 phút tiếp theo
	nextSchedule := func() time.Time {
		now := time.Now()
		// Cắt lẻ đến giờ, sau đó thêm 1 giờ + 4 phút (ví dụ: 8:30 → 9:02:00)
		next := now.Truncate(time.Hour).Add(time.Hour + 4*time.Minute)
		return next
	}
	// Tạo timer với thời gian đến lần chạy tiếp theo (9:02:00 nếu now là 8:30:00)
	timer := time.NewTimer(time.Until(nextSchedule()))
	defer timer.Stop()
	go s.Run()
	for {
		select {
		case <-timer.C:
			go s.Run()
			timer.Reset(time.Until(nextSchedule()))
		case <-s.stopChan:
			utils.Logger.Info().Msg("Scheduler stopped")
			return
		}
	}
}

func (s *Scheduler3) Run() {
	if err := s.analyzeService.AnalyzeAndNotifyVolumes(s.channelID); err != nil {
		utils.Logger.Error().Err(err).Msg("failed to analyze and notify volumes")
	}
	utils.Logger.Info().Msg("analyze and notify completed")
}
func (s *Scheduler3) Stop() {
	close(s.stopChan)
}
