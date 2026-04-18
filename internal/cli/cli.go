package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ccy/devices-monitor/internal/common"
	"github.com/ccy/devices-monitor/pkg/logger"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type CLI struct {
	serverURL string
	token     string
	userID    string
	configDir string
	logger    *logger.Logger
}

func NewCLI(serverURL string) *CLI {
	usr, _ := user.Current()
	configDir := filepath.Join(usr.HomeDir, ".ccy")

	return &CLI{
		serverURL: serverURL,
		configDir: configDir,
		logger:    logger.NewLogger(),
	}
}

func (c *CLI) loadConfig() error {
	tokenFile := filepath.Join(c.configDir, "token")
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return err
	}

	var config struct {
		Token  string `json:"token"`
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	c.token = config.Token
	c.userID = config.UserID
	return nil
}

func (c *CLI) saveConfig() error {
	if err := os.MkdirAll(c.configDir, 0700); err != nil {
		return err
	}

	config := struct {
		Token  string `json:"token"`
		UserID string `json:"user_id"`
	}{
		Token:  c.token,
		UserID: c.userID,
	}

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	tokenFile := filepath.Join(c.configDir, "token")
	return os.WriteFile(tokenFile, data, 0600)
}

func (c *CLI) Register(email, password string) error {
	c.logger.Info("Attempting to register user: %s", email)

	req := common.RegisterRequest{
		Email:    email,
		Password: password,
	}

	data, err := json.Marshal(req)
	if err != nil {
		c.logger.Error("Failed to marshal registration request: %v", err)
		return err
	}

	resp, err := http.Post(c.serverURL+"/api/register", "application/json", bytes.NewReader(data))
	if err != nil {
		c.logger.Error("Failed to connect to server: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("Registration failed with status %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("registration failed: %s", string(body))
	}

	var registerResp common.RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&registerResp); err != nil {
		c.logger.Error("Failed to decode registration response: %v", err)
		return err
	}

	c.token = registerResp.Token
	c.userID = registerResp.UserID

	if err := c.saveConfig(); err != nil {
		c.logger.Error("Failed to save config: %v", err)
		return err
	}

	c.logger.Info("User registered successfully: %s", email)
	fmt.Println("Registration successful")
	fmt.Printf("User ID: %s\n", registerResp.UserID)
	return nil
}

func (c *CLI) Login(email, password string) error {
	req := common.LoginRequest{
		Email:    email,
		Password: password,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.serverURL+"/api/login", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed: %s", string(body))
	}

	var loginResp common.LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return err
	}

	c.token = loginResp.Token
	c.userID = loginResp.UserID

	if err := c.saveConfig(); err != nil {
		return err
	}

	fmt.Println("Login successful")
	return nil
}

func (c *CLI) RegisterDevice(name, identifier string) error {
	if err := c.loadConfig(); err != nil {
		return fmt.Errorf("not logged in. Please run 'ccy login' first")
	}

	req := common.DeviceRegisterRequest{
		Name:       name,
		Identifier: identifier,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", c.serverURL+"/api/devices", bytes.NewReader(data))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to register device: %s", string(body))
	}

	var registerResp common.DeviceRegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&registerResp); err != nil {
		return err
	}

	fmt.Printf("Device registered successfully\n")
	fmt.Printf("Device ID: %s\n", registerResp.DeviceID)
	return nil
}

func (c *CLI) ListDevices() error {
	if err := c.loadConfig(); err != nil {
		return fmt.Errorf("not logged in. Please run 'ccy login' first")
	}

	req, err := http.NewRequest("GET", c.serverURL+"/api/devices", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get devices")
	}

	var devices []*common.Device
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return err
	}

	fmt.Println("Devices:")
	fmt.Printf("%-20s %-30s %-10s\n", "Name", "ID", "Status")
	fmt.Println(strings.Repeat("-", 60))
	for _, device := range devices {
		status := "Offline"
		if device.Online {
			status = "Online"
		}
		fmt.Printf("%-20s %-30s %-10s\n", device.Name, device.ID, status)
	}

	return nil
}

func (c *CLI) GetStatus(deviceID string) error {
	if err := c.loadConfig(); err != nil {
		return fmt.Errorf("not logged in. Please run 'ccy login' first")
	}

	req, err := http.NewRequest("GET", c.serverURL+"/api/devices/"+deviceID, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get device status")
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	status := resp.Header.Get("X-Device-Status")
	if status == "offline" {
		fmt.Println("Device is offline (showing last snapshot):")
	}

	deviceData, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(deviceData))

	return nil
}

func (c *CLI) GetNetworkInfo(deviceID string) error {
	if err := c.loadConfig(); err != nil {
		return fmt.Errorf("not logged in. Please run 'ccy login' first")
	}

	cmd := common.Command{
		Type:      "net",
		DeviceID:  deviceID,
		Timestamp: 0,
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.serverURL+"/api/devices/"+deviceID, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))

	return nil
}

func (c *CLI) SSH(deviceID string) error {
	if err := c.loadConfig(); err != nil {
		return fmt.Errorf("not logged in. Please run 'ccy login' first")
	}

	webrtcCLI := NewWebRTCCLI(c)
	return webrtcCLI.Connect(deviceID)
}

func (c *CLI) createRequest(method, endpoint string, data []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, c.serverURL+endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *CLI) Logout() error {
	tokenFile := filepath.Join(c.configDir, "token")
	if err := os.Remove(tokenFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	c.token = ""
	c.userID = ""
	fmt.Println("Logged out successfully")
	return nil
}

func (c *CLI) StartAgent() error {
	fmt.Println("Starting agent...")
	fmt.Println("This would launch the agent as a background service")
	return nil
}
