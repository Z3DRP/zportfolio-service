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
	zlg.WithReportCaller(true),
)

func run() {
	zserver, err := routes.NewServer()
	if err != nil {
		logger.MustDebug(fmt.Sprintf("fatal error creating server: %s", err))
		panic(err)
	}

	if err := zserver.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.MustDebug(fmt.Sprintf("fatal server error: %s", err))
	}

	routes.HandleShutdown(zserver)
}

func main() {
	run()
}
