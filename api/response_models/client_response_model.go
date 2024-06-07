package response_models

import (
	"bhc_test_task/api/requests_models"
)

type ClientReqResponse struct {
	Message string                                 `json:"message"`
	Error   string                                 `json:"error,omitempty"`
	Client  requests_models.CreateClientReqModel   `json:"client,omitempty"`
	Clients []requests_models.CreateClientReqModel `json:"clients,omitempty"`
}

