package utils

import (
	"fmt"

	"github.com/Z3DRP/zportfolio-service/internal/dtos"
	"github.com/gorilla/websocket"
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
	return fmt.Sprintf("%v %v operation failed :: %v", d.Operation, d.Store, d.Err)
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

type InvalidDataErr struct {
	FieldName    string
	ExpectedType interface{}
	RecievedType interface{}
	Err          error
}

func NewInvalidDataErr(fname string, expectedType, recievedType interface{}, e error) InvalidDataErr {
	return InvalidDataErr{
		FieldName:    fname,
		ExpectedType: expectedType,
		RecievedType: recievedType,
		Err:          e,
	}
}

func (i InvalidDataErr) Error() string {
	return fmt.Sprintf("invalid data, expected type: [%T] recieved type: [%T] for %v", i.ExpectedType, i.RecievedType, i.FieldName)
}

func (i InvalidDataErr) Unwrap() error {
	return i.Err
}

type MissingDataErr struct {
	FieldName    string
	ExpectedType string
	Err          error
}

func NewMissingDataErr(fldname, expctedType string, e error) MissingDataErr {
	return MissingDataErr{
		FieldName:    fldname,
		ExpectedType: expctedType,
		Err:          e,
	}
}

func (m MissingDataErr) Error() string {
	return fmt.Sprintf("error missing %v, expected %v", m.FieldName, m.ExpectedType)
}

func (m MissingDataErr) Unwrap() error {
	return m.Err
}

type UserNotFoundErr struct {
	Identifier string
	Err        error
}

func (u UserNotFoundErr) Error() string {
	return fmt.Sprintf("user %v not found :: %v", u.Identifier, u.Err)
}

func (u UserNotFoundErr) Unwrap() error {
	return u.Err
}

func NewUserNotFoundErr(id string, e error) UserNotFoundErr {
	return UserNotFoundErr{Identifier: id, Err: e}
}

type InvalidOperationErr struct {
	Operation   string
	Description string
	Err         error
}

func (i InvalidOperationErr) Error() string {
	return fmt.Sprintf("error %v not allowed, %v", i.Operation, i.Description)
}

func (i InvalidOperationErr) Unwrap() error {
	return i.Err
}

func NewInvalidOperationErr(op, desrp string, e error) InvalidOperationErr {
	return InvalidOperationErr{
		Operation:   op,
		Description: desrp,
		Err:         e,
	}
}

type FailedToSendErr struct {
	Conn *websocket.Conn
	Err  error
}

func (f FailedToSendErr) Error() string {
	return fmt.Sprintf("failed to send err message for conn, %v :: %v", f.Conn, f.Err)
}

func (f FailedToSendErr) Unwrap() error {
	return f.Err
}

func NewFailedToSendErr(c *websocket.Conn, e error) FailedToSendErr {
	return FailedToSendErr{Conn: c, Err: e}
}

type NoCacheResultErr struct {
	Obj        string
	Identifier string
	Err        error
}

func (nc NoCacheResultErr) Error() string {
	return fmt.Sprintf("no cache results for %v: %v", nc.Obj, nc.Identifier)
}

func (nc NoCacheResultErr) Unwrap() error {
	return nc.Err
}

func NewNoCacheResultErr(obj, idnfr string, e error) NoCacheResultErr {
	return NoCacheResultErr{
		Obj:        obj,
		Identifier: idnfr,
		Err:        e,
	}
}

type PermissionErr struct {
	Action      string
	Description string
	Err         error
}

func (p PermissionErr) Error() string {
	return fmt.Sprintf("invalid operation: %v, %v", p.Action, p.Description)
}

func (p PermissionErr) Unwrap() error {
	return p.Err
}

func NewPermissionErr(act, desc string, e error) PermissionErr {
	return PermissionErr{
		Action:      act,
		Description: desc,
		Err:         e,
	}
}

type WbsConnClosedErr struct {
	Conn *websocket.Conn
	Err  error
}

func (c WbsConnClosedErr) Error() string {
	return fmt.Sprintf("websocket connection is closed; connection: %v", c.Conn.RemoteAddr())
}

func (c WbsConnClosedErr) Unwrap() error {
	return c.Err
}

func NewWbsConnClosedErr(c *websocket.Conn, e error) WbsConnClosedErr {
	return WbsConnClosedErr{
		Conn: c,
		Err:  e,
	}
}

type InvalidUserErr struct {
	Id  string
	Err error
}

func (i InvalidUserErr) Error() string {
	return fmt.Sprintf("invalid user :%v does not exist :: %v", i.Id, i.Err)
}

func (i InvalidUserErr) Unwrap() error {
	return i.Err
}

func NewInvalidUser(id string, e error) InvalidUserErr {
	return InvalidUserErr{
		Id:  id,
		Err: e,
	}
}

type ErrFailedSendEventResponse struct {
	EventType string
	Err       error
	Conn      *websocket.Conn
}

func (e ErrFailedSendEventResponse) Error() string {
	return fmt.Sprintf("failed to send %v on %v", e.EventType, e.Conn)
}

func (e ErrFailedSendEventResponse) Unwrap() error {
	return e.Err
}

func NewFailedSendEventResponse(conn *websocket.Conn, evntType string, e error) *ErrFailedSendEventResponse {
	return &ErrFailedSendEventResponse{
		EventType: evntType,
		Err:       e,
		Conn:      conn,
	}
}

type ErrZypherFailure struct {
	Process string
	Err     error
}

func (e ErrZypherFailure) Error() string {
	return fmt.Sprintf("failed to calculate zypher for %v", e.Process)
}

func (e ErrZypherFailure) Unwrap() error {
	return e.Err
}

func NewZypherFailureErr(prcss string, e error) *ErrZypherFailure {
	return &ErrZypherFailure{
		Process: prcss,
		Err:     e,
	}
}
