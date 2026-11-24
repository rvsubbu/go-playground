package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

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
		log.Fatal("server creation failure")
	}
	log.Printf("Basic API Server listening on port %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func chiSetup() *http.Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(util.IncrRequests)

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
		Handler:           r,
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
		Handler:           util.IncrRequests(router),
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
		Handler:           util.IncrRequests(mux),
		ReadHeaderTimeout: util.BasicConf.Server.ReadTimeout,
	}
}
