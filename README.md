# warp-awg-gen

Generates AmneziaWG 2.0 configuration files for Cloudflare WARP with full CPS camouflage support.

[Readme in Russian](README.ru.md)

## New project
- **[warp-cli](https://github.com/Skiro1/warp-cli)**

## Features

- Cloudflare WARP registration via the official API
- WARP+ license key activation (not verified)
- Registration caching (90 days) -- no API call needed for subsequent runs
- CPS (Camouflage Packet Sending) for DPI bypass using real WARP QUIC packets
- Protocol presets for I2-I5 randomization
- Custom AWG parameters (Jc, Jmin, Jmax, S1-S4, H1-H4)
- Custom CPS expressions (I1-I5)
- Hex dump to CPS conversion (Wireshark -> I1)
- YAML configuration file support
- Automatic fastest endpoint selection via TCP latency test
- Full WARP route table (AllowedIPs)

## Usage

### Basic generation

```
warp-awg-gen.exe -o warp.conf
```

Generates a WARP config and saves it to the specified file. The first run registers with the Cloudflare WARP API.

### WARP+ license

```
warp-awg-gen.exe -l YOUR_LICENSE_KEY -o warp-plus.conf
```

### Registration caching

The first run saves the registration to a cache file. Subsequent runs reuse it, skipping the API call entirely. This is useful if the API is blocked on your network -- you only need to enable zapret once.

```
warp-awg-gen.exe --fresh-reg -o warp.conf   # force new registration
warp-awg-gen.exe -o warp.conf                # uses cache (no API call)
```

Cache file location (default): `warp-reg.json` in the current directory. Lifetime: 90 days.

### Protocol masks

Controls the structure of I2-I5 random padding.

```
warp-awg-gen.exe --mask quic -o warp.conf   # QUIC-style padding (default)
warp-awg-gen.exe --mask dns -o warp.conf    # DNS-sized padding
warp-awg-gen.exe --mask sip -o warp.conf    # SIP-sized padding
```

I1 always uses a real captured WARP QUIC payload regardless of the selected mask.

### Disable CPS

```
warp-awg-gen.exe --transport none -o warp-plain.conf
```

Generates a standard WireGuard config without CPS camouflage.

### YAML configuration

```
warp-awg-gen.exe -c example.yaml -o warp.conf
```

### Custom AWG parameters

```
warp-awg-gen.exe --jc 4 --jmin 40 --jmax 70 --s1 0 --s2 0 --s3 0 --s4 0 -o warp.conf
warp-awg-gen.exe --h1 1 --h2 2 --h3 3 --h4 4 -o warp.conf
```

### Custom CPS expressions

```
warp-awg-gen.exe --i1 "<b 0xce00000001...><r 16>" --i2 "<r 32>" -o warp.conf
```

### Hex dump to CPS

Convert a Wireshark hex dump into an I1 CPS expression. The hex data must contain at least 32 bytes (64 hex characters).

```
warp-awg-gen.exe --from-hex "ce0000000108c10f0123456789abcdef..." -o warp.conf
```

### Custom endpoint

```
warp-awg-gen.exe --endpoint engage.cloudflareclient.com:2408 -o warp.conf
warp-awg-gen.exe --endpoint 162.159.192.1:943 -o warp.conf
```

By default, the fastest endpoint is selected automatically via TCP latency test.

### Additional options

```
--dns        DNS servers (default: 1.1.1.1, 1.0.0.1, 2606:4700:4700::1111, 2606:4700:4700::1001)
--mtu        MTU (default: 1280)
--keepalive  PersistentKeepalive interval (default: 25)
--reg-cache  Registration cache file path (default: warp-reg.json)
```

## Generated config example

```
[Interface]
PrivateKey = <base64-private-key>
Address = 172.16.0.2, 2606:4700:110:xxxx:xxxx:xxxx:xxxx:xxxx
DNS = 1.1.1.1, 1.0.0.1, 2606:4700:4700::1111, 2606:4700:4700::1001
MTU = 1280
Jc = 4
Jmin = 40
Jmax = 70
S1 = 0
S2 = 0
S3 = 0
S4 = 0
H1 = 1
H2 = 2
H3 = 3
H4 = 4
I1 = <b 0xce000000010897a2...> (real WARP QUIC payload, 1250+ bytes)
I2 = <b 0x...> (random hex bytes)
I3 = <b 0x...> (random hex bytes)
I4 = <b 0x...> (random hex bytes)
I5 = <b 0x...> (random hex bytes)

[Peer]
PublicKey = bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=
AllowedIPs = <full WARP route table>
Endpoint = 162.159.192.1:943 (or fastest detected endpoint)
PersistentKeepalive = 25
```

## Troubleshooting

### Registration fails (TLS handshake timeout)

The Cloudflare WARP API may be blocked on your network. Use zapret, a VPN, or a proxy for the first run only. Subsequent runs use cached registration and do not require API access.

```
warp-awg-gen.exe --fresh-reg -o warp.conf   # enable zapret, run once
```

### Registration fails with 429

Rate limited. Wait a few minutes and try again, or use a cached registration if available.

### Config imports but does not connect

Try a different endpoint. Some IPs or ports may be blocked.

```
warp-awg-gen.exe --endpoint engage.cloudflareclient.com:2408 -o warp.conf
warp-awg-gen.exe --endpoint 162.159.192.1:943 -o warp.conf
```

If the issue persists, disable CPS and try again:

```
warp-awg-gen.exe --transport none -o warp-plain.conf
```

### "no IPv4 address assigned"

The WARP API returned an unexpected response format. This usually means the API has changed. Use `--fresh-reg` to retry.

### Config connects but traffic does not pass through

Ensure `--mtu 1280` is set. If DPI is still blocking traffic, CPS may not be sufficient on your network. Try the `--transport none` mode to verify whether CPS is the issue, or use a different mask.

### --from-hex does nothing

The hex string must contain at least 32 bytes (64 hex characters). Non-hex characters (spaces, colons, dots) are automatically stripped.

### Cache load fails (expired)

Registration cache expires after 90 days. Run with `--fresh-reg` to re-register (requires API access).

### YAML config not working

Ensure the file follows the structure shown in `example.yaml`. Boolean and numeric values must be in the correct format (no quotes around numbers, etc).

## Build

```
go build -o warp-awg-gen.exe .
```

Requires Go 1.21+ and `golang.org/x/crypto`.

## License

MIT
