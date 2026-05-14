package server

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type MySQLProto struct {
	seq       byte
	cfg       map[string]interface{}
	banner    string
	delayMs   int
}

func NewMySQLProto(cfg map[string]interface{}) *MySQLProto {
	m := &MySQLProto{seq: 0, cfg: cfg}
	if cfg != nil {
		if v, ok := cfg["banner"]; ok {
			if s, ok := v.(string); ok {
				m.banner = s
			}
		}
		if v, ok := cfg["delay_ms"]; ok {
			switch val := v.(type) {
			case float64:
				m.delayMs = int(val)
			case int:
				m.delayMs = val
			}
		}
	}
	if m.banner == "" {
		m.banner = "5.7.33-0ubuntu0.18.04.1-log"
	}
	return m
}

func (m *MySQLProto) Handle(conn net.Conn) (string, bool, string) {
	if m.delayMs > 0 {
		time.Sleep(time.Duration(m.delayMs) * time.Millisecond)
	}
	if err := m.sendHandshake(conn); err != nil {
		return "", false, ""
	}

	for attempt := 0; attempt < 3; attempt++ {
		packet, err := m.readPacket(conn)
		if err != nil || len(packet) == 0 {
			return "", false, ""
		}

		if packet[0] != 0x01 {
			continue
		}

		username, resp := m.parseAuthPacket(packet[1:])
		isAttack := username == "root" || username == "admin" || username == "test"

		m.sendError(conn, 1045, "Access denied for user '%s'@'%%' (using password: YES)", username)

		detail := fmt.Sprintf("auth_attempt:user=%s", username)
		return "mysql:" + detail, isAttack, resp
	}
	return "", false, ""
}

func (m *MySQLProto) readPacket(conn net.Conn) ([]byte, error) {
	header := make([]byte, 4)
	if _, err := conn.Read(header); err != nil {
		return nil, err
	}
	length := int(header[0]) | int(header[1])<<8 | int(header[2]<<16)
	m.seq = header[3]
	if length > 65536 || length < 0 {
		return nil, fmt.Errorf("invalid packet length")
	}
	body := make([]byte, length)
	n, err := conn.Read(body)
	if err != nil {
		return nil, err
	}
	return body[:n], nil
}

func (m *MySQLProto) parseAuthPacket(data []byte) (string, string) {
	if len(data) < 32 {
		return "", string(data)
	}
	offset := 0
	if data[0] == 0x00 {
		offset = 0
	} else if data[0] == 0x14 {
		offset = int(data[0]) + 1
	} else {
		for i := 0; i < len(data) && data[i] != 0x00; i++ {
			if data[i] == 0x14 {
				offset = i + 1 + int(data[i+1])
				break
			}
		}
		if offset == 0 {
			for i := 0; i < len(data); i++ {
				if data[i] == 0x00 {
					offset = i + 1
					break
				}
			}
		}
	}
	usernameEnd := offset
	for i := offset; i < len(data); i++ {
		if data[i] == 0x00 {
			usernameEnd = i
			break
		}
	}
	username := string(data[offset:usernameEnd])
	authResponseStart := usernameEnd + 1
	var authResponse string
	if authResponseStart < len(data) {
		authResponse = string(data[authResponseStart:])
	}
	return username, authResponse
}

func (m *MySQLProto) sendHandshake(conn net.Conn) error {
	proto := []byte{0x0a}
	proto = append(proto, []byte(m.banner)...)
	proto = append(proto, 0x00)

	threadID := make([]byte, 4)
	binary.LittleEndian.PutUint32(threadID, 12345)
	proto = append(proto, threadID...)

	authPluginData := make([]byte, 8)
	for i := range authPluginData {
		authPluginData[i] = byte(i + 1)
	}
	proto = append(proto, authPluginData...)
	proto = append(proto, 0x00)

	proto = append(proto, 0xff, 0xf7, 0xff, 0x02)
	proto = append(proto, 0x08)
	proto = append(proto, 0x02)
	proto = append(proto, make([]byte, 13)...)

	pluginName := []byte("mysql_native_password\x00")
	proto = append(proto, pluginName...)

	return m.writePacket(conn, proto)
}

func (m *MySQLProto) sendError(conn net.Conn, code uint16, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	payload := []byte{0xff}
	binary.LittleEndian.PutUint16(payload[1:3], code)
	payload = append(payload, '#')
	payload = append(payload, []byte("28000")...)
	payload = append(payload, make([]byte, 6)...)
	payload = append(payload, []byte(msg)...)
	return m.writePacket(conn, payload)
}

func (m *MySQLProto) writePacket(conn net.Conn, payload []byte) error {
	length := len(payload)
	header := make([]byte, 4)
	header[0] = byte(length)
	header[1] = byte(length >> 8)
	header[2] = byte(length >> 16)
	header[3] = m.seq
	m.seq++
	_, err := conn.Write(header)
	if err != nil {
		return err
	}
	_, err = conn.Write(payload)
	return err
}
