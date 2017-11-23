// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package config

import (
	"fmt"
	"os"
	"sync"
)

const (
	// ApplicationIDKey is the identifier of an application ID
	ApplicationIDKey = "applicationID"
	// ApplicationDescriptionKey is the configuration key of the application's
	// description
	ApplicationDescriptionKey = "applicationDesc"
	// ApplicationOwnerKey is the configuration key for an application's owner
	ApplicationOwnerKey = "applicationOwner"

	environment = "_ENVIRONMENT"
	datacenter  = "_DATACENTER"
	configdir   = "_CONFIG_DIR"
	config      = "config"
)

// TODO(ai) underscore-prefix these per Uber style
var (
	global   ConfigurationProvider
	locked   bool
	setupMux sync.Mutex

	_envPrefix          = "APP"
	configProviderFuncs = []ProviderFunc{YamlProvider(), EnvProvider()}
	cpMux               sync.Mutex
)

// Global returns the singleton configuration provider
func Global() ConfigurationProvider {
	setupMux.Lock()
	defer setupMux.Unlock()
	locked = true
	return global
}

// ServiceName returns the service's names
func ServiceName() string {
	return Global().GetValue(ApplicationIDKey).AsString()
}

// SetGlobal sets the singleton configuration provider
func SetGlobal(provider ConfigurationProvider, force bool) {
	setupMux.Lock()
	defer setupMux.Unlock()
	if locked && !force {
		panic("Global provider must be set before any configuration access")
	}
	global = provider
}

// ResetGlobal is used for tests
func ResetGlobal() {
	setupMux.Lock()
	defer setupMux.Unlock()
	global = nil
}

func getConfigFiles() []string {
	env := GetEnvironment()
	dc := os.Getenv(GetEnvironmentPrefix() + datacenter)

	var files []string
	if dc != "" && env != "" {
		files = append(files, fmt.Sprintf("./%s/%s-%s.yaml", config, env, dc))
	}
	files = append(files,
		fmt.Sprintf("./%s/%s.yaml", config, env),
		fmt.Sprintf("./%s/base.yaml", config))

	return files
}

func getResolver() FileResolver {
	paths := []string{}
	configDir := os.Getenv(GetEnvironmentPrefix() + configdir)
	if configDir != "" {
		paths = []string{configDir}
	}
	return NewRelativeResolver(paths...)
}

// YamlProvider returns function to create Yaml based configuration provider
func YamlProvider() ProviderFunc {
	return func() (ConfigurationProvider, error) {
		return NewYAMLProviderFromFiles(false, getResolver(), getConfigFiles()...), nil
	}
}

// EnvProvider returns function to create environment based config provider
func EnvProvider() ProviderFunc {
	return func() (ConfigurationProvider, error) {
		return NewEnvProvider(defaultEnvPrefix, nil), nil
	}
}

// GetEnvironment returns current environment setup for the service
func GetEnvironment() string {
	env := os.Getenv(GetEnvironmentPrefix() + environment)
	if env == "" {
		env = "development"
	}
	return env
}

// SetEnvironmentPrefix sets environment prefix for the application
func SetEnvironmentPrefix(envPrefix string) {
	_envPrefix = envPrefix
}

// GetEnvironmentPrefix returns environment prefix for the application
func GetEnvironmentPrefix() string {
	return _envPrefix
}

// ProviderFunc is used to create config providers on configuration initialization
type ProviderFunc func() (ConfigurationProvider, error)

// RegisterProviders registers configuration providers for the global config
func RegisterProviders(providerFuncs ...ProviderFunc) {
	cpMux.Lock()
	defer cpMux.Unlock()
	configProviderFuncs = append(configProviderFuncs, providerFuncs...)
}

// UnregisterProviders clears all the default providers
func UnregisterProviders() {
	cpMux.Lock()
	defer cpMux.Unlock()
	configProviderFuncs = nil
}

// InitializeGlobalConfig initializes the ConfigurationProvider for use in a service
func InitializeGlobalConfig() {
	var providers []ConfigurationProvider
	for _, providerFunc := range configProviderFuncs {
		cp, err := providerFunc()
		if err != nil {
			panic(err)
		}
		providers = append(providers, cp)
	}
	global = NewProviderGroup("global", providers...)
}
