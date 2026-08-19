package repository

import (
	"shortlink/internal/model"
	"time"

	"gorm.io/gorm"
)

type AccessLogRepo struct {
	db *gorm.DB
}

func NewAccessLogRepo(db *gorm.DB) *AccessLogRepo {
	return &AccessLogRepo{db: db}
}

func (r *AccessLogRepo) Create(log *model.AccessLog) error {
	return r.db.Create(log).Error
}

func (r *AccessLogRepo) BatchCreate(logs []*model.AccessLog) error {
	return r.db.Create(logs).Error
}

type StatsResult struct {
	TodayPV  int64     `json:"today_pv"`
	TodayUV  int64     `json:"today_uv"`
	DailyPVs []DailyPV `json:"daily_pvs"`
}

type DailyPV struct {
	Date string `json:"date"`
	PV   int64  `json:"pv"`
}

// GetTodayStats 查询今日 PV/UV
func (r *AccessLogRepo) GetTodayStats(shortCode string, todayStart time.Time) (pv, uv int64, err error) {
	// PV
	err = r.db.Model(&model.AccessLog{}).
		Where("short_code = ? AND access_time >= ?", shortCode, todayStart).
		Count(&pv).Error
	if err != nil {
		return
	}
	// UV：按 IP 去重
	err = r.db.Model(&model.AccessLog{}).
		Where("short_code = ? AND access_time >= ?", shortCode, todayStart).
		Distinct("ip").
		Count(&uv).Error
	return
}

// GetDailyPVLast7Days 最近7天每日PV
func (r *AccessLogRepo) GetDailyPVLast7Days(shortCode string, since time.Time) ([]DailyPV, error) {
	var results []DailyPV
	err := r.db.Model(&model.AccessLog{}).
		Select("DATE(access_time) as date, COUNT(*) as pv").
		Where("short_code = ? AND access_time >= ?", shortCode, since).
		Group("DATE(access_time)").
		Order("date ASC").
		Scan(&results).Error
	return results, err
}
