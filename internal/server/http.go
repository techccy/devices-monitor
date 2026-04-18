package server

import (
	"context"
	"encoding/json"
	"github.com/ccy/devices-monitor/internal/common"
	"github.com/ccy/devices-monitor/pkg/auth"
	"github.com/ccy/devices-monitor/pkg/logger"
	"github.com/ccy/devices-monitor/pkg/metrics"
	"github.com/ccy/devices-monitor/pkg/storage"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	storage          *storage.Storage
	auth             *auth.Auth
	wsManager        *WSManager
	signalingManager *SignalingManager
	logger           *logger.Logger
	metrics          *metrics.Metrics
}

func NewServer(storage *storage.Storage, auth *auth.Auth) *Server {
	wsManager := NewWSManager(storage)
	return &Server{
		storage:          storage,
		auth:             auth,
		wsManager:        wsManager,
		signalingManager: NewSignalingManager(wsManager),
		logger:           logger.NewLogger(),
		metrics:          metrics.NewMetrics(),
	}
}

func (s *Server) Start(addr string) error {
	http.Handle("/metrics", s.metrics.Handler())
	http.Handle("/health", s.metrics.Handler())
	http.HandleFunc("/api/register", s.metricsMiddleware(s.loggingMiddleware(s.handleRegister)))
	http.HandleFunc("/api/login", s.metricsMiddleware(s.loggingMiddleware(s.handleLogin)))
	http.HandleFunc("/api/devices", s.authMiddleware(s.handleDevices))
	http.HandleFunc("/api/devices/", s.authMiddleware(s.handleDeviceCommands))
	http.HandleFunc("/api/ws", s.handleWebSocket)
	http.HandleFunc("/api/webrtc/offer", s.authMiddleware(s.handleWebRTCOffer))
	http.HandleFunc("/api/webrtc/answer", s.authMiddleware(s.handleWebRTCAnswer))
	http.HandleFunc("/api/webrtc/ice", s.authMiddleware(s.handleWebRTCICE))

	return http.ListenAndServe(addr, nil)
}

func (s *Server) StartTLS(addr, certFile, keyFile string) error {
	http.HandleFunc("/api/register", s.handleRegister)
	http.HandleFunc("/api/login", s.handleLogin)
	http.HandleFunc("/api/devices", s.authMiddleware(s.handleDevices))
	http.HandleFunc("/api/devices/", s.authMiddleware(s.handleDeviceCommands))
	http.HandleFunc("/api/ws", s.handleWebSocket)

	return http.ListenAndServeTLS(addr, certFile, keyFile, nil)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.logger.Error("Invalid method for registration: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("Failed to decode registration request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		s.logger.Error("Missing email or password in registration")
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	existingUser, err := s.storage.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		s.logger.Error("User already exists with email: %s", req.Email)
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := s.auth.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password: %v", err)
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user, err := s.storage.CreateUserWithHash(req.Email, hashedPassword)
	if err != nil {
		s.logger.Error("Failed to create user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := s.auth.GenerateToken(user)
	if err != nil {
		s.logger.Error("Failed to generate token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := common.RegisterResponse{
		UserID: user.ID,
		Token:  token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	s.logger.Info("User registered successfully: %s", req.Email)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := s.storage.GetUserByEmail(req.Email)
	if err != nil || user == nil || !s.auth.CheckPassword(req.Password, user.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := s.auth.GenerateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := common.LoginResponse{
		Token:  token,
		UserID: user.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)

	switch r.Method {
	case http.MethodGet:
		devices, err := s.storage.GetDevicesByUser(claims.UserID)
		if err != nil {
			http.Error(w, "Failed to get devices", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(devices)

	case http.MethodPost:
		var req common.DeviceRegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if req.Name == "" || req.Identifier == "" {
			http.Error(w, "Name and identifier are required", http.StatusBadRequest)
			return
		}

		device, err := s.storage.CreateDevice(claims.UserID, req.Name, req.Identifier)
		if err != nil {
			http.Error(w, "Failed to create device", http.StatusInternalServerError)
			return
		}

		response := common.DeviceRegisterResponse{
			DeviceID:  device.ID,
			DeviceKey: device.DeviceKey,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		s.logger.Info("Device registered: %s for user: %s", device.ID, claims.UserID)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleDeviceCommands(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)
	deviceID := strings.TrimPrefix(r.URL.Path, "/api/devices/")

	switch r.Method {
	case http.MethodGet:
		device, err := s.storage.GetDevice(deviceID)
		if err != nil || device == nil || device.UserID != claims.UserID {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}

		if !device.Online {
			snapshot, _ := s.storage.GetSnapshot(deviceID)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Device-Status", "offline")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"device":   device,
				"snapshot": snapshot,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Device-Status", "online")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"device": device,
		})

	case http.MethodPost:
		var cmd common.Command
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		device, err := s.storage.GetDevice(deviceID)
		if err != nil || device == nil || device.UserID != claims.UserID {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}

		if !device.Online {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Device-Status", "offline")
			snapshot, _ := s.storage.GetSnapshot(deviceID)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":    "Device is offline",
				"snapshot": snapshot,
			})
			return
		}

		cmd.DeviceID = deviceID
		cmd.Timestamp = 0

		response, err := s.wsManager.SendCommand(deviceID, &cmd)
		if err != nil {
			http.Error(w, "Failed to send command", http.StatusInternalServerError)
			return
		}

		if response == nil {
			http.Error(w, "No response from device", http.StatusGatewayTimeout)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	s.wsManager.HandleConnection(w, r)
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := s.auth.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "claims", claims)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		s.logger.Info("Request: %s %s", r.Method, r.URL.Path)

		next(w, r)

		duration := time.Since(start)
		s.logger.Info("Request completed: %s %s took %v", r.Method, r.URL.Path, duration)
	}
}

func (s *Server) metricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.metrics.IncrementRequest(r.URL.Path)

		writer := &responseWriter{ResponseWriter: w}
		next(writer, r)

		if writer.status >= 400 {
			s.metrics.IncrementError(r.URL.Path)
		}
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (s *Server) handleWebRTCOffer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := r.Context().Value("claims").(*auth.Claims)

	var req common.WebRTCOffer
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	sessionID := s.signalingManager.CreateSession(req.DeviceID, claims.UserID)

	err := s.signalingManager.HandleOffer(sessionID, req.DeviceID, req.SDP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"session_id": sessionID,
	})
}

func (s *Server) handleWebRTCAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := r.Context().Value("claims").(*auth.Claims)

	var req common.WebRTCAnswer
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := s.signalingManager.HandleAnswer(req.SessionID, claims.UserID, req.SDP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleWebRTCICE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := r.Context().Value("claims").(*auth.Claims)

	var req common.WebRTCIceCandidate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	sdpMLineIndex := uint16(0)
	sdpMid := ""
	if req.SDPMLineIndex != nil {
		sdpMLineIndex = *req.SDPMLineIndex
	}
	if req.SDPMid != nil {
		sdpMid = *req.SDPMid
	}

	err := s.signalingManager.HandleICECandidate(req.SessionID, req.DeviceID, claims.UserID, req.Candidate, sdpMLineIndex, sdpMid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
