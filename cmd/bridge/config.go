package main

import (
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"path/filepath"
	"strings"
)

type PrimoBridgeConfig struct {
	PrimoSourceData string `toml:"primosourcedata"`
	PrimoDeepLink   string `toml:"primodeeplink"`
	BoxImagePath    string `toml:"boximagepath"`
	CertPem         string `toml:"certpem"`
	KeyPem          string `toml:"keypem"`
	LogFile         string `toml:"logfile"`
	LogLevel        string `toml:"loglevel"`
	LogFormat       string `toml:"logformat"`
	AccessLog       string `toml:"accesslog"`
	BaseDir         string `toml:"basedir"`
	StaticDir       string `toml:"staticdir"`
	TemplateDir     string `toml:"templatedir"`
	Addr            string `toml:"addr"`
	AddrExt         string `toml:"addrext"`
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
