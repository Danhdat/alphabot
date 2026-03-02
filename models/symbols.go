package models

import (
	"time"

	"gorm.io/gorm"
)

type AlphaAPIResponse struct {
	Code    string        `json:"code"`
	Message interface{}   `json:"message"` // Can be null, so interface{} is safer
	Data    []AlphaSymbol `json:"data"`
}

// AlphaSymbol represents a token from the Alpha API.
// Note: Numerical values are stored as strings as the API returns them that way.
type AlphaSymbol struct {
	ID                uint   `gorm:"primaryKey"`
	TokenID           string `gorm:"unique;not null" json:"tokenId"`
	ChainID           string `gorm:"not null" json:"chainId"`
	ChainName         string `json:"chainName"`
	ContractAddress   string `gorm:"unique;not null" json:"contractAddress"`
	Symbol            string `gorm:"not null" json:"symbol"`
	PercentChange24h  string `gorm:"type:varchar(255)" json:"percentChange24h"`
	Volume24h         string `gorm:"type:varchar(255)" json:"volume24h"`
	MarketCap         string `gorm:"type:varchar(255)" json:"marketCap"`
	FDV               string `gorm:"type:varchar(255)" json:"fdv"`
	Liquidity         string `gorm:"type:varchar(255)" json:"liquidity"`
	TotalSupply       string `gorm:"type:varchar(255)" json:"totalSupply"`
	CirculatingSupply string `gorm:"type:varchar(255)" json:"circulatingSupply"`
	Holders           string `json:"holders"`
	ListingCex        bool   `json:"listingCex"`
	HotTag            bool   `json:"hotTag"`
	CanTransfer       bool   `json:"canTransfer"`
	Offline           bool   `json:"offline"`
	AlphaID           string `json:"alphaId"`
	Offsell           bool   `json:"offsell"`
	PriceHigh24h      string `gorm:"type:varchar(255)" json:"priceHigh24h"`
	PriceLow24h       string `gorm:"type:varchar(255)" json:"priceLow24h"`
	OnlineTge         bool   `json:"onlineTge"`
	OnlineAirdrop     bool   `json:"onlineAirdrop"`
	CexOffDisplay     bool   `json:"cexOffDisplay"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (AlphaSymbol) TableName() string {
	return "alpha_symbols"
}

func (s *AlphaSymbol) BeforeCreate(tx *gorm.DB) error {
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	return nil
}

func (s *AlphaSymbol) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}

type DataUpdate struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"unique;not null" json:"name"`
	LastUpdate time.Time `json:"last_update"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (d *DataUpdate) TableName() string {
	return "data_update"
}
