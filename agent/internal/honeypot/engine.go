package honeypot

import "net/http"

type AttackInfo struct {
	Type    string            `json:"type"`
	Detail  map[string]string `json:"detail"`
	Source  string            `json:"source"`
}

type Honeypot interface {
	Name() string
	Type() int
	HandleHTTP(w http.ResponseWriter, r *http.Request, body []byte, cfg map[string]interface{}) (bool, *AttackInfo)
}

func Create(t int) Honeypot {
	switch t {
	case 1:
		return &ElasticsearchHoneypot{}
	case 2:
		return &WebLogicHoneypot{}
	case 3:
		return &MySQLHoneypot{}
	case 4:
		return &RedisHoneypot{}
	case 5:
		return &MongoDBHoneypot{}
	case 6:
		return &NginxHoneypot{}
	case 7:
		return &ApacheHoneypot{}
	default:
		return nil
	}
}
