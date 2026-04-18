package server

import (
	"encoding/json"
	"github.com/ccy/devices-monitor/internal/common"
	"github.com/ccy/devices-monitor/pkg/storage"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSConnection struct {
	conn         *websocket.Conn
	deviceID     string
	send         chan []byte
	heartbeat    *time.Ticker
	lastPongTime time.Time
	mu           sync.Mutex
}

type WSManager struct {
	storage        *storage.Storage
	connections    map[string]*WSConnection
	cliConnections map[string]*WSConnection
	mu             sync.RWMutex
	pendingCmds    map[string]chan *common.CommandResponse
}

func NewWSManager(st *storage.Storage) *WSManager {
	return &WSManager{
		storage:        st,
		connections:    make(map[string]*WSConnection),
		cliConnections: make(map[string]*WSConnection),
		pendingCmds:    make(map[string]chan *common.CommandResponse),
	}
}

func (m *WSManager) HandleConnection(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")
	deviceKey := r.URL.Query().Get("device_key")

	if deviceID == "" || deviceKey == "" {
		log.Println("Missing device ID or device key")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	device, err := m.storage.GetDevice(deviceID)
	if err != nil || device == nil {
		log.Printf("Device not found: %s", deviceID)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if device.DeviceKey != deviceKey {
		log.Printf("Invalid device key for device: %s", deviceID)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	wsConn := &WSConnection{
		conn:         conn,
		send:         make(chan []byte, 256),
		deviceID:     deviceID,
		heartbeat:    time.NewTicker(30 * time.Second),
		lastPongTime: time.Now(),
	}

	m.mu.Lock()
	m.connections[deviceID] = wsConn
	m.mu.Unlock()

	m.storage.UpdateDeviceStatus(deviceID, true)
	log.Printf("Device %s connected and authenticated", deviceID)

	go wsConn.heartbeatLoop(m)
	go wsConn.readPump(m)
	go wsConn.writePump()
}

func (m *WSManager) SendCommand(deviceID string, cmd *common.Command) (*common.CommandResponse, error) {
	m.mu.RLock()
	wsConn, exists := m.connections[deviceID]
	m.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	responseChan := make(chan *common.CommandResponse, 1)
	m.mu.Lock()
	m.pendingCmds[deviceID+"."+cmd.Type] = responseChan
	m.mu.Unlock()

	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}

	wsConn.send <- data

	select {
	case resp := <-responseChan:
		return resp, nil
	default:
		return nil, nil
	}
}

func (ws *WSConnection) readPump(m *WSManager) {
	defer func() {
		ws.conn.Close()
		m.mu.Lock()
		delete(m.connections, ws.deviceID)
		m.storage.UpdateDeviceStatus(ws.deviceID, false)
		m.mu.Unlock()
		log.Printf("Device %s disconnected", ws.deviceID)
	}()

	for {
		messageType, data, err := ws.conn.ReadMessage()
		if err != nil {
			break
		}

		if messageType == websocket.TextMessage {
			var ping common.WSPing
			if err := json.Unmarshal(data, &ping); err == nil && ping.Type == "PONG" {
				ws.updatePongTime()
				log.Printf("Received PONG from device %s", ws.deviceID)
				continue
			}

			var cmd common.Command
			if err := json.Unmarshal(data, &cmd); err == nil {
				m.storage.SaveSnapshot(&common.Snapshot{
					DeviceID:  cmd.DeviceID,
					Timestamp: cmd.Timestamp,
				})
			}

			var resp common.CommandResponse
			if err := json.Unmarshal(data, &resp); err == nil {
				m.mu.Lock()
				if ch, exists := m.pendingCmds[resp.DeviceID+"."+resp.Type]; exists {
					select {
					case ch <- &resp:
					default:
					}
					delete(m.pendingCmds, resp.DeviceID+"."+resp.Type)
				}
				m.mu.Unlock()
			}
		}
	}
}

func (ws *WSConnection) writePump() {
	defer ws.conn.Close()

	for {
		select {
		case data, ok := <-ws.send:
			if !ok {
				return
			}
			if err := ws.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}
}

func (ws *WSConnection) heartbeatLoop(m *WSManager) {
	defer ws.heartbeat.Stop()

	for {
		select {
		case <-ws.heartbeat.C:
			ping := common.WSPing{
				Type:      "PING",
				Timestamp: time.Now().Unix(),
			}
			data, _ := json.Marshal(ping)
			if err := ws.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Failed to send ping to device %s: %v", ws.deviceID, err)
				return
			}

			ws.mu.Lock()
			if time.Since(ws.lastPongTime) > 90*time.Second {
				log.Printf("Device %s heartbeat timeout", ws.deviceID)
				ws.mu.Unlock()
				ws.conn.Close()
				return
			}
			ws.mu.Unlock()
		case <-time.After(90 * time.Second):
			ws.mu.Lock()
			if time.Since(ws.lastPongTime) > 90*time.Second {
				log.Printf("Device %s heartbeat timeout", ws.deviceID)
				ws.mu.Unlock()
				ws.conn.Close()
				return
			}
			ws.mu.Unlock()
		}
	}
}

func (ws *WSConnection) updatePongTime() {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.lastPongTime = time.Now()
}
