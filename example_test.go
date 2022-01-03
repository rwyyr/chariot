// MIT License
//
// Copyright (c) 2022 Roman Homoliako
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package chariot_test

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rwyyr/chariot"
)

type Config struct {
	ServerAddr string `json:"server_addr"`
}

type Server struct {
	addr string
}

func NewConfig(ctx context.Context, client *http.Client) (Config, error) {

	url := os.Getenv("CONFIG_URL")
	if url == "" {
		return Config{}, errors.New("config URL env wasn't set")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Config{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return Config{}, err
	}
	defer resp.Body.Close()

	var config Config
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func NewHTTPClient() *http.Client {

	return &http.Client{
		Timeout: 5 * time.Second,
	}
}

func NewServer(config Config) *Server {

	server := Server{
		addr: config.ServerAddr,
	}

	http.HandleFunc("/healthz", server.handleHealthz)

	return &server
}

func Example() {

	app, err := chariot.New(chariot.With(
		NewConfig,
		NewServer,
		NewHTTPClient,
	))
	if err != nil {
		log.Fatalf("Failed to create an app: %s\n", err)
	}
	defer app.Shutdown()

	if err := app.Run(); err != nil {
		log.Fatalf("Failed running the app: %s\n", err)
	}
}

func (s *Server) Run(context.Context) error {

	return http.ListenAndServe(s.addr, nil)
}

func (*Server) handleHealthz(resp http.ResponseWriter, _ *http.Request) {

	resp.Write([]byte("ok"))
}
