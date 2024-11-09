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
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return config.IsValidOrigin(r.Header.Get("origin")) },
}

func NewServer(sconfig config.ZServerConfig) (*http.Server, error) {
	// TODO refactore routes and dacstores so routes and controller so controller checks the cache and
	mux := http.NewServeMux()
	mux.HandleFunc("GET /about", getAbout)
	mux.HandleFunc("POST /zypher", getZypher)
	mux.HandleFunc("GET /schedule", handleScheduleWebsocket)
	// mux.HandleFunc("POST /task", handleCreateTask)
	// mux.HandleFunc("PUT /task", handleEditTask)
	// mux.HandleFunc("DELETE /task", handleRemoveTask)

	server := &http.Server{
		Addr:         sconfig.Address,
		ReadTimeout:  time.Second * time.Duration(sconfig.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(sconfig.WriteTimeout),
		Handler:      handlePanic(loggerMiddleware(headerMiddleware(contextMiddleware(mux, 10*time.Second)))),
	}

	return server, nil
}

// TODO update dtos to accept json and in each Task event handler use make any refactors needed
func handleScheduleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	uip := utils.GetIP(r)
	if err != nil {
		e := utils.NewWebSocketErr("upgrade http connection to websocket", err)
		utils.LogError(logger, e, zlg.Debug)
		utils.SendErrMessage(conn, e, enums.ProtocolErr)
		if err != nil {

		}
		return
	}
	defer conn.Close()

	for {
		var msg dtos.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			logger.MustDebug(fmt.Sprintf("could not read websocket message:: %v", err))
			e := utils.NewWebSocketErr("read websocket message", err)
			utils.LogError(logger, e, zlg.Debug)
			utils.SendErrMessage(conn, e, enums.ProtocolErr)
			if err != nil {
				conn.Close()
			}
			break
		}

		switch msg.Event {
		case "getSchedule":
			var payload dtos.PeriodPayloadDTO
			err = json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				e := utils.NewJsonDecodeErr(string(msg.Payload), err)
				utils.LogError(logger, e, zlg.Debug)
				utils.SendErrMessage(conn, e, enums.UnsupportedData)
				return
			}
			handleGetSchedule(r.Context(), conn, payload, uip)

		case "createTask":
			var pload dtos.TaskPayloadDTO
			err = json.Unmarshal(msg.Payload, &pload)
			if err != nil {
				e := utils.NewJsonDecodeErr(msg.Payload, err)
				utils.LogError(logger, e, zlg.Debug)
				utils.SendErrMessage(conn, e, enums.UnsupportedData)
				return
			}
			handleCreateTask(r.Context(), conn, pload, uip)
		case "editTask":
			var pload dtos.TaskPayloadDTO
			err = json.Unmarshal(msg.Payload, &pload)
			if err != nil {
				e := utils.NewJsonDecodeErr(msg.Payload, err)
				utils.LogError(logger, e, zlg.Debug)
				utils.SendErrMessage(conn, e, enums.UnsupportedData)
				return
			}
			handleEditTask(r.Context(), conn, msg.Payload, uip)
		case "deleteTask":
			var pload dtos.TaskDeletePaylaod
			err := json.Unmarshal(msg.Payload, pload)
			if err != nil {
				e := utils.NewJsonDecodeErr(msg.Payload, err)
				utils.LogError(logger, e, zlg.Debug)
				utils.SendErrMessage(conn, e, enums.UnsupportedData)
				return
			}
			handleRemoveTask(r.Context(), conn, pload, uip)
		default:
			eventErr := struct {
				Err     string
				Event   string
				ConCode int
			}{
				Err:     "unknown event",
				Event:   msg.Event,
				ConCode: enums.UnsupportedData,
			}

			logger.MustDebug(fmt.Sprintf("%v: %v", eventErr.Err, eventErr.Event))
			err = conn.WriteJSON(eventErr)
			if err != nil {
				logger.MustDebug(fmt.Sprintf("error occurred while trying to write json error to socket connection:: %v", err))
				conn.Close()
			}
		}
	}

}

