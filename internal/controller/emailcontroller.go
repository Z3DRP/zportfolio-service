package controller

import (
	"context"
	"fmt"

	"github.com/Z3DRP/zportfolio-service/internal/adapters"
	"github.com/Z3DRP/zportfolio-service/internal/dtos"
	"github.com/Z3DRP/zportfolio-service/internal/models"
	"github.com/Z3DRP/zportfolio-service/internal/services"
)

// TODO pass in user customizations instead of using default
func SendTaskRequestNotificationEmail(ctx context.Context, task models.Task, usrData adapters.UserData, emlInfo adapters.EmailInfo) error {
	tskData := adapters.NewTaskData(task)
	alertDTO := dtos.NewTaskRequestAlertDto(tskData, usrData, *adapters.NewCustomizations())
	emailRequestDto := dtos.NewZemailRequestDto(
		"zachpalmer1017@gmail.com",
		alertDTO.AlertTypeString(),
		emlInfo.Body,
		emlInfo.Cc,
		emlInfo.UseHtml,
		alertDTO,
	)
	emailServ, err := services.Initialize().Build(*emailRequestDto)
	if err != nil {
		return fmt.Errorf("could not build email client:: %v", err)
	}

	err = emailServ.(services.Service).Execute(ctx)
	if err != nil {
		return fmt.Errorf("failed to send task request email:: %w", err)
	}
	return nil
}
