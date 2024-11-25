package wsman

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Z3DRP/zportfolio-service/config"
	"github.com/Z3DRP/zportfolio-service/enums"
	"github.com/Z3DRP/zportfolio-service/internal/adapters"
	"github.com/Z3DRP/zportfolio-service/internal/controller"
	"github.com/Z3DRP/zportfolio-service/internal/dacstore"
	"github.com/Z3DRP/zportfolio-service/internal/dtos"
	"github.com/Z3DRP/zportfolio-service/internal/models"
	"github.com/Z3DRP/zportfolio-service/internal/utils"
	"github.com/Z3DRP/zportfolio-service/internal/zlogger"
)

func HandleGetSchedule(ctx context.Context, clnt *Client, evnt Event) error {
	select {
	case <-ctx.Done():
		e := utils.NewTimeoutErr("get schedule", nil)
		if err := utils.MustSendErrMessage(clnt.Connection, e, enums.Timeout); err != nil {
			clnt.Manager.logger.MustDebug(err.Error())
		}
		clnt.Connection.Close()
	default:
		var scheduleData models.Responser
		var fetchSchedEvnt EvntFetchSchedule

		if err := json.Unmarshal(evnt.Payload, &fetchSchedEvnt); err != nil {
			return utils.NewJsonDecodeErr(fetchSchedEvnt, err)
		}

		cacheClient, err := dacstore.NewRedisClient(ctx)

		if err != nil {
			return dacstore.NewRedisConnErr(cacheClient.ClientID(ctx), err)
		}
		// NOTE period data must be a iso string
		periodStart, err := time.Parse(time.RFC3339, fetchSchedEvnt.PeriodStart)

		if err != nil {
			return utils.NewTimeParseErr(fetchSchedEvnt.PeriodStart, "Period Start", err)
		}

		periodEnd, err := time.Parse(time.RFC3339, fetchSchedEvnt.PeriodEnd)

		if err != nil {
			return utils.NewTimeParseErr(fetchSchedEvnt.PeriodStart, "Period End", err)
		}

		clnt.SetPeriod(&models.Period{StartDate: periodStart, EndDate: periodEnd})
		scheduleData, err = dacstore.CheckScheduleData(ctx, cacheClient, periodStart, periodEnd)
		var noResults *dacstore.ErrNoCacheResult

		if err != nil {
			if !errors.As(err, &noResults) {
				return utils.NewCacheOpErr("read", "schedule", err)
			}
			clnt.Manager.logger.MustDebug("cache had no values for schedule")
		}

		if scheduleData == nil {
			scheduleData, err = controller.FetchSchedule(ctx, periodStart, periodEnd)
			if err != nil {
				periodStr := fmt.Sprintf("Period [start: %v, end: %v]", periodStart.String(), periodEnd.String())
				e := utils.ErrFetchRecords{RecordType: "schedule", Msg: periodStr, Err: err}
				return e
			}

			if scheduleData != nil {
				if err = dacstore.SetScheduleData(ctx, cacheClient, periodStart, periodEnd, scheduleData); err != nil {
					return utils.NewCacheOpErr("write", "schedule", err)
				}
			}
		}

		if sdata, ok := scheduleData.(*models.ScheduleResponse); ok {

			msg := dtos.NewEventDto(evnt.Type, sdata)

			if err = utils.MustSendMessage(clnt.Connection, msg); err != nil {
				return utils.NewFailedSendEventResponse(clnt.Connection, evnt.Type, err)
			}

		} else {
			return utils.NewTypeCastErr(scheduleData, models.ScheduleResponse{}, nil)
		}

	}
	return nil
}

