package server

import (
	"encoding/binary"
	"net"
	"time"
)

type MongoDBProto struct {
	cfg     map[string]interface{}
	delayMs int
}

func NewMongoDBProto(cfg map[string]interface{}) *MongoDBProto {
	m := &MongoDBProto{cfg: cfg}
	if cfg != nil {
		if v, ok := cfg["delay_ms"]; ok {
			switch val := v.(type) {
			case float64:
				m.delayMs = int(val)
			case int:
				m.delayMs = val
			}
		}
	}
	return m
}

type MongoQuery struct {
	FullCollectionName string
	Query              []byte
}

func (m *MongoDBProto) Handle(conn net.Conn) (string, bool, string) {
	if m.delayMs > 0 {
		time.Sleep(time.Duration(m.delayMs) * time.Millisecond)
	}
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	data := make([]byte, 4096)
	n, err := conn.Read(data)
	if err != nil || n == 0 {
		return "", false, ""
	}
	data = data[:n]

	if len(data) < 16 {
		return "", false, ""
	}

	msgID := binary.LittleEndian.Uint32(data[4:8])
	opCode := binary.LittleEndian.Uint32(data[12:16])

	if opCode == 1 && len(data) >= 36 {
		query := m.parseOPQuery(data[16:])
		m.sendOPReply(conn, msgID, query)
		return "mongodb:" + query.FullCollectionName, query.FullCollectionName != "", ""
	}

	return "", false, ""
}

func (m *MongoDBProto) parseOPQuery(data []byte) MongoQuery {
	var q MongoQuery
	if len(data) < 4 {
		return q
	}
	flags := binary.LittleEndian.Uint32(data[0:4])
	_ = flags
	offset := 4

	if len(data) <= offset {
		return q
	}
	nameBytes := []byte{}
	for i := offset; i < len(data); i++ {
		if data[i] == 0 {
			nameBytes = data[offset:i]
			q.FullCollectionName = string(nameBytes)
			break
		}
	}
	return q
}

func (m *MongoDBProto) sendOPReply(conn net.Conn, requestID uint32, q MongoQuery) {
	response := make([]byte, 36)
	binary.LittleEndian.PutUint32(response[0:4], 68)
	binary.LittleEndian.PutUint32(response[4:8], requestID)
	binary.LittleEndian.PutUint32(response[8:12], 0)
	binary.LittleEndian.PutUint32(response[12:16], 1)
	binary.LittleEndian.PutUint32(response[16:20], 0)
	binary.LittleEndian.PutUint32(response[20:24], 0)
	binary.LittleEndian.PutUint32(response[24:28], 0)
	binary.LittleEndian.PutUint32(response[28:32], 0)
	binary.LittleEndian.PutUint32(response[32:36], 0)

	body := []byte{
		0x16, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0xd4,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x10, 0x00, 0x00, 0x00,
		0x01,
		0x6f, 0x6b, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x80, 0x3f,
	}
	response = append(response, body...)
	binary.LittleEndian.PutUint32(response[0:4], uint32(len(response)))
	conn.Write(response)
}
