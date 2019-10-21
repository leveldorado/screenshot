package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/leveldorado/screenshot/bootstrap"
)

const (
	gracefulShutdownPeriod = 10 * time.Second
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	r, err := bootstrap.Build(ctx, os.Args)
	if err != nil {
		log.Println(fmt.Sprintf("failed to build runner: [error: %s]", err))
		os.Exit(1)
	}
	if err := r.Run(ctx); err != nil {
		log.Println(fmt.Sprintf("failed to run runner: [error: %s]", err))
		os.Exit(1)
	}
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
	cancel()
	stopContext, _ := context.WithTimeout(context.Background(), gracefulShutdownPeriod)
	if err := r.Stop(stopContext); err != nil {
		log.Println(fmt.Sprintf(`failed to stop runner with error: %s`, err))
	}
}
