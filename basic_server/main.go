package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gin-gonic/gin"

	"github.com/rvsubbu/go-playground/util"
)

var (
	version   = "0.1.0"
	gitSHA    = "unknown"
	buildTime = "unknown"
	defaultConfigFile = "config/config.yaml"
)

var buildInfo *util.BuildInfo

func init() {
	buildInfo = &util.BuildInfo{
    	Version:   version,
    	GitSHA:    gitSHA,
    	BuildTime: buildTime,
	}

	configFile := flag.String("config", "", "path to config file")
	flag.Parse()

	if *configFile == "" {
		configFile = &defaultConfigFile
	}
	err := util.ConfigInit(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(strings.ToUpper(util.BasicConf.Log.Level))); err != nil {
		log.Fatalf("invalid log level %q: %v", util.BasicConf.Log.Level, err)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))
	fmt.Printf("%+v\n", util.BasicConf)
}

func main() {
	var server *http.Server
	switch util.BasicConf.Server.WebServer {
	case "chi":
		server = chiSetup()
	case "gin":
		server = ginSetup()
	case "net/http":
		server = netHttpSetup()
	default:
		server = netHttpSetup()
	}

	if server == nil {
		slog.Error("server creation failure")
		os.Exit(1)
	}
	if err := runServer(server); err != nil {
		slog.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}

func runServer(server *http.Server) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	slog.Info("Basic API Server listening", "addr", server.Addr)
	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		slog.Info("Shutdown signal received")
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	shutdownErr := server.Shutdown(shutdownContext)
	if shutdownErr != nil {
		slog.Warn("Graceful shutdown timed out; closing active connections", "error", shutdownErr)
		if err := server.Close(); err != nil {
			return errors.Join(shutdownErr, fmt.Errorf("force close server: %w", err))
		}
	}

	if err := <-serverErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	if shutdownErr != nil {
		return fmt.Errorf("graceful shutdown: %w", shutdownErr)
	}
	return nil
}

func chiSetup() *http.Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		util.VersionHandler(buildInfo)(w, r)
	})
	r.Get("/config", func(w http.ResponseWriter, r *http.Request) {
		util.ConfigHandler(buildInfo)(w, r)
	})
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		util.HealthzHandler()(w, r)
	})
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		util.MetricsHandler()(w, r)
	})

	return &http.Server{
		Addr:              util.BasicConf.Server.Addr,
		Handler:           util.InstrumentRequests(util.BasicConf.App.Name)(r),
		ReadHeaderTimeout: util.BasicConf.Server.ReadTimeout,
	}
}

func ginSetup() *http.Server {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/version", gin.WrapF(util.VersionHandler(buildInfo)))
	router.GET("/config", gin.WrapF(util.ConfigHandler(buildInfo)))
	router.GET("/healthz", gin.WrapF(util.HealthzHandler()))
	router.GET("/metrics", gin.WrapF(util.MetricsHandler()))

	return &http.Server{
		Addr:              util.BasicConf.Server.Addr,
		Handler:           util.InstrumentRequests(util.BasicConf.App.Name)(router),
		ReadHeaderTimeout: util.BasicConf.Server.ReadTimeout,
	}
}

func netHttpSetup() *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		util.VersionHandler(buildInfo)(w, r)
	})

	mux.HandleFunc("GET /config", func(w http.ResponseWriter, r *http.Request) {
		util.ConfigHandler(buildInfo)(w, r)
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		util.HealthzHandler()(w, r)
	})

	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		util.MetricsHandler()(w, r)
	})

	return &http.Server{
		Addr:              util.BasicConf.Server.Addr,
		Handler:           util.InstrumentRequests(util.BasicConf.App.Name)(mux),
		ReadHeaderTimeout: util.BasicConf.Server.ReadTimeout,
	}
}
