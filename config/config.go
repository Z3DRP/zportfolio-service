package config

import (
	"errors"
	"fmt"

	zlg "github.com/Z3DRP/zportfolio-service/internal/zlogger"
	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
)

const LogPrefix = "var/log"
const LogName = "logs.log"

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
	ZServer       ZServerConfig `mapstructure:"zserver"`
	DatabaseStore DbStoreConfig `mapstructure:"database"`
}

type ZServerConfig struct {
	Address      string `mapstructure:"address"`
	ReadTimeout  int    `mapstructure:"readTimeout"`
	WriteTimeout int    `mapstructure:"writeTimeout"`
}

type DbStoreConfig struct {
	DbName    string `mapstructure:"dbName"`
	DbUri     string `mapstructure:"dbUri"`
	ExpCol    string `mapstructure:"expCol"`
	DetailCol string `mapstructure:"detailCol"`
	SkillCol  string `mapstructure:"skillCol"`
	DbUsr     string `mapstructure:"default"`
	DbPwd     string `mapstructure:"dbPwd"`
}

func ReadServerConfig() (*ZServerConfig, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
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
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
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
