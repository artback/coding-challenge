package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// App represents the server's internal state.
// It holds configuration about providers and content
type App struct {
	ContentClients map[Provider]Client
	Config         ContentMix
}

func (a App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	query := req.URL.Query()
	count, err := strconv.Atoi(query.Get("count"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	offset, err := strconv.Atoi(query.Get("offset"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	ip, err := getIP(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	param := Parameters{ip: ip, offset: offset, count: count}
	defer func() {
		if r := recover(); r != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}()
	json.NewEncoder(w).Encode(GetContentItems(a, param))
}

func getIP(r *http.Request) (string, error) {
	//Get IP from the X-REAL-IP header
	ip := r.Header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	//Get IP from X-FORWARDED-FOR header
	ips := r.Header.Get("X-FORWARDED-FOR")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip, nil
		}
	}

	//Get IP from RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}
	return "", fmt.Errorf("No valid ip found")
}
