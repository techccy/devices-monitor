package ssh

import (
	"encoding/json"
	"fmt"
	"github.com/ccy/devices-monitor/internal/common"
	"github.com/gorilla/websocket"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

type SSHTunnel struct {
	conn     *websocket.Conn
	deviceID string
}

func NewSSHTunnel(conn *websocket.Conn, deviceID string) *SSHTunnel {
	return &SSHTunnel{
		conn:     conn,
		deviceID: deviceID,
	}
}

func (t *SSHTunnel) Connect() error {
	fmt.Printf("Establishing SSH tunnel to device %s...\n", t.deviceID)

	cmd := common.Command{
		Type:      "ssh",
		DeviceID:  t.deviceID,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	if err := t.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return err
	}

	go t.readFromServer()
	go t.writeToServer()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	return t.conn.Close()
}

func (t *SSHTunnel) readFromServer() {
	for {
		messageType, data, err := t.conn.ReadMessage()
		if err != nil {
			fmt.Printf("\nConnection closed: %v\n", err)
			os.Exit(1)
		}

		if messageType == websocket.TextMessage {
			var msg map[string]interface{}
			if err := json.Unmarshal(data, &msg); err == nil {
				if output, ok := msg["output"].(string); ok {
					fmt.Print(output)
				}
			}
		}
	}
}

func (t *SSHTunnel) writeToServer() {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe")
	} else {
		cmd = exec.Command("/bin/bash")
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start shell: %v\n", err)
		return
	}

	if err := cmd.Wait(); err != nil {
		fmt.Printf("Shell exited: %v\n", err)
	}
}

func StartLocalProxy() error {
	fmt.Println("Starting local SSH proxy...")
	fmt.Println("SSH tunnel functionality requires:")
	fmt.Println("1. Agent to run local shell")
	fmt.Println("2. WebSocket message forwarding for I/O")
	fmt.Println("3. Terminal size handling")
	return nil
}
