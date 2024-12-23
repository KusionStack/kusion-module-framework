package registry

import "kcl-lang.io/kpm/pkg/client"

const (
	EnvKusionModuleRegistryHost     = "KUSION_MODULE_REGISTRY_HOST"
	EnvKusionModuleRegistryUsername = "KUSION_MODULE_REGISTRY_USERNAME"
	EnvKusionModuleRegistryPassword = "KUSION_MODULE_REGISTRY_PASSWORD"
)

// KusionModuleClient is the client of Kusion Module Registry.
type KusionModuleClient struct {
	*client.KpmClient
}
