package config

type Application struct {
	Addr        string
	HealthzAddr string
}

func NewApplication() (Application, error) {

	config := Application{
		Addr:        ":8080",
		HealthzAddr: ":8081",
	}

	return config, nil
}
