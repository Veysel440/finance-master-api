package http

import (
	"net"
	"net/http"
	"strings"
)

func clientIP(r *http.Request) string {
	h := r.Header.Get("X-Forwarded-For")
	if h != "" {
		if i := strings.IndexByte(h, ','); i > 0 {
			return strings.TrimSpace(h[:i])
		}
		return strings.TrimSpace(h)
	}
	if ip := r.Header.Get("X-Real-Ip"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func clientUA(r *http.Request) string { return r.UserAgent() }
