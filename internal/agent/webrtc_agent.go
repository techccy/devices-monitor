package agent

import (
	"encoding/json"
	"fmt"
	"github.com/ccy/devices-monitor/internal/common"
	"github.com/ccy/devices-monitor/pkg/config"
	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

type WebRTCAgent struct {
	agent      *Agent
	peerConn   *webrtc.PeerConnection
	dataChan   *webrtc.DataChannel
	pty        *os.File
	mu         sync.Mutex
	sessionID  string
	cmd        *exec.Cmd
	turnConfig *config.TURNServerConfig
}

func NewWebRTCAgent(agent *Agent) *WebRTCAgent {
	return &WebRTCAgent{
		agent: agent,
	}
}

func (wa *WebRTCAgent) HandleWebRTCOffer(offer common.WebRTCOffer) error {
	wa.mu.Lock()
	defer wa.mu.Unlock()

	wa.sessionID = offer.SessionID

	iceServers := []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	}

	if wa.agent.turnConfig != nil {
		iceServers = append(iceServers, webrtc.ICEServer{
			URLs:       []string{wa.agent.turnConfig.URI},
			Username:   wa.agent.turnConfig.Username,
			Credential: wa.agent.turnConfig.Password,
		})
	}

	rtcConfig := webrtc.Configuration{
		ICEServers: iceServers,
	}

	peerConn, err := webrtc.NewPeerConnection(rtcConfig)
	if err != nil {
		return fmt.Errorf("failed to create peer connection: %w", err)
	}
	wa.peerConn = peerConn

	peerConn.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		candidateJSON := c.ToJSON()
		iceMsg := common.WebRTCIceCandidate{
			Type:          "ICE_CANDIDATE",
			DeviceID:      wa.agent.deviceID,
			SessionID:     wa.sessionID,
			Candidate:     candidateJSON.Candidate,
			SDPMLineIndex: candidateJSON.SDPMLineIndex,
			SDPMid:        candidateJSON.SDPMid,
		}
		data, _ := json.Marshal(iceMsg)
		wa.agent.conn.WriteMessage(websocket.TextMessage, data)
	})

	peerConn.OnDataChannel(func(dc *webrtc.DataChannel) {
		wa.dataChan = dc
		dc.OnOpen(func() {
			log.Printf("DataChannel opened for session %s", wa.sessionID)
			wa.startPTY()
		})

		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			if msg.IsString {
				wa.handleStringMessage(msg.Data)
			} else {
				wa.handleBinaryMessage(msg.Data)
			}
		})
	})

	err = peerConn.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offer.SDP,
	})
	if err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}

	answer, err := peerConn.CreateAnswer(nil)
	if err != nil {
		return fmt.Errorf("failed to create answer: %w", err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConn)
	err = peerConn.SetLocalDescription(answer)
	if err != nil {
		return fmt.Errorf("failed to set local description: %w", err)
	}

	<-gatherComplete

	answerMsg := common.WebRTCAnswer{
		Type:      "ANSWER",
		DeviceID:  wa.agent.deviceID,
		SessionID: wa.sessionID,
		SDP:       peerConn.LocalDescription().SDP,
	}
	data, _ := json.Marshal(answerMsg)
	wa.agent.conn.WriteMessage(websocket.TextMessage, data)

	return nil
}

func (wa *WebRTCAgent) startPTY() {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	wa.cmd = exec.Command(shell)
	wa.cmd.Env = os.Environ()

	ptyFile, err := pty.Start(wa.cmd)
	if err != nil {
		log.Printf("Failed to start PTY: %v", err)
		return
	}
	wa.pty = ptyFile

	go wa.copyFromPTYToDataChannel()
}

func (wa *WebRTCAgent) copyFromPTYToDataChannel() {
	buf := make([]byte, 1024)
	for {
		n, err := wa.pty.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("PTY read error: %v", err)
			}
			break
		}

		if wa.dataChan != nil && wa.dataChan.ReadyState() == webrtc.DataChannelStateOpen {
			wa.dataChan.Send(buf[:n])
		}
	}
	wa.close()
}

func (wa *WebRTCAgent) handleStringMessage(data []byte) {
	var resizeMsg common.PTYResize
	if err := json.Unmarshal(data, &resizeMsg); err == nil && resizeMsg.Type == "RESIZE" {
		wa.resizePTY(resizeMsg.Cols, resizeMsg.Rows)
	}
}

func (wa *WebRTCAgent) handleBinaryMessage(data []byte) {
	if wa.pty != nil {
		_, err := wa.pty.Write(data)
		if err != nil {
			log.Printf("PTY write error: %v", err)
		}
	}
}

func (wa *WebRTCAgent) resizePTY(cols, rows int) {
	if wa.pty != nil {
		if err := pty.Setsize(wa.pty, &pty.Winsize{
			Cols: uint16(cols),
			Rows: uint16(rows),
		}); err != nil {
			log.Printf("Failed to resize PTY: %v", err)
		}
	}
}

func (wa *WebRTCAgent) close() {
	wa.mu.Lock()
	defer wa.mu.Unlock()

	if wa.pty != nil {
		wa.pty.Close()
		wa.pty = nil
	}

	if wa.cmd != nil && wa.cmd.Process != nil {
		wa.cmd.Process.Kill()
		wa.cmd = nil
	}

	if wa.dataChan != nil {
		wa.dataChan.Close()
		wa.dataChan = nil
	}

	if wa.peerConn != nil {
		wa.peerConn.Close()
		wa.peerConn = nil
	}

	log.Printf("WebRTC session %s closed", wa.sessionID)
}

func (wa *WebRTCAgent) HandleICECandidate(candidate common.WebRTCIceCandidate) error {
	wa.mu.Lock()
	defer wa.mu.Unlock()

	if wa.peerConn == nil {
		return fmt.Errorf("peer connection not initialized")
	}

	err := wa.peerConn.AddICECandidate(webrtc.ICECandidateInit{
		Candidate:     candidate.Candidate,
		SDPMLineIndex: candidate.SDPMLineIndex,
		SDPMid:        candidate.SDPMid,
	})
	if err != nil {
		return fmt.Errorf("failed to add ICE candidate: %w", err)
	}

	return nil
}

func (wa *WebRTCAgent) HandleSSH() {
	wa.close()
	wa.mu.Lock()
	defer wa.mu.Unlock()
}
