package tests

import (
	"bhc_test_task/api/handler"
	"bhc_test_task/api/requests_models"
	"bhc_test_task/api/response_models"
	"bhc_test_task/manager"
	"bhc_test_task/model/data_storage"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"testing"
)

const numOfClients = 4

var (
	dataStorage    = data_storage.NewLocalDataStorage()
	clientsManager = manager.NewClientsManager(dataStorage)
)

func TestCanAddClients(t *testing.T) {
	clientsIDs := make([]uint64, numOfClients)
	for i := 0; i < numOfClients; i++ {
		client := requests_models.CreateClientReqModel{
			Name:         fmt.Sprintf("client_%d", i+1),
			WorkingHours: uint16(rand.Intn(manager.MaxWorkingHours-manager.MinWorkingHours) + manager.MinWorkingHours),
			Priority:     uint8(rand.Intn(manager.MaxPriority-manager.MinPriority) + manager.MinPriority),
			LeadCap:      uint16(rand.Intn(manager.MaxLeadCapacity-manager.MinLeadCapacity) + manager.MinLeadCapacity),
		}

		clientID, err := clientsManager.CreateClient(client.Name, client.WorkingHours, client.Priority, client.LeadCap)
		if err != nil {
			t.Fatalf("failed to create a client %d; Error: %v", i, err)
		}

		clientsIDs[i] = clientID
	}

	clients, err := clientsManager.GetClientsInfo()
	if err != nil {
		t.Fatalf("failed to get a list of clients after the test. Error: %s", err)
	}

outerLoop:
	for _, clientID := range clientsIDs {
		for _, client := range clients {
			if client.ID == clientID {
				continue outerLoop
			}
		}

		t.Errorf("cannot find client with id %d after it was added", clientID)
	}

}

func TestCanFindBestClientForLead(t *testing.T) {
	clients := []requests_models.CreateClientReqModel{
		{
			Name:         "loser1",
			WorkingHours: manager.MinWorkingHours + 1,
			Priority:     manager.MinPriority + 1,
			LeadCap:      manager.MaxLeadCapacity - 1,
		},
		{
			Name:         "loser1",
			WorkingHours: manager.MinWorkingHours + 2,
			Priority:     manager.MinPriority + 2,
			LeadCap:      manager.MaxLeadCapacity - 2,
		},
		{
			Name:         "loser3",
			WorkingHours: manager.MinWorkingHours + 3,
			Priority:     manager.MinPriority + 3,
			LeadCap:      manager.MaxLeadCapacity - 3,
		},
		{
			Name:         "winner",
			WorkingHours: manager.MaxWorkingHours,
			Priority:     manager.MaxPriority,
			LeadCap:      manager.MinLeadCapacity,
		},
	}

	for i, client := range clients {
		_, err := clientsManager.CreateClient(client.Name, client.WorkingHours, client.Priority, client.LeadCap)
		if err != nil {
			t.Fatalf("failed to create client %d. Error: %s", i, err)
		}
	}

	winner, err := clientsManager.FindClientForLead(true)
	if err != nil {
		t.Fatalf("failed to find client for the lead. Error: %s", err)
	}

	if winner.Name != "winner" {
		t.Fatalf("winner's name is %s. Expected to be %s", winner.Name, "winner")
	}

	if winner.GotLeads != 1 {
		t.Fatalf("winner got %d leads. Expected to be %v", winner.GotLeads, 1)
	}
}

