package controller

import (
	"context"
	"fmt"

	"github.com/Z3DRP/zportfolio-service/enums"
	"github.com/Z3DRP/zportfolio-service/internal/adapters"
	"github.com/Z3DRP/zportfolio-service/internal/dtos"
	"github.com/Z3DRP/zportfolio-service/internal/models"
	"github.com/Z3DRP/zportfolio-service/internal/services"
)

const recipientAddr = "zachpalmer1017@gmail.com"

type ErrFailedEmailNotification struct {
	NotificationType int
	Err              error
}

func (e ErrFailedEmailNotification) Error() string {
	return fmt.Sprintf("failed to send '%v' email notification", enums.ZemailType(e.NotificationType).String())
}

func (e ErrFailedEmailNotification) Unwrap() error {
	return e.Err
}

type ErrServiceCreation struct {
	ServiceType string
	Err         error
}

func (e ErrServiceCreation) Error() string {
	return fmt.Sprintf("could not build %v service", e.ServiceType)
}

func (e ErrServiceCreation) Unwrap() error {
	return e.Err
}

type ErrUnknown struct {
	ServiceType         string
	RecievedServiceType string
	Err                 error
}

func (e ErrUnknown) Error() string {
	return fmt.Sprintf("unexpected %v service error, type (%v) could not be used for service", e.ServiceType, e.RecievedServiceType)
}

func (e ErrUnknown) Unwrap() error {
	return e.Err
}

// TODO pass in user customizations instead of using default
func SendTaskNotificationEmail(ctx context.Context, task models.Task, usrData adapters.UserData, emlInfo adapters.EmailInfo, notiType enums.ZemailType) error {
	tskData := adapters.NewTaskData(task)
	alertDTO := dtos.NewTaskNotificationDto(tskData, usrData, *adapters.NewCustomizations(), notiType)
	emailRequestDto := dtos.NewZemailRequestDto(
		recipientAddr,
		alertDTO.AlertTypeString(),
		emlInfo.Body,
		emlInfo.Cc,
		emlInfo.UseHtml,
		alertDTO,
	)

	emailServ, err := services.Initialize().Build(*emailRequestDto)
	if err != nil {
		return ErrServiceCreation{ServiceType: "email", Err: err}
	}

	serv, ok := emailServ.(services.Service)
	if !ok {
		return ErrUnknown{ServiceType: "email", RecievedServiceType: fmt.Sprintf("%T", emailServ)}
	}

	err = serv.Execute(ctx)
	if err != nil {
		return ErrFailedEmailNotification{NotificationType: emailRequestDto.EmailData.AlertType(), Err: err}
	}

	return nil
}

func SendThanksNotification(ctx context.Context, usrData adapters.UserData, emlInfo adapters.EmailInfo) error {
	alertDTO := dtos.NewThanksAlertDto(usrData, *adapters.NewCustomizations())
	emailReqDto := dtos.NewZemailRequestDto(
		recipientAddr,
		alertDTO.AlertTypeString(),
		emlInfo.Body,
		emlInfo.Cc,
		emlInfo.UseHtml,
		alertDTO,
	)

	emailServ, err := services.Initialize().Build(*emailReqDto)
	if err != nil {
		return ErrServiceCreation{ServiceType: "email", Err: err}
	}

	serv, ok := emailServ.(*services.EmailService)
	if !ok {
		return ErrUnknown{ServiceType: "email", RecievedServiceType: fmt.Sprintf("%T", emailServ)}
	}

	err = serv.Execute(ctx)
	if err != nil {
		return ErrFailedEmailNotification{NotificationType: emailReqDto.EmailData.AlertType(), Err: err}
	}

	return nil
}
