package handler

import (
	"context"
	"log"
	"net/http"
	"time"
)

type DBPinger interface {
	Ping(ctx context.Context) error
}

type PingHandler struct {
	pinger DBPinger
}

func NewPingHandler(pinger DBPinger) *PingHandler {
	return &PingHandler{pinger: pinger}
}

func (h *PingHandler) PingHandler(w http.ResponseWriter, r *http.Request) {
	if h.pinger == nil {
		log.Println("Ping failed: pinger is nil")
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := h.pinger.Ping(ctx); err != nil {
		log.Printf("Ping failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
