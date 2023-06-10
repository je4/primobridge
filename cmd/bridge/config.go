package main

import (
	"emperror.dev/errors"
	"github.com/BurntSushi/toml"
	"path/filepath"
	"strings"
	"time"
)

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

type DBMySQL struct {
	DSN            string
	ConnMaxTimeout duration
	Schema         string
}

type PrimoBridgeConfig struct {
	PrimoSourceData string  `toml:"primosourcedata"`
	PrimoDeepLink   string  `toml:"primodeeplink"`
	SiteViewerLink  string  `toml:"siteviewerlink"`
	BoxImagePath    string  `toml:"boximagepath"`
	CertPem         string  `toml:"certpem"`
	KeyPem          string  `toml:"keypem"`
	LogFile         string  `toml:"logfile"`
	LogLevel        string  `toml:"loglevel"`
	LogFormat       string  `toml:"logformat"`
	AccessLog       string  `toml:"accesslog"`
	BaseDir         string  `toml:"basedir"`
	StaticDir       string  `toml:"staticdir"`
	TemplateDir     string  `toml:"templatedir"`
	Addr            string  `toml:"addr"`
	AddrExt         string  `toml:"addrext"`
	DB              DBMySQL `toml:"db"`
}

func LoadPrimoBridgeConfig(fp string, conf *PrimoBridgeConfig) error {
	_, err := toml.DecodeFile(fp, conf)
	if err != nil {
		return errors.Wrapf(err, "error loading config file %v", fp)
	}
	conf.BaseDir = strings.TrimRight(filepath.ToSlash(conf.BaseDir), "/")
	conf.BoxImagePath = strings.TrimRight(filepath.ToSlash(conf.BoxImagePath), "/")
	return nil
}
