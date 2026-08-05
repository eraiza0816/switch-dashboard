package rtlplayground

// Host-side validation of device CLI commands, ported from rtlpctl.  The
// firmware intentionally has minimal input validation (RTLPlayground commit
// 6a28c58: "trim input validation and shorten messages to reduce BANK2
// size"), so write commands are validated here before they are sent.

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// cmdValidators maps each known device write command to its validator.
// Commands not listed here (read-only or unknown) pass through unchanged.
var cmdValidators = map[string]func([]string) error{
	"commit":        vNoArgs("commit"),
	"reset":         vNoArgs("reset"),
	"preshared_key": vPSK,
	"hostname":      vHostname,
	"passwd":        vPasswd,
	"ip":            vIPAddrCmd("ip", true),
	"gw":            vIPAddrCmd("gw", false),
	"netmask":       vIPAddrCmd("netmask", false),
	"port":          vPort,
	"mtu":           vMTU,
	"pvid":          vPvid,
	"vlan":          vVlan,
	"ingress":       vIngress,
	"isolate":       vIsolate,
	"mirror":        vMirror,
	"lag":           vLag,
	"laghash":       vLagHash,
	"eee":           vEEE,
	"bw":            vBW,
	"stp":           vStp,
	"igmp":          vIgmp,
	"telnet":        vOnOff("telnet"),
	"web":           vOnOff("web"),
	"l2":            vL2,
	"sfp":           vSFP,
	"regget":        vRegGet,
	"regset":        vRegSet,
	"sdsget":        vSdsGet,
	"sdsset":        vSdsSet,
	"phyget":        vPhyGet,
	"physet":        vPhySet,
}

// ValidateCommand checks a device CLI command string before it is sent to
// the switch.  Only known write commands are validated; unknown or read-only
// commands pass through unchanged so that future firmware commands are never
// blocked.
func ValidateCommand(text string) error {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	v, ok := cmdValidators[words[0]]
	if !ok {
		return nil
	}
	if err := v(words); err != nil {
		return fmt.Errorf("%s: %w", words[0], err)
	}
	return nil
}

func vNoArgs(cmd string) func([]string) error {
	return func(words []string) error {
		if len(words) != 1 {
			return fmt.Errorf("takes no arguments")
		}
		return nil
	}
}

// vOnOff requires exactly "on" or "off" (telnet, web).
func vOnOff(cmd string) func([]string) error {
	return func(words []string) error {
		if len(words) != 2 || words[1] != "on" && words[1] != "off" {
			return fmt.Errorf("usage: %s on|off", cmd)
		}
		return nil
	}
}

func vPSK(words []string) error {
	switch len(words) {
	case 1:
		return nil // clears the key
	case 2:
		return validatePSK(words[1])
	}
	return fmt.Errorf("expected 64 hex chars or no argument to clear, got %d words", len(words))
}

func vHostname(words []string) error {
	switch len(words) {
	case 1:
		return nil // show current name
	case 2:
		return validateDeviceHostname(words[1])
	}
	return fmt.Errorf("usage: hostname <name> (max 31 chars)")
}

func vPasswd(words []string) error {
	if len(words) != 2 {
		return fmt.Errorf("usage: passwd <password> (5-20 chars)")
	}
	return validatePassword(words[1])
}

// vIPAddrCmd validates ip/gw/netmask.  The device still validates these, but
// a pre-check catches typos before a malformed address is applied.
func vIPAddrCmd(cmd string, allowDHCP bool) func([]string) error {
	return func(words []string) error {
		switch len(words) {
		case 1:
			return nil // show
		case 2:
			if allowDHCP && words[1] == "dhcp" {
				return nil
			}
			return validateIPAddr(words[1])
		}
		return fmt.Errorf("usage: %s <ip-address>", cmd)
	}
}

