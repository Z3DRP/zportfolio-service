package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
	Id        primitive.ObjectID `bson:"_id:omitempty"`
	StartTime time.Time          `bson:"start_time"`
	EndTime   time.Time          `bson:"end_time"`
	Detail    string             `bson:"detail"`
}

type Tasklist []Task

func NewTask(start, end time.Time, details string) *Task {
	return &Task{
		StartTime: start,
		EndTime:   end,
		Detail:    details,
	}
}

func (t Task) ViewAttr() string {
	return fmt.Sprintf("Task: {Id: %v, Start: %v, End: %v, Detail: %v}", t.Id, t.StartTime.String(), t.EndTime.String(), t.Detail)
}

func (t *Task) Date() string {
	yr, month, day := t.StartTime.Date()
	return fmt.Sprintf("%v-%v-%v", day, month, yr)
}

func (t *Task) WeekDayName() string {
	return t.StartTime.Weekday().String()
}

func (t *Task) MonthName() string {
	return t.StartTime.Month().String()
}

func (t *Task) Month() int {
	return int(t.StartTime.Month())
}

func (t *Task) Day() int {
	return t.StartTime.Day()
}

func (t *Task) Yr() int {
	return t.StartTime.Year()
}

func (t *Task) FormattedDate() string {
	return fmt.Sprintf("%v, %v %v, %v", t.WeekDayName(), t.MonthName(), t.Day(), t.Yr())
}

func (t *Task) FormattedTime() string {
	return fmt.Sprintf("%v:%v - %v:%v", t.StartTime.Hour(), t.StartTime.Minute(), t.EndTime.Hour(), t.EndTime.Minute())
}

func (t *Task) WeekDay() int {
	return int(t.StartTime.Weekday())
}

func (t *Task) TaskDayKey() string {
	return fmt.Sprintf(
		"%v",
		t.StartTime.Weekday(),
	)
}

func (t *Task) TaskHrKey() int {
	return t.StartTime.Hour()
}
