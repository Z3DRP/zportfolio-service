package eml

import (
	"fmt"

	"github.com/Z3DRP/zportfolio-service/config"
	"github.com/wneessen/go-mail"
)

type Emailer interface {
	Send(*mail.Msg) bool
}

type ZEmail struct {
	To           string
	From         string
	Subject      string
	Body         string
	CC           string
	IsHtml       bool
	HtmlTemplate string // might have to change to a temple cmp
	TemplateBody string // have this be a template.. user enters contact info, email, name, msg etc to form then it is fed into a
	// a go standrd lib template to create the email body
	// NOTE TemplateBody could be extended to have a template type to offer other message body templates
	username string
	pwd      string
}

func NewDefaultZemail() *ZEmail {
	return &ZEmail{
		To:           "",
		From:         "",
		Subject:      "",
		Body:         "",
		CC:           "",
		IsHtml:       false,
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

func Build(ops ...func(*ZEmail)) (*ZEmail, error) {
	emlConfig, err := config.ReadEmailConfig()
	if err != nil {
		return &ZEmail{}, fmt.Errorf("could not find sender email settings:: %v", err)
	}

	allOps := append([]func(*ZEmail){withUsername(emlConfig.SenderAddress), withUserPwd(emlConfig.SenderPwd)}, ops...)

	return NewZemail(allOps...), nil
}

func (ze *ZEmail) Recipient(rcp string) {
	ze.To = rcp
}

func withUsername(usr string) func(*ZEmail) {
	return func(z *ZEmail) {
		z.username = usr
		z.From = usr
	}
}

func withUserPwd(pwd string) func(*ZEmail) {
	return func(z *ZEmail) {
		z.pwd = pwd
	}
}

func WithTo(to string) func(*ZEmail) {
	return func(z *ZEmail) {
		z.To = to
	}
}

func WithFrom(from string) func(*ZEmail) {
	return func(z *ZEmail) {
		z.From = from
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

func WithCc(c string) func(z *ZEmail) {
	return func(z *ZEmail) {
		z.CC = c
	}
}

func WithIsHtml(isIt bool) func(z *ZEmail) {
	return func(z *ZEmail) {
		z.IsHtml = isIt
	}
}

func WithTemplateBody(tbody string) func(z *ZEmail) {
	return func(z *ZEmail) {
		z.TemplateBody = tbody
	}
}

func (z ZEmail) Send(m *mail.Msg) bool {
	return true
}
