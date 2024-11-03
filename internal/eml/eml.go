package eml

import (
	"fmt"

	"github.com/Z3DRP/zportfolio-service/config"
	"github.com/Z3DRP/zportfolio-service/internal/dtos"
	"github.com/Z3DRP/zportfolio-service/internal/emltmpl"
)

type ZEmail struct {
	HtmlBody  *emltmpl.HtmlNotificationTemplate // might have to change to a temple cmp
	TextBody  *emltmpl.TxtNotificationTemplate  // have this be a template.. user enters contact info, email, name, msg etc to form then it is fed into a
	Recipient string
	Sender    string
	Subject   string
	// a go standrd lib template to create the email body
	// NOTE TextBody could be extended to have a template type to offer other message body templates
	username     string
	pwd          string
	smtpServer   string
	smtpPort     int
	UseHtml      bool
	TemplateData dtos.AlertDTOer
	CC           []string
}

func NewDefaultZemail() *ZEmail {
	return &ZEmail{
		Recipient: "",
		Sender:    "",
		Subject:   "",
		CC:        make([]string, 0),
		UseHtml:   false,
		username:  "",
		pwd:       "",
	}
}

func NewZemail(ops ...func(*ZEmail)) *ZEmail {
	z := NewDefaultZemail()
	for _, op := range ops {
		op(z)
	}
	return z
}

func Build(emailData dtos.AlertDTOer, ops ...func(*ZEmail)) (*ZEmail, error) {
	emlConfig, err := config.ReadEmailConfig()
	if err != nil {
		return &ZEmail{}, fmt.Errorf("could not find sender email settings:: %v", err)
	}

	allOps := append(
		[]func(*ZEmail){
			withUsername(emlConfig.SenderAddress),
			withUserPwd(emlConfig.SenderPwd),
			WithSender(emlConfig.SenderAddress),
			withSmtpServer(emlConfig.SmtpServer),
			withSmtpPort(emlConfig.SmtpPort),
			WithAlertData(emailData),
		},
		ops...,
	)
	ze := NewZemail(allOps...)
	ze.SetTemplates(emailData)
	return ze, nil
}

func CreateConfigList(emlReq dtos.ZemailRequestDto) []func(*ZEmail) {
	configOps := make([]func(*ZEmail), 0)
	configOps = append(configOps, WithUseHtml(emlReq.UseHtml))
	if emlReq.To != "" {
		configOps = append(configOps, WithRecipient(emlReq.To))
	}
	if len(emlReq.Cc) > 0 {
		configOps = append(configOps, WithCc(emlReq.Cc))
	}
	if emlReq.Subject != "" {
		configOps = append(configOps, WithSubject(emlReq.Subject))
	}

	return configOps
}

func (z *ZEmail) SetTemplates(alertReq dtos.AlertDTOer) (*ZEmail, error) {
	txtTemplate, htmlTemplate, err := emltmpl.TemplateFactory(alertReq)
	if err != nil {
		return nil, err
	}

	z.TextBody = &txtTemplate
	if z.UseHtml {
		z.HtmlBody = &htmlTemplate
	}
	return z, nil
}

func WithRecipient(rcp string) func(*ZEmail) {
	return func(z *ZEmail) {
		z.Recipient = rcp
	}
}

func withUsername(usr string) func(*ZEmail) {
	return func(z *ZEmail) {
		z.username = usr
		z.Sender = usr
	}
}

func withUserPwd(pwd string) func(*ZEmail) {
	return func(z *ZEmail) {
		z.pwd = pwd
	}
}

func withSmtpServer(addr string) func(*ZEmail) {
	return func(z *ZEmail) {
		z.smtpServer = addr
	}
}

func withSmtpPort(p int) func(*ZEmail) {
	return func(z *ZEmail) {
		z.smtpPort = p
	}
}

func WithAlertData(d dtos.AlertDTOer) func(*ZEmail) {
	return func(z *ZEmail) {
		z.TemplateData = d
	}
}

func WithSender(from string) func(*ZEmail) {
	return func(z *ZEmail) {
		z.Sender = from
	}
}

func WithSubject(sub string) func(*ZEmail) {
	return func(z *ZEmail) {
		z.Subject = sub
	}
}

func WithCc(c []string) func(z *ZEmail) {
	return func(z *ZEmail) {
		z.CC = c
	}
}

func WithUseHtml(isIt bool) func(z *ZEmail) {
	return func(z *ZEmail) {
		z.UseHtml = isIt
	}
}

func (z ZEmail) SmtpPort() int {
	return z.smtpPort
}

func (z ZEmail) SmtpServer() string {
	return z.smtpServer
}

func (z ZEmail) Username() string {
	return z.username
}

func (z ZEmail) Pwd() string {
	return z.pwd
}
