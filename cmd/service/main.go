package main

import (
	"context"
	golog "log"
	"net/http"
	"os"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/whiteforestz/iino/internal/domain/hwwatcher"
	"github.com/whiteforestz/iino/internal/domain/tg"
	"github.com/whiteforestz/iino/internal/domain/wgwatcher"
	"github.com/whiteforestz/iino/internal/pkg/logger"
	"github.com/whiteforestz/iino/internal/pkg/sig"
)

func main() {
	var err error

	ctx, cancel := sig.WithCancel(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	if err := godotenv.Load(); err != nil {
		golog.Println(err.Error())
		os.Exit(-1)
	}

	log, err := zap.NewProduction()
	if err != nil {
		golog.Println(err.Error())
		os.Exit(-1)
	}

	logger.Init(log)
	defer func() {
		if r := recover(); r != nil {
			logger.Instance().Error("can't start app", zap.Any("recover", r))
		}

		if err != nil {
			logger.Instance().Error("can't start app", zap.Error(err))
		}
	}()

	var (
		httpClient = &http.Client{}
	)

	var (
		hwWatcherCfg = hwwatcher.MustNewConfig()
		wgWatcherCfg = wgwatcher.MustNewConfig()
		tgCfg        = tg.MustNewConfig()
	)

	var (
		hwWatcherDomain = hwwatcher.New(hwWatcherCfg)
		wgWatcherDomain = wgwatcher.New(wgWatcherCfg)
		tgDomain        = tg.New(tgCfg, httpClient, hwWatcherDomain, wgWatcherDomain)
	)

	hwWatcherDomain.Listen(ctx)
	wgWatcherDomain.Listen(ctx)
	tgDomain.Listen(ctx)

	logger.Instance().Info("Started! Press CTRL-C to interrupt...")

	defer func() {
		cancel()
		hwWatcherDomain.Wait()
		tgDomain.Wait()

		logger.Instance().Info("Bye!")
	}()

	<-ctx.Done()
}
