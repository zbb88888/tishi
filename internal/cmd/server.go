package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
	"github.com/zbb88888/tishi/internal/db"
	"github.com/zbb88888/tishi/internal/scheduler"
	"github.com/zbb88888/tishi/internal/server"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "启动 API Server + Scheduler",
	Long:  "启动 HTTP API 服务和定时任务调度器，这是 tishi 的主运行模式。",
	RunE:  runServer,
}

func runServer(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	log := logger.Named("server")

	// 连接数据库
	pool, err := db.Connect(cmd.Context(), cfg.Database)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer pool.Close()
	log.Info("数据库连接成功")

	// 创建 HTTP Server
	srv := server.New(pool, log, cfg)
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      srv.Router(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// 创建 Scheduler
	sched := scheduler.New(pool, log, cfg)

	// Graceful shutdown
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 启动 Scheduler
	go func() {
		if err := sched.Start(ctx); err != nil {
			log.Error("scheduler 退出", zap.Error(err))
		}
	}()

	// 启动 HTTP Server
	go func() {
		log.Info("HTTP 服务启动", zap.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP 服务异常退出", zap.Error(err))
		}
	}()

	// 等待退出信号
	sig := <-sigCh
	log.Info("收到退出信号", zap.String("signal", sig.String()))

	// 优雅关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	cancel() // stop scheduler

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP 服务关闭失败", zap.Error(err))
		return err
	}

	log.Info("服务已优雅关闭")
	return nil
}
