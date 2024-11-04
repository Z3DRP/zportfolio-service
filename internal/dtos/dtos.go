package dtos

import (
	"fmt"

	"github.com/Z3DRP/zportfolio-service/enums"
	adp "github.com/Z3DRP/zportfolio-service/internal/adapters"
)

type DTOer interface {
	String() string
}

type AlertDTOer interface {
	AlertType() int
	AlertTypeString() string
}

type TaskRequestDTO struct {
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
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
}

type ErrMessage struct {
	Err error
}

func (e ErrMessage) Error() string {
	return fmt.Sprintf("could not read message from websocket, Err: %v", e.Err)
}

func (e ErrMessage) Unwrap() error {
	return e.Err
}
