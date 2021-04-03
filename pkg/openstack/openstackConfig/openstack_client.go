package openstackConfig

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/acceptance/openstack/networking/v2/extensions/mtu"
	"github.com/gophercloud/gophercloud/acceptance/openstack/networking/v2/extensions/portsbinding"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	geportsbinding "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/portsbinding"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/pagination"
	"k8s.io/klog"
)

type OSClient struct {
	region string
	providerClient *gophercloud.ProviderClient
	netClient *gophercloud.ServiceClient
}

//def setup_openstacksdk():

func NewOSClient(opts *gophercloud.AuthOptions, region string) (*OSClient, error){
	providerClient, err := openstack.AuthenticatedClient(*opts)
	if err != nil {
		klog.Errorf("Create Openstack Provider Client with %v Error: %s", *opts, err)
		return nil, err
	}

	netClient, err := openstack.NewNetworkV2(providerClient, gophercloud.EndpointOpts{Region: region})
	if err != nil {
		klog.Errorf("NewNetworkV2 Error: %v!\n\n" , err)
		return nil, err
	}
	osClient := &OSClient{
		providerClient: providerClient,
		region: region,
		netClient: netClient,
	}

	klog.Infof("netClient: %+v\n\n" , netClient.Endpoint)

	//for test
	//osClient.ProjectList()
	//osClient.NetworkList()
	return osClient, err
}

func (c *OSClient) CreatePort(opts geportsbinding.CreateOptsExt)(* portsbinding.PortWithBindingExt, error){
	portExt := &portsbinding.PortWithBindingExt{}
	err := ports.Create(c.netClient, opts).ExtractInto(portExt)
	if err != nil{
		klog.Errorf("Create port Failed. opts: %+v, Error: %v\n", opts, err)
	}
	return portExt, err
}

func (c *OSClient) GetPort(id string)(* portsbinding.PortWithBindingExt, error){
	portExt := &portsbinding.PortWithBindingExt{}
	err := ports.Get(c.netClient, id).ExtractInto(portExt)

	return portExt, err
}

func (c *OSClient) DeletePort(id string) error {
	return ports.Delete(c.netClient, id).Err
}

func (c *OSClient) GetNetwork(id string) (*mtu.NetworkMTU, error) {
	netExt := &mtu.NetworkMTU{}
	err := networks.Get(c.netClient, id).ExtractInto(netExt)
	return netExt, err
}

func (c *OSClient) GetSubnet(id string) (*subnets.Subnet, error) {
	return subnets.Get(c.netClient, id).Extract()
}

func (c *OSClient) ListNetwork(){
	pager := networks.List(c.netClient, networks.ListOpts{})
	pager.EachPage(func(page pagination.Page) (bool, error) {
		networks, _ := networks.ExtractNetworks(page)
		for _, n := range networks {
			klog.Infof("network: %v, %v, %v\n", n.ID, n.Name, n.Subnets)
		}
		return true, nil
	})
}

func (c *OSClient) ProjectList(){
	if client, err := openstack.NewIdentityV3(c.providerClient, gophercloud.EndpointOpts{}); err == nil{
		allPages, _ := projects.List(client, projects.ListOpts{}).AllPages()
		allProjects, _ := projects.ExtractProjects(allPages)

		for _, project := range allProjects {
			klog.Infof("ProjectName: %v\n\n", project.Name)
		}
	}
}

