package cli

import (
	"encoding/json"
	"fmt"
	"github.com/ccy/devices-monitor/internal/common"
	"github.com/ccy/devices-monitor/pkg/config"
	"github.com/pion/webrtc/v3"
	"golang.org/x/sys/unix"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"
)

const (
	TCGETS = 0x5401
	TCSETS = 0x5402
)

type WebRTCCLI struct {
	cli        *CLI
	peerConn   *webrtc.PeerConnection
	dataChan   *webrtc.DataChannel
	oldState   *unix.Termios
	done       chan struct{}
	turnConfig *config.TURNServerConfig
}

func NewWebRTCCLI(cli *CLI) *WebRTCCLI {
	return &WebRTCCLI{
		cli:  cli,
		done: make(chan struct{}),
	}
}

func (wc *WebRTCCLI) Connect(deviceID string) error {
	if err := wc.cli.loadConfig(); err != nil {
		return fmt.Errorf("not logged in. Please run 'ccy login' first")
	}

	iceServers := []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	}

	if wc.turnConfig != nil {
		iceServers = append(iceServers, webrtc.ICEServer{
			URLs:       []string{wc.turnConfig.URI},
			Username:   wc.turnConfig.Username,
			Credential: wc.turnConfig.Password,
		})
	}

	rtcConfig := webrtc.Configuration{
		ICEServers: iceServers,
	}

	peerConn, err := webrtc.NewPeerConnection(rtcConfig)
	if err != nil {
		return fmt.Errorf("failed to create peer connection: %w", err)
	}
	wc.peerConn = peerConn

	peerConn.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		candidateJSON := c.ToJSON()
		iceMsg := common.WebRTCIceCandidate{
			Type:          "ICE_CANDIDATE",
			DeviceID:      deviceID,
			SessionID:     "",
			Candidate:     candidateJSON.Candidate,
			SDPMLineIndex: candidateJSON.SDPMLineIndex,
			SDPMid:        candidateJSON.SDPMid,
		}

		data, _ := json.Marshal(iceMsg)
		wc.sendHTTPPost("/api/webrtc/ice", data)
	})

	dataChan, err := peerConn.CreateDataChannel("terminal", nil)
	if err != nil {
		return fmt.Errorf("failed to create data channel: %w", err)
	}
	wc.dataChan = dataChan

	dataChan.OnOpen(func() {
		fmt.Printf("Connected to device %s\n", deviceID)
		wc.enterRawMode()
		wc.handleInput()
	})

	dataChan.OnMessage(func(msg webrtc.DataChannelMessage) {
		if msg.IsString {
			wc.handleStringMessage(msg.Data)
		} else {
			wc.handleBinaryMessage(msg.Data)
		}
	})

	peerConn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Printf("Connection state changed: %s\n", state.String())
		if state == webrtc.PeerConnectionStateClosed || state == webrtc.PeerConnectionStateFailed {
			wc.exitRawMode()
			close(wc.done)
		}
	})

	offer, err := peerConn.CreateOffer(nil)
	if err != nil {
		return fmt.Errorf("failed to create offer: %w", err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConn)
	err = peerConn.SetLocalDescription(offer)
	if err != nil {
		return fmt.Errorf("failed to set local description: %w", err)
	}

	<-gatherComplete

	offerMsg := common.WebRTCOffer{
		Type:      "OFFER",
		DeviceID:  deviceID,
		SessionID: "",
		SDP:       peerConn.LocalDescription().SDP,
	}

	data, _ := json.Marshal(offerMsg)
	var response struct {
		SessionID string `json:"session_id"`
	}

	if err := wc.sendHTTPPostJSON("/api/webrtc/offer", data, &response); err != nil {
		return err
	}

	wc.monitorWindowSize(deviceID)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		wc.close()
	}()

	<-wc.done
	return nil
}

