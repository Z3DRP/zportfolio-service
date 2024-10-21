package models

import (
	"fmt"

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
	Id       primitive.ObjectID `bson:"_id,omitempty"`
	Length   int                `bson:"length"`
	Type     PeriodType         `bson:"type"`
	IsActive bool               `bson:"is_active"`
}

func NewPeriod(length int, ptype PeriodType, active bool) *Period {
	return &Period{
		Length:   length,
		Type:     ptype,
		IsActive: active,
	}
}

func (p Period) ViewAttr() string {
	return fmt.Sprintf("length: %v, type: %v", p.Length, p.Type)
}
