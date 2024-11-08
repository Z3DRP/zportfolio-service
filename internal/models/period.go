package models

import (
	"fmt"
	"time"

	"github.com/Z3DRP/zportfolio-service/enums"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Period struct {
	Id        primitive.ObjectID `bson:"_id,omitempty"`
	StartDate time.Time          `bson:"start_date"`
	EndDate   time.Time          `bson:"end_date"`
	Length    int                `bson:"length"`
	Type      enums.PeriodType   `bson:"type"`
	IsActive  bool               `bson:"is_active"`
}

func NewPeriod(length int, ptype enums.PeriodType, start, end time.Time, active bool) *Period {
	return &Period{
		StartDate: start,
		EndDate:   end,
		Length:    length,
		Type:      ptype,
		IsActive:  active,
	}
}

func (p Period) FmtStart() string {
	return fmt.Sprintf("%v/%v/%v", p.StartDate.Month(), p.StartDate.Day(), p.StartDate.Year())
}

func (p Period) FmtEnd() string {
	return fmt.Sprintf("%v/%v/%v", p.EndDate.Month(), p.EndDate.Day(), p.EndDate.Year())
}

func (pd Period) GetDaysInPeriod() []time.Time {
	daysInPeriod := make([]time.Time, 0)
	for d := 0; d <= pd.Length; d++ {
		nwDay := pd.StartDate.AddDate(0, 0, d)
		if nwDay.Compare(pd.EndDate) != 1 {
			daysInPeriod = append(daysInPeriod, nwDay)
		}
	}
	return daysInPeriod
}

func (p Period) ViewAttr() string {
	return fmt.Sprintf("start: %v, end: %v, length: %v, type: %v", p.StartDate, p.EndDate, p.Length, p.Type)
}
