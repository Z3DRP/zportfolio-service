package services

import (
	"context"
	"fmt"

	"github.com/Z3DRP/zportfolio-service/internal/dtos"
	"github.com/Z3DRP/zportfolio-service/internal/eml"
	"github.com/wneessen/go-mail"
)

type EmlServicer interface {
	Build(dtos.ZemailRequestDto) (EmlServicer, error)
}

type EmailService struct {
	*eml.ZEmail
}

func Initialize() EmlServicer {
	return &EmailService{}
}

func (ems *EmailService) Build(e dtos.ZemailRequestDto) (EmlServicer, error) {
	var err error
	configList := eml.CreateConfigList(e)
	ems.ZEmail, err = eml.Build(e.EmailData, configList...)

	if err != nil {
		return nil, err
	}
	return ems, nil
}

func (ems *EmailService) Execute(ctx context.Context) error {
	msg := mail.NewMsg()
	if err := msg.From(ems.Sender); err != nil {
		return err
	}
	if err := msg.To(ems.Recipient); err != nil {
		return err
	}

	msg.Subject(ems.Subject)
	if err := msg.Cc(ems.CC...); err != nil {
		return fmt.Errorf("could not set cc list:: %w", err)
	}

	if ems.UseHtml {
		if err := msg.SetBodyHTMLTemplate(*ems.HtmlBody, ems.TemplateData); err != nil {
			return fmt.Errorf("could not set html template:: %w", err)
		}
		if err := msg.AddAlternativeTextTemplate(*ems.TextBody, ems.TemplateData); err != nil {
			return fmt.Errorf("could not set alternative text template:: %w", err)
		}
	} else {
		if err := msg.SetBodyTextTemplate(*ems.TextBody, ems.TemplateData); err != nil {
			return fmt.Errorf("could not set text template:: %w", err)
		}
	}

	client, err := mail.NewClient(
		ems.SmtpServer(),
		mail.WithPort(ems.SmtpPort()),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(ems.Username()),
		mail.WithPassword(ems.Pwd()),
	)

	if err != nil {
		return fmt.Errorf("an error occurred while trying to create email client:: %w", err)
	}

	if err := client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("an error occurred while dialing and sending message:: %w", err)
	}

	return nil
}
