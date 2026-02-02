package audit

import (
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

func TestFileObserverConcurrency(t *testing.T) {
	testFile := "/tmp/test_audit_concurrent.log"
	defer os.Remove(testFile)

	observer := NewFileObserver(testFile)
	event := AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   []string{"test-metric"},
		IPAddress: "127.0.0.1",
	}

	var wg sync.WaitGroup
	// Write from multiple goroutines
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			observer.Notify(event)
		}()
	}
	wg.Wait()

	// Verify file exists and has content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	lines := 0
	for _, b := range content {
		if b == '\n' {
			lines++
		}
	}

	if lines != 10 {
		t.Errorf("Expected 10 lines, got %d", lines)
	}
}

func TestHTTPObserverWithMockServer(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observer := NewHTTPObserver(server.URL)
	event := AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   []string{"test-metric"},
		IPAddress: "127.0.0.1",
	}

	// Send notification - this should succeed with retries
	observer.Notify(event)
	time.Sleep(100 * time.Millisecond) // Allow time for retries
}

func TestHTTPObserverWithFailingServer(t *testing.T) {
	// Create a mock server that returns 500 error
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	observer := NewHTTPObserver(server.URL)
	event := AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   []string{"test-metric"},
		IPAddress: "127.0.0.1",
	}

	// Send notification - should retry and eventually fail
	observer.Notify(event)
	time.Sleep(100 * time.Millisecond)

	// Check that retries happened (default is max 3)
	if attempts < 1 {
		t.Error("Expected at least 1 attempt, got 0")
	}
}

func TestSubjectConcurrency(t *testing.T) {
	subject := NewSubject()
	testFile := "/tmp/test_audit_subject.log"
	defer os.Remove(testFile)
	observer := NewFileObserver(testFile)
	event := AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   []string{"test-metric"},
		IPAddress: "127.0.0.1",
	}

	var wg sync.WaitGroup

	// Attach observers concurrently
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			subject.Attach(observer)
		}()
	}
	wg.Wait()

	// Notify all concurrently
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			subject.NotifyAll(event)
		}()
	}
	wg.Wait()

	// Detach observers concurrently
	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			subject.Detach(observer)
		}()
	}
	wg.Wait()
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		forwarded      string
		realIP         string
		remoteAddr     string
		expectedResult string
	}{
		{
			name:           "X-Forwarded-For with multiple IPs",
			forwarded:      "203.0.113.1, 203.0.113.2, 203.0.113.3",
			realIP:         "",
			remoteAddr:     "192.168.1.1:12345",
			expectedResult: "203.0.113.1",
		},
		{
			name:           "X-Real-IP header",
			forwarded:      "",
			realIP:         "203.0.113.1",
			remoteAddr:     "192.168.1.1:12345",
			expectedResult: "203.0.113.1",
		},
		{
			name:           "RemoteAddr fallback",
			forwarded:      "",
			realIP:         "",
			remoteAddr:     "192.168.1.1:12345",
			expectedResult: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.forwarded != "" {
				req.Header.Set("X-Forwarded-For", tt.forwarded)
			}
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}
			req.RemoteAddr = tt.remoteAddr

			result := GetClientIP(req)
			if result != tt.expectedResult {
				t.Errorf("Expected %s, got %s", tt.expectedResult, result)
			}
		})
	}
}
