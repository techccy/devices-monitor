package storage

import (
	"github.com/ccy/devices-monitor/internal/common"
	"sync"
	"time"
)

type Storage struct {
	users     map[string]*common.User
	devices   map[string]*common.Device
	snapshots map[string]*common.Snapshot
	mu        sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		users:     make(map[string]*common.User),
		devices:   make(map[string]*common.Device),
		snapshots: make(map[string]*common.Snapshot),
	}
}

func (s *Storage) CreateUser(email, password string) (*common.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := &common.User{
		ID:       generateID(),
		Email:    email,
		Password: password,
	}
	s.users[user.ID] = user
	return user, nil
}

func (s *Storage) CreateUserWithHash(email, hashedPassword string) (*common.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := &common.User{
		ID:       generateID(),
		Email:    email,
		Password: hashedPassword,
	}
	s.users[user.ID] = user
	return user, nil
}

func (s *Storage) GetUserByEmail(email string) (*common.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (s *Storage) GetUser(id string) (*common.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.users[id], nil
}

func (s *Storage) CreateDevice(userID, name, identifier string) (*common.Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	device := &common.Device{
		ID:         generateID(),
		UserID:     userID,
		Name:       name,
		Identifier: identifier,
		DeviceKey:  generateID(),
		Online:     false,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}
	s.devices[device.ID] = device
	return device, nil
}

func (s *Storage) GetDevice(id string) (*common.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.devices[id], nil
}

func (s *Storage) GetDevicesByUser(userID string) ([]*common.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var devices []*common.Device
	for _, device := range s.devices {
		if device.UserID == userID {
			devices = append(devices, device)
		}
	}
	return devices, nil
}

func (s *Storage) UpdateDeviceStatus(deviceID string, online bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if device, ok := s.devices[deviceID]; ok {
		device.Online = online
		device.UpdatedAt = time.Now().Unix()
	}
	return nil
}

func (s *Storage) SaveSnapshot(snapshot *common.Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshots[snapshot.DeviceID] = snapshot

	if device, ok := s.devices[snapshot.DeviceID]; ok {
		device.UpdatedAt = time.Now().Unix()
	}
	return nil
}

func (s *Storage) GetSnapshot(deviceID string) (*common.Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.snapshots[deviceID], nil
}
