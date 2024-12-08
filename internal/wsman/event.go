package wsman

import (
	"context"
	"encoding/json"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(ctx context.Context, clnt *Client, evnt Event) error

const (
	EventFetchSchedule     = "fetch_schedule"
	EventBroadcastSchedule = "broadcast_schedule"
	EventCreateTask        = "create_task"
	EventUpdateTask        = "update_task"
	EventRemoveTask        = "remove_task"
	EventInsertResponse    = "insert_response"
	EventUpdateResponse    = "update_response"
	EventRemovedResponse   = "remove_response"
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
