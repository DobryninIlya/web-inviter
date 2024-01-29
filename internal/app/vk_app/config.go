package vk_app

type Config struct {
	BindAddr    string `toml:"bind_addr"`
	DatabaseURL string `toml:"database_url"`
	ShopID      int    `toml:"yookassa_shop_id"`
	APIKey      string `toml:"yookassa_api_key"`
	BotToken    string `toml:"bot_token"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		BindAddr: ":8080",
	}
}