func vPort(words []string) error {
	if len(words) < 3 {
		return fmt.Errorf("usage: port <port> [show|on|off|<speed> [half|full]|name <name>|duplex [half|full]]")
	}
	if err := validatePortNum(words[1], 9); err != nil {
		return err
	}
	switch words[2] {
	case "show":
		if len(words) > 3 {
			return fmt.Errorf("too many arguments after show")
		}
		return nil
	case "name":
		if len(words) != 4 {
			return fmt.Errorf("usage: port <port> name <name> (max 31 chars)")
		}
		if len(words[3]) > 31 {
			return fmt.Errorf("port name too long: %d chars (max 31)", len(words[3]))
		}
		return nil
	case "duplex":
		if len(words) > 4 {
			return fmt.Errorf("too many arguments after duplex")
		}
		if len(words) == 4 {
			return validateDuplex(words[3])
		}
		return nil // device defaults to half
	case "10m", "100m", "1g", "2g5", "5g", "10g", "auto", "on", "off":
		if len(words) == 4 {
			return validateDuplex(words[3])
		}
		if len(words) > 4 {
			return fmt.Errorf("too many arguments after %s", words[2])
		}
		return nil
	}
	return fmt.Errorf("unknown port command: %q (use show, on, off, <speed> [half|full], name, duplex)", words[2])
}

func vMTU(words []string) error {
	switch len(words) {
	case 1:
		return nil
	case 2:
		if words[1] == "show" {
			return nil
		}
		return fmt.Errorf("usage: mtu [show|<port> <size>]")
	case 3:
		if err := validatePortNum(words[1], 9); err != nil {
			return err
		}
		return validateMTU(words[2])
	}
	return fmt.Errorf("usage: mtu [show|<port> <size>]")
}

func vPvid(words []string) error {
	if len(words) != 3 {
		return fmt.Errorf("usage: pvid <port> <vlan-id> (1-4094)")
	}
	if err := validatePortNum(words[1], 9); err != nil {
		return err
	}
	return validateVLANID(words[2])
}

func vVlan(words []string) error {
	if len(words) < 2 {
		return fmt.Errorf("usage: vlan <vlan-id> [d|mgmt|<name>|<port>[t|u]...] (vlan-id 1-4094)")
	}
	if err := validateVLANID(words[1]); err != nil {
		return err
	}
	for _, w := range words[2:] {
		switch {
		case w == "d" || w == "mgmt":
			// d and mgmt are only meaningful as the sole argument
		case w[0] >= 'a' && w[0] <= 'z' || w[0] >= 'A' && w[0] <= 'Z':
			// VLAN name token; length is not bounded on the device
		default:
			if err := validateVlanPortToken(w); err != nil {
				return err
			}
		}
	}
	return nil
}

func vIngress(words []string) error {
	if len(words) < 2 {
		return fmt.Errorf("usage: ingress [<port>]t|u|a...")
	}
	for _, w := range words[1:] {
		if err := validateIngressToken(w); err != nil {
			return err
		}
	}
	return nil
}

func vIsolate(words []string) error {
	if len(words) < 3 {
		if len(words) == 2 {
			if err := validatePortToken(words[1]); err != nil {
				return err
			}
		}
		return fmt.Errorf("usage: isolate <port> [show|off|<port>...]")
	}
	if err := validatePortToken(words[1]); err != nil {
		return err
	}
	switch words[2] {
	case "show", "off":
		if len(words) > 3 {
			return fmt.Errorf("too many arguments after %s", words[2])
		}
		return nil
	}
	for _, w := range words[2:] {
		if err := validatePortToken(w); err != nil {
			return err
		}
	}
	return nil
}

func vMirror(words []string) error {
	switch len(words) {
	case 1:
		return fmt.Errorf("usage: mirror [status|off|<mirroring-port> [<port>[t|r]]...]")
	case 2:
		if words[1] == "status" || words[1] == "off" {
			return nil
		}
	}
	for _, w := range words[1:] {
		body := strings.TrimSuffix(strings.TrimSuffix(w, "t"), "r")
		if err := validatePortToken(body); err != nil {
			return fmt.Errorf("invalid mirror port: %q", w)
		}
	}
	return nil
}

func vLag(words []string) error {
	if len(words) == 1 || len(words) == 2 && words[1] == "show" {
		return nil
	}
	if len(words[1]) != 1 || words[1][0] < '1' || words[1][0] > '4' {
		return fmt.Errorf("invalid LAG group: %q (must be 1-4)", words[1])
	}
	for _, w := range words[2:] {
		if err := validatePortToken(w); err != nil {
			return err
		}
	}
	return nil
}

func vLagHash(words []string) error {
	if len(words) < 2 {
		return fmt.Errorf("usage: laghash <group> [spa|smac|dmac|sip|dip|sport|dport]...")
	}
	if len(words[1]) != 1 || words[1][0] < '0' || words[1][0] > '3' {
		return fmt.Errorf("invalid LAG group: %q (must be 0-3)", words[1])
	}
	for _, w := range words[2:] {
		switch w {
		case "spa", "smac", "dmac", "sip", "dip", "sport", "dport":
		default:
			return fmt.Errorf("bad hash type: %q (use spa, smac, dmac, sip, dip, sport, dport)", w)
		}
	}
	return nil
}

