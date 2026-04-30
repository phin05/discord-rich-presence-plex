package discord

import (
	"bytes"
	"drpp/server/config"
	"drpp/server/logger"
	"encoding/binary"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const (
	maxPayloadBytes = uint32(1 * 1024 * 1024) // 1 MB
	rateLimitWindow = 15 * time.Second
)

const isUnix = runtime.GOOS != "windows"

var RuntimeDirectoryPath = func() string {
	if config.Containerised {
		return "/run/app"
	}
	// https://github.com/discord/discord-rpc/blob/963aa9f3e5ce81a4682c6ca3d136cddda614db33/src/connection_unix.cpp#L29C33-L29C33
	for _, key := range []string{"XDG_RUNTIME_DIR", "TMPDIR", "TMP", "TEMP"} {
		if path := os.Getenv(key); path != "" {
			return path
		}
	}
	return "/tmp"
}()
var processId = os.Getpid()
var ipcPipeBase = func() string {
	if isUnix {
		return RuntimeDirectoryPath
	}
	return `\\?\pipe`
}()

type Service struct {
	clientId   string
	pipes      []string
	conn       net.Conn
	connected  bool
	mu         sync.RWMutex
	rateLimit  int
	calls      []time.Time
	ipcTimeout time.Duration
}

func NewService(clientId string, pipeNumber int, rateLimit int, ipcTimeoutSeconds int) *Service {
	var pipes []string
	var pipeNumbers []int
	if pipeNumber == -1 {
		for i := range 10 {
			pipeNumbers = append(pipeNumbers, i)
		}
	} else {
		pipeNumbers = append(pipeNumbers, pipeNumber)
	}
	for _, i := range pipeNumbers {
		pipeName := fmt.Sprintf("discord-ipc-%d", i)
		pipes = append(pipes, filepath.Join(ipcPipeBase, pipeName))
		if isUnix {
			pipes = append(pipes, filepath.Join(ipcPipeBase, "app", "com.discordapp.Discord", pipeName))
			pipes = append(pipes, filepath.Join(ipcPipeBase, ".flatpak", "com.discordapp.Discord", "xdg-run", pipeName))
			pipes = append(pipes, filepath.Join(ipcPipeBase, ".flatpak", "dev.vencord.Vesktop", "xdg-run", pipeName))
			pipes = append(pipes, filepath.Join(ipcPipeBase, "snap.discord", pipeName))
		}
	}
	return &Service{
		clientId:   clientId,
		pipes:      pipes,
		rateLimit:  rateLimit,
		ipcTimeout: time.Duration(ipcTimeoutSeconds) * time.Second,
	}
}

func (s *Service) SetActivity(activity *Activity) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	var callsInRange []time.Time
	cutoff := now.Add(-rateLimitWindow)
	for _, call := range s.calls {
		if call.After(cutoff) {
			callsInRange = append(callsInRange, call)
		}
	}
	s.calls = callsInRange
	if len(s.calls) >= s.rateLimit {
		return fmt.Errorf("rate limit exceeded, retry after %s", s.calls[0].Add(rateLimitWindow).Sub(now).Round(time.Second))
	}
	s.calls = append(s.calls, now)
	deadline := now.Add(s.ipcTimeout)
	if !s.connected {
		if err := s.connect(deadline); err != nil {
			return fmt.Errorf("connect: %w", err)
		}
	}
	if err := s.write(
		1,
		&frame{
			Cmd: frameCmdSetActivity,
			Args: &frameArgs{
				Pid:      processId,
				Activity: activity,
			},
			Nonce: time.Now().Format(time.RFC3339Nano),
		},
		deadline,
	); err != nil {
		s.disconnect()
		return fmt.Errorf("write: %w", err)
	}
	resp, err := s.read(deadline)
	if err != nil {
		s.disconnect()
		return fmt.Errorf("read: %w", err)
	}
	if resp.Evt == "ERROR" {
		return errors.New(resp.Data.Message)
	}
	return nil
}

func (s *Service) connect(deadline time.Time) error {
	if s.connected {
		return errors.New("already connected")
	}
	for _, pipe := range s.pipes {
		var err error
		s.conn, err = dial(pipe, time.Second)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				logger.Error(err, "Failed to connect on pipe %q", pipe)
			}
			continue
		}
		s.connected = true
		if err := s.write(0, &handshake{V: 1, ClientId: s.clientId}, deadline); err != nil {
			logger.Error(err, "Failed to write handshake")
			s.disconnect()
			continue
		}
		resp, err := s.read(deadline)
		if err != nil {
			logger.Error(err, "Failed to read handshake response")
			s.disconnect()
			continue
		}
		if resp.Evt != "READY" {
			logger.Error(errors.New(resp.Message), "Handshake error")
			s.disconnect()
			continue
		}
		logger.Info("Connected on pipe %q as user %q", pipe, resp.Data.User.Username)
		return nil
	}
	return fmt.Errorf("failed to connect, tried %d pipes", len(s.pipes))
}

func (s *Service) Disconnect() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.connected {
		return
	}
	s.disconnect()
	logger.Info("Disconnected")
}

func (s *Service) disconnect() {
	if err := s.conn.Close(); err != nil {
		logger.Error(err, "Failed to close connection")
	}
	s.connected = false
}

func (s *Service) read(deadline time.Time) (*ipcResponse, error) {
	if !s.connected {
		return nil, errors.New("not connected")
	}
	if err := s.conn.SetReadDeadline(deadline); err != nil {
		return nil, fmt.Errorf("set read deadline: %w", err)
	}
	header := make([]byte, 8)
	if _, err := io.ReadFull(s.conn, header); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	op := binary.LittleEndian.Uint32(header[0:4])
	length := binary.LittleEndian.Uint32(header[4:8])
	if length > maxPayloadBytes {
		return nil, fmt.Errorf("payload size %d exceeds maximum of %d", length, maxPayloadBytes)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(s.conn, payload); err != nil {
		return nil, fmt.Errorf("read payload: %w", err)
	}
	logger.Debug("[READ] Op: %d, Payload: %s", op, string(payload))
	var response ipcResponse
	if err := json.Unmarshal(payload, &response); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}
	return &response, nil
}

func (s *Service) write(op uint32, payload any, deadline time.Time) error {
	if !s.connected {
		return errors.New("not connected")
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	logger.Debug("[WRITE] Op: %d, Payload: %s", op, string(payloadBytes))
	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.LittleEndian, op); err != nil {
		return fmt.Errorf("write op to buffer: %w", err)
	}
	payloadLen := uint32(len(payloadBytes)) //nolint:gosec
	if err := binary.Write(&buffer, binary.LittleEndian, payloadLen); err != nil {
		return fmt.Errorf("write length to buffer: %w", err)
	}
	if _, err := buffer.Write(payloadBytes); err != nil {
		return fmt.Errorf("write payload to buffer: %w", err)
	}
	if err := s.conn.SetWriteDeadline(deadline); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}
	if _, err := s.conn.Write(buffer.Bytes()); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}