func (wc *WebRTCCLI) sendHTTPPost(endpoint string, data []byte) error {
	req, err := wc.cli.createRequest("POST", endpoint, data)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	return nil
}

func (wc *WebRTCCLI) sendHTTPPostJSON(endpoint string, data []byte, response interface{}) error {
	req, err := wc.cli.createRequest("POST", endpoint, data)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(response)
}

func (wc *WebRTCCLI) enterRawMode() error {
	fd := int(os.Stdin.Fd())

	oldState, err := unix.IoctlGetTermios(fd, TCGETS)
	if err != nil {
		return err
	}
	wc.oldState = oldState

	newState := *oldState
	newState.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG
	newState.Cc[unix.VMIN] = 1
	newState.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(fd, TCSETS, &newState); err != nil {
		return err
	}

	return nil
}

func (wc *WebRTCCLI) exitRawMode() {
	if wc.oldState != nil {
		fd := int(os.Stdin.Fd())
		unix.IoctlSetTermios(fd, TCSETS, wc.oldState)
		wc.oldState = nil
	}
}

func (wc *WebRTCCLI) handleInput() {
	buf := make([]byte, 1024)
	for {
		select {
		case <-wc.done:
			return
		default:
			n, err := os.Stdin.Read(buf)
			if err != nil {
				if err == io.EOF {
					wc.close()
					return
				}
				continue
			}

			if wc.dataChan != nil && wc.dataChan.ReadyState() == webrtc.DataChannelStateOpen {
				wc.dataChan.Send(buf[:n])
			}
		}
	}
}

func (wc *WebRTCCLI) handleStringMessage(data []byte) {
	var resizeMsg common.PTYResize
	if err := json.Unmarshal(data, &resizeMsg); err == nil && resizeMsg.Type == "RESIZE" {
		wc.resizePTY(resizeMsg.Cols, resizeMsg.Rows)
	}
}

func (wc *WebRTCCLI) handleBinaryMessage(data []byte) {
	os.Stdout.Write(data)
}

func (wc *WebRTCCLI) resizePTY(cols, rows int) {
	ws, err := getWinsize()
	if err == nil {
		if ws.Cols != uint16(cols) || ws.Rows != uint16(rows) {
			if err := setWinsize(uint16(cols), uint16(rows)); err != nil {
				fmt.Printf("Failed to resize terminal: %v\n", err)
			}
		}
	}
}

func (wc *WebRTCCLI) monitorWindowSize(deviceID string) {
	lastCols, lastRows := 0, 0

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ws, err := getWinsize()
				if err != nil {
					continue
				}

				if int(ws.Cols) != lastCols || int(ws.Rows) != lastRows {
					lastCols = int(ws.Cols)
					lastRows = int(ws.Rows)

					if wc.dataChan != nil && wc.dataChan.ReadyState() == webrtc.DataChannelStateOpen {
						resizeMsg := common.PTYResize{
							Type: "RESIZE",
							Cols: lastCols,
							Rows: lastRows,
						}
						data, _ := json.Marshal(resizeMsg)
						wc.dataChan.Send(data)
					}
				}
			case <-wc.done:
				return
			}
		}
	}()
}

func (wc *WebRTCCLI) close() {
	wc.exitRawMode()

	if wc.dataChan != nil {
		wc.dataChan.Close()
		wc.dataChan = nil
	}

	if wc.peerConn != nil {
		wc.peerConn.Close()
		wc.peerConn = nil
	}

	close(wc.done)
}

type Winsize struct {
	Rows uint16
	Cols uint16
}

func getWinsize() (*Winsize, error) {
	ws := &Winsize{}
	fd := int(os.Stdout.Fd())
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(ws)))
	if err != 0 {
		return nil, err
	}
	return ws, nil
}

func setWinsize(cols, rows uint16) error {
	ws := &Winsize{Rows: rows, Cols: cols}
	fd := int(os.Stdout.Fd())
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
	if err != 0 {
		return err
	}
	return nil
}
