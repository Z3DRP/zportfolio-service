package config

import (
	"errors"
	"fmt"

	zlg "github.com/Z3DRP/zportfolio-service/internal/zlogger"
	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
)

const LogPrefix = "var/log"
const LogName = "portfolio.log"

var logfile = zlg.NewLogFile(
	zlg.WithFilename(fmt.Sprintf("%s/%s", LogPrefix, LogName)),
)
var logger = zlg.NewLogger(
	logfile,
	zlg.WithJsonFormatter(true),
	zlg.WithLevel("trace"),
	zlg.WithReportCaller(true),
)

func NewLogger(lf *lumberjack.Logger, lvl string, jsonfmtr, reportCaller bool) *zlg.Zlogrus {
	return zlg.NewLogger(
		logfile,
		zlg.WithJsonFormatter(jsonfmtr),
		zlg.WithLevel(lvl),
		zlg.WithReportCaller(reportCaller),
	)
}

type Configurations struct {
	ZServer       ZServerConfig
	DatabaseStore DbStoreConfig
}

type ZServerConfig struct {
	Address      string
	ReadTimeout  int
	WriteTimeout int
	Static       string
}

type DbStoreConfig struct {
	DbName string
	DbUri  string
	DbCol  string
}

func ReadServerConfig() (*ZServerConfig, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetConfigType("yml")
	var configs Configurations

	if err := viper.ReadInConfig(); err != nil {
		emsg := fmt.Sprintf("error reading config file, %v", err)
		logger.MustDebug(emsg)
		return nil, errors.New(emsg)
	}

	err := viper.Unmarshal(&configs)
	if err != nil {
		emsg := fmt.Sprintf("unable to decode config to json, %v", err)
		logger.MustDebug(emsg)
		return nil, errors.New(emsg)
	}
	return &configs.ZServer, nil
}

func ReadDbConfig() (*DbStoreConfig, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetConfigType("yml")
	var configs Configurations

	if err := viper.ReadInConfig(); err != nil {
		emsg := fmt.Sprintf("error reading config file, %v", err)
		logger.MustDebug(emsg)
		return nil, errors.New(emsg)
	}

	err := viper.Unmarshal(&configs)
	if err != nil {
		emsg := fmt.Sprintf("unable to decode config to json, %v", err)
		logger.MustDebug(emsg)
		return nil, errors.New(emsg)
	}
	return &configs.DatabaseStore, nil
}
