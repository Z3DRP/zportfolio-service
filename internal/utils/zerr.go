package utils

import (
	"fmt"
)

type ErrJsonDecode struct {
	RecievedType interface{}
	Err          error
}

func (e ErrJsonDecode) Error() string {
	return fmt.Sprintf("could not decode: %v of type: %T into json", e.RecievedType, e.RecievedType)
}

func (e ErrJsonDecode) Unwrap() error {
	return e.Err
}

func NewJsonEncodeErr(rtype interface{}, e error) ErrJsonEncode {
	return ErrJsonEncode{
		RecievedType: rtype,
		Err:          e,
	}
}

type ErrJsonEncode struct {
	RecievedType interface{}
	Err          error
}

func (e ErrJsonEncode) Error() string {
	return fmt.Sprintf("could not encode: %v of type: %T into json", e.RecievedType, e.RecievedType)
}

func (e ErrJsonEncode) Unwrap() error {
	return e.Err
}

func NewJsonDecodeErr(rtype interface{}, e error) ErrJsonDecode {
	return ErrJsonDecode{
		RecievedType: rtype,
		Err:          e,
	}
}

type ErrTimeParse struct {
	Value   string
	VarName string
	Err     error
}

func NewTimeParseErr(val, vnam string, e error) ErrTimeParse {
	return ErrTimeParse{
		Value:   val,
		VarName: vnam,
		Err:     e,
	}
}

func (e ErrTimeParse) Error() string {
	return fmt.Sprintf("error could not parse time from: %v for %v", e.Value, e.VarName)
}

func (e ErrTimeParse) Unwrap() error {
	return e.Err
}

type ErrCacheOp struct {
	Operation string
	Object    string
	Err       error
}

func (e ErrCacheOp) Error() string {
	return fmt.Sprintf("an error occurred while %v %v cache:: %v", e.Operation, e.Object, e.Err)
}

func (e ErrCacheOp) Unwrap() error {
	return e.Err
}

func NewCacheOpErr(op, obj string, e error) ErrCacheOp {
	return ErrCacheOp{
		Operation: op,
		Object:    obj,
		Err:       e,
	}
}

type ErrFetchRecords struct {
	RecordType string
	Msg        string
	Err        error
}

func (e ErrFetchRecords) Error() string {
	return fmt.Sprintf("could not retrieve %v for %v", e.RecordType, e.Msg)
}

func (e ErrFetchRecords) Unwrap() error {
	return e.Err
}

type ErrTypeCast struct {
	RecievedType interface{}
	ExpectedType interface{}
	Err          error
}

func (e ErrTypeCast) Error() string {
	return fmt.Sprintf("could not cast type: [%T] into type: [%T] ", e.RecievedType, e.ExpectedType)
}

func (e ErrTypeCast) Unwrap() error {
	return e.Err
}

func NewTypeCastErr(rtype, etype interface{}, e error) ErrTypeCast {
	return ErrTypeCast{
		RecievedType: rtype,
		ExpectedType: etype,
		Err:          e,
	}
}

type ErrWebsocketConnection struct {
	Operation string
	Err       error
}

func (e ErrWebsocketConnection) Error() string {
	return fmt.Sprintf("could not %v", e.Operation)
}

func (e ErrWebsocketConnection) Unwrap() error {
	return e.Err
}

func NewWebSocketErr(operation string, e error) ErrWebsocketConnection {
	return ErrWebsocketConnection{
		Operation: operation,
		Err:       e,
	}
}

type ErrTimeout struct {
	Operation string
	Err       error
}

func (e ErrTimeout) Error() string {
	return fmt.Sprintf("%v", e.Operation)
}

func (e ErrTimeout) Unwrap() error {
	return e.Err
}

func NewTimeoutErr(op string, e error) ErrTimeout {
	return ErrTimeout{
		Operation: op,
		Err:       e,
	}
}

type ErrConfigFile struct {
	ConfigType string
	Err        error
}

func (e ErrConfigFile) Error() string {
	return fmt.Sprintf("an error occurred while reading %v :: %v", e.ConfigType, e.Err)
}

func (e ErrConfigFile) Unwrap() error {
	return e.Err
}

func NewConfigFileErr(ctype string, e error) ErrConfigFile {
	return ErrConfigFile{
		ConfigType: ctype,
		Err:        e,
	}
}

type IdGeneratorErr struct {
	Object string
	Err    error
}

func (e IdGeneratorErr) Error() string {
	return fmt.Sprintf("could not generate %v id :: %v", e.Object, e.Err)
}

func (e IdGeneratorErr) Unwrap() error {
	return e.Err
}

func NewIdGenErr(obj string, e error) IdGeneratorErr {
	return IdGeneratorErr{
		Object: obj,
		Err:    e,
	}
}

type DbErr struct {
	Operation string
	Store     string
	Err       error
}

func (d DbErr) Error() string {
	return fmt.Sprintf("an error occurred while %v %v :: %v", d.Operation, d.Store, d.Err)
}

func (d DbErr) Unwrap() error {
	return d.Err
}

func NewDbErr(op, store string, e error) DbErr {
	return DbErr{
		Operation: op,
		Store:     store,
		Err:       e,
	}
}

type NotificationFailedErr struct {
	Err              error
	NotificationData string
	UserData         string
}

func (n NotificationFailedErr) Error() string {
	return fmt.Sprintf("failed to send notification: NotificationData: %v, UserData: %v :: %v", n.NotificationData, n.UserData, n.Err)
}

func (n NotificationFailedErr) Unwrap() error {
	return n.Err
}

func NewNotificationErr(notiData, usrData string, e error) NotificationFailedErr {
	return NotificationFailedErr{
		Err:              e,
		NotificationData: notiData,
		UserData:         usrData,
	}
}
