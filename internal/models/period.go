package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PeriodType int

const (
	Weekly = iota + 1
	BiWeekly
	Monthly
)

func (p PeriodType) String() string {
	return [...]string{"Weekly", "BiWeekly", "Monthly"}[p-1]
}

func (p PeriodType) Eindex() int {
	return int(p)
}

type Period struct {
	Id        primitive.ObjectID `bson:"_id,omitempty"`
	StartDate time.Time          `bson:"start_date"`
	EndDate   time.Time          `bson:"end_date"`
	Length    int                `bson:"length"`
	Type      PeriodType         `bson:"type"`
	IsActive  bool               `bson:"is_active"`
}

func NewPeriod(length int, ptype PeriodType, start, end time.Time, active bool) *Period {
	return &Period{
		StartDate: start,
		EndDate:   end,
		Length:    length,
		Type:      ptype,
		IsActive:  active,
	}
}

func (p Period) ViewAttr() string {
	return fmt.Sprintf("start: %v, end: %v, length: %v, type: %v", p.StartDate, p.EndDate, p.Length, p.Type)
}
