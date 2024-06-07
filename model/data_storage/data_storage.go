package data_storage

import (
	"bhc_test_task/model/data_model"
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

type DataStorage interface {
	CreateClient(name string, workingHours uint16, priority uint8, leadCap uint16) (uint64, error)
	GetClientsInfo() ([]data_model.Client, error)
	GetClientInfo(id uint64) (data_model.Client, error)
	UpdateClients(clients ...data_model.Client) error
}

func NewLocalDataStorage() DataStorage {
	return &LocalDataStorage{
		mutex:   new(sync.RWMutex),
		clients: make(map[uint64]data_model.Client),
	}
}

type LocalDataStorage struct {
	mutex   *sync.RWMutex
	clients map[uint64]data_model.Client
}

func (storage *LocalDataStorage) CreateClient(name string, workingHours uint16, priority uint8, leadCap uint16) (uint64, error) {
	client := data_model.NewClient(name, workingHours, priority, leadCap)
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	if _, ok := storage.clients[client.ID]; ok {
		return 0, errors.Wrapf(ErrClientAlreadyExists, fmt.Sprintf("client with id %v already exists", client.ID))
	}

	storage.clients[client.ID] = client

	return client.ID, nil
}

func (storage *LocalDataStorage) GetClientsInfo() ([]data_model.Client, error) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()

	res := make([]data_model.Client, 0, len(storage.clients))

	for _, client := range storage.clients {
		res = append(res, client)
	}

	return res, nil
}

func (storage *LocalDataStorage) GetClientInfo(id uint64) (data_model.Client, error) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()

	client, ok := storage.clients[id]
	if ok {
		return client, nil
	}

	return data_model.Client{}, ErrClientDoesNotExist
}

func (storage *LocalDataStorage) UpdateClients(clients ...data_model.Client) error {
	storage.mutex.Lock()
	for i := range clients {
		client := clients[i]
		storage.clients[client.ID] = client
	}
	storage.mutex.Unlock()

	return nil
}
