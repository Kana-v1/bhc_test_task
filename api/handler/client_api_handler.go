package handler

import (
	"bhc_test_task/api/requests_models"
	"bhc_test_task/api/response_models"
	"bhc_test_task/manager"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	CreateClientEndpoint      = "/client/create"
	GetClientInfoEndpoint     = "/client/info"
	GetClientsInfoEndpoint    = "/clients/info"
	FindClientForLeadEndpoint = "/clients/findForLead"
)

type Handler struct {
	manager *manager.ClientsManager
}

func SetupHandler(manager *manager.ClientsManager) {
	handler := &Handler{manager}

	http.HandleFunc(CreateClientEndpoint, handler.CreateNewClient)        // post/put localhost:2000/client/create
	http.HandleFunc(GetClientInfoEndpoint, handler.GetClientInfo)         // get localhost:2000/client/info&id=123
	http.HandleFunc(GetClientsInfoEndpoint, handler.GetClientsInfo)       // get localhost:2000/clients/info
	http.HandleFunc(FindClientForLeadEndpoint, handler.FindClientForLead) // get/post localhost:2000/clients/findForLead
}

func (handler *Handler) CreateNewClient(w http.ResponseWriter, r *http.Request) {
	var response response_models.ClientReqResponse
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		response.Error = "method is not allowed"
		response.Message = fmt.Sprintf("only %s and %s methods are allowed", http.MethodPost, http.MethodPut)
		writeResponse(http.StatusMethodNotAllowed, response, w)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error = err.Error()
		response.Message = ErrFailedToReadBody.Error()
		writeResponse(http.StatusBadRequest, response, w)
		return
	}

	var createClientReq requests_models.CreateClientReqModel
	if err := json.Unmarshal(body, &createClientReq); err != nil {
		response.Error = err.Error()
		response.Message = ErrFailedToReadBody.Error()
		writeResponse(http.StatusBadRequest, response, w)
		return
	}

	statusCode := http.StatusOK
	response.Message = "Client has been created"

	response.Client.ID, err = handler.manager.CreateClient(
		createClientReq.Name,
		createClientReq.WorkingHours,
		createClientReq.Priority,
		createClientReq.LeadCap,
	)

	if err != nil {
		if errors.Is(err, manager.ErrInvalidInput) {
			statusCode = http.StatusBadRequest
			response.Message = "Invalid input"
		} else {
			statusCode = http.StatusInternalServerError
			response.Message = "Internal error"
		}
		response.Error = err.Error()
	}

	writeResponse(statusCode, response, w)
}

func (handler *Handler) GetClientInfo(w http.ResponseWriter, r *http.Request) {
	var response response_models.ClientReqResponse
	if r.Method != http.MethodGet {
		response.Error = "method is not allowed"
		response.Message = fmt.Sprintf("only %s method is allowed", http.MethodGet)
		writeResponse(http.StatusMethodNotAllowed, response, w)
		return
	}

	clientID, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		response.Message = "Failed to extract clientID"
		response.Error = err.Error()
		writeResponse(http.StatusBadRequest, response, w)
		return
	}

	statusCode := http.StatusOK
	client, err := handler.manager.GetClientInfo(clientID)
	if err != nil {
		response.Error = err.Error()
		response.Message = "Failed to get client info"
		statusCode = http.StatusInternalServerError
		if errors.Is(err, manager.ErrInvalidInput) {
			statusCode = http.StatusBadRequest
		}
	} else {
		response.Client = requests_models.ConvertDataClientToAPIClient(client)
	}

	writeResponse(statusCode, response, w)
}

func (handler *Handler) GetClientsInfo(w http.ResponseWriter, r *http.Request) {
	var response response_models.ClientReqResponse
	if r.Method != http.MethodGet {
		response.Error = "method is not allowed"
		response.Message = fmt.Sprintf("only %s method is allowed", http.MethodGet)
		writeResponse(http.StatusMethodNotAllowed, response, w)
		return
	}

	statusCode := http.StatusOK
	clients, err := handler.manager.GetClientsInfo()
	if err != nil {
		statusCode = http.StatusInternalServerError
		response.Error = err.Error()
		response.Message = "Failed to get clients info"
	} else {
		response.Clients = make([]requests_models.CreateClientReqModel, len(clients))
		for i := range clients {
			response.Clients[i] = requests_models.ConvertDataClientToAPIClient(clients[i])
		}
	}

	writeResponse(statusCode, response, w)
}

func (handler *Handler) FindClientForLead(w http.ResponseWriter, r *http.Request) {
	var response response_models.ClientReqResponse
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		response.Error = "method is not allowed"
		response.Message = fmt.Sprintf("only %s method is allowed", http.MethodGet)
		writeResponse(http.StatusMethodNotAllowed, response, w)
		return
	}

	shouldIssueLead := r.Method == http.MethodGet
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			response.Message = ErrFailedToReadBody.Error()
			response.Error = err.Error()
			writeResponse(http.StatusBadRequest, response, w)
			return
		}

		var reqData requests_models.FindClientForLeadReqModel
		if err := json.Unmarshal(body, &reqData); err != nil {
			response.Message = ErrFailedToReadBody.Error()
			response.Error = err.Error()
			writeResponse(http.StatusBadRequest, response, w)
		}

		shouldIssueLead = reqData.IssueLead
	}

	response.Message = "Client for the lead has been found"
	statusCode := http.StatusOK

	client, err := handler.manager.FindClientForLead(shouldIssueLead)
	if err != nil {
		response.Message = "Failed to find client to issue the lead to"

		if errors.Is(err, manager.ErrDataLayer) {
			statusCode = http.StatusInternalServerError
			response.Error = "Internal error"
		} else {
			statusCode = http.StatusBadRequest
			response.Error = err.Error()
		}
	}

	response.Client = requests_models.ConvertDataClientToAPIClient(client)
	writeResponse(statusCode, response, w)
}

func writeResponse(statusCode int, response any, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	data, err := json.Marshal(response)
	if err != nil {
		data = []byte("Failed to marshal response")
	}

	w.Write(data)
}
