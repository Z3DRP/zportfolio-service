package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Availabler interface {
	FormattedTime() string
}

type Day int

const (
	Sunday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

func (d Day) String() string {
	return [...]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}[d]
}

func (d Day) Eindex() int {
	return int(d)
}

func NewDay(d int) Day {
	switch d {
	case 1:
		return Monday
	case 2:
		return Tuesday
	case 3:
		return Wednesday
	case 4:
		return Thursday
	case 5:
		return Friday
	case 6:
		return Saturday
	default:
		return Sunday
	}
}

type AvailableDay struct {
	WeekDay  int
	FromHour int
	FromMin  int
	ToHour   int
	ToMinute int
}

func (a *AvailableDay) FormattedTime() string {
	return fmt.Sprintf("%v:%v - %v:%v", a.FromHour, a.FromMin, a.ToHour, a.ToMinute)
}

func NewAvailableDay(wkday, frmHr, frmMin, toHr, toMin int) Availabler {
	return &AvailableDay{
		WeekDay:  wkday,
		FromHour: frmHr,
		FromMin:  frmMin,
		ToHour:   toHr,
		ToMinute: toMin,
	}
}

type Availability struct {
	Id            primitive.ObjectID `bson:"_id,omitempty"`
	Day           int                `bson:"day"`
	DayName       string             `bson:"day_name"`
	AvailableFrom time.Time          `bson:"available_from"`
	AvailableTo   time.Time          `bson:"available_to"`
	CreatedAt     time.Time          `bson:"created_at"`
	Newest        bool               `bson:"newest"`
}

func (a *Availability) FormattedTime() string {
	return fmt.Sprintf("%v:%v - %v:%v", a.AvailableFrom.Hour(), a.AvailableFrom.Minute(), a.AvailableFrom.Hour(), a.AvailableTo.Minute())
}

func (a Availability) ViewAttr() string {
	return fmt.Sprintf("Availibility:: Id: %v, Day: %v, DayName: %v, From: %v, To: %v, Newest: %v", a.Id, a.Day, a.DayName, a.AvailableFrom, a.AvailableTo, a.Newest)
}

func NewAvailability(openDay AvailableDay) Availabler {
	curYr := time.Now().Year()
	avbFrom := time.Date(curYr, STD_MONTH, STD_DAY, openDay.FromHour, openDay.FromMin, STD_SEC, STD_SEC, time.Local)
	avbTo := time.Date(curYr, STD_MONTH, STD_DAY, openDay.FromHour, openDay.FromMin, STD_SEC, STD_SEC, time.Local)
	day := NewDay(openDay.WeekDay)
	return &Availability{
		Day:           openDay.WeekDay,
		DayName:       day.String(),
		AvailableFrom: avbFrom,
		AvailableTo:   avbTo,
	}
}

func (d Availability) Time() string {
	return fmt.Sprintf("%v:%v - %v:%v", d.AvailableFrom.Hour(), d.AvailableFrom.Minute(), d.AvailableTo.Hour(), d.AvailableTo.Minute())
}

func (d Availability) DayKey() int {
	return d.AvailableFrom.Day()
}

func (d Availability) HourKey() int {
	return d.AvailableFrom.Hour()
}

func (d Availability) WeekDayKey() int {
	return int(d.AvailableFrom.Weekday())
}

func (d Availability) StartDay() string {
	return d.AvailableFrom.Weekday().String()
}

func (d Availability) EndDay() string {
	return d.AvailableFrom.Weekday().String()
}

func (d Availability) StartTime() string {
	return fmt.Sprintf("%v:%v", d.AvailableFrom.Hour(), d.AvailableFrom.Minute())
}

func (d Availability) EndTime() string {
	return fmt.Sprintf("%v:%v", d.AvailableTo.Hour(), d.AvailableTo.Minute())
}

func (d Availability) FormattedAvailability() string {
	return fmt.Sprintf("%v %v %v, %v - %v", d.StartDay(), d.AvailableFrom.Month().String(), d.AvailableFrom.Day(), d.AvailableFrom.Year(), d.FormattedTime())
}
