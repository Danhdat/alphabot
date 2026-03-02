package services

import (
	"alphabot/models"
	"alphabot/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const alphaAPIURL = "https://www.binance.com/bapi/defi/v1/public/wallet-direct/buw/wallet/cex/alpha/all/token/list"

type FetcherService struct{}

func NewFetcherService() *FetcherService {
	return &FetcherService{}
}

func (s *FetcherService) fetchAlphaFromAPI() ([]models.AlphaSymbol, error) {
	resp, err := http.Get(alphaAPIURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var apiResponse models.AlphaAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}
	var symbols []models.AlphaSymbol
	for _, alphaSymbol := range apiResponse.Data {
		symbol := models.AlphaSymbol{
			TokenID:           alphaSymbol.TokenID,
			ChainID:           alphaSymbol.ChainID,
			ChainName:         alphaSymbol.ChainName,
			ContractAddress:   alphaSymbol.ContractAddress,
			Symbol:            alphaSymbol.Symbol,
			PercentChange24h:  alphaSymbol.PercentChange24h,
			Volume24h:         alphaSymbol.Volume24h,
			MarketCap:         alphaSymbol.MarketCap,
			FDV:               alphaSymbol.FDV,
			Liquidity:         alphaSymbol.Liquidity,
			TotalSupply:       alphaSymbol.TotalSupply,
			CirculatingSupply: alphaSymbol.CirculatingSupply,
			Holders:           alphaSymbol.Holders,
			ListingCex:        alphaSymbol.ListingCex,
			HotTag:            alphaSymbol.HotTag,
			CanTransfer:       alphaSymbol.CanTransfer,
			Offline:           alphaSymbol.Offline,
			AlphaID:           alphaSymbol.AlphaID,
			Offsell:           alphaSymbol.Offsell,
			PriceHigh24h:      alphaSymbol.PriceHigh24h,
			PriceLow24h:       alphaSymbol.PriceLow24h,
			OnlineTge:         alphaSymbol.OnlineTge,
			OnlineAirdrop:     alphaSymbol.OnlineAirdrop,
			CexOffDisplay:     alphaSymbol.CexOffDisplay,
		}
		symbols = append(symbols, symbol)
	}
	return symbols, nil
}

func (s *FetcherService) FetchAlphaSymbols() error {
	if !models.NewCommonRepository().ShouldUpdate("alpha_symbols") {
		utils.Logger.Info().Msg("alpha_symbols is up to date")
		return nil
	}
	symbols, err := s.fetchAlphaFromAPI()
	if err != nil {
		utils.Logger.Error().Err(err).Msg("failed to fetch alpha symbols")
		return err
	}
	if err := models.NewAlphaSymbolRepository().SaveToDatabaseAlpha(symbols); err != nil {
		utils.Logger.Error().Err(err).Msg("failed to save alpha symbols to database")
		return err
	}
	// Filter out symbols with CexOffDisplay set to true
	var filteredSymbols []models.AlphaSymbol
	for _, symbol := range symbols {
		if !symbol.CexOffDisplay || symbol.ChainID != "8453" {
			filteredSymbols = append(filteredSymbols, symbol)
		}
	}

	if err := models.NewCommonRepository().UpdateLastUpdateTime("alpha_symbols"); err != nil {
		utils.Logger.Error().Err(err).Msg("failed to update last update time")
		return err
	}
	return nil
}

type Scheduler1 struct {
	fetchService *FetcherService
	stopChan     chan struct{}
}

func NewScheduler1(fetchService *FetcherService) *Scheduler1 {
	return &Scheduler1{
		fetchService: fetchService,
		stopChan:     make(chan struct{}),
	}
}
func (s *Scheduler1) Run() {
	utils.Logger.Info().Msg("scheduler started")
	if err := s.fetchService.FetchAlphaSymbols(); err != nil {
		utils.Logger.Error().Err(err).Msg("failed to fetch alpha symbols")
	}
	utils.Logger.Info().Msg("alpha symbols fetched successfully")
}
func (s *Scheduler1) Start() {
	nextSchedule := func() time.Time {
		now := time.Now()
		currentHour := now.Truncate(time.Hour)
		next := currentHour.Add(2*time.Hour + 30*time.Minute)
		if next.Before(now) {
			next = next.Add(2 * time.Hour)
		}
		return next
	}
	ticker := time.NewTimer(time.Until(nextSchedule()))
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ticker.C:
				s.Run()
				ticker.Reset(time.Until(nextSchedule()))
			case <-s.stopChan:
				utils.Logger.Info().Msg("scheduler stopped")
				return
			}
		}
	}()
}

func (s *Scheduler1) Stop() {
	close(s.stopChan)
}