func vEEE(words []string) error {
	if len(words) < 2 {
		return fmt.Errorf("usage: eee [on|off|status] [port] [100m|1g|2g5]")
	}
	switch words[1] {
	case "on", "off", "status":
	default:
		return fmt.Errorf("unknown eee command: %q (use on, off, status)", words[1])
	}
	for _, w := range words[2:] {
		switch w {
		case "100m", "1g", "2g5":
		default:
			if err := validatePortNum(w, 9); err != nil {
				return fmt.Errorf("invalid eee argument: %q (use a port 1-9 or 100m, 1g, 2g5)", w)
			}
		}
	}
	return nil
}

func vBW(words []string) error {
	if len(words) < 3 {
		return fmt.Errorf("usage: bw [in|out|status] <port> [<hexvalue>|off|drop|fc]")
	}
	switch words[1] {
	case "in", "out", "status":
	default:
		return fmt.Errorf("unknown bw direction: %q (use in, out, status)", words[1])
	}
	if err := validatePortNum(words[2], 9); err != nil {
		return err
	}
	switch len(words) {
	case 3:
		if words[1] == "status" {
			return nil
		}
		return fmt.Errorf("usage: bw [in|out|status] <port> [<hexvalue>|off|drop|fc]")
	case 4:
		switch words[3] {
		case "off", "drop", "fc":
			if words[1] == "in" || words[1] == "out" {
				return nil
			}
			return fmt.Errorf("bw %s does not take a value", words[1])
		}
		if words[1] == "in" || words[1] == "out" {
			return validateHexArg(words[3], 1, 4)
		}
		return fmt.Errorf("bw %s does not take a value", words[1])
	}
	return fmt.Errorf("too many arguments")
}

func vStp(words []string) error {
	if len(words) > 2 {
		return fmt.Errorf("usage: stp [on|off]")
	}
	if len(words) == 2 && words[1] != "on" && words[1] != "off" {
		return fmt.Errorf("usage: stp [on|off]")
	}
	return nil
}

func vIgmp(words []string) error {
	if len(words) > 2 {
		return fmt.Errorf("usage: igmp [on|off|show]")
	}
	if len(words) == 2 && words[1] != "on" && words[1] != "off" && words[1] != "show" {
		return fmt.Errorf("usage: igmp [on|off|show]")
	}
	return nil
}

func vL2(words []string) error {
	switch len(words) {
	case 1:
		return nil // show learned entries
	case 2:
		if words[1] == "forget" {
			return nil
		}
		return fmt.Errorf("usage: l2 [forget|del <idx>]")
	case 3:
		if words[1] != "del" {
			return fmt.Errorf("usage: l2 [forget|del <idx>]")
		}
		return validateL2Idx(words[2])
	}
	return fmt.Errorf("usage: l2 [forget|del <idx>]")
}

func vSFP(words []string) error {
	if len(words) == 1 {
		return nil // list all slots
	}
	if err := validateSfpSlot(words[1]); err != nil {
		return err
	}
	if len(words) == 2 {
		return nil // show slot info
	}
	sub := words[2]
	switch sub {
	case "1g", "2g5", "10g", "100m", "auto":
		if len(words) > 3 {
			return fmt.Errorf("too many arguments after %s", sub)
		}
		return nil
	case "describe", "dump", "save", "restore", "fix", "patch", "clone":
		return validatePwArgs(words[3:])
	case "checksum":
		return validatePwArgs(words[3:]) // [--fix] [--pw <hex8>]
	case "write":
		if len(words) < 5 {
			return fmt.Errorf("usage: sfp <slot> write <offset> <value> [--pw <hex8>]")
		}
		if err := validateHexArg(words[3], 1, 1); err != nil {
			return fmt.Errorf("invalid write offset: %w", err)
		}
		if err := validateHexArg(words[4], 1, 1); err != nil {
			return fmt.Errorf("invalid write value: %w", err)
		}
		return validatePwArgs(words[5:])
	case "bulk":
		if len(words) != 4 {
			return fmt.Errorf("usage: sfp <slot> bulk <512hexchars>")
		}
		if len(words[3]) != 512 || !isHexStr(words[3]) {
			return fmt.Errorf("bulk data must be exactly 512 hex chars")
		}
		return nil
	}
	return fmt.Errorf("unknown sfp subcommand: %q (use 1g, 2g5, 10g, 100m, auto, describe, dump, save, restore, fix, patch, clone, checksum, write, bulk)", sub)
}

