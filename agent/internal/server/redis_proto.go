package server

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type RedisProto struct {
	cfg   map[string]interface{}
	banner string
	delayMs int
}

func NewRedisProto(cfg map[string]interface{}) *RedisProto {
	r := &RedisProto{cfg: cfg}
	if cfg != nil {
		if v, ok := cfg["banner"]; ok {
			if s, ok := v.(string); ok {
				r.banner = s
			}
		}
		if v, ok := cfg["delay_ms"]; ok {
			switch val := v.(type) {
			case float64:
				r.delayMs = int(val)
			case int:
				r.delayMs = val
			}
		}
	}
	if r.banner == "" {
		r.banner = "Redis server v=6.2.5 sha=00000000:0 malloc=jemalloc-5.1.0 bits=64 build=0\r\n"
	}
	return r
}

type RedisCmd struct {
	Name   string
	Args   []string
	Raw    string
}

func (r *RedisProto) Handle(conn net.Conn) (string, bool, string) {
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	cmds, err := r.readCommands(conn)
	if err != nil || len(cmds) == 0 {
		return "", false, ""
	}

	var detail, lastCmd string
	isAttack := false

	for _, cmd := range cmds {
		lastCmd = cmd.Raw
		detail = r.evalCmd(cmd)
		if cmd.Name == "AUTH" || cmd.Name == "FLUSHALL" || cmd.Name == "FLUSHDB" || cmd.Name == "CONFIG" {
			isAttack = true
		}
		if cmd.Name == "QUIT" {
			r.writeSimpleString(conn, "OK")
			break
		}
	}

	if len(cmds) > 0 {
		last := cmds[len(cmds)-1]
		if last.Name == "PING" {
			r.writeSimpleString(conn, "PONG")
		} else if last.Name == "ECHO" && len(last.Args) > 0 {
			r.writeBulkString(conn, last.Args[0])
		} else if last.Name == "INFO" {
			r.writeInfo(conn)
		} else if last.Name == "DBSIZE" {
			r.writeInteger(conn, 1)
		} else if last.Name == "ROLE" {
			r.writeArray(conn, []string{"master", "0", "0"})
		} else if last.Name == "TYPE" && len(last.Args) > 0 {
			r.writeSimpleString(conn, "string")
		} else if last.Name == "EXISTS" {
			r.writeInteger(conn, 0)
		} else if last.Name == "LEN" && len(last.Args) > 0 {
			r.writeInteger(conn, 0)
		} else {
			r.writeError(conn, "ERR unknown command '%s'", last.Name)
		}
	}

	return "redis:" + detail, isAttack, lastCmd
}

func (r *RedisProto) evalCmd(cmd RedisCmd) string {
	switch cmd.Name {
	case "AUTH":
		return "auth:" + strings.Join(cmd.Args, ",")
	case "SELECT":
		return "select:" + strings.Join(cmd.Args, ",")
	case "GET", "SET", "DEL", "KEYS":
		return cmd.Name + ":" + strings.Join(cmd.Args, ",")
	default:
		return cmd.Name
	}
}

func (r *RedisProto) readCommands(conn net.Conn) ([]RedisCmd, error) {
	var cmds []RedisCmd
	buf := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return cmds, nil
	}
	data := string(buf[:n])
	parts := strings.Split(data, "\r\n")

	i := 0
	for i < len(parts) {
		if len(parts[i]) == 0 {
			i++
			continue
		}
		if parts[i][0] == '*' {
			count := 0
			if _, err := parseRESPInt(parts[i], &count); err == nil && count > 0 && i+1 < len(parts) && parts[i+1][0] == '$' {
				if i+2 < len(parts) && parts[i+2][0] == '*' {
					cmdName := strings.ToUpper(parts[i+2][1:])
					args := []string{}
					pos := i + 3
					for len(args) < count-1 && pos+1 < len(parts) {
						if parts[pos][0] == '$' {
							pos++
							if pos < len(parts) {
								args = append(args, parts[pos])
							}
						}
						pos++
					}
					cmds = append(cmds, RedisCmd{Name: cmdName, Args: args, Raw: strings.Join(args, " ")})
				} else if i+2 < len(parts) {
					cmds = append(cmds, RedisCmd{Name: strings.ToUpper(parts[i+2][1:]), Raw: parts[i+2][1:]})
				}
				i++
			}
		}
		i++
	}
	return cmds, nil
}

func parseRESPInt(s string, out *int) (bool, error) {
	if len(s) == 0 {
		return false, nil
	}
	val := 0
	for i := 1; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false, nil
		}
		val = val*10 + int(s[i]-'0')
	}
	*out = val
	return true, nil
}

func (r *RedisProto) writeSimpleString(conn net.Conn, s string) {
	conn.Write([]byte("+" + s + "\r\n"))
}

func (r *RedisProto) writeBulkString(conn net.Conn, s string) {
	conn.Write([]byte("$" + itoa(len(s)) + "\r\n" + s + "\r\n"))
}

func (r *RedisProto) writeError(conn net.Conn, format string, args ...interface{}) {
	conn.Write([]byte("-ERR " + fmt.Sprintf(format, args...) + "\r\n"))
}

func (r *RedisProto) writeInteger(conn net.Conn, n int) {
	conn.Write([]byte(":" + itoa(n) + "\r\n"))
}

func (r *RedisProto) writeArray(conn net.Conn, elems []string) {
	conn.Write([]byte("*" + itoa(len(elems)) + "\r\n"))
	for _, e := range elems {
		conn.Write([]byte("$" + itoa(len(e)) + "\r\n" + e + "\r\n"))
	}
}

func (r *RedisProto) writeInfo(conn net.Conn) {
	info := "# Server\nredis_version:6.2.5\nredis_mode:standalone\n"
	info += "os:Linux 5.4.0 x86_64\narch_bits:64\n"
	info += "tcp_port:6379\nuptime_in_seconds:3600\n"
	r.writeBulkString(conn, info)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string('0'+byte(n%10)) + result
		n /= 10
	}
	return result
}
