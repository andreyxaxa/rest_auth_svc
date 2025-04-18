package server

type Config struct {
	Addr         string `toml:"addr"`
	LogLevel     string `toml:"log_level"`
	DatabaseURL  string `toml:"database_url"`
	JwtSecretKey string `toml:"jwt_secret"`
}

func NewConfig() *Config {
	return &Config{
		Addr:     ":8080",
		LogLevel: "debug",
		//JwtSecretKey: "mega-super-ultra-xxl-turbo-secret-key123",
	}
}
