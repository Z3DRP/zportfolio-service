package wsman

import "encoding/json"

type Event struct {
	Type    string          `json:"type"`
	Period  string          `json:"period"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event Event, client *Client) error

const (
	EventFetchSchedule = "fetch_schedule"
	EventCreateTask    = "create_task"
	EventUpdateTask    = "update_task"
	EventRemoveTask    = "remove_task"
)

type EvntTaskDelete struct {
	Uid string `json:"uid"`
	Tid string `json:"tid"`
}

type EvntFetchSchedule struct {
	PeriodStart string `json:"periodStart"`
	PeriodEnd   string `json:"periodEnd"`
}

type EvntTaskUpsert struct {
	Start   string `json:"start"`
	End     string `json:"end"`
	Detail  string `json:"detail"`
	Uid     string `json:"uid"`
	UsrName string `json:"userName"`
	Tid     string `json:"tid"`
	Company string `json:"company"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Roles   string `json:"roles"`
}
