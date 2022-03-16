package main

import (
	"context"
	golog "log"
	"net/http"
	"os"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/whiteforestz/iino/internal/domain/sysmon"
	"github.com/whiteforestz/iino/internal/domain/tg"
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
		if err != nil {
			logger.Instance().Error("can't start app", zap.Error(err))
		}
	}()

	httpClient := &http.Client{}

	tgCfg, err := tg.NewConfig()
	if err != nil {
		return
	}

	sysMonDomain := sysmon.New()
	tgDomain := tg.New(httpClient, sysMonDomain, *tgCfg)

	sysMonDomain.Listen(ctx)
	tgDomain.Listen(ctx)

	logger.Instance().Info("Started! Press CTRL-C to interrupt...")

	defer func() {
		cancel()
		sysMonDomain.Wait()
		tgDomain.Wait()

		logger.Instance().Info("Bye!")
	}()

	<-ctx.Done()
}
