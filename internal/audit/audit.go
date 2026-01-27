package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type AuditEvent struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type Observer interface {
	Notify(event AuditEvent)
}

type FileObserver struct {
	filePath string
}

func NewFileObserver(filePath string) *FileObserver {
	return &FileObserver{
		filePath: filePath,
	}
}

func (o *FileObserver) Notify(event AuditEvent) {
	file, err := os.OpenFile(o.filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Failed to open audit file: %v", err)
		return
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal audit event: %v", err)
		return
	}

	if _, err := fmt.Fprintln(file, string(data)); err != nil {
		log.Printf("Failed to write to audit file: %v", err)
	}
}

type HTTPObserver struct {
	url string
}

func NewHTTPObserver(url string) *HTTPObserver {
	return &HTTPObserver{
		url: url,
	}
}

func (o *HTTPObserver) Notify(event AuditEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal audit event: %v", err)
		return
	}

	resp, err := http.Post(o.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Failed to send audit event: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Audit server returned non-OK status: %d", resp.StatusCode)
	}
}

type Subject struct {
	observers []Observer
}

func NewSubject() *Subject {
	return &Subject{
		observers: make([]Observer, 0),
	}
}

func (s *Subject) Attach(observer Observer) {
	s.observers = append(s.observers, observer)
}

func (s *Subject) Detach(observer Observer) {
	for i, obs := range s.observers {
		if obs == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

func (s *Subject) NotifyAll(event AuditEvent) {
	for _, observer := range s.observers {
		observer.Notify(event)
	}
}

func GetClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func CreateAuditEvent(r *http.Request, metrics []string) AuditEvent {
	return AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   metrics,
		IPAddress: GetClientIP(r),
	}
}
