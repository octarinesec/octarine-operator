package controllers

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/processors"
)

type DefaultGatewayCreator struct {
	creator processors.APIGatewayCreator
}

func NewDefaultGatewayCreator() *DefaultGatewayCreator {
	return &DefaultGatewayCreator{
		creator: processors.NewDefaultGatewayCreator(),
	}
}

func (creator *DefaultGatewayCreator) CreateGateway(cbContainersAgent *cbcontainersv1.CBContainersAgent, accessToken string) Gateway {
	return creator.CreateGateway(cbContainersAgent, accessToken)
}
