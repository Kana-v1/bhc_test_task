package manager

import (
	"bhc_test_task/model/data_model"
	"bhc_test_task/model/data_storage"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	MaxWorkingHours = 160
	MinWorkingHours = 10

	MaxPriority = 10
	MinPriority = 1

	MaxLeadCapacity = 100
	MinLeadCapacity = 1
)

type ClientsManager struct {
	clientSearchMutex *sync.Mutex // used while searching for client for the lead
	dataStorage       data_storage.DataStorage
}

func NewClientsManager(dataStorage data_storage.DataStorage) *ClientsManager {
	return &ClientsManager{
		clientSearchMutex: new(sync.Mutex),
		dataStorage:       dataStorage,
	}
}

func (manager *ClientsManager) CreateClient(name string, workingHours uint16, priority uint8, leadCap uint16) (uint64, error) {
	if workingHours > MaxWorkingHours || workingHours < MinWorkingHours {
		return 0, errors.Wrapf(ErrInvalidInput, "working hours(%d) beyond the limits [%d; %d]", workingHours, MinWorkingHours, MaxWorkingHours)
	}

	if priority > MaxPriority || priority < MinPriority {
		return 0, errors.Wrapf(ErrInvalidInput, "priority(%d) beyond the limits [%d; %d]", priority, MinPriority, MaxPriority)
	}

	if leadCap > MaxLeadCapacity || leadCap < MinLeadCapacity {
		return 0, errors.Wrapf(ErrInvalidInput, "leadCap(%d) beyond the limits [%d; %d]", leadCap, MinLeadCapacity, MaxLeadCapacity)
	}

	clientID, err := manager.dataStorage.CreateClient(name, workingHours, priority, leadCap)
	if err != nil && errors.Is(err, data_storage.ErrClientAlreadyExists) {
		return 0, errors.Wrap(ErrInvalidInput, err.Error())
	}

	return clientID, err
}

func (manager *ClientsManager) GetClientsInfo() ([]data_model.Client, error) {
	return manager.dataStorage.GetClientsInfo()
}

func (manager *ClientsManager) GetClientInfo(id uint64) (data_model.Client, error) {
	client, err := manager.dataStorage.GetClientInfo(id)
	if err != nil && errors.Is(err, data_storage.ErrClientDoesNotExist) {
		return client, errors.Wrap(ErrInvalidInput, err.Error())
	}

	return client, err
}

// every new request should search for the client or 2 reqs simultaneously should get the same result?
func (manager *ClientsManager) FindClientForLead(issueLead bool) (data_model.Client, error) {
	manager.clientSearchMutex.Lock()
	defer manager.clientSearchMutex.Unlock()

	if err := manager.resetSpoiledLeadCaps(); err != nil {
		return data_model.Client{}, errors.Wrap(ErrDataLayer, err.Error())
	}

	winner, err := manager.findClientForLeadAlg()
	if err != nil {
		return data_model.Client{}, err
	}

	if issueLead {
		winner.GotLeads += 1
		if err := manager.dataStorage.UpdateClients(winner); err != nil {
			return winner, errors.Wrapf(ErrDataLayer, "failed to issue a lead. Error: %s", err.Error())
		}
	}

	return winner, nil
}

func (manager *ClientsManager) resetSpoiledLeadCaps() error {
	clients, err := manager.dataStorage.GetClientsInfo()
	if err != nil {
		return errors.Wrap(ErrDataLayer, err.Error())
	}

	clientsToUpdate := make([]data_model.Client, 0, len(clients))
	timeNow := time.Now().UTC()
	for _, client := range clients {
		if timeNow.After(client.GotLeadAt.Add(data_model.ClientTTL)) {
			client.GotLeads = 0
			client.GotLeadAt = timeNow
			clientsToUpdate = append(clientsToUpdate, client)
		}
	}

	return manager.dataStorage.UpdateClients(clientsToUpdate...)
}

func (manager *ClientsManager) findClientForLeadAlg() (data_model.Client, error) {
	clients, err := manager.dataStorage.GetClientsInfo()
	if err != nil {
		return data_model.Client{}, errors.Wrap(ErrDataLayer, err.Error())
	}

	if len(clients) == 0 {
		return data_model.Client{}, errors.New("no clients have been added before trying to issue the lead")
	}

	type clientValueTuple struct {
		client data_model.Client
		value  float64
	}

	clientsTupleCh := make(chan clientValueTuple, len(clients))
	for i := range clients {
		go func(client data_model.Client) {
			tuple := clientValueTuple{client: data_model.Client{}, value: -1}
			if clientValue, ok := manager.getClientValue(client); ok {
				tuple = clientValueTuple{client: client, value: clientValue}
			}

			clientsTupleCh <- tuple
		}(clients[i])
	}

	winners := make([]data_model.Client, 1)
	curMaxValue := float64(0)
	counter := 0
	for tuple := range clientsTupleCh {
		if tuple.value > curMaxValue {
			curMaxValue = tuple.value
			winners = []data_model.Client{tuple.client}
		} else if tuple.value == curMaxValue {
			winners = append(winners, tuple.client)
		}

		counter++
		if counter >= len(clients) {
			close(clientsTupleCh)
		}
	}

	if len(winners) < 1 {
		return data_model.Client{}, errors.Errorf("none of the clients can get a lead")
	}

	return winners[rand.Intn(len(winners))], nil // select random winner if there are > 1 winner with the same value
}

// alg should be reworked if we're wanna give leads to the client evenly during the ClientTTL
func (manager *ClientsManager) getClientValue(client data_model.Client) (value float64, shouldProceed bool) {
	weightsSum := float64(1)
	priorityWeight := weightsSum * 0.5
	workingHoursWeight := weightsSum * 0.2
	leadsWeight := weightsSum * 0.3

	priorityValue := math.Pow(float64(client.Priority), 2) * priorityWeight            // bigger priority - bigger value
	workingHoursValue := math.Log10(float64(client.WorkingHours)) * workingHoursWeight // more hours - bigger value

	leadsGap := float64(client.LeadCap - client.GotLeads)
	if leadsGap == 0 {
		return 0, false
	}

	/*
		more leadsGap - bigger value
		less leadCap - bigger value
	*/
	leadsCountValue := leadsGap * (1 / float64(client.LeadCap)) * leadsWeight

	return priorityValue + workingHoursValue + leadsCountValue, true
}
