package enums

type ZemailType int

const (
	TaskRequest = iota
	TaskEdit
	TaskDelete
	ThankYou
)

func (zt ZemailType) String() string {
	return [...]string{"Task Request", "Thank You", "TaskEdit", "Task Delete"}[zt]
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
