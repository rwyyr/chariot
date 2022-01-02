package server

import (
	"context"
	"net/http"

	"github.com/rwyyr/chariot/example/module/internal/config"
)

type API struct {
	addr string
}

func NewAPI(config config.Application) *API {

	api := API{
		addr: config.Addr,
	}

	http.HandleFunc("/foo", api.handleFoo)

	return &api
}

func (a *API) Run(ctx context.Context) error {

	return http.ListenAndServe(a.addr, nil)
}

func (*API) handleFoo(resp http.ResponseWriter, _ *http.Request) {

	resp.Write([]byte("bar"))
}
