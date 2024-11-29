package module

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"kusionstack.io/kusion-module-framework/pkg/module/proto"
)

const PluginKey = "module-default"

// Module is the interface that we're exposing as a kusion module plugin.
type Module interface {
	Generate(ctx context.Context, req *proto.GeneratorRequest) (*proto.GeneratorResponse, error)
}

type GRPCClient struct {
	client proto.ModuleClient
}

func (c *GRPCClient) Generate(ctx context.Context, req *proto.GeneratorRequest) (*proto.GeneratorResponse, error) {
	return c.client.Generate(ctx, req)
}

type GRPCServer struct {
	// This is the real implementation
	Impl Module
	proto.UnimplementedModuleServer
}

func (s *GRPCServer) Generate(ctx context.Context, req *proto.GeneratorRequest) (res *proto.GeneratorResponse, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.WithStack(err)
			res = &proto.GeneratorResponse{}
		}
	}()
	res, err = s.Impl.Generate(ctx, req)
	return
}

type GRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins that are written in Go.
	Impl Module
}

// GRPCServer is going to be invoked by the go-plugin framework
func (p *GRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterModuleServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

// GRPCClient is going to be invoked by the go-plugin framework
func (p *GRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewModuleClient(c)}, nil
}
