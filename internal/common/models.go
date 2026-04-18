package common

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Device struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	DeviceKey  string `json:"device_key,omitempty"`
	Online     bool   `json:"online"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

type Snapshot struct {
	DeviceID      string  `json:"device_id"`
	Timestamp     int64   `json:"timestamp"`
	CPUUsage      float64 `json:"cpu_usage"`
	MemoryUsage   float64 `json:"memory_usage"`
	DiskUsage     float64 `json:"disk_usage"`
	DiskRemaining int64   `json:"disk_remaining"`
	NetworkStatus string  `json:"network_status"`
}

type NetworkInfo struct {
	DeviceID   string   `json:"device_id"`
	Timestamp  int64    `json:"timestamp"`
	Latency    int64    `json:"latency_ms"`
	ExternalIP string   `json:"external_ip"`
	InternalIP string   `json:"internal_ip"`
	Routes     []string `json:"routes"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

type DeviceRegisterRequest struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

type DeviceRegisterResponse struct {
	DeviceID  string `json:"device_id"`
	DeviceKey string `json:"device_key"`
}

type Command struct {
	Type      string      `json:"type"`
	DeviceID  string      `json:"device_id"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

type CommandResponse struct {
	Type      string      `json:"type"`
	DeviceID  string      `json:"device_id"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
}
