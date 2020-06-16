package controller

import (
	"github.com/octarinesec/octarine-operator/pkg/controller/octarine"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, octarine.Add)
}
