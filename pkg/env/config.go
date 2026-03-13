package env

import "time"

type Config struct {
	ENV     string        `env:"ENV,required"`
	MAXUSER int           `env:"MAX_USERS,required"`
	TIMEOUT time.Duration `env:"TIMEOUT,required"`
}
