package enums

type Enum interface {
	String() string
	Index() int
}

type WsConnCode int

const (
	NormalClosure = iota + 1000
	GoingAway
	ProtocolErr
	UnsupportedData
	UnsupportedPayload = 1007
	PolicyViolation    = 1008
	CloseToLarge       = 1009
	ServerError        = 1011
	ServerRestart      = 1012
	TryAgainLater      = 1013
	BadGateway         = 1014
	CacheError         = 4000
	NoCacheErr         = 4001
	DatabaseError      = 4002
	JsonDecodeError    = 4003
	JsonEncodeError    = 4004
	Timeout            = 4005
	TypeCastErr        = 4006
	TimeParseErr       = 4007
	MissingData        = 4008
	PermissionErr      = 4009
)

var wsConnCodeStrings = map[WsConnCode]string{
	NormalClosure:      "Normal Closure",
	GoingAway:          "Going Away",
	ProtocolErr:        "Protocol Error",
	UnsupportedData:    "Unsupported",
	UnsupportedPayload: "Unsupported Payload",
	PolicyViolation:    "Policy Violation",
	CloseToLarge:       "Close To Large",
	ServerError:        "Sever Error",
	ServerRestart:      "Server Restart",
	TryAgainLater:      "Try Again Later",
	BadGateway:         "Bad Gateway",
	CacheError:         "Cache Error",
	NoCacheErr:         "No Cache Results",
	DatabaseError:      "Database Error",
	JsonDecodeError:    "Json Decode Error",
	JsonEncodeError:    "Json Encode Error",
	Timeout:            "Timeout",
	TypeCastErr:        "Type Cast Error",
	TimeParseErr:       "Time Parse Error",
	MissingData:        "Missing Data",
	PermissionErr:      "Permission Error",
}

func (w WsConnCode) String() string {
	return wsConnCodeStrings[w]
}

func (w WsConnCode) Index() int {
	return int(w)
}

type ZemailType int

const (
	TaskCreated = iota
	TaskEdit
	TaskDelete
	ThankYou
)

func (zt ZemailType) String() string {
	return [...]string{"Task Created", "Task Edited", "Task Deleteed", "Thank You"}[zt]
}

func (zt ZemailType) Index() int {
	return int(zt)
}

type TaskType int

const (
	Phone = iota
	InPerson
	MicrosoftTeams
	Zoom
	GoogleMeets
)

func (tt TaskType) String() string {
	return [...]string{"Phone", "In Person", "Microsoft Teams", "Zoom", "Google Meets"}[tt]
}

func (tt TaskType) Index() int {
	return int(tt)
}

type PeriodType int

const (
	Weekly = iota + 1
	BiWeekly
	Monthly
)

func (p PeriodType) String() string {
	return [...]string{"Weekly", "BiWeekly", "Monthly"}[p-1]
}

func (p PeriodType) Eindex() int {
	return int(p)
}

type DbOp int

const (
	Read DbOp = iota
	Insert
	Update
	Upsert
	Delete
)

func (d DbOp) String() string {
	return [...]string{"Read", "Insert", "Update", "Upsert", "Delete"}[d]
}

func (d DbOp) Index() int {
	return int(d)
}
