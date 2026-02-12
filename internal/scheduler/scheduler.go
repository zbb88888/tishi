// Package scheduler manages cron-based task orchestration.
package scheduler

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/analyzer"
	"github.com/zbb88888/tishi/internal/collector"
	"github.com/zbb88888/tishi/internal/config"
	"github.com/zbb88888/tishi/internal/content"
)

// Scheduler orchestrates periodic tasks.
type Scheduler struct {
	pool      *pgxpool.Pool
	log       *zap.Logger
	cfg       *config.Config
	cron      *cron.Cron
	collector *collector.Collector
	analyzer  *analyzer.Analyzer
	generator *content.Generator
	taskLock  *taskLock
}

// taskLock prevents overlapping task execution.
type taskLock struct {
	mu      sync.Mutex
	running map[string]bool
}

func newTaskLock() *taskLock {
	return &taskLock{running: make(map[string]bool)}
}

func (l *taskLock) tryLock(name string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.running[name] {
		return false
	}
	l.running[name] = true
	return true
}

func (l *taskLock) unlock(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.running, name)
}

// New creates a new Scheduler.
func New(pool *pgxpool.Pool, log *zap.Logger, cfg *config.Config) *Scheduler {
	return &Scheduler{
		pool:      pool,
		log:       log.Named("scheduler"),
		cfg:       cfg,
		cron:      cron.New(cron.WithLogger(cron.VerbosePrintfLogger(zap.NewStdLog(log)))),
		collector: collector.New(pool, log.Named("collector"), cfg),
		analyzer:  analyzer.New(pool, log.Named("analyzer"), cfg),
		generator: content.NewGenerator(pool, log.Named("generator"), cfg),
		taskLock:  newTaskLock(),
	}
}

// Start begins the scheduler and blocks until ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) error {
	// Daily collection
	if _, err := s.cron.AddFunc(s.cfg.Scheduler.CollectCron, func() {
		s.runTask(ctx, "collect", func(ctx context.Context) error {
			return s.collector.Run(ctx)
		})
	}); err != nil {
		return err
	}

	// Daily analysis
	if _, err := s.cron.AddFunc(s.cfg.Scheduler.AnalyzeCron, func() {
		s.runTask(ctx, "analyze", func(ctx context.Context) error {
			return s.analyzer.Run(ctx)
		})
	}); err != nil {
		return err
	}

	// Weekly report
	if _, err := s.cron.AddFunc(s.cfg.Scheduler.WeeklyCron, func() {
		s.runTask(ctx, "weekly", func(ctx context.Context) error {
			return s.generator.Run(ctx, "weekly")
		})
	}); err != nil {
		return err
	}

	// Monthly report
	if _, err := s.cron.AddFunc(s.cfg.Scheduler.MonthlyCron, func() {
		s.runTask(ctx, "monthly", func(ctx context.Context) error {
			return s.generator.Run(ctx, "monthly")
		})
	}); err != nil {
		return err
	}

	s.cron.Start()
	s.log.Info("调度器已启动",
		zap.String("collect", s.cfg.Scheduler.CollectCron),
		zap.String("analyze", s.cfg.Scheduler.AnalyzeCron),
		zap.String("weekly", s.cfg.Scheduler.WeeklyCron),
		zap.String("monthly", s.cfg.Scheduler.MonthlyCron),
	)

	<-ctx.Done()
	s.cron.Stop()
	s.log.Info("调度器已停止")
	return nil
}

// runTask executes a task with locking and error handling.
func (s *Scheduler) runTask(ctx context.Context, name string, fn func(context.Context) error) {
	if !s.taskLock.tryLock(name) {
		s.log.Warn("任务正在运行中, 跳过", zap.String("task", name))
		return
	}
	defer s.taskLock.unlock(name)

	s.log.Info("任务开始", zap.String("task", name))
	if err := fn(ctx); err != nil {
		s.log.Error("任务失败", zap.String("task", name), zap.Error(err))
		return
	}
	s.log.Info("任务完成", zap.String("task", name))
}
