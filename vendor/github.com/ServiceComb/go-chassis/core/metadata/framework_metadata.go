package metadata

import (
	"sync"
)

// constant for sdk version, name, registration component
const (
	SdkVersion               = "1.0.3"
	SdkName                  = "Go-Chassis"
	SdkRegistrationComponent = "SDK"
)

// variables of micro-service framework, mutex variable
var (
	msFramework *Framework
	Once        = &sync.Once{}
)

// Framework is for to represents name, version, registration
type Framework struct {
	Name     string
	Version  string
	Register string
}

// SetName is to set the framework name
func (f *Framework) SetName(name string) {
	if f != nil {
		f.Name = name
	}
}

// SetVersion to set the version of framework
func (f *Framework) SetVersion(version string) {
	if f != nil {
		f.Version = version
	}
}

// SetRegister to register the framework
func (f *Framework) SetRegister(register string) {
	if f != nil {
		f.Register = register
	}
}

// NewFramework returns the object of msFramework
func NewFramework() *Framework {
	Once.Do(func() {
		msFramework = new(Framework)
		msFramework.Name = SdkName
		msFramework.Version = SdkVersion
		msFramework.Register = SdkRegistrationComponent

	})
	return msFramework
}
