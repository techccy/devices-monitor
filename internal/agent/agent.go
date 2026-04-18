package agent

import (
	"encoding/json"
	"fmt"
	"github.com/ccy/devices-monitor/internal/common"
	"github.com/gorilla/websocket"
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type Agent struct {
	serverURL string
	deviceID  string
	deviceKey string
	conn      *websocket.Conn
	done      chan struct{}
	heartbeat int
}

func NewAgent(serverURL, deviceID, deviceKey string, heartbeat int) *Agent {
	return &Agent{
		serverURL: serverURL,
		deviceID:  deviceID,
		deviceKey: deviceKey,
		done:      make(chan struct{}),
		heartbeat: heartbeat,
	}
}

func (a *Agent) Start() error {
	hostname, _ := os.Hostname()
	identifier := hostname + "-" + runtime.GOOS

	fmt.Printf("Starting agent for device %s (%s)\n", hostname, identifier)

	if err := a.connect(); err != nil {
		return err
	}

	go a.heartbeatLoop()
	go a.readLoop()

	<-a.done
	return nil
}

func (a *Agent) connect() error {
	backoff := 1 * time.Second
	maxBackoff := 60 * time.Second

	serverURL := a.serverURL
	if a.deviceID != "" && a.deviceKey != "" {
		serverURL += "?device_id=" + a.deviceID + "&device_key=" + a.deviceKey
	}

	for {
		conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
		if err == nil {
			a.conn = conn
			fmt.Println("Connected to server")
			return nil
		}

		fmt.Printf("Connection failed: %v, retrying in %v\n", err, backoff)
		time.Sleep(backoff)

		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

func (a *Agent) heartbeatLoop() {
	interval := time.Duration(a.heartbeat) * time.Second
	if interval == 0 {
		interval = 5 * time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.sendSnapshot(); err != nil {
				fmt.Printf("Failed to send snapshot: %v\n", err)
			}
		case <-a.done:
			return
		}
	}
}

func (a *Agent) readLoop() {
	for {
		select {
		case <-a.done:
			return
		default:
			messageType, data, err := a.conn.ReadMessage()
			if err != nil {
				fmt.Printf("Read error: %v, reconnecting...\n", err)
				if err := a.connect(); err != nil {
					continue
				}
				continue
			}

			if messageType == websocket.TextMessage {
				var cmd common.Command
				if err := json.Unmarshal(data, &cmd); err == nil {
					go a.handleCommand(cmd)
				}
			}
		}
	}
}

func (a *Agent) sendSnapshot() error {
	snapshot := a.collectSnapshot()
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	return a.conn.WriteMessage(websocket.TextMessage, data)
}

func (a *Agent) handleCommand(cmd common.Command) {
	response := common.CommandResponse{
		Type:      cmd.Type,
		DeviceID:  a.deviceID,
		Timestamp: time.Now().Unix(),
	}

	switch cmd.Type {
	case "status":
		response.Data = a.collectSnapshot()
	case "net":
		response.Data = a.collectNetworkInfo()
	case "ssh":
		response.Data = map[string]interface{}{
			"output": a.handleSSH(),
		}
	case "exec":
		if cmdData, ok := cmd.Data.(map[string]interface{}); ok {
			if command, ok := cmdData["command"].(string); ok {
				response.Data = a.executeCommand(command)
			}
		}
	}

	data, err := json.Marshal(response)
	if err == nil {
		a.conn.WriteMessage(websocket.TextMessage, data)
	}
}

func (a *Agent) collectSnapshot() *common.Snapshot {
	cpu, mem, disk := a.getSystemMetrics()

	return &common.Snapshot{
		DeviceID:      a.deviceID,
		Timestamp:     time.Now().Unix(),
		CPUUsage:      cpu,
		MemoryUsage:   mem,
		DiskUsage:     disk,
		DiskRemaining: 0,
		NetworkStatus: "connected",
	}
}

func (a *Agent) collectNetworkInfo() *common.NetworkInfo {
	hostname, _ := os.Hostname()
	addrs, _ := net.LookupHost(hostname)

	var internalIP string
	if len(addrs) > 0 {
		for _, addr := range addrs {
			if !net.ParseIP(addr).IsLoopback() {
				internalIP = addr
				break
			}
		}
	}

	return &common.NetworkInfo{
		DeviceID:   a.deviceID,
		Timestamp:  time.Now().Unix(),
		Latency:    a.measureLatency(),
		ExternalIP: "unknown",
		InternalIP: internalIP,
		Routes:     []string{"default"},
	}
}

func (a *Agent) executeCommand(cmd string) string {
	var output []byte
	var err error

	if runtime.GOOS == "windows" {
		output, err = exec.Command("cmd", "/c", cmd).CombinedOutput()
	} else {
		output, err = exec.Command("sh", "-c", cmd).CombinedOutput()
	}

	if err != nil {
		return fmt.Sprintf("Error: %v\nOutput: %s", err, string(output))
	}
	return string(output)
}

func (a *Agent) handleSSH() string {
	return "SSH session started. Interactive terminal forwarding..."
}

func (a *Agent) getSystemMetrics() (cpu, mem, disk float64) {
	return 50.0, 60.0, 70.0
}

func (a *Agent) measureLatency() int64 {
	return 50
}

func (a *Agent) Stop() {
	close(a.done)
	if a.conn != nil {
		a.conn.Close()
	}
}
