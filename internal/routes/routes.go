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
	"github.com/Z3DRP/zportfolio-service/enums"
	"github.com/Z3DRP/zportfolio-service/internal/adapters"
	"github.com/Z3DRP/zportfolio-service/internal/controller"
	"github.com/Z3DRP/zportfolio-service/internal/dacstore"
	"github.com/Z3DRP/zportfolio-service/internal/dtos"
	"github.com/Z3DRP/zportfolio-service/internal/middleware"
	"github.com/Z3DRP/zportfolio-service/internal/models"
	"github.com/Z3DRP/zportfolio-service/internal/utils"
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
	// TODO refactore routes and dacstores so routes and controller so controller checks the cache and
	mux := http.NewServeMux()
	mux.HandleFunc("GET /about", getAbout)
	mux.HandleFunc("POST /zypher", getZypher)
	mux.HandleFunc("GET /schedule", getSchedule)
	mux.HandleFunc("POST /task", createTask)
	mux.HandleFunc("PUT /task", editTask)
	mux.HandleFunc("DELETE /task", removeTask)

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
		cacheClient, err := dacstore.NewRedisClient(r.Context())
		if err != nil {
			logger.MustDebug(fmt.Sprintf("cache error: %s", err))
			http.Error(w, "request cache error", http.StatusInternalServerError)
			return
		}
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

		scheduleData, err = dacstore.CheckScheduleData(r.Context(), cacheClient, periodStart, periodEnd)
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

			if scheduleData != nil {
				err = dacstore.SetScheduleData(r.Context(), cacheClient, periodStart, periodEnd, scheduleData)
			}

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
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		// logger.MustDebug(fmt.Sprintf("request: %s method: %s timed out", r.URL, r.Method))
		// http.Error(w, "request time out", http.StatusRequestTimeout)
		handleRequestTimeout(r, w)
		return
	default:
		// var tsk map[string]interface{}
		tsk := dtos.TaskRequestDTO{}
		settings, err := config.ReadZypherSettings()
		if err != nil {
			// logger.MustDebug(fmt.Sprintf("an error occurred while reading zypher config:: %v", err))
			// http.Error(w, fmt.Sprintf("an error occurred while reading zypher config:: %v", err), http.StatusInternalServerError)
			handleConfigReadErr(config.NewConfigReadError("zypher", err), w)
			return
		}

		cacheClient, err := dacstore.NewRedisClient(r.Context())
		if err != nil {
			// logger.MustDebug(fmt.Sprintf("could not connect to redis:: %v", err))
			// http.Error(w, fmt.Sprintf("could not connect to redis:: %v", err), http.StatusInternalServerError)
			handleRedisConErr(dacstore.NewRedisConnErr(cacheClient.ClientID(r.Context()), err), w)
			return
		}

		err = json.NewDecoder(r.Body).Decode(&tsk)
		if err != nil {
			// logger.MustDebug(fmt.Sprintf("could not parse create task request body:: %v", err))
			// http.Error(w, fmt.Sprintf("could not parse create task request body:: %v", err), http.StatusInternalServerError)
			handleJsonDecodeErr("create task", err, w)
			return
		}

		taskStart, err := time.Parse(time.RFC3339, tsk.Start)
		if err != nil {
			// logger.MustDebug(fmt.Sprintf("invalid type for task start date:: %v", err))
			// http.Error(w, fmt.Sprintf("invalid type for task start date:: %v", err), http.StatusBadRequest)
			handleTaskTimeParseErr("start", err, w)
			return
		}

		taskEnd, err := time.Parse(time.RFC3339, tsk.End)
		if err != nil {
			// logger.MustDebug(fmt.Sprintf("invalid type for task end date:: %v", err))
			// http.Error(w, fmt.Sprintf("invalid type for task end date:: %v", err), http.StatusBadRequest)
			handleTaskTimeParseErr("end", err, w)
			return
		}

		uip := utils.GetIP(r)
		uid, err := dacstore.CheckUserData(r.Context(), cacheClient, uip)
		var noResults *dacstore.ErrNoCacheResult

		if err != nil {
			if !errors.As(err, &noResults) {
				// logger.MustDebug(fmt.Sprintf("error occurred while reading user cache:: %v", err))
				// http.Error(w, fmt.Sprintf("error occurred while reading user cache:: %v", err), http.StatusInternalServerError)
				handleCacheReadErr("user", err, w)
				return
			}
		}

		if uid == "" {
			uid, err = controller.CalculateZypher(uip, settings.Shift, settings.ShiftCount, settings.HashCount, settings.Alternate, settings.IgnSpace, settings.RestrictHash)
			// add user to cache so when trying to edit tasks id can be checked
			if err != nil {
				// logger.MustDebug(fmt.Sprintf("could not generate user id:: %v", err))
				// http.Error(w, fmt.Sprintf("could not generate user id:: %v", err), http.StatusInternalServerError)
				handleIdGeneratorErr("user", err, w)
				return
			}

			if err = dacstore.SetUserData(r.Context(), cacheClient, uip, uid); err != nil {
				// logger.MustDebug(fmt.Sprintf("error occurred while creating user cache:: %v", err))
				// http.Error(w, fmt.Sprintf("error occurred while creating user cache:: %v", err), http.StatusInternalServerError)
				handleCacheSetErr("user", err, w)
				return
			}
		}

		nwTask, err := controller.CreateTask(r.Context(), taskStart, taskEnd, tsk.Detail, uid)
		if err != nil {
			// logger.MustDebug(fmt.Sprintf("error occurred while usr: %v tried creating task:: %v", uid, err))
			// http.Error(w, fmt.Sprintf("error occurred while usr: %v tried creating task:: %v", uid, err), http.StatusInternalServerError)
			handleTaskActionErr("createing", err, uid, w)
			return
		}

		if tskRes, ok := nwTask.(*models.TaskInsertResponse); ok {
			// TODO call SendTaskRequestNotificationEmail
			usrData := adapters.NewUserData(tsk.UsrName, tsk.Company, tsk.Email, tsk.Phone, tsk.Roles)
			emlData := adapters.NewEmailInfo(tsk.Cc, tsk.Body, tsk.UseHtml)
			err = controller.SendTaskNotificationEmail(r.Context(), *tskRes.NwTask, usrData, emlData, enums.ZemailType(0))
			notificationSent := true

			if err != nil {
				notificationSent = false
				logger.MustDebug(fmt.Sprintf("failed to send task create notification:: %v", err))
			}

			response := map[string]any{
				"notificationSent":  notificationSent,
				"notificationError": err,
				"taskResult":        tskRes,
			}

			go func(udata adapters.UserData, eData adapters.EmailInfo) {
				err = controller.SendThanksNotification(r.Context(), udata, eData)
				if err != nil {
					logger.MustDebug(fmt.Sprintf("could not send thank you notification to '%v' at '%v", udata.Name, udata.Email))
				}
			}(usrData, emlData)

			if err := json.NewEncoder(w).Encode(response); err != nil {
				// logger.MustDebug(fmt.Sprintf("could not encode task response into json:: %v", err))
				// http.Error(w, fmt.Sprintf("could not encode task response into json:: %v", err), http.StatusInternalServerError)
				handleJsonEncodeErr("task insert response", err, w)
				return
			}

		} else {
			// logger.MustDebug(fmt.Sprintf("could not cast Type[%T] as Task Insert Response:: %v", nwTask, err))
			// http.Error(w, fmt.Sprintf("could not cast Type[%T] as Task Insert Response:: %v", nwTask, err), http.StatusInternalServerError)
			handleTypeCaseErr("task insert response", nwTask, err, w)
			return
		}
	}
}

func removeTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		logger.MustDebug(fmt.Sprintf("request: %s, method: %s, timed out", r.URL, r.Method))
		http.Error(w, "request time out", http.StatusRequestTimeout)
		return
	default:
		tskId := r.URL.Query().Get("taskid")
		if tskId == "" {
			logger.MustDebug("invalid task delete request:: Missing task id")
			http.Error(w, "invalid task delete request:: Missing task id", http.StatusBadRequest)
			return
		}

		var userId string
		err := json.NewDecoder(r.Body).Decode(&userId)
		if err != nil {
			logger.MustDebug(fmt.Sprintf("could not json decode request body:: %v", err))
			http.Error(w, fmt.Sprintf("could not json decode request body:: %v", err), http.StatusInternalServerError)
			return
		}

		if cacheClient, err := dacstore.NewRedisClient(r.Context()); err != nil {
			logger.MustDebug(fmt.Sprintf("could not connect to redis client:: %v", err))
			http.Error(w, fmt.Sprintf("could not connect to redis client:: %v", err), http.StatusInternalServerError)
			return
		} else {
			uid, err := dacstore.CheckUserData(r.Context(), cacheClient, utils.GetIP(r))
			var noResults *dacstore.ErrNoCacheResult
			if err != nil {
				if errors.As(err, &noResults) {
					logger.MustDebug("invalid remove request user does not own any tasks")
					http.Error(w, "invalid remove request user does not own any tasks", http.StatusBadRequest)
				}
				logger.MustDebug(fmt.Sprintf("error checking user cache:: %v", err))
				http.Error(w, fmt.Sprintf("error checking user cache:: %v", err), http.StatusInternalServerError)
				return
			}

			if userId != uid {
				logger.MustDebug("invalid remove request user must own task to remove it")
				http.Error(w, "invalid remove request user must own task to remove it", http.StatusBadRequest)
				return
			}

			delCount, err := controller.RemoveTask(r.Context(), tskId, uid)
			if err != nil {
				logger.MustDebug(fmt.Sprintf("error deleting task:: %v", err))
				http.Error(w, fmt.Sprintf("error deleting task:: %v", err), http.StatusInternalServerError)
				return
			}

			usrData := adapters.NewUserData()
			emlData := adapters.NewEmailInfo()
			w.WriteHeader(http.StatusOK)
			response := map[string]int64{
				"deleteCount": delCount,
			}
			json.NewEncoder(w).Encode(response)
		}

	}
}

func editTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	taskId := r.URL.Query().Get("taskid")
	if taskId == "" {
		logger.MustDebug("invalid edit task request:: missing task id")
		http.Error(w, "invalid edit task request:: missing task id", http.StatusBadRequest)
		return
	}

	taskReq := dtos.TaskRequestDTO{}
	err := json.NewDecoder(r.Body).Decode(&taskReq)
	if err != nil {
		logger.MustDebug(fmt.Sprintf("could not json decode request body:: %v", err))
		http.Error(w, fmt.Sprintf("could not json decode request body:: %v", err), http.StatusInternalServerError)
		return
	}

	taskStart, err := time.Parse(time.RFC3339, taskReq.Start)
	if err != nil {
		logger.MustDebug(fmt.Sprintf("could not parse time from:: %v", taskReq.Start))
		http.Error(w, fmt.Sprintf("could not parse time from:: %v", taskReq.Start), http.StatusBadRequest)
		return
	}

	taskEnd, err := time.Parse(time.RFC3339, taskReq.End)
	if err != nil {
		logger.MustDebug(fmt.Sprintf("could not parse time from:: %v", taskReq.End))
		http.Error(w, fmt.Sprintf("could not parse time from:: %v", taskReq.End), http.StatusBadRequest)
		return
	}

	if taskReq.Uid == "" {
		logger.MustDebug("error missing user id")
		http.Error(w, "error missing user id", http.StatusBadRequest)
		return
	}

	cacheClient, err := dacstore.NewRedisClient(r.Context())
	if err != nil {
		logger.MustDebug(fmt.Sprintf("could not connect to redis:: %v", err))
		http.Error(w, fmt.Sprintf("could not connect to redis:: %v", err), http.StatusInternalServerError)
		return
	}

	validUid, err := dacstore.CheckUserData(r.Context(), cacheClient, utils.GetIP(r))
	var noResults *dacstore.ErrNoCacheResult

	if err != nil {
		if errors.As(err, &noResults) {
			logger.MustDebug("invalid edit request user does not own any tasks")
			http.Error(w, "invalid edit request user does not own any tasks", http.StatusBadRequest)
			return
		}
		logger.MustDebug(fmt.Sprintf("error occurred while checking user cache:: %v", err))
		http.Error(w, fmt.Sprintf("error occurred hwile checking user cache:: %v", err), http.StatusInternalServerError)
		return
	}

	if taskReq.Uid != validUid {
		logger.MustDebug("invalid edit request user must own task to remove it")
		http.Error(w, "invalid edit request user must own task to remove it", http.StatusBadRequest)
		return
	}

	results, err := controller.EditTask(r.Context(), taskId, taskReq.Uid, taskStart, taskEnd, taskReq.Detail)
	if err != nil {
		logger.MustDebug(fmt.Sprintf("error occurred while editing task: %v :: %v", taskId, err))
		http.Error(w, fmt.Sprintf("error occurred while editing task: %v :: %v", taskId, err), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(results); err != nil {
		logger.MustDebug(fmt.Sprintf("could not json encode edit task results:: %v", err))
		http.Error(w, fmt.Sprintf("could not json encode edit task results:: %v", err), http.StatusInternalServerError)
		return
	}
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
		settings, err := config.ReadZypherSettings()
		if err != nil {
			// logger.MustDebug(fmt.Sprintf("an error occurred while reading zypher config:: %v", err))
			// http.Error(w, fmt.Sprintf("an error occurred while reading zypher config:: %v", err), http.StatusInternalServerError)
			handleConfigReadErr(config.NewConfigReadError("zypher", err), w)
			return
		}

		uip := utils.GetIP(r)
		uid, err := dacstore.CheckUserData(r.Context(), cacheClient, uip)
		var noResults *dacstore.ErrNoCacheResult

		if err != nil {
			if !errors.As(err, &noResults) {
				// logger.MustDebug(fmt.Sprintf("error occurred while reading user cache:: %v", err))
				// http.Error(w, fmt.Sprintf("error occurred while reading user cache:: %v", err), http.StatusInternalServerError)
				handleCacheReadErr("user", err, w)
				return
			}
		}

		if uid == "" {
			uid, err = controller.CalculateZypher(uip, settings.Shift, settings.ShiftCount, settings.HashCount, settings.Alternate, settings.IgnSpace, settings.RestrictHash)
			// add user to cache so when trying to edit tasks id can be checked
			if err != nil {
				handleIdGeneratorErr("user", err, w)
				return
			}

			if err = dacstore.SetUserData(r.Context(), cacheClient, uip, uid); err != nil {
				handleCacheSetErr("user", err, w)
				return
			}

			if res, err := controller.CreateVisitor(r.Context(), 1, uip, false); err != nil {
				logger.MustDebug(fmt.Sprintf("failed to create visitor record:: %v", err))
			} else {
				logger.MustDebug(fmt.Sprintf("successfully created visitor record:: %v", res.PrintRes()))
			}
		}

		//TODO update visitor count
		ads

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
