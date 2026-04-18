package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ServerConfig struct {
	Addr             string            `json:"addr"`
	TLSAddr          string            `json:"tls_addr"`
	CertFile         string            `json:"cert_file"`
	KeyFile          string            `json:"key_file"`
	Secret           string            `json:"secret"`
	TURNServerConfig *TURNServerConfig `json:"turn_server,omitempty"`
}

type TURNServerConfig struct {
	URI      string `json:"uri"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type AgentConfig struct {
	ServerURL        string            `json:"server_url"`
	DeviceID         string            `json:"device_id"`
	DeviceKey        string            `json:"device_key"`
	Heartbeat        int               `json:"heartbeat"`
	TURNServerConfig *TURNServerConfig `json:"turn_server,omitempty"`
}

type CLIConfig struct {
	ServerURL        string            `json:"server_url"`
	TURNServerConfig *TURNServerConfig `json:"turn_server,omitempty"`
}

func LoadServerConfig(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func LoadAgentConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config AgentConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func LoadCLIConfig(path string) (*CLIConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config CLIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveServerConfig(path string, config *ServerConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func SaveAgentConfig(path string, config *AgentConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func SaveCLIConfig(path string, config *CLIConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func GetDefaultServerConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".ccy", "server.json")
}

func GetDefaultAgentConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".ccy", "agent.json")
}

func GetDefaultCLIConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".ccy", "cli.json")
}
