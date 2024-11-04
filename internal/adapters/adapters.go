package adapters

import (
	"strings"

	"github.com/Z3DRP/zportfolio-service/internal/models"
)

type TaskRequest struct {
}
type TaskData struct {
	FmtTaskInfo string
	Details     string
	Method      string
}

func NewTaskData(task models.Task) TaskData {
	return TaskData{
		FmtTaskInfo: task.FormattedDateTime(),
		Details:     task.Detail,
		Method:      task.Method.String(),
	}
}

type UserData struct {
	Name    string
	Company string
	Email   string
	Phone   string
	Roles   []string
}

func NewUserData(nm, cm, em, ph, roles string) UserData {
	return UserData{
		Name:    nm,
		Company: cm,
		Email:   em,
		Phone:   ph,
		Roles:   strings.Split(roles, ","),
	}
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
