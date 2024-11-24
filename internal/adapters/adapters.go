package adapters

import (
	"fmt"
	"strings"
)

type TaskRequest struct {
}
type TaskData struct {
	FmtTaskInfo string
	Details     string
	Method      string
}

func NewTaskData(fmtDateTime, details, method string) TaskData {
	return TaskData{
		FmtTaskInfo: fmtDateTime,
		Details:     details,
		Method:      method,
	}
}

type UserData struct {
	Name    string
	Company string
	Email   string
	Phone   string
	Roles   string
}

func NewUserData(nm, cm, em, ph, roles string) UserData {
	return UserData{
		Name:    nm,
		Company: cm,
		Email:   em,
		Phone:   ph,
		Roles:   roles,
	}
}

func (u UserData) String() string {
	return fmt.Sprintf("%#v\n", u)
}

type EmailInfo struct {
	Cc      []string
	Body    string
	UseHtml bool
}

func NewEmailInfo(cc, body string, uHtml bool) EmailInfo {
	return EmailInfo{
		Cc:      strings.Split(cc, ","),
		Body:    body,
		UseHtml: uHtml,
	}
}

func DefaultEmlInfo() EmailInfo {
	return EmailInfo{
		Cc:      make([]string, 0),
		Body:    "",
		UseHtml: true,
	}
}

func (e EmailInfo) String() string {
	return fmt.Sprintf("%#v\n", e)
}

type Customizations struct {
	BodyColor   string
	TextColor   string
	BannerImage string // use base 64 encode image
}

func DefaultCustomizations() Customizations {
	return Customizations{
		BodyColor:   "#ffffff",
		TextColor:   "#000000",
		BannerImage: "",
	}
}

func NewCustomizations(ops ...func(*Customizations)) *Customizations {
	cs := DefaultCustomizations()
	for _, op := range ops {
		op(&cs)
	}
	return &cs
}

func WithBody(bc string) func(*Customizations) {
	return func(c *Customizations) {
		c.BodyColor = bc
	}
}

func WithTextColor(tc string) func(*Customizations) {
	return func(c *Customizations) {
		c.TextColor = tc
	}
}

func WithBannerImage(bi string) func(*Customizations) {
	return func(c *Customizations) {
		c.BannerImage = bi
	}
}
