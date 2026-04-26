package config

import (
	"log"

	"gopkg.in/ini.v1"
)

type ConfigList struct {
	Port      string
	SQLDriver string
	DBName    string
	LogFile   string
	Static    string
}

func LoadConfig() *ConfigList {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalln(err)
	}

	return &ConfigList{
		Port:      cfg.Section("web").Key("Port").MustString("8080"),
		SQLDriver: cfg.Section("db").Key("driver").String(),
		DBName:    cfg.Section("db").Key("name").String(),
		LogFile:   cfg.Section("web").Key("logfile").String(),
		Static:    cfg.Section("web").Key("static").String(),
	}
}
