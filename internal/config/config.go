package config

type Config struct {
	Env string `env:"ENV,required"`
}