// validatePwArgs checks optional "--pw <hex8>" arguments.  Any other tokens
// are rejected to catch typos in SFP commands.
func validatePwArgs(words []string) error {
	if len(words) == 0 {
		return nil
	}
	if words[0] == "--fix" {
		return validatePwArgs(words[1:])
	}
	if len(words) == 2 && words[0] == "--pw" {
		if len(words[1]) != 8 || !isHexStr(words[1]) {
			return fmt.Errorf("--pw must be 8 hex chars")
		}
		return nil
	}
	return fmt.Errorf("unexpected argument: %q (use [--fix] [--pw <hex8>])", words[0])
}

func vRegGet(words []string) error {
	if len(words) != 2 {
		return fmt.Errorf("usage: regget <hexvalue>")
	}
	return validateHexArg(words[1], 1, 2)
}

func vRegSet(words []string) error {
	if len(words) != 3 {
		return fmt.Errorf("usage: regset <hexvalue> <hexvalue>")
	}
	if err := validateHexArg(words[1], 1, 2); err != nil {
		return fmt.Errorf("invalid register: %w", err)
	}
	return validateHexArg(words[2], 1, 4)
}

func vSdsGet(words []string) error {
	if len(words) != 4 {
		return fmt.Errorf("usage: sdsget <sds-id> <hex:page> <hex:reg>")
	}
	return validateSdsArgs(words[1:])
}

func vSdsSet(words []string) error {
	if len(words) != 5 {
		return fmt.Errorf("usage: sdsset <sds-id> <hex:page> <hex:reg> <hex:val>")
	}
	if err := validateSdsArgs(words[1:]); err != nil {
		return err
	}
	return validateHexArg(words[4], 1, 2)
}

func validateSdsArgs(words []string) error {
	if _, err := strconv.Atoi(words[0]); err != nil {
		return fmt.Errorf("invalid sds-id: %q (must be a number 0-255)", words[0])
	}
	if err := validateHexArg(words[1], 1, 1); err != nil {
		return fmt.Errorf("invalid page: %w", err)
	}
	return validateHexArg(words[2], 1, 1)
}

func vPhyGet(words []string) error {
	if len(words) != 4 {
		return fmt.Errorf("usage: phyget <phy-id> <dev-id> <hex:reg>")
	}
	return validatePhyArgs(words[1:])
}

func vPhySet(words []string) error {
	if len(words) != 5 {
		return fmt.Errorf("usage: physet <phy-id> <dev-id> <hex:reg> <hex:val>")
	}
	if err := validatePhyArgs(words[1:]); err != nil {
		return err
	}
	return validateHexArg(words[4], 1, 2)
}

func validatePhyArgs(words []string) error {
	for _, w := range words[:2] {
		if _, err := strconv.Atoi(w); err != nil {
			return fmt.Errorf("invalid id: %q (must be a number 0-255)", w)
		}
	}
	return validateHexArg(words[2], 1, 2)
}

func isHexStr(s string) bool {
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F') {
			return false
		}
	}
	return true
}

// validatePSK verifies a pre-shared key: exactly 64 hex chars (32 bytes).
// The firmware no longer validates this (invalid chars yield 0xff bytes),
// so the host side is the enforcing side.
func validatePSK(s string) error {
	if len(s) != 64 {
		return fmt.Errorf("pre-shared key must be exactly 64 hex chars (got %d)", len(s))
	}
	if !isHexStr(s) {
		return fmt.Errorf("pre-shared key must be hex: invalid character in %q", s)
	}
	return nil
}

