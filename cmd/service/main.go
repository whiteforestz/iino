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
	"github.com/whiteforestz/iino/internal/domain/persistor"
	"github.com/whiteforestz/iino/internal/domain/tglistener"
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
		persistorCfg  = persistor.MustNewConfig()
		hwWatcherCfg  = hwwatcher.MustNewConfig()
		wgWatcherCfg  = wgwatcher.MustNewConfig()
		tgListenerCfg = tglistener.MustNewConfig()
	)

	var (
		persistorDomain  = persistor.New(persistorCfg)
		hwWatcherDomain  = hwwatcher.New(hwWatcherCfg)
		wgWatcherDomain  = wgwatcher.New(wgWatcherCfg, persistorDomain)
		tgListenerDomain = tglistener.New(
			tgListenerCfg,
			httpClient,
			hwWatcherDomain,
			wgWatcherDomain,
		)
	)

	if err := persistorDomain.Prepare(); err != nil {
		return
	}

	hwWatcherDomain.Listen(ctx)
	wgWatcherDomain.Listen(ctx)
	tgListenerDomain.Listen(ctx)

	logger.Instance().Info("Started! Press CTRL-C to interrupt...")

	defer func() {
		cancel()
		hwWatcherDomain.Wait()
		tgListenerDomain.Wait()

		if err := persistorDomain.Clean(); err != nil {
			logger.Instance().Error("can't clean persistor", zap.Error(err))
		}

		logger.Instance().Info("Bye!")
	}()

	<-ctx.Done()
}