func TestFullAPI(t *testing.T) {
	handler.SetupHandler(manager.NewClientsManager(data_storage.NewLocalDataStorage()))

	port := "8080"
	serverAddr := "http://localhost:" + port
	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			t.Errorf("failed to start a server. Error: %s", err)
		}
	}()

	// 1. create clients
	{
		clients := []requests_models.CreateClientReqModel{
			{
				Name:         "loser1",
				WorkingHours: manager.MinWorkingHours + 1,
				Priority:     manager.MinPriority + 1,
				LeadCap:      manager.MaxLeadCapacity - 1,
			},
			{
				Name:         "loser1",
				WorkingHours: manager.MinWorkingHours + 2,
				Priority:     manager.MinPriority + 2,
				LeadCap:      manager.MaxLeadCapacity - 2,
			},
			{
				Name:         "loser3",
				WorkingHours: manager.MinWorkingHours + 3,
				Priority:     manager.MinPriority + 3,
				LeadCap:      manager.MaxLeadCapacity - 3,
			},
			{
				Name:         "winner",
				WorkingHours: manager.MaxWorkingHours,
				Priority:     manager.MaxPriority,
				LeadCap:      manager.MinLeadCapacity,
			},
		}

		for i, client := range clients {
			data, err := json.Marshal(client)
			if err != nil {
				t.Fatalf("failed to marshal client #%d. Error: %s", i, err)
			}

			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", serverAddr, handler.CreateClientEndpoint), bytes.NewReader(data))
			if err != nil {
				t.Fatalf("cannot create request for creating users. Error: %s", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("failed to send creating client request for client #%d. Error: %s", i, err)
			}

			var res response_models.ClientReqResponse
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("cannot read response body for client #%d. Error %v", i, err)
			}

			if err := json.Unmarshal(body, &res); err != nil {
				t.Fatalf("failed to unmarshal body for client #%d. Error: %v", i, err)
			}

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("failed to create user. Got status code %d, expected %d. Message: %s. Error: %s", resp.StatusCode, http.StatusOK, res.Message, res.Error)
			}
		}
	}

	// 2. check that clients exist
	{
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", serverAddr, handler.GetClientsInfoEndpoint), nil)
		if err != nil {
			t.Fatalf("cannot create request for getting users info. Error: %s", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to send get users info request for client. Error: %s", err)
		}

		var res response_models.ClientReqResponse
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("cannot read response body after get clients info request. Error %v", err)
		}

		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("failed to unmarshal body after get clients info request. Error: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("failed to get users info. Got status code %d, expected %d. Message: %s. Error: %s", resp.StatusCode, http.StatusOK, res.Message, res.Error)
		}

		if res.Clients == nil || len(res.Clients) < 4 {
			t.Fatalf("got invalid number of clients after all of them have been added. Expected to be %d, actual: %d", 4, len(res.Clients))
		}
	}

	var winnerID uint64
	// 3. find the client for the lead
	{
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", serverAddr, handler.FindClientForLeadEndpoint), nil)
		if err != nil {
			t.Fatalf("cannot create request for finding for lead winner. Error: %s", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to send find lead winner request. Error: %s", err)
		}

		var res response_models.ClientReqResponse
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("cannot read response body after find lead winner request. Error %v", err)
		}

		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("failed to unmarshal body after find lead winner request. Error: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("failed to find lead winner. Got status code %d, expected %d. Message: %s. Error: %s", resp.StatusCode, http.StatusOK, res.Message, res.Error)
		}

		if res.Client.ID == 0 {
			t.Fatal("lead winner was not found, though had to")
		}

		if res.Client.Name != "winner" {
			t.Fatalf("winner has to have name '%s', actually has %s", "winner", res.Client.Name)
		}

		winnerID = res.Client.ID
	}

	// 4. make sure the winner exists (and check getting client info by id while we're here)
	{
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s?id=%d", serverAddr, handler.GetClientInfoEndpoint, winnerID), nil)
		if err != nil {
			t.Fatalf("cannot create request for getting user info. Error: %s", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to send get user info request. Error: %s", err)
		}

		var res response_models.ClientReqResponse
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("cannot read response body after get user info request request. Error %v", err)
		}

		if err := json.Unmarshal(body, &res); err != nil {
			t.Fatalf("failed to unmarshal body after get user info request request. Error: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("failed to get user info request. Got status code %d, expected %d. Message: %s. Error: %s", resp.StatusCode, http.StatusOK, res.Message, res.Error)
		}

		if res.Client.ID == 0 {
			t.Fatal("user was not found by id, though had to")
		}

		if res.Client.Name != "winner" {
			t.Fatalf("winner has to have name '%s', actually has %s", "winner", res.Client.Name)
		}

		if res.Client.ID != winnerID {
			t.Fatalf("winner id that we got with request is not equal to the actual one. Expected id: %d, got: %d", winnerID, res.Client.ID)
		}
	}
}
