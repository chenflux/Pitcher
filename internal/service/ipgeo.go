package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type IPGeoInfo struct {
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	City        string `json:"city"`
	ISP         string `json:"isp"`
	Org         string `json:"org"`
	AS          string `json:"as"`
	Source      string `json:"source"`
}

type IPGeoService struct {
	cache   map[string]IPGeoInfo
	mu      sync.RWMutex
	baseURL string
}

func NewIPGeoService() *IPGeoService {
	return &IPGeoService{
		cache:   make(map[string]IPGeoInfo),
		baseURL: "http://ip-api.com/json",
	}
}

type ipAPIResponse struct {
	Status      string `json:"status"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	Region      string `json:"regionName"`
	City        string `json:"city"`
	ISP         string `json:"isp"`
	Org         string `json:"org"`
	AS          string `json:"as"`
	Query       string `json:"query"`
}

func (s *IPGeoService) Lookup(ip string) IPGeoInfo {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return s.empty()
	}

	if net.ParseIP(ip) == nil {
		return s.empty()
	}

	if ip == "127.0.0.1" || ip == "::1" || strings.HasPrefix(ip, "localhost") {
		return IPGeoInfo{
			Country:     "Local",
			CountryCode: "LC",
			City:        "",
			ISP:         "Localhost",
			Org:         "",
			AS:          "",
			Source:      "local",
		}
	}

	if s.isPrivate(ip) {
		return IPGeoInfo{
			Country:     "Private",
			CountryCode: "PR",
			City:        "",
			ISP:         "Private Network",
			Org:         "",
			AS:          "",
			Source:      "local",
		}
	}

	s.mu.RLock()
	if cached, ok := s.cache[ip]; ok {
		s.mu.RUnlock()
		return cached
	}
	s.mu.RUnlock()

	info := s.lookupOnline(ip)

	s.mu.Lock()
	s.cache[ip] = info
	if len(s.cache) > 10000 {
		s.cache = make(map[string]IPGeoInfo)
	}
	s.mu.Unlock()

	return info
}

func (s *IPGeoService) lookupOnline(ip string) IPGeoInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/%s?fields=status,country,countryCode,city,isp,org,as", s.baseURL, ip)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return s.fallback(ip)
	}
	req.Header.Set("User-Agent", "HoneyWatch/1.0")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return s.fallback(ip)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return s.fallback(ip)
	}

	var result ipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return s.fallback(ip)
	}

	if result.Status != "success" {
		return s.fallback(ip)
	}

	log.Printf("[IPGeo] Lookup %s -> %s/%s %s", ip, result.Country, result.CountryCode, result.City)

	return IPGeoInfo{
		Country:     result.Country,
		CountryCode: result.CountryCode,
		City:        result.City,
		ISP:         result.ISP,
		Org:         result.Org,
		AS:          result.AS,
		Source:      "ip-api.com",
	}
}

func (s *IPGeoService) fallback(ip string) IPGeoInfo {
	info := IPGeoInfo{
		Country:     "Unknown",
		CountryCode: "XX",
		City:        "",
		ISP:         "",
		Org:         "",
		AS:          "",
		Source:      "iana",
	}

	firstOctet := s.getFirstOctet(ip)
	if firstOctet < 0 {
		return info
	}

	switch {
	case firstOctet >= 1 && firstOctet <= 3:
		info.Country = "North America"
		info.CountryCode = "US"
	case firstOctet >= 4 && firstOctet <= 11:
		info.Country = "North America"
		info.CountryCode = "US"
	case firstOctet >= 14 && firstOctet <= 16:
		info.Country = "Asia Pacific"
		info.CountryCode = "AU"
	case firstOctet >= 36 && firstOctet <= 61:
		info.Country = "China"
		info.CountryCode = "CN"
	case firstOctet >= 62 && firstOctet <= 95:
		info.Country = "Europe"
		info.CountryCode = "EU"
	case firstOctet >= 96 && firstOctet <= 126:
		info.Country = "North America"
		info.CountryCode = "US"
	case firstOctet >= 175 && firstOctet <= 191:
		info.Country = "Latin America"
		info.CountryCode = "BR"
	case firstOctet >= 192 && firstOctet <= 223:
		info.Country = "Asia Pacific"
		info.CountryCode = "JP"
	case firstOctet >= 224 && firstOctet <= 239:
		info.Country = "Multicast"
		info.CountryCode = "MC"
	}

	return info
}

func (s *IPGeoService) isPrivate(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}

	privateBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"fc00::/7",
		"fe80::/10",
	}

	for _, block := range privateBlocks {
		_, ipNet, err := net.ParseCIDR(block)
		if err == nil && ipNet.Contains(parsed) {
			return true
		}
	}

	return false
}

func (s *IPGeoService) getFirstOctet(ip string) int {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return -1
	}

	parts := strings.Split(parsed.String(), ".")
	if len(parts) < 1 {
		return -1
	}

	octet := 0
	for _, c := range parts[0] {
		if c < '0' || c > '9' {
			return -1
		}
		octet = octet*10 + int(c-'0')
	}
	return octet
}

func (s *IPGeoService) empty() IPGeoInfo {
	return IPGeoInfo{
		Country:     "Unknown",
		CountryCode: "XX",
		City:        "",
		ISP:         "",
		Org:         "",
		AS:          "",
		Source:      "none",
	}
}