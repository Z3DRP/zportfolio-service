package services

import (
	"github.com/Z3DRP/zportfolio-service/internal/dtos"
	"github.com/Z3DRP/zportfolio-service/internal/eml"
)

type EmailService struct {
	*eml.ZEmail
	dtos.DTOer
}

func (ems *EmailService) Initialize() Service {
	return &EmailService{}
}

func (ems *EmailService) Build(e dtos.ZemailRequestDto, dto dtos.DTOer) (Service, error) {
	var err error
	configList := eml.CreateConfigList(e)
	ems.ZEmail, err = eml.NewDefaultZemail().Build(configList...)
	if err != nil {
		return nil, err
	}
	return ems, nil
}

func (ems *EmailService) Execute() error {
	return nil
}
