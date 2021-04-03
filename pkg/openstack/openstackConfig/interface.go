package openstackConfig

import (
	"github.com/gophercloud/gophercloud/acceptance/openstack/networking/v2/extensions/mtu"
	"github.com/gophercloud/gophercloud/acceptance/openstack/networking/v2/extensions/portsbinding"
	geportsbinding "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/portsbinding"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
)

type Interface interface {
	DeletePort(id string) error
	CreatePort(opts geportsbinding.CreateOptsExt)(* portsbinding.PortWithBindingExt, error)
	GetPort(id string)(* portsbinding.PortWithBindingExt, error)

	GetNetwork(id string) (*mtu.NetworkMTU, error)
	GetSubnet(id string) (*subnets.Subnet, error)
}
