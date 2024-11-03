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
	Start  string
	End    string
	Detail string
	Uid    string
}

type TaskRequestAlertDTO struct {
	adp.TaskData
	adp.UserData
	adp.Customizations
	alertType enums.ZemailType
}

func NewTaskRequestAlertDto(tskData adp.TaskData, usrData adp.UserData, cstms adp.Customizations) *TaskRequestAlertDTO {
	return &TaskRequestAlertDTO{
		TaskData:       tskData,
		UserData:       usrData,
		Customizations: cstms,
		alertType:      enums.ZemailType(0),
	}
}

func (ts TaskRequestAlertDTO) String() string {
	return fmt.Sprintf("%#v\n", ts)
}

func (ts TaskRequestAlertDTO) AlertType() int {
	return ts.alertType.Index()
}

func (ts TaskRequestAlertDTO) AlertTypeString() string {
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
