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
	"log"
	"net/http"
	"time"

	"github.com/rwyyr/chariot"
)

func ExampleNew() {
	app, err := chariot.New(chariot.With(
		NewConfig,
		NewServer,
		NewHTTPClient,
	))
	if err != nil {
		log.Fatalf("Failed to create an app: %s\n", err)
	}
	defer app.Shutdown()
}

func ExampleApp_Run() {
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

func ExampleApp_Shutdown() {
	app, err := chariot.New(chariot.With(
		NewConfig,
		NewServer,
		NewHTTPClient,
	))
	if err != nil {
		log.Fatalf("Failed to create an app: %s\n", err)
	}
	defer app.Shutdown()
}

func ExampleApp_Retrieve() {
	app, err := chariot.New(chariot.With(
		NewConfig,
		NewServer,
		NewHTTPClient,
	))
	if err != nil {
		log.Fatalf("Failed to create an app: %s\n", err)
	}
	defer app.Shutdown()

	var config Config
	if !app.Retrieve(&config) {
		log.Fatalln("Failed to retrieve a config")
	}

	log.Printf("The server will listen on %s\n", config.ServerAddr)
}

func ExampleFuncRunner() {
	type server chariot.Runner

	newServer := func(config Config) server {

		http.HandleFunc("/foo", func(resp http.ResponseWriter, _ *http.Request) {

			resp.Write([]byte("bar"))
		})

		return chariot.FuncRunner(func(context.Context) error {

			return http.ListenAndServe(config.ServerAddr, nil)
		})
	}

	app, err := chariot.New(chariot.With(
		NewConfig,
		newServer,
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

func ExampleWith() {
	app, err := chariot.New(chariot.With(
		NewConfig,
		NewServer,
		NewHTTPClient,
	))
	if err != nil {
		log.Fatalf("Failed to create an app: %s\n", err)
	}
	defer app.Shutdown()
}

func ExampleWithComponents() {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	app, err := chariot.New(
		chariot.With(
			NewConfig,
			NewServer,
		),
		chariot.WithComponents(&client),
	)
	if err != nil {
		log.Fatalf("Failed to create an app: %s\n", err)
	}
	defer app.Shutdown()
}