func handleGetSchedule(ctx context.Context, conn *websocket.Conn, data dtos.PeriodPayloadDTO, uip string) {
	select {
	case <-ctx.Done():
		e := utils.NewTimeoutErr("get schedule", nil)
		utils.LogError(logger, e, zlg.Debug)
		if err := utils.SendErrMessage(conn, e, enums.Timeout); err != nil {
			// Try to write message of why it is closing
			_ = utils.WriteCloseMessage(conn, e, enums.Timeout)
		}
		conn.Close()
	default:
		var scheduleData models.Responser
		cacheClient, err := dacstore.NewRedisClient(ctx)
		if err != nil {
			e := &dacstore.ErrRedisConnect{ClientId: cacheClient.ClientID(ctx), Err: err}
			utils.LogError(logger, e, zlg.Debug)
			if e2 := utils.SendErrMessage(conn, e, enums.CacheError); err != nil {
				_ = utils.WriteCloseMessage(conn, e2, enums.CacheError)
				conn.Close()
			}
			return
		}
		// NOTE period data must be a iso string
		periodStart, err := time.Parse(time.RFC3339, data.PeriodStart)
		if err != nil {
			e := utils.NewTimeParseErr(data.PeriodStart, "Period Start", err)
			utils.LogError(logger, e, zlg.Debug)
			if e2 := utils.SendErrMessage(conn, e, enums.UnsupportedData); e2 != nil {
				_ = utils.WriteCloseMessage(conn, e, enums.UnsupportedData)
				conn.Close()
			}
			return
		}

		periodEnd, err := time.Parse(time.RFC3339, data.PeriodEnd)
		if err != nil {
			e := utils.NewTimeParseErr(data.PeriodStart, "Period End", err)
			utils.LogError(logger, e, zlg.Debug)
			utils.SendErrMessage(conn, e, enums.UnsupportedData)
			return
		}

		scheduleData, err = dacstore.CheckScheduleData(ctx, cacheClient, periodStart, periodEnd)
		var noResults *dacstore.ErrNoCacheResult

		if err != nil {
			if !errors.As(err, &noResults) {
				e := utils.NewCacheOpErr("read", "schedule", err)
				utils.LogError(logger, e, zlg.Debug)
				utils.SendErrMessage(conn, e, enums.CacheError)
			}
			utils.WriteLog(logger, "schedule cache had no values", zlg.Debug)
			return
		}

		if scheduleData == nil {
			scheduleData, err = controller.FetchSchedule(ctx, periodStart, periodEnd)
			if err != nil {
				periodStr := fmt.Sprintf("Period [start: %v, end: %v]", periodStart.String(), periodEnd.String())
				e := utils.ErrFetchRecords{RecordType: "schedule", Msg: periodStr, Err: err}
				utils.LogError(logger, e, zlg.Debug)
				utils.SendErrMessage(conn, e, enums.DatabaseError)
				return
			}

			if scheduleData != nil {
				err = dacstore.SetScheduleData(ctx, cacheClient, periodStart, periodEnd, scheduleData)
			}

			if err != nil {
				e := utils.NewCacheOpErr("write", "schedule", err)
				utils.LogError(logger, e, zlg.Debug)
				utils.SendErrMessage(conn, e, enums.CacheError)
				return
			}
		}

		if sdata, ok := scheduleData.(*models.ScheduleResponse); ok {
			rawScheduleDto, err := json.Marshal(dtos.NewScheduleDto(sdata))
			if err != nil {
				e := utils.NewJsonEncodeErr(dtos.NewScheduleDto(sdata), err)
				utils.LogError(logger, e, zlg.Debug)
				utils.SendErrMessage(conn, e, enums.JsonEncodeError)
			}
			msg := dtos.Message{
				Event:   "getSchedule",
				Payload: rawScheduleDto,
			}

			utils.SendMessage(conn, msg)
			if err != nil {
				// still try to write err msg
				utils.WriteCloseMessage(conn, err, enums.ServerError)
				conn.Close()
			}

		} else {
			//TODO type case error
			e := utils.NewTypeCastErr(scheduleData, models.ScheduleResponse{}, nil)
			utils.LogError(logger, e, zlg.Debug)
			utils.WriteCloseMessage(conn, e, enums.TypeCastErr)
		}

	}
}

