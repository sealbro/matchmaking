package api

type PrivateApiConfig struct {
	Address string `env:"PRIVATE_ADDRESS, default=:8081"`
}
