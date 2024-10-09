package server

import (
	"github.com/hashicorp/go-plugin"
	"kusionstack.io/kusion/pkg/modules"

	"kusionstack.io/kusion-module-framework/pkg/module"
)

// HandshakeConfig is a common handshake that is shared by plugin and host.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "MODULE_PLUGIN",
	MagicCookieValue: "ON",
}

func Start(m module.FrameworkModule) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			modules.PluginKey: &modules.GRPCPlugin{Impl: &module.FrameworkModuleWrapper{Module: m}},
		},
		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
