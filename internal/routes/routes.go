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
	zlg.WithFilename(fmt.Sprintf("%v/%v", config.LogPrefix, "routes.log")),
)
var logger = zlg.NewLogger(
	logfile,
	zlg.WithJsonFormatter(true),
	zlg.WithLevel("trace"),
	zlg.WithReportCaller(false),
)

func NewServer(sconfig config.ZServerConfig) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /about", getAbout)
	mux.HandleFunc("POST /zypher", getZypher)
	mux.HandleFunc("GET /schedule", getSchedule)
	mux.HandleFunc("POST /task", createTask)

	server := &http.Server{
		Addr:         sconfig.Address,
		ReadTimeout:  time.Second * time.Duration(sconfig.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(sconfig.WriteTimeout),
		Handler:      handlePanic(loggerMiddleware(headerMiddleware(contextMiddleware(mux, 10*time.Second)))),
	}

	return server, nil
}

func getSchedule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		logger.MustDebug(fmt.Sprintf("request: %s, method: %s, timed out", r.URL, r.Method))
		http.Error(w, "request time out", http.StatusRequestTimeout)
		return
	default:
		var scheduleData models.Responser
		// NOTE period data must be a iso string
		pstart := r.URL.Query().Get("pstart")
		pend := r.URL.Query().Get("pend")
		periodStart, err := time.Parse(time.RFC3339, pstart)
		if err != nil {
			logger.MustDebug(fmt.Sprintf("error parsing period start: %v", err))
			http.Error(w, "error parsing period start", http.StatusBadRequest)
			return
		}

		periodEnd, err := time.Parse(time.RFC3339, pend)
		if err != nil {
			logger.MustDebug(fmt.Sprintf("error parsing period end: %v", err))
			http.Error(w, "error parsing period end", http.StatusBadRequest)
			return
		}

		scheduleData, err = dacstore.CheckSchedule(r.Context(), periodStart, periodEnd)
		var noResults *dacstore.ErrNoCacheResult

		if err != nil {
			logger.MustDebug(fmt.Sprintf("error reading schedule cache: %v", err))
			if !errors.As(err, &noResults) {
				logger.MustDebug(fmt.Sprintf("an unexpected error occurred while reading schedule cache: %v", err))
				http.Error(w, "an unexpected error occurred while reading schedule cache", http.StatusInternalServerError)
				return
			}
		}

		if scheduleData == nil {
			scheduleData, err = controller.FetchSchedule(r.Context(), periodStart, periodEnd)
			if err != nil {
				periodStr := fmt.Sprintf("Period [start: %v, end: %v]", periodStart.String(), periodEnd.String())
				logger.MustDebug(fmt.Sprintf("error reading schedule for %v from database: %v", periodStr, err))
				emsg := fmt.Sprintf("coult not retrieve schedule for %v: %v", periodStr, err)
				http.Error(w, emsg, http.StatusInternalServerError)
				return
			}

			err = dacstore.SetSchedule(r.Context(), periodStart, periodEnd)
			if err != nil {
				logger.MustDebug(fmt.Sprintf("an error occurred while cache schedule for Period: [start: %v, end: %v]:: %v", periodStart.String(), periodEnd.String(), err))
				http.Error(w, fmt.Sprintf("could not cache schedule for Period: [start: %v, end: %v]:: %v", periodStart.String(), periodEnd.String(), err), http.StatusInternalServerError)
				return
			}
		}

		if sdata, ok := scheduleData.(*models.ScheduleResponse); ok {
			if err := json.NewEncoder(w).Encode(sdata); err != nil {
				logger.MustDebug(fmt.Sprintf("could not encode schedule response:: %v", err))
				http.Error(w, fmt.Sprintf("could not encode schedule response:: %v", err), http.StatusInternalServerError)
				return
			}
		} else {
			logger.MustDebug(fmt.Sprintf("could not cast type: [%T] into type schedule response", scheduleData))
			http.Error(w, fmt.Sprintf("could not cast type: [%T] into type schedule response", scheduleData), http.StatusInternalServerError)
			return
		}

	}
}

func createTask(w http.ResponseWriter, r *http.Request) {

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
		return
	default:
		var portfolioData models.Responser
		cacheClient, err := dacstore.NewRedisClient(r.Context())
		logger.MustTrace(fmt.Sprintf("cache client created..: %v", cacheClient))

		if err != nil {
			logger.MustDebug(fmt.Sprintf("cache error: %s", err))
			http.Error(w, "request cache error", http.StatusInternalServerError)
			return
		}

		portfolioData, err = dacstore.CheckPortfolioData(r.Context(), cacheClient)
		var noResult *dacstore.ErrNoCacheResult

		if err != nil {
			logger.MustDebug(fmt.Sprintf("error returned from cahce check.. err: %v", err))
			if !errors.As(err, &noResult) {
				logger.MustDebug(fmt.Sprintf("an unexpected error occurred while retrieving cache data:: %v", err))
				http.Error(w, "an unexpected error occurred retrieving cache", http.StatusInternalServerError)
				return
			}
		}

		if portfolioData == nil {
			logger.MustTrace("portfolio data was nil from cache.. fetching from database ....")
			portfolioData, err = controller.GetPortfolioData(r.Context())

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
			err = dacstore.SetPortfolioData(r.Context(), cacheClient, portfolioData)
			if err != nil {
				logger.MustDebug(fmt.Sprintf("could not set portfolio cache data: %s", err))
				http.Error(w, "an error occurred while setting portfolio cache data", http.StatusInternalServerError)
				return
			}
		}

		logger.MustDebug(fmt.Sprintf("respons data:: %v", portfolioData))
		if pdata, ok := portfolioData.(*models.PortfolioResponse); ok {
			// json.NewEncoder writes the data to request or errors
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

func headerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if config.IsValidOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Content-Type", "application/json")
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
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
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
