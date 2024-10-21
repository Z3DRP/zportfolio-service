package main

import (
	"fmt"
	"net/http"

	"github.com/Z3DRP/zportfolio-service/config"
	"github.com/Z3DRP/zportfolio-service/internal/routes"
	zlg "github.com/Z3DRP/zportfolio-service/internal/zlogger"
)

var logfile = zlg.NewLogFile(
	zlg.WithFilename(fmt.Sprintf("%v/routes/%v", config.LogPrefix, config.LogName)),
)
var logger = zlg.NewLogger(
	logfile,
	zlg.WithJsonFormatter(true),
	zlg.WithLevel("trace"),
	zlg.WithReportCaller(false),
)

func run() {
	var serverConfig, scfgErr = config.ReadServerConfig()
	if scfgErr != nil {
		logger.MustDebug(fmt.Sprintf("error parsing config file, %s", scfgErr))
		// fmt.Errorf("error occured while reading config: %s", scfgErr)
		return
	}
	zserver, err := routes.NewServer(*serverConfig)
	if err != nil {
		logger.MustDebug(fmt.Sprintf("fatal error creating server: %s", err))
		panic(err)
	}

	// err = http.ListenAndServeTLS(serverConfig.Address, serverConfig.Fchain, serverConfig.Pkey, nil)
	// if err != nil {
	// 	logger.MustDebug(fmt.Sprintf("could not start tls server: %s", err))
	// }
	if err := zserver.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.MustDebug(fmt.Sprintf("fatal server error: %s", err))
	}

	logger.MustDebug("server is live and running")

	routes.HandleShutdown(zserver)
}

func main() {
	run()
}
