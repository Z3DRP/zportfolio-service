package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/Z3DRP/zportfolio-service/config"
	"github.com/Z3DRP/zportfolio-service/internal/dacstore"
	"github.com/Z3DRP/zportfolio-service/internal/models"
	zlg "github.com/Z3DRP/zportfolio-service/internal/zlogger"
)

var logfile = zlg.NewLogFile(
	zlg.WithFilename(fmt.Sprintf("%v/%v", config.LogPrefix, "controller.log")),
)
var logger = config.NewLogger(logfile, "trace", true, false)

func GetExperiences(ctx context.Context) ([]models.Experience, error) {
	dbStore, err := dacstore.CreateExperienceStore(ctx)
	xpData := make([]models.Experience, 0)
	if err != nil {
		return []models.Experience{}, err
	}

	xp, err := dbStore.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	for _, exp := range xp {
		if exp, ok := exp.(models.Experience); ok {
			xpData = append(xpData, exp)
		} else {
			return nil, fmt.Errorf("invalid type for experience: %T", exp)
		}
	}
	return xpData, nil
}

func GetDetails(ctx context.Context) (models.Detail, error) {
	dbStore, err := dacstore.CreateDetailStore(ctx)
	logger.MustTrace(fmt.Sprintf("detail store created: %v", dbStore))

	var detail models.Detail
	if err != nil {
		logger.MustTrace(fmt.Sprintf("error during detail store creation: %s", err))
		return models.Detail{}, err
	}

	details, err := dbStore.FetchByName(ctx, "Zach Palmer")
	if err != nil {
		return models.Detail{}, err
	}

	detail, ok := details.(models.Detail)
	if !ok {
		return models.Detail{}, fmt.Errorf("invalid type for detail: %T", detail)
	}

	return detail, nil
}

func GetSkills(ctx context.Context) ([]models.Skill, error) {
	dbStore, err := dacstore.CreateSkillStore(ctx)
	skillsData := make([]models.Skill, 0)
	if err != nil {
		return []models.Skill{}, err
	}

	skills, err := dbStore.Fetch(ctx)
	if err != nil {
		return []models.Skill{}, err
	}

	for _, skill := range skills {
		if skill, ok := skill.(models.Skill); ok {
			skillsData = append(skillsData, skill)
		}
	}
	return skillsData, nil
}

func GetPortfolioData(ctx context.Context) (models.Responser, error) {
	skills, err := GetSkills(ctx)
	if err != nil {
		skillErr := fmt.Errorf("%w: %s", dacstore.ErrFetchSkill, err)
		return nil, skillErr
	}

	xp, err := GetExperiences(ctx)
	if err != nil {
		xpErr := fmt.Errorf("%w: %s", dacstore.ErrFetchExperience, err)
		return nil, xpErr
	}

	details, err := GetDetails(ctx)
	if err != nil {
		detailErr := fmt.Errorf("%w: %s", dacstore.ErrFetchDetails, err)
		return nil, detailErr
	}

	return models.NewPortfolioResponse(details, xp, skills), nil
}

// TODO fetchSchedule and createTask, and getTasks
// NOTE fetchSchedule will have to call all of the 'builder' methods to build out the hrly schedule and stuff

func FetchSchedule(ctx context.Context, start, end time.Time) (models.Responser, error) {

}

func createTask(ctx context.Context, start, end time.Time, details string) (models.Responser, error) {

}
