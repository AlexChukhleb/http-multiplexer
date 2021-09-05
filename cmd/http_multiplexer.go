package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ach/http-multiplexer/internal/app"
	"github.com/ach/http-multiplexer/internal/delivery/http"
)

func main() {
	a := app.NewApp(&app.Config{
		MaxOutgoingConnections: getEnvInt("MaxOutgoingConnections", 4),
		MaxLinksPerRequest:     getEnvInt("MaxLinksPerRequest", 20),
		RequestTimeout:         time.Millisecond * time.Duration(getEnvInt("RequestTimeoutMS", 1000)),
	})

	server := http.NewServer(&http.Config{
		MaxIncomingConnections: getEnvInt("MaxIncomingConnections", 100),
		Port:                   getEnvInt("HttpPort", 8066),
	}, a)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		osCall := <-quit
		log.Printf("os call:%+v", osCall)
		cancelFunc()
	}()

	err := server.Run(ctx)
	if err != nil {
		log.Print(err.Error())
	}
}

func getEnvStr(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	value := getEnvStr(key, strconv.FormatInt(int64(defaultValue), 10))
	i64, err := strconv.ParseInt(value, 10, 0)
	if err != nil {
		return defaultValue
	}
	return int(i64)
}
