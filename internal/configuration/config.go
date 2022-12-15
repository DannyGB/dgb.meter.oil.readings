package configuration

import "os"

type Configuration struct {
	MONGO_CONNECTION string
	MONGO_DB         string
	MONGO_COLLECTION string
	HTTP_PORT        string
	CORS_CLIENTS     string
}

func NewConfig(env MeterEnvironment) Configuration {

	if env == "" {
		panic("No environment supplied to load config")
	}

	configuration := &Configuration{}
	configuration.CORS_CLIENTS = os.Getenv("METER_OIL_READINGS_CORS_CLIENTS")
	configuration.HTTP_PORT = os.Getenv("METER_OIL_READINGS_HTTP_PORT")
	configuration.MONGO_COLLECTION = os.Getenv("METER_OIL_READINGS_MONGO_COLLECTION")
	configuration.MONGO_DB = os.Getenv("METER_READINGS_MONGO_DB")
	configuration.MONGO_CONNECTION = os.Getenv("METER_READINGS_MONGO_CONNECTION")

	return *configuration
}
