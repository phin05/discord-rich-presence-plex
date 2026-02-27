package api

import (
	"drpp/server/logger"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

var localNets = parseCidrs([]string{"127.0.0.0/8", "::1/128"})

func parseCidrs(cidrs []string) []*net.IPNet {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, net, err := net.ParseCIDR(cidr)
		if err != nil {
			logger.Warning("Invalid CIDR %q: %v", cidr, err)
			continue
		}
		nets = append(nets, net)
	}
	return nets
}

func ipCheckMiddleware(next http.Handler, allowedNetworks []string, trustedProxies []string) http.Handler {
	allowedNets := parseCidrs(allowedNetworks)
	trustedProxyNets := parseCidrs(trustedProxies)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil || host == "" {
			if host == "" {
				err = errors.New("empty host")
			}
			handleError(w, r, fmt.Errorf("split address %s: %w", r.RemoteAddr, err))
			return
		}
		ip := net.ParseIP(host)
		if ip == nil {
			handleError(w, r, fmt.Errorf("parse IP %s", host))
			return
		}
		xff := r.Header.Get("X-Forwarded-For")
		if xff != "" && isAllowed(ip, trustedProxyNets) {
			xffIps := strings.Split(xff, ",")
			// Iterate in reverse to find the first non-trusted-proxy IP, which would be the client's real IP
			for i := len(xffIps) - 1; i >= 0; i-- {
				xffIp := strings.TrimSpace(xffIps[i])
				ip = net.ParseIP(xffIp)
				if ip == nil {
					handleError(w, r, ErrBadRequest(fmt.Sprintf("Failed to parse X-Forwarded-For IP %s", xffIp), nil))
					return
				}
				if !isAllowed(ip, trustedProxyNets) {
					break
				}
			}
		}
		if !isAllowed(ip, localNets) && !isAllowed(ip, allowedNets) {
			logger.Warning("%s %s: Denying IP %s", r.Method, r.URL.Path, ip)
			handleError(w, r, ErrForbidden(fmt.Sprintf("IP %s not allowed", ip)))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isAllowed(ip net.IP, nets []*net.IPNet) bool {
	for _, net := range nets {
		if net.Contains(ip) {
			return true
		}
	}
	return false
}

func securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubdomains")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "frame-ancestors 'none'; default-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self' https://api.github.com https://plex.tv https://clients.plex.tv")
		next.ServeHTTP(w, r)
	})
}

const devOrigin = "http://localhost:5173"

func devCorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") != devOrigin {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", devOrigin)
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
