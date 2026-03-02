package services

import (
	"alphabot/models"
	"alphabot/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type AutoVolumeService struct {
	volumeRepo *models.AutoVolumeRecordRepository
	alphaRepo  *models.AlphaSymbolRepository
}

func NewAutoVolumeService() *AutoVolumeService {
	return &AutoVolumeService{
		volumeRepo: models.NewAutoVolumeRecordRepository(),
		alphaRepo:  models.NewAlphaSymbolRepository(),
	}
}

func (s *AutoVolumeService) fetchAlphaKlines(symbol string) ([][]interface{}, error) {
	url := fmt.Sprintf("https://www.binance.com/bapi/defi/v1/public/alpha-trade/klines?interval=1h&limit=23&symbol=%s", symbol)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Cấu trúc response mới của API Alpha
	var alphaResponse struct {
		Code    string          `json:"code"`
		Message interface{}     `json:"message"`
		Data    [][]interface{} `json:"data"`
		Success bool            `json:"success"`
	}

	if err := json.Unmarshal(body, &alphaResponse); err != nil {
		return nil, fmt.Errorf("Error parsing alpha klines: %w", err)
	}

	if !alphaResponse.Success || alphaResponse.Code != "000000" {
		return nil, fmt.Errorf("API Alpha returned error: %v", alphaResponse.Message)
	}

	return alphaResponse.Data, nil
}

func (s *AutoVolumeService) FetchAndSaveAllSymbolsVolume() error {
	alphaSymbols, err := s.alphaRepo.GetAllAlphaSymbols()
	if err != nil {
		return fmt.Errorf("Error fetching alpha symbols: %w", err)
	}
	for _, alphaSymbol := range alphaSymbols {
		var klines [][]interface{}
		var err error
		klines, err = s.fetchAlphaKlines(alphaSymbol + "USDT")
		if err != nil {
			utils.Logger.Error().Msgf("Error fetching klines for alpha symbol %s: %v", alphaSymbol, err)
			continue
		}
		originalSymbol, err := s.alphaRepo.GetNameByAlphaSymbol(alphaSymbol)
		if err != nil {
			utils.Logger.Error().Msgf("Error fetching original symbol for alpha symbol %s: %v", alphaSymbol, err)
			continue
		}
		// Xử lý klines - ĐẢM BẢO CÓ ĐỦ 22 NẾN ĐÃ ĐÓNG
		// API trả về 23 nến: [nến 1, nến 2, ..., nến 22, nến 23]
		// Trong đó nến 23 là nến đang hình thành (chưa đóng)

		// Loại bỏ nến cuối cùng (nến 23 - đang hình thành) để chỉ lấy nến đã đóng
		if len(klines) > 1 {
			klines = klines[:len(klines)-1] // Kết quả: [nến 1, nến 2, ..., nến 22]
		}
		// Lấy 22 nến đã đóng (từ nến 1 đến nến 22)
		recentKlines := klines
		if len(klines) > 22 {
			recentKlines = klines[len(klines)-22:] // Lấy 22 nến cuối cùng
		}
		loc := time.FixedZone("UTC+7", 7*60*60)
		var records []models.AutoVolumeRecord
		for _, k := range recentKlines {
			openTime := utils.ParseKlineValue(k[0])
			quoteAssetVolumeStr := k[7].(string)
			quoteAssetVolume, _ := strconv.ParseFloat(quoteAssetVolumeStr, 64)
			openPriceStr := k[1].(string)
			openPrice, _ := strconv.ParseFloat(openPriceStr, 64)
			closePriceStr := k[4].(string)
			closePrice, _ := strconv.ParseFloat(closePriceStr, 64)
			highPriceStr := k[2].(string)
			highPrice, _ := strconv.ParseFloat(highPriceStr, 64)
			lowPriceStr := k[3].(string)
			lowPrice, _ := strconv.ParseFloat(lowPriceStr, 64)

			record := models.AutoVolumeRecord{
				Symbol:           originalSymbol,
				OpenTime:         openTime,
				QuoteAssetVolume: quoteAssetVolume,
				OpenPrice:        openPrice,
				ClosePrice:       closePrice,
				HighPrice:        highPrice,
				LowPrice:         lowPrice,
				CreatedAt:        time.Now().In(loc),
				UpdatedAt:        time.Now().In(loc),
				AlphaID:          alphaSymbol,
			}
			records = append(records, record)
		}
		if err := s.volumeRepo.ReplaceAllForSymbol(originalSymbol, records); err != nil {
			utils.Logger.Error().Msgf("Error saving DB %s: %v", originalSymbol, err)
		} else {
			utils.Logger.Info().Msgf("Updated %d records volume for %s (original: %s)",
				len(records), originalSymbol, alphaSymbol)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

type Scheduler2 struct {
	autoVolumeService *AutoVolumeService
	stopChan          chan struct{}
}

func NewScheduler2(autoVolumeService *AutoVolumeService) *Scheduler2 {
	return &Scheduler2{
		autoVolumeService: autoVolumeService,
		stopChan:          make(chan struct{}),
	}
}

func (s *Scheduler2) Start() {
	utils.Logger.Info().Msg("Scheduler Volume started")
	nextHour := func() time.Time {
		now := time.Now()
		next := now.Truncate(time.Hour).Add(time.Hour + 2*time.Minute)
		return next
	}
	timer := time.NewTimer(time.Until(nextHour()))
	defer timer.Stop()
	go s.Run()
	for {
		select {
		case <-timer.C:
			go s.Run()
			timer.Reset(time.Until(nextHour()))
		case <-s.stopChan:
			utils.Logger.Info().Msg("Scheduler stopped")
			return
		}
	}
}

func (s *Scheduler2) Run() {
	utils.Logger.Info().Msg("Running update")
	if err := s.autoVolumeService.FetchAndSaveAllSymbolsVolume(); err != nil {
		utils.Logger.Error().Msgf("Error when updating data: %v", err)
	}
	utils.Logger.Info().Msg("Update completed")

}

func (s *Scheduler2) Stop() {
	close(s.stopChan)
}
