package application

import (
	"errors"
	api "github.com/lxc/incus/v6/shared/api"

	"log/slog"
)

// CreateNetworks creates the default network for a stack
func (c *Compose) CreateNetworks() error {
	slog.Info("Creating networks")

	for key, network := range c.ComposeProject.Networks {
		if network.External {
			continue
		}
		var nettype string
		var uplink string

		if ok, err := network.Extensions.Get("x-incus-type", &nettype); !ok || err != nil {
			nettype = c.Network.Type
		}

		if ok, err := network.Extensions.Get("x-incus-uplink", &uplink); !ok || err != nil {
			uplink = c.Network.Uplink
		}

		slog.Info("Creating network",
			slog.String("key", key),
			slog.String("name", network.Name),
			slog.String("type", nettype),
			slog.String("uplink", uplink),
		)

		var apiNetwork api.NetworksPost

		apiNetwork.Name = network.Name
		apiNetwork.Type = nettype
		apiNetwork.Config = map[string]string{}

		if nettype == "ovn" {
			apiNetwork.Config["network"] = uplink
		}

		// Parse remote
		resources, err := c.ParseServers(network.Name)
		if err != nil {
			return err
		}

		resource := resources[0]
		client := resource.server

		// Create the network
		err = client.CreateNetwork(apiNetwork)
		if err != nil {
			return err
		}

		slog.Info("Network created", slog.String("name", network.Name))
	}

	return nil
}

// DestroyNetworks destroys the default network for a stack
func (c *Compose) DestroyNetworks() error {
	slog.Info("Destroying networks")

	var funcError error

	for key, network := range c.ComposeProject.Networks {
		if network.External {
			continue
		}

		slog.Info("Destroying network", slog.String("key", key), slog.String("name", network.Name))

		resources, err := c.ParseServers(network.Name)
		if err != nil {
			funcError = errors.Join(funcError, err)
		}

		resource := resources[0]

		// Delete the network
		err = resource.server.DeleteNetwork(resource.name)
		if err != nil {
			funcError = errors.Join(funcError, err)
		}
		slog.Info("Destroyed network", slog.String("name", network.Name))
	}
	return funcError
}
