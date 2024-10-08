package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/Z3DRP/zportfolio-service/config"
	"github.com/Z3DRP/zportfolio-service/internal/controller"
	"github.com/Z3DRP/zportfolio-service/internal/dacstore"
	"github.com/Z3DRP/zportfolio-service/internal/middleware"
	"github.com/Z3DRP/zportfolio-service/internal/models"
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

func NewServer() (*http.Server, error) {
	var serverConfig, scfgErr = config.ReadServerConfig()
	if scfgErr != nil {
		logger.MustDebug(fmt.Sprintf("error parsing config file, %s", scfgErr))
		return nil, fmt.Errorf("error occured while reading config: %s", scfgErr)
	}

	mux := http.NewServeMux()
	// TODO add serverConfig to Mux
	mux.HandleFunc("GET /about", getAbout)
	mux.HandleFunc("POST /zypher", getZypher)

	server := &http.Server{
		Addr:         serverConfig.Address,
		ReadTimeout:  time.Second * time.Duration(serverConfig.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(serverConfig.WriteTimeout),
		Handler:      handlePanic(loggerMiddleware(headerMiddleware(contextMiddleware(mux, 10*time.Second)))),
	}

	return server, nil
}

func headerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Headers", "Content-Type")
		next.ServeHTTP(w, r)
	})
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
			logger.MustDebug("invalid shift param")
			http.Error(w, "invalid 'shift' parameter", http.StatusBadRequest)
			return
		}
		shfCount, err := parseInt(shiftCount)
		if err != nil {
			logger.MustDebug("invalid shift count param")
			http.Error(w, "invalid 'shiftCount' parameter", http.StatusBadRequest)
			return
		}
		hshCount, err := parseInt(hashCount)
		if err != nil {
			logger.MustDebug("invalid hash count")
			http.Error(w, "invalid 'hashCount' parameter", http.StatusBadRequest)
			return
		}

		alt, err := parseBool(alternate)
		if err != nil {
			logger.MustDebug("invalid alternate param")
			http.Error(w, "invalid 'alternate' parameter", http.StatusBadRequest)
			return
		}

		ignSpace, err := parseBool(ignoreSpace)
		if err != nil {
			logger.MustDebug("invalid ignore space param")
			http.Error(w, "invalid 'ignoreSpace' parameter", http.StatusBadRequest)
			return
		}

		rstHsh, err := parseBool(restrictHashShift)
		if err != nil {
			logger.MustDebug("invalid restrict hash param")
			http.Error(w, "invalid 'restrictHash' parameter", http.StatusBadRequest)
			return
		}
		result, err := controller.CalculateZypher(txt, shf, shfCount, hshCount, *alt, *ignSpace, *rstHsh)
		if err != nil {
			logger.MustDebug(fmt.Sprintf("error occurred while calculating zypher: %s", err))
			http.Error(w, "error occured while calculating hash", http.StatusInternalServerError)
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
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		logger.MustDebug(fmt.Sprintf("request: %s method: %s timed out", r.URL, r.Method))
		http.Error(w, "request time out", http.StatusRequestTimeout)
	default:
		var portfolioData models.Responser
		portfolioData, err := controller.GetPortfolioData(r.Context())
		if err != nil {
			if errors.Is(err, dacstore.ErrFetchSkill) {
				logger.MustDebug(fmt.Sprintf("could not retrieve skill data, %s", err))
				http.Error(w, "could not retrieve skill data", http.StatusInternalServerError)
				return
			} else if errors.Is(err, dacstore.ErrFetchExperience) {
				logger.MustDebug(fmt.Sprintf("could not retrieve experience data, %s", err))
				http.Error(w, "could not retrieve experience data", http.StatusInternalServerError)
				return
			} else if errors.Is(err, dacstore.ErrFetchDetails) {
				logger.MustDebug(fmt.Sprintf("could not retrieve details data, %s", err))
				http.Error(w, "could not retrieve details data", http.StatusInternalServerError)
				return
			}

			logger.MustDebug(fmt.Sprintf("an unexpected error occurred while fetching portfolio data: %s", err))
			http.Error(w, "an unexpecting error occurred while fetching data", http.StatusInternalServerError)
			return
		}

		logger.MustDebug(fmt.Sprintf("respons data:: %v", portfolioData))
		if pdata, ok := portfolioData.(*models.PortoflioResponse); ok {
			if err := json.NewEncoder(w).Encode(pdata); err != nil {
				logger.MustDebug(fmt.Sprintf("could not encode portfolio response: %s", err))
				http.Error(w, "could not encode response", http.StatusInternalServerError)
				return
			}
		} else {
			msg := fmt.Sprintf("could not cast type: [%T] into portfolio data", portfolioData)
			logger.MustDebug(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
	}
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
