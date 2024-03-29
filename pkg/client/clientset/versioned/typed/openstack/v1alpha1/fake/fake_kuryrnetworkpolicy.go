/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	v1alpha1 "projectkuryr/kuryr/pkg/apis/openstack/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeKuryrNetworkPolicies implements KuryrNetworkPolicyInterface
type FakeKuryrNetworkPolicies struct {
	Fake *FakeOpenstackV1alpha1
	ns   string
}

var kuryrnetworkpoliciesResource = schema.GroupVersionResource{Group: "openstack.org", Version: "v1alpha1", Resource: "kuryrnetworkpolicies"}

var kuryrnetworkpoliciesKind = schema.GroupVersionKind{Group: "openstack.org", Version: "v1alpha1", Kind: "KuryrNetworkPolicy"}

// Get takes name of the kuryrNetworkPolicy, and returns the corresponding kuryrNetworkPolicy object, and an error if there is any.
func (c *FakeKuryrNetworkPolicies) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.KuryrNetworkPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(kuryrnetworkpoliciesResource, c.ns, name), &v1alpha1.KuryrNetworkPolicy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KuryrNetworkPolicy), err
}

// List takes label and field selectors, and returns the list of KuryrNetworkPolicies that match those selectors.
func (c *FakeKuryrNetworkPolicies) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.KuryrNetworkPolicyList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(kuryrnetworkpoliciesResource, kuryrnetworkpoliciesKind, c.ns, opts), &v1alpha1.KuryrNetworkPolicyList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.KuryrNetworkPolicyList{ListMeta: obj.(*v1alpha1.KuryrNetworkPolicyList).ListMeta}
	for _, item := range obj.(*v1alpha1.KuryrNetworkPolicyList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested kuryrNetworkPolicies.
func (c *FakeKuryrNetworkPolicies) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(kuryrnetworkpoliciesResource, c.ns, opts))

}

// Create takes the representation of a kuryrNetworkPolicy and creates it.  Returns the server's representation of the kuryrNetworkPolicy, and an error, if there is any.
func (c *FakeKuryrNetworkPolicies) Create(ctx context.Context, kuryrNetworkPolicy *v1alpha1.KuryrNetworkPolicy, opts v1.CreateOptions) (result *v1alpha1.KuryrNetworkPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(kuryrnetworkpoliciesResource, c.ns, kuryrNetworkPolicy), &v1alpha1.KuryrNetworkPolicy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KuryrNetworkPolicy), err
}

// Update takes the representation of a kuryrNetworkPolicy and updates it. Returns the server's representation of the kuryrNetworkPolicy, and an error, if there is any.
func (c *FakeKuryrNetworkPolicies) Update(ctx context.Context, kuryrNetworkPolicy *v1alpha1.KuryrNetworkPolicy, opts v1.UpdateOptions) (result *v1alpha1.KuryrNetworkPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(kuryrnetworkpoliciesResource, c.ns, kuryrNetworkPolicy), &v1alpha1.KuryrNetworkPolicy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KuryrNetworkPolicy), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeKuryrNetworkPolicies) UpdateStatus(ctx context.Context, kuryrNetworkPolicy *v1alpha1.KuryrNetworkPolicy, opts v1.UpdateOptions) (*v1alpha1.KuryrNetworkPolicy, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(kuryrnetworkpoliciesResource, "status", c.ns, kuryrNetworkPolicy), &v1alpha1.KuryrNetworkPolicy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KuryrNetworkPolicy), err
}

// Delete takes name of the kuryrNetworkPolicy and deletes it. Returns an error if one occurs.
func (c *FakeKuryrNetworkPolicies) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(kuryrnetworkpoliciesResource, c.ns, name), &v1alpha1.KuryrNetworkPolicy{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeKuryrNetworkPolicies) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(kuryrnetworkpoliciesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.KuryrNetworkPolicyList{})
	return err
}

// Patch applies the patch and returns the patched kuryrNetworkPolicy.
func (c *FakeKuryrNetworkPolicies) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.KuryrNetworkPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(kuryrnetworkpoliciesResource, c.ns, name, pt, data, subresources...), &v1alpha1.KuryrNetworkPolicy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.KuryrNetworkPolicy), err
}
