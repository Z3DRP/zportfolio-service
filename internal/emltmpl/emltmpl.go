package emltmpl

import (
	"errors"
	"fmt"
	htemplate "html/template"
	"text/template"
	"time"

	"github.com/Z3DRP/zportfolio-service/enums"
	"github.com/Z3DRP/zportfolio-service/internal/dtos"
)

const TSKTXT_IDENTIFIER = "tsk-alert-txt"
const TSKHTML_IDENTIFIER = "tsk-alert-html"
const THKTXT_IDENTIFIER = "thk-alert-txt"
const THKHTML_IDENTIFIER = "thk-alert-html"

type TxtNotificationTemplate *template.Template
type HtmlNotificationTemplate *htemplate.Template

type ErrEmailTypeErr struct {
	ExpectedEmailType enums.ZemailType
	ActualEmailType   enums.ZemailType
}

func (emt ErrEmailTypeErr) Error() string {
	return fmt.Sprintf("invalid error template type, expected: %v, got: %v", emt.ExpectedEmailType.String(), emt.ActualEmailType.String())
}

func NewErrEmailType(expType, actType enums.ZemailType) ErrEmailTypeErr {
	return ErrEmailTypeErr{
		ExpectedEmailType: expType,
		ActualEmailType:   actType,
	}
}

// TODO add customizations for colors

func NewTaskAlertTmpl(alertInfo dtos.TaskRequestAlertDTO) (TxtNotificationTemplate, HtmlNotificationTemplate, error) {
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
	<body style="font-family: sans-serif, Arial; color:{{.TextColor}} background-color:{{.BackgroundColor}} line-height:1.6;">
	{{if .Image}}
	<header>
		<img src="data:image/png;base64,{{.Image}}" alt="Banner image" style="width:100%; height:auto" />
	</header>
	{{end}}
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

func NewThanksNotificationTmpl(alertInfo dtos.ThanksAlertDTO) (TxtNotificationTemplate, HtmlNotificationTemplate, error) {
	txtNoti := template.New(fmt.Sprintf("%v%v", THKTXT_IDENTIFIER, time.Now().UnixMilli()))
	htmlNoti := htemplate.New(fmt.Sprintf("%v%v", THKHTML_IDENTIFIER, time.Now().UnixMilli()))

	txtTmpl := `
		Hello {{.Name}},\n
		Thank you for requesting a time to speak with me about upcoming opportunities offered or filled by your company.
		If you need a updated copy of my resume one can be found at zachpalmer.dev/resume.\n
		I am looking forward to speaking with you.\n\n

		Best regards,\n
		Zach Palmer
	`

	txtNoti, err := txtNoti.Parse(txtTmpl)
	if err != nil {
		return nil, nil, err
	}

	htmlTmpl := `
	<!DOCTYPE html>
	<html>
		<head>
			<meta charset="UTF-8">
			<title>Task Request Notification</title>
		</head>
		<body style="font-family: sans-serif, Arial; color: {{.TextColor}}; background-color: {{.BackgroundColor}}; line-height:1.6;">
			{{if .Image}}
				<header>
					<img src="data:image/png;base64,{{.Image}}" alt="Banner image" style="width:100%; height:auto" />
				</header>
			{{end}}	
			<h2 style="margin: 2px;">Hello {{.Name}},</h2><br>
			<p>
				Thank you for requesting a time to speak with me about upcoming opportunities offered or filled by your company.
				If you need a updated copy of my resume one can be found at <strong>zachpalmer.dev/resume.</strong>
			</p>
			<p>
				I am looking forward to speaking with you.
			</p>

			<h5 style="margin-top: 2px;">Best regards,</h5>
			<h5>Zach Palmer</h5>
		</body>
	</html>
	`

	htmlNoti, err = htmlNoti.Parse(htmlTmpl)
	if err != nil {
		return nil, nil, err
	}

	return txtNoti, htmlNoti, nil
}

// TODO on this templateFacotry method handle the errors for the casting from DTOers to the structs

func TemplateFactory(emailData dtos.AlertDTOer) (TxtNotificationTemplate, HtmlNotificationTemplate, error) {
	switch emailData.AlertTypeString() {
	case "Task Request":
		if data, ok := emailData.(dtos.TaskRequestAlertDTO); ok {
			return NewTaskAlertTmpl(data)
		}
		return nil, nil, NewErrEmailType(enums.ZemailType(0), enums.ZemailType(emailData.AlertType()))
	case "Task Edit":
		return nil, nil, errors.ErrUnsupported
	case "Task Delete":
		return nil, nil, errors.ErrUnsupported
	case "Thank You":
		if data, ok := emailData.(dtos.ThanksAlertDTO); ok {
			return NewThanksNotificationTmpl(data)
		}
		return nil, nil, NewErrEmailType(enums.ZemailType(3), enums.ZemailType(emailData.AlertType()))
	default:
		return nil, nil, errors.New("invalid template type")
	}
}
