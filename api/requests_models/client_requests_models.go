package requests_models

import "bhc_test_task/model/data_model"

type CreateClientReqModel struct {
	ID           uint64 `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	WorkingHours uint16 `json:"working_hours,omitempty"`
	Priority     uint8  `json:"priority,omitempty"`
	LeadCap      uint16 `json:"lead_capacity,omitempty"`
}

type FindClientForLeadReqModel struct {
	IssueLead bool `json:"issue_lead"`
}

func ConvertDataClientToAPIClient(client data_model.Client) CreateClientReqModel {
	return CreateClientReqModel{
		ID:           client.ID,
		Name:         client.Name,
		WorkingHours: client.WorkingHours,
		Priority:     client.Priority,
		LeadCap:      client.LeadCap,
	}
}
