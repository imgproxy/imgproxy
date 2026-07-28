// Package privatenet classifies IP addresses as private, extending the stdlib
// net.IP.IsPrivate() check to also cover address ranges and embedding schemes
// it misses, for the purposes of SSRF guards.
//
// This package provides:
//   - IsPrivate: extends net.IP.IsPrivate() to include RFC 6598 CGNAT (100.64.0.0/10)
//   - IsLoopback: checks if an address is loopback or unspecified
//   - IsLinkLocal: checks if an address is link-local
//   - Unwrap: decodes IPv6 addresses that carry an embedded IPv4 address
//     per RFC 4291 (IPv4-compatible, IPv4-mapped), RFC 3056 (6to4), RFC 6052/8215 (NAT64),
//     RFC 4380 (Teredo), and RFC 5214 (ISATAP).
//
// The Unwrap function is used to detect SSRF bypasses where an attacker wraps
// a disallowed IPv4 address (loopback, link-local, private, CGNAT) inside an
// IPv6 embedding form that, on its own, doesn't appear private.
//
// All check functions also check any IPv4 address embedded via Unwrap,
// ensuring no escapes via IPv6 embedding.
package privatenet

import (
	"bytes"
	"net"
)

var (
	// RFC 6598 - Carrier-Grade NAT / shared address space
	cgnatPrefix = mustParseCIDR("100.64.0.0/10")
	// RFC 6052 - NAT64 well-known prefix
	nat64WellKnown = mustParseCIDR("64:ff9b::/96")
	// RFC 8215 - NAT64 local-use prefix
	nat64LocalUse6Bytes = [6]byte{0x64, 0xff, 0x9b, 0x01, 0x00, 0x00}
	// unwrappers holds all IPv6-embedding unwrapping functions
	unwrappers = []func(net.IP) (net.IP, bool){
		unwrapIPv4Compatible,
		unwrap6to4,
		unwrapNAT64WellKnown,
		unwrapNAT64LocalUse,
		unwrapTeredo,
		unwrapISATAP,
	}
)

func mustParseCIDR(s string) *net.IPNet {
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return ipnet
}

// IsPrivate reports whether ip is a private address per net.IP.IsPrivate(),
// or falls within RFC 6598 CGNAT shared address space (100.64.0.0/10),
// including via an embedded IPv4 address.
func IsPrivate(ip net.IP) bool {
	return checkAddr(ip, isPrivateAddr)
}

// IsLoopback reports whether ip is a loopback or unspecified address,
// including via an embedded IPv4 address.
func IsLoopback(ip net.IP) bool {
	return checkAddr(ip, isLoopbackAddr)
}

// IsLinkLocal reports whether ip is a link-local address (unicast or multicast),
// including via an embedded IPv4 address.
func IsLinkLocal(ip net.IP) bool {
	return checkAddr(ip, isLinkLocalAddr)
}

// Unwrap detects if ip is an IPv6 address carrying an embedded IPv4 address
// and returns the decoded IPv4 address. Returns nil if ip is not an IPv6-embeds-IPv4
// form, or if ip is already an IPv4 address. Follows the errors.Unwrap convention.
//
// Recognizes:
//   - RFC 4291 §2.5.5 IPv4-compatible (::a.b.c.d) and IPv4-mapped (::ffff:a.b.c.d)
//   - RFC 3056 6to4 (2002:a.b.c.d::/16) — bits 16–47 encode the IPv4 address
//   - RFC 6052 NAT64 well-known (64:ff9b::a.b.c.d/96) — last 32 bits
//   - RFC 8215 NAT64 local-use (/48 layout per RFC 6052 §2.2)
//   - RFC 4380 Teredo (2001::/32) — last 32 bits XORed with 0xFFFFFFFF
//   - RFC 5214 ISATAP (interface ID 0000:5EFE:a.b.c.d or 0200:5EFE:a.b.c.d)
func Unwrap(ip net.IP) net.IP {
	// IPv4 is not an embedding scheme.
	if ip.To4() != nil {
		return nil
	}

	// Ensure we're working with a 16-byte form (IPv6).
	if len(ip) != net.IPv6len {
		return nil
	}

	for _, fn := range unwrappers {
		if result, ok := fn(ip); ok {
			return result
		}
	}

	return nil
}

