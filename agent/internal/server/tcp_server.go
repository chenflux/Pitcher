package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type TCPServer struct {
	listener net.Listener
	addr     string
	handler  func(conn net.Conn)
	running  bool
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

func NewTCPServer(host string, port int, handler func(conn net.Conn)) *TCPServer {
	return &TCPServer{
		addr:    fmt.Sprintf("%s:%d", host, port),
		handler: handler,
		stopCh:  make(chan struct{}),
	}
}

func NewTCPServerWithHandler(host string, port int, handler func(conn net.Conn) (string, bool, string)) *TCPServer {
	wrapper := func(conn net.Conn) {
		handler(conn)
	}
	return &TCPServer{
		addr:    fmt.Sprintf("%s:%d", host, port),
		handler: wrapper,
		stopCh:  make(chan struct{}),
	}
}

func (s *TCPServer) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.addr, err)
	}
	s.listener = ln
	s.running = true

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for s.running {
			conn, err := ln.Accept()
			if err != nil {
				if !s.running {
					return
				}
				log.Printf("Accept error on %s: %v", s.addr, err)
				continue
			}
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				conn.SetDeadline(time.Now().Add(30 * time.Second))
				s.handler(conn)
				conn.Close()
			}()
		}
	}()

	log.Printf("TCP server listening on %s", s.addr)
	return nil
}

func (s *TCPServer) Stop() {
	s.running = false
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
}

func writeMySQLPacket(seq byte, payload []byte) []byte {
	length := len(payload)
	header := []byte{
		byte(length), byte(length >> 8), byte(length >> 16),
		seq,
	}
	return append(header, payload...)
}

func WriteMySQLHandshake(conn net.Conn) {
	payload := make([]byte, 0, 128)
	payload = append(payload, 0x0a)
	payload = append(payload, []byte("5.7.33-0ubuntu0.18.04.1-log\x00")...)
	payload = append(payload, 0x01, 0x00, 0x00, 0x00)
	payload = append(payload, []byte("!'\"#$%&")...)
	payload = append(payload, 0x00)
	payload = append(payload, 0xff, 0xf7, 0xff, 0x02)
	payload = append(payload, 0x08)
	payload = append(payload, 0x02)
	payload = append(payload, make([]byte, 13)...)
	payload = append(payload, []byte("mysql_native_password\x00")...)

	conn.Write(writeMySQLPacket(0x00, payload))
}

func WriteMySQLError(conn net.Conn) {
	msg := []byte("Access denied for user 'root'@'%' (using password: YES)")
	payload := []byte{0xff}
	payload = append(payload, 0x15, 0x04)
	payload = append(payload, []byte("#28000")...)
	payload = append(payload, msg...)

	conn.Write(writeMySQLPacket(0x02, payload))
}

func WriteRedisError(conn net.Conn) {
	conn.Write([]byte("-ERR wrong number of arguments for 'info' command\r\n"))
}

func WriteMongoReply(conn net.Conn) {
	resp := []byte{
		0x22, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0xd4,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x16, 0x00, 0x00, 0x00,
		0x02,
		'o', 'k', 0x00,
		0x00, 0x00, 0x00, 0x01,
	}
	conn.Write(resp)
}
