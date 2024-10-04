package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/Z3DRP/zportfolio-service/config"
	"github.com/Z3DRP/zportfolio-service/internal/controller"
	"github.com/Z3DRP/zportfolio-service/internal/middleware"
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

func NewServer() (*http.Server, error) {
	var serverConfig, scfgErr = config.ReadServerConfig()
	if scfgErr != nil {
		logger.MustDebug(fmt.Sprintf("error parsing config file, %s", scfgErr))
		return nil, fmt.Errorf("error occured while reading config: %s", scfgErr)
	}

	mux := http.NewServeMux()
	// TODO add serverConfig to Mux
	mux.HandleFunc("GET http://localhost:8081/about/", getAbout)
	mux.HandleFunc("POST https://localhost:8081/zypher/", getZypher)

	server := &http.Server{
		Addr:         serverConfig.Address,
		ReadTimeout:  time.Duration(serverConfig.ReadTimeout),
		WriteTimeout: time.Duration(serverConfig.WriteTimeout),
		Handler:      handlePanic(loggerMiddleware(contextMiddleware(mux, time.Duration(serverConfig.WriteTimeout)))),
	}

	return server, nil
}

func contextMiddleware(next http.Handler, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &middleware.WrappedWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		next.ServeHTTP(wrapped, r)
		logger.MustDebug(fmt.Sprintf("Method: %s, URI: %s, IP: %s, Duration: %v, Status: %v", r.Method, r.RequestURI, r.RemoteAddr, start, wrapped.StatusCode))
	})
}

func HandleShutdown(server *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	logger.MustDebug("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.MustFatal(fmt.Sprintf("Server forced shutdown: %v", err))
	}
	logger.MustDebug("Server exited")
}

func handlePanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.MustDebug(fmt.Sprintf("panic recovered %v", rec))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func getZypher(w http.ResponseWriter, r *http.Request) {
	// TODO getZypher will perform the zyphash func and return result
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		http.Error(w, "request time out", http.StatusRequestTimeout)
	default:
		txt := r.URL.Query().Get("txt")
		shift := r.URL.Query().Get("shft")
		shiftCount := r.URL.Query().Get("shftcount")
		hashCount := r.URL.Query().Get("hshcount")
		alternate := r.URL.Query().Get("alt")
		ignoreSpace := r.URL.Query().Get("ignspace")
		restrictHashShift := r.URL.Query().Get("restricthash")

		shf, err := parseInt(shift)
		if err != nil {
			http.Error(w, "invalid 'shift' parameter", http.StatusBadRequest)
			return
		}
		shfCount, err := parseInt(shiftCount)
		if err != nil {
			http.Error(w, "invalid 'shiftCount' parameter", http.StatusBadRequest)
			return
		}
		hshCount, err := parseInt(hashCount)
		if err != nil {
			http.Error(w, "invalid 'hashCount' parameter", http.StatusBadRequest)
			return
		}

		alt, err := parseBool(alternate)
		if err != nil {
			http.Error(w, "invalid 'alternate' parameter", http.StatusBadRequest)
			return
		}

		ignSpace, err := parseBool(ignoreSpace)
		if err != nil {
			http.Error(w, "invalid 'ignoreSpace' parameter", http.StatusBadRequest)
			return
		}

		rstHsh, err := parseBool(restrictHashShift)
		if err != nil {
			http.Error(w, "invalid 'restrictHash' parameter", http.StatusBadRequest)
			return
		}
		result, err := controller.CalculateZypher(txt, shf, shfCount, hshCount, *alt, *ignSpace, *rstHsh)
		if err != nil {
			http.Error(w, "error occured while calculating hash", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		response := map[string]string{
			"result": result,
		}
		json.NewEncoder(w).Encode(response)
	}
}

func getAbout(w http.ResponseWriter, r *http.Request) {
	// TODO call getPortfolioData then check for and log any errors or return data
}

func parseInt(param string) (int, error) {
	arg, err := strconv.Atoi(param)
	if err != nil {
		return -1, err
	}
	return arg, nil
}

func parseBool(param string) (*bool, error) {
	arg, err := strconv.ParseBool(param)
	if err != nil {
		return nil, err
	}
	return &arg, nil
}
