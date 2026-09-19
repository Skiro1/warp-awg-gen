package awg

import (
	"encoding/base64"
	"fmt"
	"strings"

	"warp-awg-gen/internal/warp"
)

var warpAllowedIPs = "" +
	"0.0.0.0/0, ::/0"


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
	    
	    b.WriteString(fmt.Sprintf("ContentPaddingAddition = 10-100 \n"))
	    b.WriteString(fmt.Sprintf("RekeyAfterTime = 100-120 \n"))
        b.WriteString(fmt.Sprintf("RekeyTimeout = 3-7 \n"))
        b.WriteString(fmt.Sprintf("RejectAfterTime = 150-180 \n"))
        b.WriteString(fmt.Sprintf("KeepaliveTimeout = 5-15 \n"))
        b.WriteString(fmt.Sprintf("MaxHandshakeAttempts = 15-20 \n"))
        b.WriteString(fmt.Sprintf("RandomTrailers = on \n"))
        b.WriteString(fmt.Sprintf("DisableCookies = on \n"))


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
