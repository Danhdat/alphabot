package models

import (
	"time"

	"gorm.io/gorm"
)

type AlphaSymbolRepository struct {
	db *gorm.DB
}

func NewAlphaSymbolRepository() *AlphaSymbolRepository {
	return &AlphaSymbolRepository{db: DB}
}

func (r *AlphaSymbolRepository) GetAllAlphaSymbols() ([]string, error) {
	var symbols []AlphaSymbol
	err := r.db.Find(&symbols).Error
	if err != nil {
		return nil, err
	}
	var result []string
	for _, s := range symbols {
		if !s.CexOffDisplay {
			result = append(result, s.AlphaID)
		}
	}
	return result, nil
}

func (r *AlphaSymbolRepository) SaveToDatabaseAlpha(symbols []AlphaSymbol) error {
	// Xoá dữ liệu cũ
	if err := r.db.Unscoped().Where("1 = 1").Delete(&AlphaSymbol{}).Error; err != nil {
		return err
	}
	// Lưu dữ liệu mới
	if len(symbols) > 0 {
		if err := r.db.Create(&symbols).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *AlphaSymbolRepository) GetAllAlphaName() ([]string, error) {
	var symbols []AlphaSymbol
	err := r.db.Find(&symbols).Error
	if err != nil {
		return nil, err
	}
	var result []string
	for _, s := range symbols {
		if !s.CexOffDisplay {
			result = append(result, s.Symbol)
		}
	}
	return result, nil
}

func (r *AlphaSymbolRepository) GetBySymbol(symbol string) (*AlphaSymbol, error) {
	var alpha AlphaSymbol
	err := r.db.Where("symbol = ?", symbol).First(&alpha).Error
	if err != nil {
		return nil, err
	}
	return &alpha, nil
}

func (r *AlphaSymbolRepository) GetNameByAlphaSymbol(symbol string) (string, error) {
	var alphaSymbol AlphaSymbol
	err := r.db.Where("alpha_id = ?", symbol).First(&alphaSymbol).Error
	return alphaSymbol.Symbol, err
}

type CommonRepository struct {
	db *gorm.DB
}

func NewCommonRepository() *CommonRepository {
	return &CommonRepository{db: DB}
}

func (r *CommonRepository) UpdateLastUpdateTime(tableName string) error {
	var dataUpdate DataUpdate
	result := r.db.Model(&DataUpdate{}).Where("name = ?", tableName).First(&dataUpdate)
	if result.Error != nil {
		return r.db.Create(&DataUpdate{
			Name:       tableName,
			LastUpdate: time.Now(),
		}).Error
	}
	dataUpdate.LastUpdate = time.Now()
	return r.db.Save(&dataUpdate).Error
}

const updateInterval = 1 * 24 * time.Hour // 1 ngày
func (r *CommonRepository) ShouldUpdate(tableName string) bool {
	var dataUpdate DataUpdate
	err := r.db.Model(&DataUpdate{}).Where("name = ?", tableName).First(&dataUpdate).Error
	if err != nil {
		return true
	}
	return time.Since(dataUpdate.LastUpdate) > updateInterval
}

type AutoVolumeRecordRepository struct {
	db *gorm.DB
}

func NewAutoVolumeRecordRepository() *AutoVolumeRecordRepository {
	return &AutoVolumeRecordRepository{db: DB}
}

func (r *AutoVolumeRecordRepository) ReplaceAllForSymbol(symbol string, records []AutoVolumeRecord) error {
	// Bắt đầu transaction
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Xóa tất cả dữ liệu cũ của symbol
	if err := tx.Unscoped().Where("symbol = ?", symbol).Delete(&AutoVolumeRecord{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Thêm dữ liệu mới
	if len(records) > 0 {
		if err := tx.Create(&records).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction
	return tx.Commit().Error
}

func (r *AutoVolumeRecordRepository) Create(record *AutoVolumeRecord) error {
	return r.db.Create(record).Error
}

func (r *AutoVolumeRecordRepository) GetLastNBySymbol(symbol string, n int) ([]AutoVolumeRecord, error) {
	var records []AutoVolumeRecord
	err := r.db.Where("symbol = ?", symbol).Order("open_time DESC").Limit(n).Find(&records).Error
	return records, err
}

type NotificationLogRepository struct {
	db *gorm.DB
}

func NewNotificationLogRepository() *NotificationLogRepository {
	return &NotificationLogRepository{db: DB}
}
func (r *NotificationLogRepository) Create(log *NotificationLog) error {
	return r.db.Create(log).Error
}
func (r *NotificationLogRepository) CountBySymbolToday(symbol string) (int64, error) {
	var count int64
	loc := time.FixedZone("UTC+7", 7*60*60)
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	err := r.db.Model(&NotificationLog{}).
		Where("symbol = ? AND created_at >= ?", symbol, today).
		Count(&count).Error
	return count, err
}

func (r *NotificationLogRepository) CountBySymbolThisWeek(symbol string) (int64, error) {
	var count int64
	loc := time.FixedZone("UTC+7", 7*60*60)
	now := time.Now().In(loc)
	year, week := now.ISOWeek()

	// Tính thứ 2 đầu tuần và chủ nhật cuối tuần
	startOfWeek := firstDayOfISOWeek(year, week, loc)
	endOfWeek := startOfWeek.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	err := r.db.Model(&NotificationLog{}).
		Where("symbol = ? AND created_at >= ? AND created_at <= ?",
			symbol,
			startOfWeek,
			endOfWeek,
		).
		Count(&count).Error

	return count, err
}

func (r *NotificationLogRepository) GetLogsThisWeek() ([]NotificationLog, error) {
	loc := time.FixedZone("UTC+7", 7*60*60)
	now := time.Now().In(loc)
	year, week := now.ISOWeek()
	startOfWeek := firstDayOfISOWeek(year, week, loc)
	endOfWeek := startOfWeek.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	var logs []NotificationLog
	err := r.db.Where("created_at >= ? AND created_at <= ?", startOfWeek, endOfWeek).Find(&logs).Error
	return logs, err
}

func firstDayOfISOWeek(year, week int, loc *time.Location) time.Time {
	date := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}
	isoYear, isoWeek := date.ISOWeek()
	for isoYear < year || isoWeek < week {
		date = date.AddDate(0, 0, 7)
		isoYear, isoWeek = date.ISOWeek()
	}
	return date
}
