package main

import (
	"context"
	"database/sql"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"github.com/je4/primobridge/v2/pkg/bridge"
	"github.com/je4/primobridge/v2/pkg/mediathek"
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

/*
https://slsp-fhnw.primo.exlibrisgroup.com/discovery/fulldisplay?context=L&vid=41SLSP_FNW:testfabiano&search_scope=DN_and_CI&tab=41SLSP_FHNW_DN_and_CI&docid=alma990038750950205518
*/

func main() {
	var err error
	var configfile = flag.String("cfg", "/etc/primobridge.toml", "configuration file")

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

	var staticFS, templateFS, boxImageFS fs.FS

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

	if config.BoxImagePath == "" {
		boxImageFS, err = fs.Sub(web.StaticFS, "static/3dthumb/jpg")
		if err != nil {
			logger.Panicf("cannot get subtree of embedded template: %v", err)
		}
	} else {
		boxImageFS = os.DirFS(config.BoxImagePath)
	}

	var db *sql.DB
	logger.Debugf("connecting mysql database")
	db, err = sql.Open("mysql", config.DB.DSN)
	if err != nil {
		// don't write dsn in error message due to password inside
		logger.Panicf("error connecting to database: %v", err)
		return
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		logger.Panicf("cannot ping database: %v", err)
		return
	}
	db.SetConnMaxLifetime(time.Duration(config.DB.ConnMaxTimeout.Duration))

	mapper, err := mediathek.NewMediathekMapper(db, boxImageFS, config.SiteViewerLink, logger)
	if err != nil {
		logger.Panicf("cannot instianziate mapper: %v", err)
	}

	srv, err := bridge.NewServer(
		"PrimoBridge",
		config.Addr,
		config.AddrExt,
		config.PrimoSourceData,
		config.PrimoDeepLink,
		staticFS,
		templateFS,
		mapper,
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