// validateDeviceHostname verifies a switch hostname: 1-31 chars from
// [a-zA-Z0-9_-].  The firmware no longer validates this and truncates at the
// first invalid character (e.g. a dot), which would silently change the name.
func validateDeviceHostname(s string) error {
	if s == "" {
		return fmt.Errorf("hostname is empty")
	}
	if len(s) > 31 {
		return fmt.Errorf("hostname too long: %d chars (max 31)", len(s))
	}
	for _, c := range s {
		ok := c == '-' || c == '_' ||
			(c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
		if !ok {
			return fmt.Errorf("hostname contains invalid character %q (allowed: letters, digits, '-', '_')", c)
		}
	}
	return nil
}

// validatePassword verifies the web login password: 5-20 chars, matching the
// firmware checks (min 5 in parse_passwd, max 20 per docs).
func validatePassword(s string) error {
	if len(s) < 5 {
		return fmt.Errorf("password too short: %d chars (min 5)", len(s))
	}
	if len(s) > 20 {
		return fmt.Errorf("password too long: %d chars (max 20)", len(s))
	}
	return nil
}

// validateIPAddr checks an IPv4 address.  The firmware still validates
// addresses on the device (dotted-quad only), this is a pre-check to catch
// typos early.
func validateIPAddr(s string) error {
	if ip := net.ParseIP(s); ip == nil || ip.To4() == nil {
		return fmt.Errorf("invalid IP address: %q", s)
	}
	return nil
}

// validatePortNum checks a single-digit physical port number (1..maxPort).
func validatePortNum(s string, maxPort int) error {
	if len(s) != 1 || s[0] < '1' || s[0] > '9' {
		return fmt.Errorf("invalid port: %q (must be a single digit 1-%d)", s, maxPort)
	}
	if int(s[0]-'0') > maxPort {
		return fmt.Errorf("invalid port: %q (must be 1-%d)", s, maxPort)
	}
	return nil
}

// validatePortToken checks a physical port (1-9) or "10" for the CPU port.
// Used by commands that accept the 2-digit CPU port token (vlan, isolate,
// mirror, lag).
func validatePortToken(s string) error {
	if s == "10" {
		return nil
	}
	if len(s) != 1 || s[0] < '1' || s[0] > '9' {
		return fmt.Errorf("invalid port: %q (must be 1-9, or 10 for CPU)", s)
	}
	return nil
}

func validateDuplex(s string) error {
	if s != "half" && s != "full" {
		return fmt.Errorf("invalid duplex: %q (use half or full)", s)
	}
	return nil
}

// validateMTU checks an MTU value.  The firmware only enforces the 16383
// maximum; the 64 minimum matches its own TODO.
func validateMTU(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil || n < 64 || n > 16383 {
		return fmt.Errorf("invalid MTU: %q (must be a number 64-16383)", s)
	}
	return nil
}

// validateSfpSlot checks an SFP slot number.  The firmware accepts only
// slots 1-2; whether a machine actually has the slot is checked on-device.
func validateSfpSlot(s string) error {
	if len(s) != 1 || s[0] < '1' || s[0] > '2' {
		return fmt.Errorf("invalid SFP slot: %q (must be 1 or 2)", s)
	}
	return nil
}

// validateHexArg checks a hex value of minBytes..maxBytes bytes.
func validateHexArg(s string, minBytes, maxBytes int) error {
	if !isHexStr(s) {
		return fmt.Errorf("invalid hex value: %q", s)
	}
	minChars, maxChars := minBytes*2-1, maxBytes*2
	if len(s) < minChars || len(s) > maxChars {
		return fmt.Errorf("invalid hex value: %q (must be %d-%d hex chars)", s, minChars, maxChars)
	}
	return nil
}

// validateIngressToken checks an ingress filter token: "t", "u" or "a" for
// all ports, or "<port><t|u|a>" for a single port.
func validateIngressToken(s string) error {
	if len(s) == 1 {
		if s[0] == 't' || s[0] == 'u' || s[0] == 'a' {
			return nil
		}
	} else if len(s) == 2 && s[0] >= '1' && s[0] <= '9' {
		if s[1] == 't' || s[1] == 'u' || s[1] == 'a' {
			return nil
		}
	}
	return fmt.Errorf("invalid ingress token: %q (use [<port>]t|u|a)", s)
}

// validateVlanPortToken checks a VLAN member token: port (1-9 or 10=CPU)
// with an optional t/u suffix.
func validateVlanPortToken(s string) error {
	body := strings.TrimSuffix(strings.TrimSuffix(s, "t"), "u")
	return validatePortToken(body)
}

func validateVLANID(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 4094 {
		return fmt.Errorf("invalid VLAN ID: %q (must be a number 1-4094)", s)
	}
	return nil
}

func validateL2Idx(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 || n > 4095 {
		return fmt.Errorf("invalid L2 index: %q (must be a number 0-4095)", s)
	}
	return nil
}
