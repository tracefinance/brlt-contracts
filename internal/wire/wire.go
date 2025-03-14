//go:build wireinject

package wire

import (
	"vault0/internal/api"

	"github.com/google/wire"
)

// Container holds all application dependencies organized by layer
type Container struct {
	Core     *Core
	Server   *api.Server
	Services *Services
}

// NewContainer creates a new dependency injection container
func NewContainer(
	core *Core,
	server *api.Server,
	services *Services,
) *Container {
	return &Container{
		Core:     core,
		Server:   server,
		Services: services,
	}
}

// ContainerSet combines all dependency sets
var ContainerSet = wire.NewSet(
	CoreSet,
	ServerSet,
	ServicesSet,
	NewContainer,
)

// InitializeContainer creates a new container with all dependencies wired up
// BuildContainer is a placeholder function that will be replaced by wire with the actual implementation
func BuildContainer() (*Container, error) {
	wire.Build(ContainerSet)
	return nil, nil
}
