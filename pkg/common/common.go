package common

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

var (
	DATABASE_URL   = ""
	WsAddr         = "0.0.0.0:9998"
	CurrentVersion = ""
)

type EnvConfigSpec struct {
	Debug                bool   `envconfig:"DEBUG" default:"false"`
	DebugRequests        bool   `envconfig:"DEBUG_REQUESTS" default:"false"`
	DebugSQL             bool   `envconfig:"DEBUG_SQL" default:"false"`
	DebugWS              bool   `envconfig:"DEBUG_WS" default:"false"`
	UIUsername           string `envconfig:"UI_USERNAME" required:"false"`
	UIPassword           string `envconfig:"UI_PASSWORD" required:"false"`
	DatabaseURL          string `envconfig:"DATABASE_URL" required:"false" default:""`
	WsAddr               string `envconfig:"XBVR_WS_ADDR" required:"false" default:""`
	WebPort              int    `envconfig:"XBVR_WEB_PORT" required:"false" default:"0"`
	DBConnectionPoolSize int    `envconfig:"DB_CONNECTION_POOL_SIZE" required:"false" default:"0"`
	ConcurrentScrapers   int    `envconfig:"CONCURRENT_SCRAPERS" required:"false" default:"9999"`
}

var EnvConfig EnvConfigSpec

func init() {
	godotenv.Load()

	err := envconfig.Process("", &EnvConfig)
	if err != nil {
		Log.Fatal(err.Error())
	}
}
