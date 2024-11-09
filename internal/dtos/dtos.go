package dtos

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Z3DRP/zportfolio-service/enums"
	adp "github.com/Z3DRP/zportfolio-service/internal/adapters"
	"github.com/Z3DRP/zportfolio-service/internal/models"
)

type DTOer interface {
	String() string
}

type AlertDTOer interface {
	AlertType() int
	AlertTypeString() string
}

type TaskAlertDTO struct {
	adp.TaskData
	adp.UserData
	adp.Customizations
	alertType enums.ZemailType
}

func NewTaskNotificationDto(tskData adp.TaskData, usrData adp.UserData, cstms adp.Customizations, notificationType enums.ZemailType) *TaskAlertDTO {
	return &TaskAlertDTO{
		TaskData:       tskData,
		UserData:       usrData,
		Customizations: cstms,
		alertType:      enums.ZemailType(notificationType),
	}
}

func (ts TaskAlertDTO) String() string {
	return fmt.Sprintf("%#v\n", ts)
}

func (ts TaskAlertDTO) AlertType() int {
	return ts.alertType.Index()
}

func (ts TaskAlertDTO) AlertTypeString() string {
	return ts.alertType.String()
}

type ThanksAlertDTO struct {
	adp.UserData
	adp.Customizations
	alertType enums.ZemailType
}

func NewThanksAlertDto(usrData adp.UserData, cstms adp.Customizations) *ThanksAlertDTO {
	return &ThanksAlertDTO{
		UserData:       usrData,
		Customizations: cstms,
		alertType:      enums.ZemailType(3),
	}
}

func (th ThanksAlertDTO) String() string {
	return fmt.Sprintf("%#v\n", th)
}

func (ta ThanksAlertDTO) AlertTypeString() string {
	return ta.alertType.String()
}

func (ta ThanksAlertDTO) AlertType() int {
	return ta.alertType.Index()
}

type ZemailRequestDto struct {
	To         string
	Subject    string
	Cc         []string
	CustomBody string
	UseHtml    bool
	EmailData  AlertDTOer
}

func (ze ZemailRequestDto) String() string {
	return fmt.Sprintf("%#v\n", ze)
}

func NewZemailRequestDto(to, sub, bdy string, cc []string, useHtml bool, data AlertDTOer) *ZemailRequestDto {
	return &ZemailRequestDto{
		To:         to,
		Subject:    sub,
		Cc:         cc,
		CustomBody: bdy,
		UseHtml:    useHtml,
		EmailData:  data,
	}
}

type Message struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

type SocketErrMsg struct {
	ErrMsg      string
	CodeMessage string
	ConnCode    int
}

type DayDto struct {
	Day       int
	WeekDay   string
	ShortDate string
	LongDate  string
	FmtDate   string
}

func NewDayDto(day int, wkDay, shDate, lnDate, fmDate string) DayDto {
	return DayDto{
		Day:       day,
		WeekDay:   wkDay,
		ShortDate: shDate,
		LongDate:  lnDate,
		FmtDate:   fmDate,
	}
}

type PeriodDto struct {
	StartDate string
	EndDate   string
	FmtString string
	Days      []DayDto
}

func NewPeriodDto(p models.Period) PeriodDto {
	periodDays := p.GetDaysInPeriod()
	days := make([]DayDto, p.Length)
	for _, day := range periodDays {
		shrtDate := fmt.Sprintf("%v/%v/%v", day.Month(), day.Day(), day.Year())
		lngDate := fmt.Sprintf("%v %v %v", day.Weekday().String(), day.Month().String(), day.Day())
		days = append(days, NewDayDto(day.Day(), day.Weekday().String(), shrtDate, lngDate, day.String()))
	}

	return PeriodDto{
		StartDate: p.FmtStart(),
		EndDate:   p.FmtEnd(),
		FmtString: fmt.Sprintf("%v - %v", p.FmtStart(), p.FmtEnd()),
		Days:      days,
	}
}

type ScheduleDto struct {
	Availability   map[int]AvailabilityDto
	Agenda         []models.HourlySchedule
	DaysAvailable  map[int]bool
	HoursAvailable map[string]bool // keys are composites of weekday and hour
	CurrentPeriod  PeriodDto
}

func initScheduleDto(sch models.Schedule, p models.Period) ScheduleDto {
	avbs := make(map[int]AvailabilityDto, len(sch.Availability))
	for _, av := range sch.Availability {
		if _, ok := avbs[av.WeekDayKey()]; !ok {
			avbs[av.WeekDayKey()] = NewAvailabilityDto(av)
		}
	}
	return ScheduleDto{
		Availability:   avbs,
		Agenda:         sch.HourlyAgenda,
		DaysAvailable:  sch.DaysAvailable,
		HoursAvailable: sch.HoursAvailable,
		CurrentPeriod:  NewPeriodDto(p),
	}
}

func NewScheduleDto(s *models.ScheduleResponse) ScheduleDto {
	return initScheduleDto(*s.Agenda, s.CurrentPeriod)
}

type AvailabilityDto struct {
	Day             int
	WeekDay         int
	DayName         string
	AvaiableFrom    string
	FmtAvaiableFrom string
	AvailableTo     string
	FmtAvaiableTo   string
}

func NewAvailabilityDto(avb models.Availability) AvailabilityDto {
	return AvailabilityDto{
		Day:             avb.Day,
		WeekDay:         avb.WeekDayKey(),
		DayName:         avb.DayName,
		AvaiableFrom:    fmt.Sprintf("%v:%v", avb.AvailableFrom.Hour(), avb.AvailableFrom.Minute()),
		FmtAvaiableFrom: Get12HrFmt(avb.AvailableFrom),
		AvailableTo:     fmt.Sprintf("%v:%v", avb.AvailableTo.Hour(), avb.AvailableTo.Minute()),
		FmtAvaiableTo:   Get12HrFmt(avb.AvailableTo),
	}
}

type PeriodPayloadDTO struct {
	PeriodStart string
	PeriodEnd   string
}

type TaskPayloadDTO struct {
	Start       string
	End         string
	Detail      string
	Uid         string
	UsrName     string
	Company     string
	Email       string
	Phone       string
	Roles       string
	Cc          string
	Body        string
	BodyColor   string
	TextColor   string
	BannerImage string
	UseHtml     bool
}

type TaskDeletePaylaod struct {
	Uid string
	Tid string
}

type UserDto struct {
	adp.UserData
	Ip  string
	Uid string
}

func NewUserDto(usr adp.UserData, ip, uid string) UserDto {
	return UserDto{
		UserData: usr,
		Ip:       ip,
		Uid:      uid,
	}
}

func (u UserDto) String() string {
	return fmt.Sprintf("%#v\n", u)
}

func Get12HrFmt(t time.Time) string {
	hr := convertHr(t.Hour())
	antem := getAntemeridiem(hr)
	return fmt.Sprintf("%v:%v %s", hr, t.Minute(), antem)
}

func getAntemeridiem(hr int) string {
	antemeridiem := "AM"
	if hr >= 12 {
		antemeridiem = "PM"
	}
	return antemeridiem
}

func convertHr(h int) int {
	hr := h % 12
	if hr == 0 {
		hr = 12
	}
	return hr
}