// checkAddr reports whether check matches ip, or the IPv4 address embedded in
// ip (if any). Unwrap yields at most one embedding layer — its result, when
// non-nil, always resolves as bare IPv4 and Unwrap on it returns nil — so a
// single extra check is sufficient; there is nothing further to unwrap.
func checkAddr(ip net.IP, check func(net.IP) bool) bool {
	if ip == nil {
		return false
	}
	if check(ip) {
		return true
	}
	if unwrapped := Unwrap(ip); unwrapped != nil {
		return check(unwrapped)
	}
	return false
}

func isPrivateAddr(ip net.IP) bool {
	return ip.IsPrivate() || cgnatPrefix.Contains(ip)
}

func isLoopbackAddr(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsUnspecified()
}

func isLinkLocalAddr(ip net.IP) bool {
	return ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

func unwrapIPv4Compatible(ip net.IP) (net.IP, bool) {
	// RFC 4291 §2.5.5 IPv4-compatible (::a.b.c.d / ::/96).
	// Exclude IPv4-mapped (::ffff:a.b.c.d / ::ffff:0:0/96) which stdlib already unwraps.
	// Check first 12 bytes are all zero.
	for i := range 12 {
		if ip[i] != 0 {
			return nil, false
		}
	}
	return net.IPv4(ip[12], ip[13], ip[14], ip[15]), true
}

func unwrap6to4(ip net.IP) (net.IP, bool) {
	// RFC 3056 6to4 (2002::/16). IPv4 is in bits 16–47 (bytes 2–5).
	if ip[0] == 0x20 && ip[1] == 0x02 {
		return net.IPv4(ip[2], ip[3], ip[4], ip[5]), true
	}
	return nil, false
}

func unwrapNAT64WellKnown(ip net.IP) (net.IP, bool) {
	// RFC 6052 NAT64 well-known (64:ff9b::/96). IPv4 is in the last 32 bits.
	if nat64WellKnown.Contains(ip) {
		return net.IPv4(ip[12], ip[13], ip[14], ip[15]), true
	}
	return nil, false
}

func unwrapNAT64LocalUse(ip net.IP) (net.IP, bool) {
	// RFC 8215 NAT64 local-use (/48 layout per RFC 6052 §2.2).
	// Prefix bytes 0–5 are 64:ff9b:1 (the /48 prefix).
	// The IPv4 address occupies bytes 6–9 after the prefix, with byte 8 reserved (must be 0).
	// So IPv4 is encoded as bytes 6, 7, 9, 10.
	if bytes.HasPrefix(ip, nat64LocalUse6Bytes[:]) {
		return net.IPv4(ip[6], ip[7], ip[9], ip[10]), true
	}
	return nil, false
}

func unwrapTeredo(ip net.IP) (net.IP, bool) {
	// RFC 4380 Teredo (2001::/32).
	// IPv4 is in the last 32 bits, XORed with 0xFFFFFFFF.
	if ip[0] == 0x20 && ip[1] == 0x01 && ip[2] == 0 && ip[3] == 0 {
		return net.IPv4(ip[12]^0xFF, ip[13]^0xFF, ip[14]^0xFF, ip[15]^0xFF), true
	}
	return nil, false
}

func unwrapISATAP(ip net.IP) (net.IP, bool) {
	// RFC 5214 ISATAP (any prefix:0000:5EFE:a.b.c.d or any prefix:0200:5EFE:a.b.c.d).
	// Check bytes 8–11 for the 0000:5EFE or 0200:5EFE pattern.
	if (ip[8] == 0 || ip[8] == 0x02) && ip[9] == 0 &&
		ip[10] == 0x5e && ip[11] == 0xfe {
		return net.IPv4(ip[12], ip[13], ip[14], ip[15]), true
	}
	return nil, false
}
