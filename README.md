![](chariot.png)

# Overview

A minimalistic, zero-dependency, runtime-based DI application framework. It gives one means to engineer a service
application as a set of distinct components. Dependencies between components are resolved automatically. Runnable
components—servers, message consumers, and whatnots—are done so automatically in parallel. A proper clean-up mechanism
is provided. No more bloated mains. No more high effort of wiring a new part of an application with existing ones. The
framework simplifies one thing and does so right.

# Example

```go
package main

import (
    "log"
    "context"
    "net/http"

    "github.com/rwyyr/chariot"
)

type Config struct {
    Addr string
}

type Server struct {
    addr string
}

// The context is provided by default, and is configurable.
func NewConfig(context.Context) Config {

    return Config{
        Addr: ":8080",
    }
}

// The dependency is resolved automatically.
func NewServer(config Config) (*Server, error) {

    server := Server{
        addr: config.Addr,
    }

    http.HandleFunc("/hello", server.handleHello)

    // Returning an error terminates the initialization.
    return &server, nil
}

// Having the method makes a component runnable. The context is provided by default, and is configurable.
func (s *Server) Run(context.Context) error {

    // Returning an error terminates the run phase. The context passed to other runners is canceled thus.
    return http.ListenAndServe(s.addr, nil)
}

// Having the method registers a component for the clean-up phase. The context is provided by default, and is
// configurable.
func (*Server) Shutdown(context.Context) {

    // Do a clean-up.
}

func (*Server) handleHello(resp http.ResponseWriter, _ *http.Request) {

    resp.Write([]byte("world"))
}

func main() {

    // The initialization phase.
    app, err := chariot.New(
        NewConfig,
        NewServer,
    )
    if err != nil {
        log.Fatalf("Failed to create a new app: %s\n", err)
    }
    // The clean-up phase.
    defer app.Shutdown()

    // The run phase.
    if err := app.Run(); err != nil {
        log.Fatalf("Failed running the app: %s\n", err)
    }
}
```
