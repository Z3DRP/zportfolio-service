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

type ErrConfigRead struct {
	FileType     string
	Path         string
	ConfigObject string
	Err          error
}

func (ec *ErrConfigRead) Error() string {
	return fmt.Sprintf("An error occurred while reading %v config file: FileType: %v, Path: %v :: %v", ec.ConfigObject, ec.FileType, ec.Path, ec.Err)
}

func (ec *ErrConfigRead) Unwrap() error {
	return ec.Err
}

func NewConfigReadError(configObj string, e error) *ErrConfigRead {
	return &ErrConfigRead{
		FileType:     "yaml",
		Path:         "./config",
		ConfigObject: configObj,
		Err:          e,
	}
}

type Configurations struct {
	ZServer        ZServerConfig `mapstructure:"zserver"`
	DatabaseStore  DbStoreConfig `mapstructure:"database"`
	ZypherSettings ZypherConfig  `mapstructure:"zysettings"`
	ZEmailSettings ZEmailConfig  `mapstructure:"zemailsettings"`
}

type ZServerConfig struct {
	Address      string `mapstructure:"address"`
	ReadTimeout  int    `mapstructure:"readTimeout"`
	WriteTimeout int    `mapstructure:"writeTimeout"`
	Fchain       string `mapstructure:"fchain"`
	Pkey         string `mapstructure:"pkey"`
}

type DbStoreConfig struct {
	DbName          string `mapstructure:"dbName"`
	DbUri           string `mapstructure:"dbUri"`
	ExpCol          string `mapstructure:"expCol"`
	DetailCol       string `mapstructure:"detailCol"`
	SkillCol        string `mapstructure:"skillCol"`
	PeriodCol       string `mapstructure:"periodCol"`
	TaskCol         string `mapstructure:"taskCol"`
	VisitorCol      string `mapstructure:"visitorCol"`
	AvailabilityCol string `mapstructure:"availablityCol"`
	DbUsr           string `mapstructure:"default"`
	DbPwd           string `mapstructure:"dbPwd"`
}

type ZypherConfig struct {
	Shift        int  `mapstructure:"shift"`
	ShiftCount   int  `mapstructure:"shiftCount"`
	HashCount    int  `mapstructure:"hashCount"`
	Alternate    bool `mapstructure:"alternate"`
	IgnSpace     bool `mapstructure:"ignSpace"`
	RestrictHash bool `mapstructure:"restrictHash"`
}

type ZEmailConfig struct {
	SenderAddress   string `mapstructure:"senderAddress"`
	SenderPwd       string `mapstructure:"senderPwd"`
	RecieverAddress string `mapstructure:"recieverAddress"`
	SmtpServer      string `mapstructure:"smtpServer"`
	SmtpPort        int    `mapstructure:"smtpPort"`
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

func ReadZypherSettings() (ZypherConfig, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	var configs Configurations

	if err := viper.ReadInConfig(); err != nil {
		emsg := fmt.Sprintf("error reading config file, %v", err)
		logger.MustDebug(emsg)
		return ZypherConfig{}, errors.New(emsg)
	}

	err := viper.Unmarshal(&configs)
	if err != nil {
		emsg := fmt.Sprintf("unable to decode config to json:: %v", err)
		logger.MustDebug(emsg)
		return ZypherConfig{}, errors.New(emsg)
	}
	return configs.ZypherSettings, nil
}

func ReadEmailConfig() (*ZEmailConfig, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	var configs Configurations

	if err := viper.ReadInConfig(); err != nil {
		emsg := fmt.Sprintf("error reading config file, %v", err)
		logger.MustDebug(emsg)
		return nil, errors.New(emsg)
	}

	err := viper.Unmarshal(&configs)
	if err != nil {
		emsg := fmt.Sprintf("unable to decode config to json:: %v", err)
		logger.MustDebug(emsg)
		return nil, errors.New(emsg)
	}
	return &configs.ZEmailSettings, nil
}

func IsValidOrigin(origin string) bool {
	return true
	//	validOrigin := map[string]bool{
	//		"http://localhost:3000":      true,
	//		"https://localhost:3000":     true,
	//		"http://zachpalmer.dev":      true,
	//		"https://zachpalmer.dev":     true,
	//		"http://www.zachpalmer.dev":  true,
	//		"https://www.zachpalmer.dev": true,
	//		"www.zachpalmer.dev":         true,
	//	}
	//
	// return validOrigin[origin]
}