func handleCreateTask(ctx context.Context, conn *websocket.Conn, data dtos.TaskPayloadDTO, uip string) {
	select {
	case <-ctx.Done():
		e := utils.NewTimeoutErr("create task", nil)
		utils.LogError(logger, e, zlg.Debug)
		if err := utils.SendErrMessage(conn, e, enums.Timeout); err != nil {
			utils.WriteCloseMessage(conn, e, enums.Timeout)
		}
		conn.Close()
	default:
		settings, err := config.ReadZypherSettings()
		if err != nil {
			e := utils.NewConfigFileErr("zypher config", err)
			utils.LogError(logger, e, zlg.Debug)
			if e2 := utils.SendErrMessage(conn, e, enums.ServerError); e2 != nil {
				_ = utils.WriteMessage(conn, e.Error())
				conn.Close()
			}
			return
		}

		cacheClient, err := dacstore.NewRedisClient(ctx)
		if err != nil {
			utils.LogError(logger, err, zlg.Debug)
			utils.SendErrMessage(conn, err, enums.CacheError)
			return
		}

		taskStart, err := time.Parse(time.RFC3339, data.Start)
		if err != nil {
			e := utils.NewTimeParseErr(data.Start, "start date", err)
			utils.LogError(logger, e, zlg.Debug)
			utils.SendErrMessage(conn, e, enums.TimeParseErr)
			return
		}

		taskEnd, err := time.Parse(time.RFC3339, data.End)
		if err != nil {
			e := utils.NewTimeParseErr(data.End, "end date", err)
			utils.LogError(logger, e, zlg.Debug)
			utils.SendErrMessage(conn, e, enums.TimeParseErr)
			return
		}

		usrInfo, err := dacstore.CheckUserData(ctx, cacheClient, uip)
		var noResults *dacstore.ErrNoCacheResult

		if err != nil {
			if !errors.As(err, &noResults) {
				e := utils.NewCacheOpErr("reading", "user", err)
				utils.LogError(logger, e, zlg.Debug)
				utils.SendErrMessage(conn, e, enums.CacheError)
				return
			}
		}

		usr, ok := usrInfo.(dtos.UserDto)
		if !ok {
			e := utils.NewTypeCastErr(usrInfo, dtos.UserDto{}, nil)
			utils.LogError(logger, e, zlg.Debug)
			utils.SendErrMessage(conn, e, enums.ServerError)
			return
		}

		var uidKey string
		if usr.Uid == "" {
			uidKey, err = controller.CalculateZypher(uip, settings.Shift, settings.ShiftCount, settings.HashCount, settings.Alternate, settings.IgnSpace, settings.RestrictHash)
			// add user to cache so when trying to edit tasks id can be checked
			if err != nil {
				utils.LogError(logger, err, zlg.Debug)
				utils.SendErrMessage(conn, err, enums.ServerError)
				return
			}

			if err = dacstore.SetUserData(ctx, cacheClient, data.UsrName, data.Company, data.Phone, data.Email, data.Roles, uip, uidKey); err != nil {
				e := utils.NewCacheOpErr("writing", "user", err)
				utils.LogError(logger, e, zlg.Debug)
				utils.SendErrMessage(conn, e, enums.CacheError)
				return
			}
		}

		nwTask, err := controller.CreateTask(ctx, taskStart, taskEnd, data.Detail, usr.Uid)
		if err != nil {
			e := utils.NewDbErr(enums.Insert.String(), "user", err)
			utils.LogError(logger, e, zlg.Debug)
			utils.SendErrMessage(conn, e, enums.DatabaseError)
			return
		}

		if tskRes, ok := nwTask.(*models.TaskInsertResponse); ok {
			// TODO call SendTaskRequestNotificationEmail
			usrData := adapters.NewUserData(data.UsrName, data.Company, data.Email, data.Phone, data.Roles)
			emlData := adapters.NewEmailInfo(data.Cc, data.Body, data.UseHtml)
			err = controller.SendTaskNotificationEmail(ctx, *tskRes.NwTask, usrData, emlData, enums.ZemailType(0))
			notificationSent := true

			if err != nil {
				notificationSent = false
				e := utils.NewNotificationErr(emlData.String(), usrData.String(), err)
				utils.LogError(logger, e, zlg.Debug)
			}

			response := map[string]any{
				"notificationSent":  notificationSent,
				"notificationError": err,
				"taskResult":        tskRes,
			}

			go func(udata adapters.UserData, eData adapters.EmailInfo) {
				err = controller.SendThanksNotification(ctx, udata, eData)
				if err != nil {
					utils.LogError(logger, fmt.Errorf("could not send thank you notification to '%v' at '%v'", udata.Name, udata.Email), zlg.Debug)
				}
			}(usrData, emlData)

			pload, err := json.Marshal(response)
			if err != nil {
				e := utils.NewJsonEncodeErr(response, err)
				utils.LogError(logger, e, zlg.Debug)
				utils.SendErrMessage(conn, e, enums.JsonEncodeError)
			}

			msg := dtos.Message{Event: "createTask", Payload: pload}
			err = utils.SendMessage(conn, msg)
			if err != nil {
				_ = utils.WriteCloseMessage(conn, err, enums.JsonEncodeError)
				conn.Close()
			}

		} else {
			e := utils.NewTypeCastErr(nwTask, models.TaskInsertResponse{}, nil)
			utils.LogError(logger, e, zlg.Debug)
			if err = utils.SendErrMessage(conn, e, enums.TypeCastErr); err != nil {
				_ = utils.WriteCloseMessage(conn, err, enums.TypeCastErr)
			}
			return
		}
	}
}

