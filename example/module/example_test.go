package module_test

import (
	"log"

	"github.com/rwyyr/chariot"
	"github.com/rwyyr/chariot/example/module/internal/config"
	"github.com/rwyyr/chariot/example/module/internal/server"
)

func Module() chariot.Module {

	return chariot.WithOptions(
		config.Module(),
		server.Module(),
	)
}

func Example() {

	app, err := chariot.New(Module())
	if err != nil {
		log.Fatalf("Failed to create an app: %s\n", err)
	}
	defer app.Shutdown()

	if err := app.Run(); err != nil {
		log.Fatalf("Failed running the app: %s\n", err)
	}
}
