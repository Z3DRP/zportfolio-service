package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"

	"github.com/Z3DRP/zportfolio-service/config"
	"github.com/Z3DRP/zportfolio-service/internal/controller"
	"github.com/Z3DRP/zportfolio-service/internal/dacstore"
	"github.com/Z3DRP/zportfolio-service/internal/dtos"
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

func GetAbout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		logger.MustDebug(fmt.Sprintf("request: %s method: %s timed out", r.URL, r.Method))
		http.Error(w, "request time out", http.StatusRequestTimeout)
		return
	default:
		var portfolioData models.Responser
		cacheClient, err := dacstore.NewRedisClient(r.Context())
		logger.MustDebug("calling cache")
		if err != nil {
			logger.MustDebug(fmt.Sprintf("cache error: %s", err))
			http.Error(w, "request cache error", http.StatusInternalServerError)
			return
		}

		logger.MustDebug("connected without error to redis")

		portfolioData, err = dacstore.CheckPortfolioData(r.Context(), cacheClient)
		var noResult *dacstore.ErrNoCacheResult

		logger.MustDebug("about to check cache")

		if err != nil {
			logger.MustDebug(fmt.Sprintf("error returned from cahce check.. err: %v", err))
			if !errors.As(err, &noResult) {
				logger.MustDebug(fmt.Sprintf("an unexpected error occurred while retrieving cache data:: %v", err))
				http.Error(w, "an unexpected error occurred retrieving cache", http.StatusInternalServerError)
				return
			}
		}

		logger.MustDebug("no error from cache check")

		if portfolioData == nil {
			logger.MustDebug("portfolio data was nil from cache.. fetching from database ....")
			portfolioData, err = controller.GetPortfolioData(r.Context())

			logger.MustDebug("fected from db")

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
			logger.MustDebug("no error from database call")
			err = dacstore.SetPortfolioData(r.Context(), cacheClient, portfolioData)
			if err != nil {
				logger.MustDebug(fmt.Sprintf("could not set portfolio cache data: %s", err))
				http.Error(w, "an error occurred while setting portfolio cache data", http.StatusInternalServerError)
				return
			}

			logger.MustDebug("no error setting cache")
		}

		logger.MustDebug("cache was not nil")

		//settings, err := config.ReadZypherSettings()
		if err != nil {
			logger.MustDebug(fmt.Sprintf("an error occurred while reading zypher config:: %v", err))
			return
		}

		logger.MustDebug("about to update vistor count")

		//go updateVisitorCount(settings, cacheClient, r, logger)

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

func updateVisitorCount(settings config.ZypherConfig, cacheClient *redis.Client, r *http.Request, logger *zlg.Zlogrus) {
	uip := utils.GetIP(r)
	logDebug := func(logr *zlg.Zlogrus, err error) {
		logr.MustDebug(err.Error())
	}
	usr, err := dacstore.CheckUserData(r.Context(), cacheClient, uip)
	var noResults *dacstore.ErrNoCacheResult

	if err != nil {
		if !errors.As(err, &noResults) {
			logDebug(logger, err)
			return
		}
	}
	usrDto, ok := usr.(dtos.UserDto)
	if !ok {
		logDebug(logger, utils.NewTypeCastErr(usr, usrDto, nil))
		return
	}

	if usrDto.Uid == "" {
		nwUid, err := controller.CalculateZypher(uip, settings.Shift, settings.ShiftCount, settings.HashCount, settings.Alternate, settings.IgnSpace, settings.RestrictHash)
		// add user to cache so when trying to edit tasks id can be checked
		if err != nil {
			logDebug(logger, err)
			return
		}

		if err = dacstore.SetUserData(r.Context(), cacheClient, usrDto.Name, usrDto.Company, usrDto.Phone, usrDto.Email, usrDto.Roles, uip, nwUid); err != nil {
			logDebug(logger, err)
			return
		}

		if res, err := controller.CreateVisitor(r.Context(), 1, nwUid, uip, false); err != nil {
			logDebug(logger, err)
			return
		} else {
			logger.MustDebug(fmt.Sprintf("the following visitor has been created:: %v", res.PrintRes()))
			return
		}
	}

	_, _, err = controller.EditVisitorCount(r.Context(), uip)
	if err != nil {
		logDebug(logger, err)
		return
	}

	logger.MustDebug(fmt.Sprintf("visitor %v has been updated successfully", usrDto.Uid))
}