func handleRemoveTask(ctx context.Context, conn *websocket.Conn, data dtos.TaskDeletePaylaod, uip string) {
	select {
	case <-ctx.Done():
		e := utils.NewTimeoutErr("delete task", nil)
		utils.LogError(logger, e, zlg.Debug)
		if err := utils.SendErrMessage(conn, e, enums.Timeout); err != nil {
			_ = utils.WriteCloseMessage(conn, e, enums.Timeout)
		}
		conn.Close()
	default:
		if data.Tid == "" {
			e := utils.NewMissingDataErr("task id", "string", nil)
			utils.LogError(logger, e, zlg.Debug)
			if err := utils.SendErrMessage(conn, e, enums.MissingData); err != nil {
				_ = utils.WriteCloseMessage(conn, e, enums.ServerError)
			}
			conn.Close()
			return
		}

		if data.Uid == "" {
			e := utils.NewMissingDataErr("user id", "string", nil)
			utils.LogError(logger, e, zlg.Debug)
			if err := utils.SendErrMessage(conn, e, enums.MissingData); err != nil {
				_ = utils.WriteCloseMessage(conn, e, enums.ServerError)
			}
			conn.Close()
			return
		}

		if cacheClient, err := dacstore.NewRedisClient(ctx); err != nil {
			e := dacstore.NewRedisConnErr(cacheClient.ClientID(ctx), err)
			utils.LogError(logger, e, zlg.Debug)
			if err = utils.SendErrMessage(conn, e, enums.CacheError); err != nil {
				_ = utils.WriteCloseMessage(conn, e, enums.ServerError)
				conn.Close()
				return
			}
		} else {
			usrInfo, err := dacstore.CheckUserData(ctx, cacheClient, uip)
			var noResults *dacstore.ErrNoCacheResult
			if err != nil {
				if errors.As(err, &noResults) {
					e := utils.NewInvalidOperationErr("task remove", "user does not own any tasks", err)
					utils.LogError(logger, e, zlg.Debug)
					utils.SendErrMessage(conn, e, enums.ServerError)
				}
				utils.LogError(logger, err, zlg.Debug)
				_ = utils.SendErrMessage(conn, err, enums.CacheError)
				return
			}

			usr, ok := usrInfo.(dtos.UserDto)
			if !ok {
				e := utils.NewTypeCastErr(usrInfo, dtos.UserDto{}, nil)
				utils.LogError(logger, e, zlg.Debug)
				_ = utils.SendErrMessage(conn, e, enums.ServerError)
				return
			}

			if data.Uid != usr.Uid {
				e := utils.NewInvalidOperationErr("task removal", "user must own task to remove it", nil)
				utils.LogError(logger, e, zlg.Debug)
				_ = utils.SendErrMessage(conn, e, enums.PermissionErr)
				return
			}

			delCount, err := controller.RemoveTask(ctx, data.Tid, usr.Uid)
			if err != nil {
				dbErr := utils.NewDbErr("delete", "task", err)
				utils.LogError(logger, dbErr, zlg.Debug)
				e := utils.SendErrMessage(conn, dbErr, enums.DatabaseError)
				if e != nil {
					fts := utils.NewFaildToSendErr("after delete task error", e)
					_ = utils.WriteCloseMessage(conn, fts, enums.ServerError)
					conn.Close()
				}
				return
			}

			response := map[string]int64{
				"deleteCount": delCount,
			}

			pload, err := json.Marshal(response)
			if err != nil {
				em := utils.NewJsonEncodeErr(response, err)
				utils.LogError(logger, em, zlg.Debug)
				e := utils.SendErrMessage(conn, em, enums.JsonEncodeError)
				if e != nil {
					_ = utils.WriteCloseMessage(conn, e, enums.JsonEncodeError)
					conn.Close()
					return
				}
			}

			msg := dtos.Message{
				Event:   "deleteTask",
				Payload: pload,
			}

			err = utils.SendMessage(conn, msg)
			if err != nil {
				failErr := utils.NewFailedMessageErr(msg, "delete task", err)
				utils.LogError(logger, failErr, zlg.Debug)
				_ = utils.WriteCloseMessage(conn, failErr, enums.ServerError)
				conn.Close()
			}

		}

	}
}

