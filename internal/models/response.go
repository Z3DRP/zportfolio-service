package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Responser interface {
	PrintRes() string
}

type PortfolioResponse struct {
	AboutDetails           Detail
	ProfessionalExperience []Experience
	Skills                 []Skill
}

func (pr PortfolioResponse) PrintRes() string {
	return fmt.Sprintf("About: %v; ProfessionalExperience: %v; Skills: %v", pr.AboutDetails, pr.ProfessionalExperience, pr.Skills)
}

func NewPortfolioResponse(d Detail, e []Experience, s []Skill) *PortfolioResponse {
	return &PortfolioResponse{
		AboutDetails:           d,
		ProfessionalExperience: e,
		Skills:                 s,
	}
}

type ScheduleResponse struct {
	CurrentPeriod Period
	Agenda        *Schedule
}

func NewScheduleResponse(curPeriod Period, sched *Schedule) *ScheduleResponse {
	return &ScheduleResponse{
		CurrentPeriod: curPeriod,
		Agenda:        sched,
	}
}

func (sr *ScheduleResponse) PrintRes() string {
	return fmt.Sprintf("CurrentPeriod: %v, Agenda: %v", sr.CurrentPeriod, sr.Agenda.Agenda)
}

type TaskInsertResponse struct {
	Result primitive.ObjectID
	NwTask *Task
}

func NewTaskInsertResponse(res primitive.ObjectID, tsk *Task) *TaskInsertResponse {
	return &TaskInsertResponse{
		Result: res,
		NwTask: tsk,
	}
}

func (tr *TaskInsertResponse) PrintRes() string {
	return fmt.Sprintf("Result: %v, Task: %v, Timestamp: %v", tr.Result, tr.NwTask, tr.Result.Timestamp())
}
