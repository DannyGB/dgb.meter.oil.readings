//go:build wireinject
// +build wireinject

package main

import (
	"dgb/meter.oil.readings/internal/application"
	"dgb/meter.oil.readings/internal/configuration"
	"dgb/meter.oil.readings/internal/database"

	"github.com/google/wire"
)

func CreateApi() *application.ReadingApi {

	panic(wire.Build(
		configuration.NewMeterEnvironment,
		configuration.NewConfig,
		application.NewResponse,
		database.NewRepository,
		application.NewMiddleware,
		application.NewApi,
	))
}