func handleEditTask(ctx context.Context, conn *websocket.Conn, data interface{}, uip string) {
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

	usrInfo, err := dacstore.CheckUserData(ctx, cacheClient, utils.GetIP(r))
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

	usr, ok := usrInfo.(dtos.UserDto)
	if !ok {
		e := utils.NewTypeCastErr(usrInfo, dtos.UserDto{}, nil)
		utils.LogError(logger, e, zlg.Debug)
		utils.SendErrMessage(conn, e, enums.ServerError)
	}

	if taskReq.Uid != usr.Uid {
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
			logger.MustDebug(fmt.Sprintf("an error occurred while reading zypher config:: %v", err))
			return
		}

		go updateVisitorCount(settings, cacheClient, r)

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

func updateVisitorCount(settings config.ZypherConfig, cacheClient *redis.Client, r *http.Request) {
	uip := utils.GetIP(r)
	uid, err := dacstore.CheckUserData(r.Context(), cacheClient, uip)
	var noResults *dacstore.ErrNoCacheResult

	if err != nil {
		if !errors.As(err, &noResults) {
			logger.MustDebug(fmt.Sprintf("error occurred while reading user cache:: %v", err))
			return
		}
	}

	if uid == "" {
		uid, err = controller.CalculateZypher(uip, settings.Shift, settings.ShiftCount, settings.HashCount, settings.Alternate, settings.IgnSpace, settings.RestrictHash)
		// add user to cache so when trying to edit tasks id can be checked
		if err != nil {
			logger.MustDebug(fmt.Sprintf("an error occurred while calculting visitor zypher:: %v", err))
			return
		}

		if err = dacstore.SetUserData(r.Context(), cacheClient, uip, uid); err != nil {
			logger.MustDebug(fmt.Sprintf("an error occurred while setting visitor id in cache:: %v", err))
			return
		}

		if res, err := controller.CreateVisitor(r.Context(), 1, uip, false); err != nil {
			logger.MustDebug(fmt.Sprintf("failed to create visitor record:: %v", err))
		} else {
			logger.MustDebug(fmt.Sprintf("successfully created visitor record:: %v", res.PrintRes()))
		}
	}

	_, _, err = controller.EditVisitorCount(r.Context(), uip)
	if err != nil {
		logger.MustDebug("visitor successfully updated")
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
