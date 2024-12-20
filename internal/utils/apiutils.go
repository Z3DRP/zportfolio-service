package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Z3DRP/zportfolio-service/enums"
	"github.com/Z3DRP/zportfolio-service/internal/dtos"
	"github.com/Z3DRP/zportfolio-service/internal/zlogger"
	"github.com/gorilla/websocket"
)

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func GenerateID(idType string) (string, error) {
	var prefix string
	switch strings.ToLower(idType) {
	case "task":
		prefix = "TSK"
	case "user":
		prefix = "USR"
	case "availability":
		prefix = "AVB"
	default:
		prefix = ""
	}
	id, err := createID(12, 9)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v-%v", prefix, id), nil
}

func createID(length int, mx int64) (string, error) {
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

func SendMessage[P dtos.Payloader](conn *websocket.Conn, msg P) error {
	err := conn.WriteJSON(msg)
	if err != nil {
		return err
	}
	return nil
}

func MustSendMessage[P dtos.Payloader](conn *websocket.Conn, msg dtos.EventDto[P]) error {
	err := conn.WriteJSON(msg)
	if err != nil {
		if IsConnClosed(err) {
			return NewWbsConnClosedErr(conn, err)
		}
		_ = WriteCloseMessage(conn, err, enums.ServerError)
		conn.Close()
		return NewFailedSendEventResponse(conn, msg.Type, err)
	}
	return nil
}

func MustSendErrMessage(conn *websocket.Conn, e error, conCode int) error {
	emsg := dtos.SocketErrMsg{
		ErrMsg:      e.Error(),
		ConnCode:    conCode,
		CodeMessage: enums.WsConnCode(conCode).String(),
	}

	err := conn.WriteJSON(emsg)
	if err != nil {
		if IsConnClosed(err) {
			return NewWbsConnClosedErr(conn, err)
		}
		_ = WriteCloseMessage(conn, err, enums.ServerError)
		conn.Close()
		return NewFailedToSendErr(conn, err)
	}
	return nil
}

func WriteCloseMessage(conn *websocket.Conn, e error, conCode int) error {
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(conCode, e.Error()))
	if err != nil {
		return err
	}
	return nil
}

func WriteMessage(conn *websocket.Conn, msg string) error {
	err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		return err
	}
	return nil
}

func LogError(logger *zlogger.Zlogrus, e error, lvl zlogger.LogLevel) {
	msg := fmt.Sprintf("%v", e)
	switch lvl {
	case zlogger.Trace:
		logger.MustTrace(msg)
	case zlogger.Debug:
		logger.MustDebug(msg)
	case zlogger.Info:
		logger.MustInfo(msg)
	case zlogger.Fatal:
		logger.MustFatal(msg)
	case zlogger.Error:
		logger.MustError(msg)
	case zlogger.Panic:
		logger.MustPanic(msg)
	}
}

func WriteLog(logger *zlogger.Zlogrus, s string, lvl zlogger.LogLevel) {
	msg := fmt.Sprintf("%v", s)
	switch lvl {
	case zlogger.Trace:
		logger.MustTrace(msg)
	case zlogger.Debug:
		logger.MustDebug(msg)
	case zlogger.Info:
		logger.MustInfo(msg)
	case zlogger.Fatal:
		logger.MustFatal(msg)
	case zlogger.Error:
		logger.MustError(msg)
	case zlogger.Panic:
		logger.MustPanic(msg)
	}
}

func JoinBy(s []string, seperator string) string {
	return strings.Join(s, seperator)
}

func SplitBy(s, delmtr string) []string {
	return strings.Split(s, delmtr)
}

func IsConnClosed(e error) bool {
	return errors.Is(e, io.EOF) || errors.Is(e, net.ErrClosed)
}

type InvalidEventErr struct {
	Event string
}

func (i InvalidEventErr) Error() string {
	return fmt.Sprintf("error unknown event type: %v", i.Event)
}

func NewInvalidEventErr(evnt string) InvalidEventErr {
	return InvalidEventErr{
		Event: evnt,
	}
}

func IsGreaterThanOrEqual(firstDate, secondDate time.Time) bool {
	return firstDate.UTC().Compare(secondDate) >= 0
}

func IsLessThanOrEqual(firstDate, secondDate time.Time) bool {
	return firstDate.UTC().Compare(secondDate) <= 0
}

func IsGreaterThan(firstDate, secondDate time.Time) bool {
	return firstDate.UTC().Compare(secondDate) == 1
}

func IsLessThan(firstDate, secondDate time.Time) bool {
	return firstDate.UTC().Compare(secondDate) == -1
}

func IsEqual(firstDate, secondDate time.Time) bool {
	return firstDate.UTC().Compare(secondDate) == 0
}

func IsBetweenInclusive(targetDate, startDate, endDate time.Time) bool {
	return IsGreaterThanOrEqual(targetDate, startDate) && IsLessThanOrEqual(targetDate, endDate)
}

func IsBetweenExclusive(targetDate, startDate, endDate time.Time) bool {
	return IsGreaterThan(targetDate, startDate) && IsLessThan(targetDate, endDate)
}

func IsInRange(targetStart, targetEnd, rangeStart, rangeEnd time.Time) bool {
	return IsBetweenInclusive(targetStart, rangeStart, rangeEnd) && IsBetweenExclusive(targetEnd, rangeStart, rangeEnd)
}
