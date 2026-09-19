package validator

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net"
)

type Validator struct {
	secret     string
	allowedIPs []*net.IPNet
}

func New(secret string, allowedIPs []*net.IPNet) *Validator {
	return &Validator{
		secret:     secret,
		allowedIPs: allowedIPs,
	}
}

func (v *Validator) ValidateSignature(signature string, body []byte) bool {
	if v.secret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(v.secret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

func (v *Validator) ValidateIP(addr string) bool {
	if len(v.allowedIPs) == 0 {
		return true
	}

	ipStr := extractIP(addr)
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	for _, ipNet := range v.allowedIPs {
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}

func extractIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
