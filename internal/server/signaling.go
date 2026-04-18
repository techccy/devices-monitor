package server

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"
)

type SignalingSession struct {
	SessionID     string
	DeviceID      string
	CliID         string
	OfferSDP      string
	AnswerSDP     string
	ICECandidates []string
	CreatedAt     int64
}

type SignalingManager struct {
	sessions map[string]*SignalingSession
	mu       sync.RWMutex
}

func NewSignalingManager(wsManager *WSManager) *SignalingManager {
	wsManagerInstance = wsManager
	return &SignalingManager{
		sessions: make(map[string]*SignalingSession),
	}
}

func (sm *SignalingManager) CreateSession(deviceID, cliID string) string {
	sessionID := uuid.New().String()
	session := &SignalingSession{
		SessionID:     sessionID,
		DeviceID:      deviceID,
		CliID:         cliID,
		ICECandidates: []string{},
		CreatedAt:     time.Now().Unix(),
	}

	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.mu.Unlock()

	log.Printf("Created WebRTC signaling session %s for device %s and CLI %s", sessionID, deviceID, cliID)
	return sessionID
}

func (sm *SignalingManager) HandleOffer(sessionID, deviceID, sdp string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	if session.DeviceID != deviceID {
		return ErrUnauthorized
	}

	session.OfferSDP = sdp
	log.Printf("Received WebRTC offer for session %s", sessionID)

	return nil
}

func (sm *SignalingManager) HandleAnswer(sessionID, cliID, sdp string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	if session.CliID != cliID {
		return ErrUnauthorized
	}

	session.AnswerSDP = sdp
	log.Printf("Received WebRTC answer for session %s", sessionID)

	return nil
}

func (sm *SignalingManager) HandleICECandidate(sessionID, deviceID, cliID, candidate string, sdpMLineIndex uint16, sdpMid string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	if session.DeviceID != deviceID && session.CliID != cliID {
		return ErrUnauthorized
	}

	iceData := map[string]interface{}{
		"candidate":     candidate,
		"sdpMLineIndex": sdpMLineIndex,
		"sdpMid":        sdpMid,
	}
	iceJSON, _ := json.Marshal(iceData)
	session.ICECandidates = append(session.ICECandidates, string(iceJSON))

	log.Printf("Received ICE candidate for session %s", sessionID)
	return nil
}

func (sm *SignalingManager) GetSession(sessionID string) (*SignalingSession, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

func (sm *SignalingManager) DeleteSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, sessionID)
	log.Printf("Deleted WebRTC signaling session %s", sessionID)
}

func (sm *SignalingManager) ForwardToDevice(sessionID string, message interface{}) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	return sm.sendToDevice(session.DeviceID, message)
}

func (sm *SignalingManager) ForwardToCLI(sessionID string, message interface{}) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	return sm.sendToCLI(session.CliID, message)
}

func (sm *SignalingManager) sendToDevice(deviceID string, message interface{}) error {
	wsManagerInstance.mu.RLock()
	wsConn, exists := wsManagerInstance.connections[deviceID]
	wsManagerInstance.mu.RUnlock()

	if !exists {
		return ErrDeviceOffline
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	wsConn.send <- data
	return nil
}

func (sm *SignalingManager) sendToCLI(cliID string, message interface{}) error {
	cliConn, exists := wsManagerInstance.cliConnections[cliID]
	if !exists {
		return ErrCLINotConnected
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	cliConn.send <- data
	return nil
}

var (
	ErrSessionNotFound = &SignalingError{Code: "SESSION_NOT_FOUND", Message: "Session not found"}
	ErrUnauthorized    = &SignalingError{Code: "UNAUTHORIZED", Message: "Unauthorized access"}
	ErrDeviceOffline   = &SignalingError{Code: "DEVICE_OFFLINE", Message: "Device is offline"}
	ErrCLINotConnected = &SignalingError{Code: "CLI_NOT_CONNECTED", Message: "CLI is not connected"}
)

type SignalingError struct {
	Code    string
	Message string
}

func (e *SignalingError) Error() string {
	return e.Message
}

var wsManagerInstance *WSManager
