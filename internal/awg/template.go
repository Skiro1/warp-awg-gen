package awg

import (
	"encoding/base64"
	"fmt"
	"strings"

	"warp-awg-gen/internal/warp"
)

var warpAllowedIPs = "" +
	"1.0.0.0/8, 2.0.0.0/7, 4.0.0.0/6, 8.0.0.0/7, 11.0.0.0/8, 12.0.0.0/6, " +
	"16.0.0.0/4, 32.0.0.0/3, 64.0.0.0/3, 96.0.0.0/4, 112.0.0.0/5, " +
	"120.0.0.0/6, 124.0.0.0/7, 126.0.0.0/8, 128.0.0.0/3, 160.0.0.0/5, " +
	"168.0.0.0/8, 169.0.0.0/9, 169.128.0.0/10, 169.192.0.0/11, " +
	"169.224.0.0/12, 169.240.0.0/13, 169.248.0.0/14, 169.252.0.0/15, " +
	"169.255.0.0/16, 170.0.0.0/7, 172.0.0.0/12, 172.32.0.0/11, " +
	"172.64.0.0/10, 172.128.0.0/9, 173.0.0.0/8, 174.0.0.0/7, " +
	"176.0.0.0/4, 192.0.0.0/9, 192.128.0.0/11, 192.160.0.0/13, " +
	"192.169.0.0/16, 192.170.0.0/15, 192.172.0.0/14, 192.176.0.0/12, " +
	"192.192.0.0/10, 193.0.0.0/8, 194.0.0.0/7, 196.0.0.0/6, " +
	"200.0.0.0/5, 208.0.0.0/4, 224.0.0.0/4, ::/1, 8000::/2, " +
	"c000::/3, e000::/4, f000::/5, f800::/6, fe00::/9, fec0::/10, ff00::/8"

var defaultDNS = "1.1.1.1, 1.0.0.1, 2606:4700:4700::1111, 2606:4700:4700::1001"

type WireConfig struct {
	WarpConfig *warp.RegistrationResponse
	PrivateKey []byte
	Params     *AWGParams
	CPS        *CPSPackets
	DNS        string
	Endpoint   string
	Keepalive  int
	MTU        int
}

func (c *WireConfig) Build() (string, error) {
	if c.WarpConfig == nil {
		return "", fmt.Errorf("WARP config is nil")
	}
	if len(c.WarpConfig.Config.Addresses.V4) == 0 {
		return "", fmt.Errorf("no IPv4 address assigned")
	}
	if len(c.WarpConfig.Config.Addresses.V6) == 0 {
		return "", fmt.Errorf("no IPv6 address assigned")
	}

	var b strings.Builder

	b.WriteString("[Interface]\n")
	b.WriteString(fmt.Sprintf("PrivateKey = %s\n", base64.StdEncoding.EncodeToString(c.PrivateKey)))

	ip4 := stripCIDR(c.WarpConfig.Config.Addresses.V4)
	ip6 := stripCIDR(c.WarpConfig.Config.Addresses.V6)
	b.WriteString(fmt.Sprintf("Address = %s, %s\n", ip4, ip6))

	dns := c.DNS
	if dns == "" {
		dns = defaultDNS
	}
	b.WriteString(fmt.Sprintf("DNS = %s\n", dns))

	if c.MTU > 0 {
		b.WriteString(fmt.Sprintf("MTU = %d\n", c.MTU))
	}

	if c.Params != nil {
		if c.Params.Jc > 0 {
			b.WriteString(fmt.Sprintf("Jc = %d\n", c.Params.Jc))
		}
		if c.Params.Jmin > 0 {
			b.WriteString(fmt.Sprintf("Jmin = %d\n", c.Params.Jmin))
		}
		if c.Params.Jmax > 0 {
			b.WriteString(fmt.Sprintf("Jmax = %d\n", c.Params.Jmax))
		}
		b.WriteString(fmt.Sprintf("S1 = %d\n", c.Params.S1))
		b.WriteString(fmt.Sprintf("S2 = %d\n", c.Params.S2))
		b.WriteString(fmt.Sprintf("S3 = %d\n", c.Params.S3))
		b.WriteString(fmt.Sprintf("S4 = %d\n", c.Params.S4))
		if c.Params.H1 != "" {
			b.WriteString(fmt.Sprintf("H1 = %s\n", c.Params.H1))
		}
		if c.Params.H2 != "" {
			b.WriteString(fmt.Sprintf("H2 = %s\n", c.Params.H2))
		}
		if c.Params.H3 != "" {
			b.WriteString(fmt.Sprintf("H3 = %s\n", c.Params.H3))
		}
		if c.Params.H4 != "" {
			b.WriteString(fmt.Sprintf("H4 = %s\n", c.Params.H4))
		}
	}

	if c.CPS != nil {
		if c.CPS.I1 != "" {
			b.WriteString(fmt.Sprintf("I1 = %s\n", c.CPS.I1))
		}
		if c.CPS.I2 != "" {
			b.WriteString(fmt.Sprintf("I2 = %s\n", c.CPS.I2))
		}
		if c.CPS.I3 != "" {
			b.WriteString(fmt.Sprintf("I3 = %s\n", c.CPS.I3))
		}
		if c.CPS.I4 != "" {
			b.WriteString(fmt.Sprintf("I4 = %s\n", c.CPS.I4))
		}
		if c.CPS.I5 != "" {
			b.WriteString(fmt.Sprintf("I5 = %s\n", c.CPS.I5))
		}
	}

	peerPubKey := ""
	endpoint := c.Endpoint
	if len(c.WarpConfig.Config.Peers) > 0 {
		peerPubKey = c.WarpConfig.Config.Peers[0].PublicKey
		if endpoint == "" {
			endpoint, _ = SelectFastestEndpoint()
		}
	}

	b.WriteString("\n[Peer]\n")
	b.WriteString(fmt.Sprintf("PublicKey = %s\n", peerPubKey))
	b.WriteString(fmt.Sprintf("AllowedIPs = %s\n", warpAllowedIPs))
	b.WriteString(fmt.Sprintf("Endpoint = %s\n", endpoint))
	if c.Keepalive > 0 {
		b.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", c.Keepalive))
	}

	return b.String(), nil
}

func stripCIDR(addr string) string {
	if idx := strings.IndexByte(addr, '/'); idx >= 0 {
		return addr[:idx]
	}
	return addr
}