func HandleCreateTask(ctx context.Context, clnt *Client, evnt Event) error {
	select {
	case <-ctx.Done():
		e := utils.NewTimeoutErr("create task", nil)
		if err := utils.MustSendErrMessage(clnt.Connection, e, enums.Timeout); err != nil {
			return utils.NewFailedToSendErr(clnt.Connection, &e)
		}
	default:
		var createEvnt EvntTaskUpsert
		uip := clnt.Connection.RemoteAddr().String()
		settings, err := config.ReadZypherSettings()
		if err != nil {
			return utils.NewConfigFileErr("zypher config", err)
		}

		if err = json.Unmarshal(evnt.Payload, &createEvnt); err != nil {
			return utils.NewJsonDecodeErr(createEvnt, err)
		}

		cacheClient, err := dacstore.NewRedisClient(ctx)
		if err != nil {
			return dacstore.NewRedisConnErr(cacheClient.ClientID(ctx), err)
		}

		taskStart, err := time.Parse(time.RFC3339, createEvnt.Start)
		if err != nil {
			return utils.NewTimeParseErr(createEvnt.Start, "start date", err)
		}

		taskEnd, err := time.Parse(time.RFC3339, createEvnt.End)
		if err != nil {
			return utils.NewTimeParseErr(createEvnt.End, "end date", err)
		}

		usrInfo, err := dacstore.CheckUserData(ctx, cacheClient, uip)
		var noResults *dacstore.ErrNoCacheResult

		if err != nil {
			if !errors.As(err, &noResults) {
				return utils.NewCacheOpErr("reading", "user", err)
			}
		}

		usr, ok := usrInfo.(dtos.UserDto)
		if !ok {
			return utils.NewTypeCastErr(usrInfo, dtos.UserDto{}, nil)
		}

		var uidKey string
		if usr.Uid == "" {
			uidKey, err = controller.CalculateZypher(uip, settings.Shift, settings.ShiftCount, settings.HashCount, settings.Alternate, settings.IgnSpace, settings.RestrictHash)
			// add user to cache so when trying to edit tasks id can be checked
			if err != nil {
				return utils.NewZypherFailureErr("user id creation", err)
			}

			if err = dacstore.SetUserData(ctx, cacheClient, createEvnt.UsrName, createEvnt.Company, createEvnt.Phone, createEvnt.Email, createEvnt.Roles, uip, uidKey); err != nil {
				return utils.NewCacheOpErr("writing", "user", err)
			}
		}

		nwTask, err := controller.CreateTask(ctx, taskStart, taskEnd, createEvnt.Detail, usr.Uid)
		if err != nil {
			return utils.NewDbErr(enums.Insert.String(), "user", err)
		}

		if tskRes, ok := nwTask.(*models.TaskInsertResponse); ok {
			usrData := adapters.NewUserData(createEvnt.UsrName, createEvnt.Company, createEvnt.Email, createEvnt.Phone, createEvnt.Roles)
			emlData := adapters.DefaultEmlInfo()

			go func(logr *zlogger.Zlogrus, udata adapters.UserData, eData adapters.EmailInfo) {
				if err := controller.SendTaskNotificationEmail(ctx, *tskRes.NwTask, usrData, emlData, enums.ZemailType(0)); err != nil {
					logr.MustDebug(err.Error())
				}
			}(clnt.Manager.logger, usrData, emlData)

			go func(logr *zlogger.Zlogrus, udata adapters.UserData, eData adapters.EmailInfo) {
				if err = controller.SendThanksNotification(ctx, udata, eData); err != nil {
					logr.MustDebug(fmt.Sprintf("could not send thank you notification to '%v' at '%v'", udata.Name, udata.Email))
				}
			}(clnt.Manager.logger, usrData, emlData)

			msg := dtos.NewEventDto(evnt.Type, *tskRes)

			if err = utils.MustSendMessage(clnt.Connection, msg); err != nil {
				return utils.NewFailedSendEventResponse(clnt.Connection, evnt.Type, err)
			}
			clnt.Manager.logger.MustDebug(fmt.Sprintf("task %v created successfully by usr %v", tskRes.NwTask.Id, usr.Uid))

		} else {
			return utils.NewTypeCastErr(nwTask, models.TaskInsertResponse{}, nil)
		}
	}
	return nil
}

