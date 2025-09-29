package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"scanner/src/web"
	"strings"
	"syscall"
	"time"
)

func main() { // 设置日志
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	stopCh := make(chan struct{})
	apiServer := web.ListenAndServe(getServerPort(), web.AddWebRoutes)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	close(stopCh)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	apiServer.Shutdown(ctx)
}

func getServerPort() string {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "5050"
	}

	if strings.ContainsAny(port, ":") {
		return port
	}

	return fmt.Sprintf(":%s", port)
}
