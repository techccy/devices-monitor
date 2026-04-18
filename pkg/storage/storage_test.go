package storage

import (
	"github.com/ccy/devices-monitor/internal/common"
	"testing"
)

func TestCreateUser(t *testing.T) {
	st := NewStorage()

	user, err := st.CreateUser("test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.ID == "" {
		t.Fatal("User ID is empty")
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", user.Email)
	}

	if user.Password != "password123" {
		t.Errorf("Expected password password123, got %s", user.Password)
	}
}

func TestGetUserByEmail(t *testing.T) {
	st := NewStorage()

	_, err := st.CreateUser("test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	user, err := st.GetUserByEmail("test@example.com")
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	if user == nil {
		t.Fatal("User not found")
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", user.Email)
	}
}

func TestGetUserByEmailNotFound(t *testing.T) {
	st := NewStorage()

	user, err := st.GetUserByEmail("nonexistent@example.com")
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	if user != nil {
		t.Fatal("Expected nil user, got user")
	}
}

func TestCreateDevice(t *testing.T) {
	st := NewStorage()

	user, err := st.CreateUser("test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	device, err := st.CreateDevice(user.ID, "My Laptop", "laptop-123")
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	if device.ID == "" {
		t.Fatal("Device ID is empty")
	}

	if device.Name != "My Laptop" {
		t.Errorf("Expected name My Laptop, got %s", device.Name)
	}

	if device.Identifier != "laptop-123" {
		t.Errorf("Expected identifier laptop-123, got %s", device.Identifier)
	}

	if device.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, device.UserID)
	}
}

func TestGetDevice(t *testing.T) {
	st := NewStorage()

	user, err := st.CreateUser("test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	createdDevice, err := st.CreateDevice(user.ID, "My Laptop", "laptop-123")
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	device, err := st.GetDevice(createdDevice.ID)
	if err != nil {
		t.Fatalf("Failed to get device: %v", err)
	}

	if device.ID != createdDevice.ID {
		t.Errorf("Expected device ID %s, got %s", createdDevice.ID, device.ID)
	}
}

func TestGetDevicesByUser(t *testing.T) {
	st := NewStorage()

	user, err := st.CreateUser("test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = st.CreateDevice(user.ID, "Device 1", "device-1")
	if err != nil {
		t.Fatalf("Failed to create device 1: %v", err)
	}

	_, err = st.CreateDevice(user.ID, "Device 2", "device-2")
	if err != nil {
		t.Fatalf("Failed to create device 2: %v", err)
	}

	devices, err := st.GetDevicesByUser(user.ID)
	if err != nil {
		t.Fatalf("Failed to get devices by user: %v", err)
	}

	if len(devices) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(devices))
	}
}

func TestUpdateDeviceStatus(t *testing.T) {
	st := NewStorage()

	user, err := st.CreateUser("test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	device, err := st.CreateDevice(user.ID, "My Laptop", "laptop-123")
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	err = st.UpdateDeviceStatus(device.ID, true)
	if err != nil {
		t.Fatalf("Failed to update device status: %v", err)
	}

	updatedDevice, err := st.GetDevice(device.ID)
	if err != nil {
		t.Fatalf("Failed to get updated device: %v", err)
	}

	if !updatedDevice.Online {
		t.Error("Expected device to be online")
	}
}

func TestSaveSnapshot(t *testing.T) {
	st := NewStorage()

	user, err := st.CreateUser("test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	device, err := st.CreateDevice(user.ID, "My Laptop", "laptop-123")
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	snapshot := &common.Snapshot{
		DeviceID:      device.ID,
		Timestamp:     1234567890,
		CPUUsage:      50.0,
		MemoryUsage:   60.0,
		DiskUsage:     70.0,
		DiskRemaining: 1000000000000,
		NetworkStatus: "connected",
	}

	err = st.SaveSnapshot(snapshot)
	if err != nil {
		t.Fatalf("Failed to save snapshot: %v", err)
	}

	retrievedSnapshot, err := st.GetSnapshot(device.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve snapshot: %v", err)
	}

	if retrievedSnapshot.DeviceID != device.ID {
		t.Errorf("Expected device ID %s, got %s", device.ID, retrievedSnapshot.DeviceID)
	}

	if retrievedSnapshot.CPUUsage != 50.0 {
		t.Errorf("Expected CPU usage 50.0, got %f", retrievedSnapshot.CPUUsage)
	}
}