func HandleRemoveTask(ctx context.Context, clnt *Client, envt Event) error {
	select {
	case <-ctx.Done():
		return utils.NewTimeoutErr("delete task", nil)
	default:
		var usr dtos.UserDto
		var rmvEvnt dtos.TaskDeletePaylaod
		if err := json.Unmarshal(envt.Payload, &rmvEvnt); err != nil {
			return utils.NewJsonDecodeErr(rmvEvnt, err)
		}

		if rmvEvnt.Tid == "" {
			return utils.NewMissingDataErr("task id", "string", nil)
		}

		if rmvEvnt.Uid == "" {
			return utils.NewMissingDataErr("user id", "string", nil)
		}

		if cacheClient, err := dacstore.NewRedisClient(ctx); err != nil {
			return dacstore.NewRedisConnErr(cacheClient.ClientID(ctx), err)
		} else {

			usrInfo, err := dacstore.CheckUserData(ctx, cacheClient, clnt.Connection.RemoteAddr().String())
			var noResults *dacstore.ErrNoCacheResult

			if err != nil {
				if errors.As(err, &noResults) {
					return utils.NewInvalidOperationErr("task remove", "user does not own any tasks", err)
				}
				return err
			}

			if castedUsr, ok := usrInfo.(dtos.UserDto); !ok {
				return utils.NewTypeCastErr(usrInfo, dtos.UserDto{}, nil)
			} else {
				// stupid issue with compiler not liking assigning usr to value straight from type cast
				usr = castedUsr
			}

		}

		if rmvEvnt.Uid != usr.Uid {
			return utils.NewInvalidOperationErr("task removal", "user must own task to remove it", nil)
		}

		delCountRes, err := controller.RemoveTask(ctx, rmvEvnt.Tid, usr.Uid)
		if err != nil {
			return utils.NewDbErr("delete", "task", err)
		}

		msg := dtos.NewEventDto(envt.Type, delCountRes)

		if err = utils.MustSendMessage(clnt.Connection, msg); err != nil {
			return utils.NewFailedToSendErr(clnt.Connection, err)
		}

	}
	return nil
}

func HandleEditTask(ctx context.Context, clnt *Client, envt Event) error {
	var upsertEvnt EvntTaskUpsert
	if err := json.Unmarshal(envt.Payload, &upsertEvnt); err != nil {
		return utils.NewJsonDecodeErr(upsertEvnt, err)
	}

	if upsertEvnt.Tid == "" {
		return utils.NewMissingDataErr("task id", "string", nil)
	}

	taskStart, err := time.Parse(time.RFC3339, upsertEvnt.Start)
	if err != nil {
		return utils.NewTimeParseErr(upsertEvnt.Start, "start date", err)
	}

	taskEnd, err := time.Parse(time.RFC3339, upsertEvnt.End)
	if err != nil {
		return utils.NewTimeParseErr(upsertEvnt.End, "end date", err)
	}

	if upsertEvnt.Uid == "" {
		return utils.NewMissingDataErr("user id", "string", nil)
	}

	cacheClient, err := dacstore.NewRedisClient(ctx)
	if err != nil {
		return dacstore.NewRedisConnErr(cacheClient.ClientID(ctx), err)
	}

	usrInfo, err := dacstore.CheckUserData(ctx, cacheClient, clnt.Connection.RemoteAddr().String())
	var noResults *dacstore.ErrNoCacheResult

	if err != nil {
		if errors.As(err, &noResults) {
			return utils.NewInvalidUser(upsertEvnt.Uid, err)
		}
	}

	usr, ok := usrInfo.(dtos.UserDto)
	if !ok {
		return utils.NewTypeCastErr(usrInfo, dtos.UserDto{}, nil)
	}

	if upsertEvnt.Uid != usr.Uid {
		return utils.NewPermissionErr("edit task", "user must own task to remove it", nil)
	}

	results, err := controller.EditTask(ctx, upsertEvnt.Tid, upsertEvnt.Uid, taskStart, taskEnd, upsertEvnt.Detail)
	if err != nil {
		return utils.NewDbErr("edit task", "task", err)
	}

	rsltRes, ok := results.(*models.TaskEditResponse)
	if !ok {
		return utils.NewTypeCastErr(rsltRes, results, nil)
	}

	msg := dtos.NewEventDto(envt.Type, rsltRes)

	if err = utils.MustSendMessage(clnt.Connection, msg); err != nil {
		return utils.NewFailedSendEventResponse(clnt.Connection, envt.Type, err)
	}

	return nil
}
