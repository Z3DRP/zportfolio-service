package dtos

import (
	"fmt"

	"github.com/Z3DRP/zportfolio-service/enums"
	adp "github.com/Z3DRP/zportfolio-service/internal/adapters"
)

type DTOer interface {
	String() string
}

type TaskRequestAlertDTO struct {
	adp.TaskData
	adp.UserData
	adp.Customizations
}

func (ts TaskRequestAlertDTO) String() string {
	return fmt.Sprintf("%#v\n", ts)
}

type ThanksAlertDTO struct {
	adp.UserData
	adp.Customizations
}

func (th ThanksAlertDTO) String() string {
	return fmt.Sprintf("%#v\n", th)
}

type ZemailRequestDto struct {
	To        string
	From      string
	Subject   string
	Bcc       []string
	Body      string
	UseHtml   bool
	EmailType enums.ZemailType
}

func NewZemailRequest(to, frm, sub, bdy string, bcc []string, useHtml bool, eType enums.ZemailType) *ZemailRequestDto {
	return &ZemailRequestDto{
		To:        to,
		From:      frm,
		Subject:   sub,
		Bcc:       bcc,
		Body:      bdy,
		UseHtml:   useHtml,
		EmailType: eType,
	}
}
