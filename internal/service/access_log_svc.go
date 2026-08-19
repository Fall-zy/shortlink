package service

import (
	"shortlink/internal/model"
	"shortlink/internal/repository"
	"shortlink/internal/utils"
	"sync"
	"time"

	"go.uber.org/zap"
)

type AccessLogSvc struct {
	repo   *repository.AccessLogRepo
	buffer chan *model.AccessLog
	wg     sync.WaitGroup
}

func NewAccessLogSvc(repo *repository.AccessLogRepo, bufferSize int) *AccessLogSvc {
	svc := &AccessLogSvc{
		repo:   repo,
		buffer: make(chan *model.AccessLog, bufferSize),
	}
	svc.startWorker(2)
	return svc
}

func (s *AccessLogSvc) startWorker(n int) {
	for i := 0; i < n; i++ {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			for log := range s.buffer {
				if err := s.repo.Create(log); err != nil {
					utils.Logger.Error("写入访问日志失败",
						zap.String("short_code", log.ShortCode),
						zap.Error(err))
				}
			}
		}()
	}
}

func (s *AccessLogSvc) AsyncLog(log *model.AccessLog) {
	select {
	case s.buffer <- log:
	default:
		utils.Logger.Warn("访问日志缓冲期已满,丢弃日志", zap.String("short_code", log.ShortCode))
	}
}

func (s *AccessLogSvc) Shutdown() {
	close(s.buffer)
	s.wg.Wait()
	utils.Logger.Info("访问日志写入协程已全部退出")
}

func (s *AccessLogSvc) GetStats(shortCode string) (*repository.StatsResult, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	sevenDaysAgo := todayStart.AddDate(0, 0, -6) // 包含今天共7天

	pv, uv, err := s.repo.GetTodayStats(shortCode, todayStart)
	if err != nil {
		return nil, err
	}
	dailyPVs, err := s.repo.GetDailyPVLast7Days(shortCode, sevenDaysAgo)
	if err != nil {
		return nil, err
	}
	return &repository.StatsResult{
		TodayPV:  pv,
		TodayUV:  uv,
		DailyPVs: dailyPVs,
	}, nil
}
