package emltmpl

import (
	"fmt"
	htemplate "html/template"
	"text/template"
	"time"

	"github.com/Z3DRP/zportfolio-service/internal/adapters"
)

const TSKTXT_IDENTIFIER = "tsk-alert-txt"
const TSKHTML_IDENTIFIER = "tsk-alert-html"
const THKTXT_IDENTIFIER = "thk-alert-txt"
const THKHTML_IDENTIFIER = "thk-alert-html"

type TaskNotificationTxtTmpl *template.Template
type TaskNotificationHtmlTmpl *htemplate.Template

// TODO add customizations for colors

func NewTaskAlertTxtTmpl(alertInfo adapters.TaskRequestAlertDTO) (TaskNotificationTxtTmpl, TaskNotificationHtmlTmpl, error) {
	txtNotification := template.New(fmt.Sprintf("%v%v", TSKTXT_IDENTIFIER, time.Now().UnixMilli()))
	htmlNotification := htemplate.New(fmt.Sprintf("%v%v", TSKHTML_IDENTIFIER, time.Now().UnixMilli()))

	tmplateTxt := `Task Request Notification\n
	The following task request:\n
	Date: {{.FmtTaskInfo}}
	Method: {{.Method}}
	Details: {{.Details}}\n\n
	
	Was created by:\n
	Name: {{.Name}}
	Company: {{.Company}}
	Email: {{.Email}}
	Phone: {{.Phone}}\n\n

	{{if .Roles}} 
	To discuss the following role(s):\n 
	{{range .Roles}}- {{.}} {{end}}\n
	{{end}}
	`
	tmplateHtml := `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>Task Request Notification</title>
	</head>
	{{if .Image}}
	<header>
		<img src="data:image/png;base64,{{.Image}}" alt="Banner image" style="width:100%; height:autp" />
	</header>
	{{end}}
	<body style="font-family: sans-serif, Arial; color:{{.BodyColor}} background-color:{{.BackgroundColor}} line-height:1.6;">
	<h1 style="margin-bottom: 8px;">Task Request Notification</h1><br>
	<h3>The following task request:</h3><br>
	<div style="margin: 4px;">
		<p><strong>Date:</strong> {{.FmtTaskInfo}}</p>
		<p><strong>Method:</strong> {{.Method}}</p>
		<p><strong>Details:</strong> {{.Details}}</p>
	</div>
	
	<h3>Was created by:</h3><br>
	<div style="margin: 4px;">
		<p><strong>Name:</strong> {{.Name}}</p>
		<p><strong>Company:</strong> {{.Company}}</p>
		<p><strong>Email:</strong> {{.Email}}</p>
		<p><strong>Phone:</strong> {{.Phone}}</p>
	</div>

	{{if .Roles}} 
		<h3>To discuss the following role(s):</h3><br>
		<ul style="list-style-type: none; padding-left: 0;">
			{{range .Roles}}
				<li>{{.}}</li>
			{{end}}
		</ul>
	{{end}}
	</body>
	</html>
	`
	txtNotification, err := txtNotification.Parse(tmplateTxt)
	if err != nil {
		return nil, nil, err
	}

	htmlNotification, err = htmlNotification.Parse(tmplateHtml)
	if err != nil {
		return nil, nil, err
	}

	return txtNotification, htmlNotification, err
}

func NewThanksNotificationTmpl(alertInfo adapters.ThanksAlertDTO) (TaskNotificationTxtTmpl, TaskNotificationHtmlTmpl, error) {
	txtNoti := template.New(fmt.Sprintf("%v%v", THKTXT_IDENTIFIER, time.Now().UnixMilli()))
	htmlNoti := htemplate.New(fmt.Sprintf("%v%v", THKHTML_IDENTIFIER, time.Now().UnixMilli()))

}
