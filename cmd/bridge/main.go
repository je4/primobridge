package main

import (
	"context"
	"flag"
	"github.com/je4/primobridge/v2/pkg/bridge"
	"github.com/je4/primobridge/v2/web"
	lm "github.com/je4/utils/v2/pkg/logger"
	"io"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var err error
	var configfile = flag.String("cfg", "/etc/tbbs.toml", "configuration file")

	flag.Parse()

	var config = &PrimoBridgeConfig{
		LogFile:   "",
		LogLevel:  "DEBUG",
		LogFormat: `%{time:2006-01-02T15:04:05.000} %{module}::%{shortfunc} [%{shortfile}] > %{level:.5s} - %{message}`,
		Addr:      "localhost:80",
		AddrExt:   "http://localhost:80/",
	}
	if err := LoadPrimoBridgeConfig(*configfile, config); err != nil {
		log.Printf("cannot load config file: %v", err)
	}

	// create logger instance
	logger, lf := lm.CreateLogger("Salon Digital", config.LogFile, nil, config.LogLevel, config.LogFormat)
	defer lf.Close()

	var accessLog io.Writer
	var f *os.File
	if config.AccessLog == "" {
		accessLog = os.Stdout
	} else {
		f, err = os.OpenFile(config.AccessLog, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			logger.Panicf("cannot open file %s: %v", config.AccessLog, err)
			return
		}
		defer f.Close()
		accessLog = f
	}

	var staticFS, templateFS fs.FS

	if config.StaticDir == "" {
		staticFS, err = fs.Sub(web.StaticFS, "static")
		if err != nil {
			logger.Panicf("cannot get subtree of static: %v", err)
		}
	} else {
		staticFS = os.DirFS(config.StaticDir)
	}

	if config.TemplateDir == "" {
		templateFS, err = fs.Sub(web.TemplateFS, "template")
		if err != nil {
			logger.Panicf("cannot get subtree of embedded template: %v", err)
		}
	} else {
		templateFS = os.DirFS(config.TemplateDir)
	}

	srv, err := bridge.NewServer(
		"PrimoBridge",
		config.Addr,
		config.AddrExt,
		config.PrimoSourceData,
		config.PrimoDeepLink,
		staticFS,
		templateFS,
		logger,
		accessLog,
	)
	if err != nil {
		logger.Panicf("cannot initialize server: %v", err)
	}
	go func() {
		if err := srv.ListenAndServe(config.CertPem, config.KeyPem); err != nil {
			log.Fatalf("server died: %v", err)
		}
	}()

	end := make(chan bool, 1)

	// process waiting for interrupt signal (TERM or KILL)
	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)

		signal.Notify(sigint, syscall.SIGTERM)
		signal.Notify(sigint, syscall.SIGKILL)

		<-sigint

		// We received an interrupt signal, shut down.
		logger.Infof("shutdown requested")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		srv.Shutdown(ctx)

		end <- true
	}()

	<-end
	logger.Info("server stopped")

}
