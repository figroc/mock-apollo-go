package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/lalamove/mock-apollo-go/internal/routes/apollo"
	"github.com/lalamove/nui/nlogger"
	"github.com/sirupsen/logrus"
)

var (
	filePath     string
	configPort   int
	internalPort int
	pollTimeout  time.Duration
	logger       nlogger.Provider
)

func init() {
	flag.StringVar(&filePath, "file", "./configs/example.yaml", "config filepath")
	flag.IntVar(&internalPort, "internal-port", 9090, "internal HTTP server port")
	flag.IntVar(&configPort, "config-port", 8070, "config HTTP server port")
	flag.DurationVar(&pollTimeout, "poll-timeout", time.Minute, "long poll timeout")
	flag.Parse()
	validateInput()
	logger = nlogger.NewProvider(newLogger(logrus.InfoLevel))
}

func validateInput() {
	if _, err := os.Stat(filePath); err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	// internal server for telemetry and ctrl
	internalRouter := httprouter.New()
	ctrlRoutes(internalRouter)
	pprofRoutes(internalRouter)
	internalSrv := &http.Server{
		Addr:    ":" + strconv.Itoa(internalPort),
		Handler: internalRouter,
	}
	go func() {
		if err := internalSrv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// public server for serving config via Apollo APIs
	router := httprouter.New()
	a, err := apollo.New(ctx, apollo.Config{
		ConfigPath:  filePath,
		PollTimeout: pollTimeout,
		Log:         logger,
	})
	if err != nil {
		log.Fatal(err)
	}
	a.Routes(router)
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(configPort),
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// graceful shutdown
	<-termChan
	cancel()
	internalSrv.Close()
	srv.Close()
	logger.Get().Info("shutting down")
}
