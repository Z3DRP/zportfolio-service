package eml

import (
	"fmt"

	"github.com/Z3DRP/zportfolio-service/config"
	"github.com/Z3DRP/zportfolio-service/internal/dtos"
)

type ZEmail struct {
	Recipient    string
	Sender       string
	Subject      string
	Body         string
	BCC          []string
	UseHtml      bool
	HtmlTemplate string // might have to change to a temple cmp
	TemplateBody string // have this be a template.. user enters contact info, email, name, msg etc to form then it is fed into a
	// a go standrd lib template to create the email body
	// NOTE TemplateBody could be extended to have a template type to offer other message body templates
	username string
	pwd      string
}

func NewDefaultZemail() *ZEmail {
	return &ZEmail{
		Recipient:    "",
		Sender:       "",
		Subject:      "",
		Body:         "",
		BCC:          make([]string, 0),
		UseHtml:      false,
		HtmlTemplate: "",
		username:     "",
		pwd:          "",
	}
}

func NewZemail(ops ...func(*ZEmail)) *ZEmail {
	z := NewDefaultZemail()
	for _, op := range ops {
		op(z)
	}
	return z
}

func (ze *ZEmail) Build(ops ...func(*ZEmail)) (*ZEmail, error) {
	emlConfig, err := config.ReadEmailConfig()
	if err != nil {
		return &ZEmail{}, fmt.Errorf("could not find sender email settings:: %v", err)
	}

	allOps := append([]func(*ZEmail){withUsername(emlConfig.SenderAddress), withUserPwd(emlConfig.SenderPwd)}, ops...)

	return NewZemail(allOps...), nil
}

func CreateConfigList(emlReq dtos.ZemailRequestDto) []func(*ZEmail) {
	configOps := make([]func(*ZEmail), 0)
	if emlReq.To != "" {
		configOps = append(configOps, WithRecipient(emlReq.To))
	}
	if emlReq.From != "" {
		configOps = append(configOps, WithSender(emlReq.From))
	}
	//TODO remove body from ZEMail struct
	// if emlReq.Body != "" {
	//	configOps = append(configOps, WithBody(emlReq.Body))
	// }
	configOps = append(configOps, WithUseHtml(emlReq.UseHtml))

	//TODO based on the type of email generate the corresponding template then set the text or html on struct
	if emlReq.UseHtml {
		configOps = append(configOps, WithHtmlTemplateBody(emlReq.Body))
	} else {
		configOps = append(configOps, WithTemplateBody(emlReq.Body))
	}
	if len(emlReq.Bcc) > 0 {
		configOps = append(configOps, WithBCc(emlReq.Bcc))
	}
	if emlReq.Subject != "" {
		configOps = append(configOps, WithSubject(emlReq.Subject))
	}

	return configOps
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

func WithBody(bdy string) func(z *ZEmail) {
	return func(z *ZEmail) {
		z.Body = bdy
	}
}

func WithBCc(c []string) func(z *ZEmail) {
	return func(z *ZEmail) {
		z.BCC = c
	}
}

func WithUseHtml(isIt bool) func(z *ZEmail) {
	return func(z *ZEmail) {
		z.UseHtml = isIt
	}
}

func WithTemplateBody(tbody string) func(z *ZEmail) {
	return func(z *ZEmail) {
		z.TemplateBody = tbody
	}
}

func WithHtmlTemplateBody(tbody string) func(z *ZEmail) {
	return func(z *ZEmail) {
		z.HtmlTemplate = tbody
	}
}
