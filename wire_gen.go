// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"lab-api/app"
	"lab-api/app/api"
	"lab-api/app/api/controller"
	"lab-api/app/api/service"
	"lab-api/app/system"
	controller2 "lab-api/app/system/controller"
	service2 "lab-api/app/system/service"
	"lab-api/bootstrap"
	"lab-api/common"
)

// Injectors from wire.go:

func App(config common.Config) (*app.App, error) {
	engine := bootstrap.HttpServer(config)
	db, err := bootstrap.InitializeDatabase(config)
	if err != nil {
		return nil, err
	}
	client, err := bootstrap.InitializeRedis(config)
	if err != nil {
		return nil, err
	}
	dependency := &service.Dependency{
		Db:    db,
		Redis: client,
	}
	index := service.NewIndex(dependency)
	controllerDependency := &controller.Dependency{
		IndexService: index,
	}
	controllerIndex := controller.NewIndex(controllerDependency)
	apiDependency := &api.Dependency{
		Index: controllerIndex,
	}
	routes := api.NewRoutes(engine, apiDependency)
	serviceDependency := service2.Dependency{
		Config: config,
		Db:     db,
		Redis:  client,
	}
	serviceIndex := service2.NewIndex(serviceDependency)
	dependency2 := &controller2.Dependency{
		Db:           db,
		IndexService: serviceIndex,
	}
	index2 := controller2.NewIndex(dependency2)
	admin := controller2.NewAdmin(dependency2)
	systemDependency := &system.Dependency{
		Config: config,
		Index:  index2,
		Admin:  admin,
	}
	systemRoutes := system.NewRoutes(engine, systemDependency)
	appApp := app.NewApp(engine, routes, systemRoutes)
	return appApp, nil
}
