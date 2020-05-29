package controller

import (
	v2alpha2 "github.com/rh-messaging/activemq-artemis-operator/pkg/controller/broker/v2alpha2/activemqartemisaddress"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, v2alpha2.Add)
}
