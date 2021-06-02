// Copyright 2019 Antrea Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	corev1 "k8s.io/api/core/v1"
	kuryrv1alpha1 "projectkuryr/kuryr/pkg/apis/openstack/v1alpha1"
	binding "projectkuryr/kuryr/pkg/ovs/openflow"
)

type MatchKey struct {
	ofProtocol    binding.Protocol
	valueCategory AddressCategory
	keyString     string
}

func (m *MatchKey) GetOFProtocol() binding.Protocol {
	return m.ofProtocol
}

func (m *MatchKey) GetValueCategory() AddressCategory {
	return m.valueCategory
}

func (m *MatchKey) GetKeyString() string {
	return m.keyString
}

func NewMatchKey(proto binding.Protocol, valueCategory AddressCategory, keyString string) *MatchKey {
	return &MatchKey{
		keyString:     keyString,
		ofProtocol:    proto,
		valueCategory: valueCategory,
	}
}

type AddressCategory uint8

const (
	IPAddr AddressCategory = iota
	IPNetAddr
	OFPortAddr
	L4PortAddr
	UnSupported
)

type AddressType int

const (
	SrcAddress AddressType = iota
	DstAddress
)

type Address interface {
	GetMatchValue() string
	GetMatchKey(addrType AddressType) *MatchKey
	GetValue() interface{}
}

// PolicyRule groups configurations to set up conjunctive match for egress/ingress policy rules.
type PolicyRule struct {
	Direction     kuryrv1alpha1.Direction
	From          []Address
	To            []Address
	Service       []kuryrv1alpha1.Service
	Action        *kuryrv1alpha1.RuleAction
	Priority      *uint16
	Name          string
	FlowID        uint32
	TableID       binding.TableIDType
	PolicyRef     *kuryrv1alpha1.NetworkPolicyReference
	EnableLogging bool
}

// IsAntreaNetworkPolicyRule returns if a PolicyRule is created for Antrea NetworkPolicy types.
func (r *PolicyRule) IsAntreaNetworkPolicyRule() bool {
	return true
}

// Priority is a struct that is composed of Antrea NetworkPolicy priority, rule priority and Tier priority.
// It is used as the basic unit for priority sorting.
type Priority struct {
	TierPriority   int32
	PolicyPriority float64
	RulePriority   int32
}

func (p *Priority) Less(p2 Priority) bool {
	if p.TierPriority == p2.TierPriority {
		if p.PolicyPriority == p2.PolicyPriority {
			return p.RulePriority > p2.RulePriority
		}
		return p.PolicyPriority > p2.PolicyPriority
	}
	return p.TierPriority > p2.TierPriority
}

func (p *Priority) Equals(p2 Priority) bool {
	return p.TierPriority == p2.TierPriority && p.PolicyPriority == p2.PolicyPriority && p.RulePriority == p2.RulePriority
}

// InSamePriorityZone returns true if two Priorities are of the same Tier and same priority at policy level.
func (p *Priority) InSamePriorityZone(p2 Priority) bool {
	return p.PolicyPriority == p2.PolicyPriority && p.TierPriority == p2.TierPriority
}

// IsConsecutive returns true if two Priorties are immediately next to each other.
func (p *Priority) IsConsecutive(p2 Priority) bool {
	if !p.InSamePriorityZone(p2) {
		return false
	}
	return p.RulePriority-p2.RulePriority == 1 || p2.RulePriority-p.RulePriority == 1
}

// ByPriority sorts a list of Priority by their relative TierPriority, PolicyPriority and RulePriority, in that order.
// It implements sort.Interface.
type ByPriority []Priority

func (bp ByPriority) Len() int           { return len(bp) }
func (bp ByPriority) Swap(i, j int)      { bp[i], bp[j] = bp[j], bp[i] }
func (bp ByPriority) Less(i, j int) bool { return bp[i].Less(bp[j]) }

type RuleMetric struct {
	Bytes, Packets, Sessions uint64
}

func (m *RuleMetric) Merge(m1 *RuleMetric) {
	m.Bytes += m1.Bytes
	m.Packets += m1.Packets
	m.Sessions += m1.Sessions
}

// A BitRange is a representation of a range of values from base value with a
// bitmask applied.
type BitRange struct {
	Value uint16
	Mask  *uint16
}

// EntityReference represents a reference to either a Pod or an ExternalEntity.
type EntityReference struct {
	// Pod maintains the reference to the Pod.
	Pod *corev1.Pod
	// ExternalEntity maintains the reference to the ExternalEntity.
	ExternalEntity *kuryrv1alpha1.ExternalEntityReference
}
