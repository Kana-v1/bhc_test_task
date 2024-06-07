package data_model

import (
	"sync/atomic"
	"time"
)

const ClientTTL = time.Hour * 24 * 7 // 1 week

var userID atomic.Uint64

type Client struct {
	ID           uint64    `json:"id,omitempty"`
	Name         string    `json:"name"`
	WorkingHours uint16    `json:"working_hours"` // working hours per ClientTTL
	Priority     uint8     `json:"priority"`
	LeadCap      uint16    `json:"lead_capacity"` // max amount of leads per ClientTTL
	GotLeads     uint16    `json:"-"`             // got leads during the last ClientTTL
	GotLeadAt    time.Time `json:"-"`             // time the last lead was received
}

func NewClient(name string, workingHours uint16, priority uint8, leadCap uint16) Client {
	return Client{
		ID:           userID.Add(1),
		Name:         name,
		WorkingHours: workingHours,
		Priority:     priority,
		LeadCap:      leadCap,
		GotLeadAt:    time.Now().UTC(),
	}
}
