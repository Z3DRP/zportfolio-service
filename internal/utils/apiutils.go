package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"

	adp "github.com/Z3DRP/zportfolio-service/internal/adapters"
	"github.com/Z3DRP/zportfolio-service/internal/dtos"
	"github.com/Z3DRP/zportfolio-service/internal/models"
)

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func GenerateTID() (string, error) {
	bid, err := generateID(12, 9)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("TSK-%v", bid), nil
}

func generateID(length int, mx int64) (string, error) {
	max := big.NewInt(mx)
	ilen := length
	id := ""
	errors := make([]error, 0)

	for i := 0; i <= ilen; i++ {
		n, err := rand.Int(rand.Reader, max)

		if err != nil {
			errors = append(errors, err)
		}

		id += fmt.Sprintf("%v", n)
	}

	if len(errors) > 0 {
		return "", fmt.Errorf("error occurred while generating id:: %w", errors[0])
	}

	return id, nil
}

func GenToken() ([]byte, error) {
	ilen := 32
	asci := make([]byte, ilen)
	_, err := rand.Read(asci)
	if err != nil {
		return nil, fmt.Errorf("error occurred while generating id:: %w", err)
	}
	return asci, nil
}

func ConvertTaskRequest(task models.Task, tr dtos.TaskRequestDTO) (adp.TaskData, adp.UserData, adp.EmailInfo, adp.Customizations) {
	tskData := adp.NewTaskData(task)
	usrData := adp.NewUserData(tr.UsrName, tr.Company, tr.Email, tr.Phone, tr.Roles)
	emlInfo := adp.NewEmailInfo(tr.Cc, tr.Body, tr.UseHtml)
	custInfo := adp.NewCustomizations()
	return tskData, usrData, emlInfo, *custInfo
}
