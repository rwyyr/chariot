package server

import (
	"context"
	"net/http"

	"github.com/rwyyr/chariot/example/module/config"
)

type Healthz struct {
	addr string
}

func NewHealthz(config config.Application) *Healthz {

	healthz := Healthz{
		addr: config.HealthzAddr,
	}

	http.HandleFunc("/healthz", healthz.handleHealthz)

	return &healthz
}

func (h *Healthz) Run(context.Context) error {

	return http.ListenAndServe(h.addr, nil)
}

func (*Healthz) handleHealthz(resp http.ResponseWriter, _ *http.Request) {

	resp.Write([]byte("ok"))
}
