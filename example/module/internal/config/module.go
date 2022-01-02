package config

import (
	"github.com/rwyyr/chariot"
)

func Module() chariot.Module {

	return chariot.With(NewApplication)
}
