package controller

import (
	"fmt"
	"text/template"

	"github.com/Z3DRP/zportfolio-service/internal/eml"
)

func SendTaskRequestNotificationEmail(taskData map[string]string) error {
	emailer, err := eml.Build()
	if err != nil {
		return fmt.Errorf("could not build email client:: %v", err)
	}
	t := template.New("t")

}
