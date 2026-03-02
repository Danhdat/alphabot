package models

import (
	"math"
	"time"

	"github.com/shopspring/decimal"
)

type AutoVolumeRecord struct {
	ID               uint    `gorm:"primaryKey"`
	Symbol           string  `gorm:"index;not null"`
	OpenTime         float64 `gorm:"not null"`
	QuoteAssetVolume float64 `gorm:"not null"`
	OpenPrice        float64 `gorm:"not null"`
	ClosePrice       float64 `gorm:"not null"`
	HighPrice        float64 `gorm:"not null"`
	LowPrice         float64 `gorm:"not null"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	AlphaID          string
}

func (AutoVolumeRecord) TableName() string {
	return "auto_volume_records"
}

func (r *AutoVolumeRecord) Candlestick() float64 {
	// 1 : green 0: red
	if r.ClosePrice > r.OpenPrice {
		return 1
	} else {
		return 0
	}
}

func (r *AutoVolumeRecord) CandlestickBody() float64 {
	return math.Abs(r.ClosePrice - r.OpenPrice)
}

func (r *AutoVolumeRecord) CandlestickLength() float64 {
	return r.HighPrice - r.LowPrice
}

func (r *AutoVolumeRecord) CandlestickUpperShadow() float64 {
	if r.ClosePrice > r.OpenPrice {
		return r.HighPrice - r.ClosePrice
	} else {
		return r.HighPrice - r.OpenPrice
	}
}

func (r *AutoVolumeRecord) CandlestickLowerShadow() float64 {
	if r.ClosePrice > r.OpenPrice {
		return r.OpenPrice - r.LowPrice
	} else {
		return r.ClosePrice - r.LowPrice
	}
}

func (r *AutoVolumeRecord) IsCandlestickBodyLong(avgLength float64, multiplier float64) bool {
	return r.CandlestickBody() > avgLength*multiplier
}

func (r *AutoVolumeRecord) IsCandlestickBodyShort(avgLength float64, multiplier float64) bool {
	return r.CandlestickBody() < avgLength*multiplier
}

func (r *AutoVolumeRecord) CandlestBodyMidpoint() float64 {
	if r.OpenPrice >= r.ClosePrice { // green candlestick
		return r.OpenPrice + (r.ClosePrice-r.OpenPrice)/2
	} else { // red candlestick
		return r.ClosePrice + (r.OpenPrice-r.ClosePrice)/2
	}
}

type VolumeAnalysis struct {
	CurrentVolume  decimal.Decimal
	VolumeSMA21    decimal.Decimal
	VolumeRatio    decimal.Decimal
	VolumeSignal   string
	VolumeStrength string
	Confirmation   string
}

const (
	VOLUME_SMA_PERIOD = 21  // SMA của Volume (21 kỳ)
	VOLUME_SPIKE_1_5X = 1.5 // Volume spike 1.5x
	VOLUME_SPIKE_2X   = 2.0 // Volume spike 2x
	VOLUME_SPIKE_3X   = 3.0 // Volume spike 3x
)
