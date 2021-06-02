package types

const (
	// HostLocalSourceBit is the bit of the iptables fwmark space to mark locally generated packets.
	// Value must be within the range [0, 31].
	HostLocalSourceBit = 0
)

var (
	// HostLocalSourceMark is the mark generated from HostLocalSourceBit.
	HostLocalSourceMark = uint32(1 << HostLocalSourceBit)

	// SNATIPMarkMask is the bits of packet mark that stores the ID of the
	// SNAT IP for a "Pod -> external" egress packet, that is to be SNAT'd.
	SNATIPMarkMask = uint32(0xFF)
)
