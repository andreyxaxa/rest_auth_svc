package server

type Config struct {
	Addr string `toml:"addr"`
}

func NewConfig() *Config {
	return &Config{
		Addr: ":8080",
	}
}
