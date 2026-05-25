package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"warp-awg-gen/internal/awg"
	"warp-awg-gen/internal/config"
	"warp-awg-gen/internal/pcap"
	"warp-awg-gen/internal/warp"
)

func main() {
	flags := parseFlags()
	cfg := loadConfig(flags)

	if flags.Transport != "" && flags.Transport != "none" && cfg.Mask == "random" {
		cfg.Mask = flags.Transport
	}

	if cfg.FromHex != "" {
		cps := pcap.HexToCPS(cfg.FromHex)
		if cps != "" {
			cfg.CPS.I1 = &cps
			if cfg.Mask == "random" {
				cfg.Mask = "none"
			}
		} else {
			log.Println("Warning: --from-hex data too short (need >= 32 bytes/64 hex chars), ignoring")
		}
	}

	var kp *warp.KeyPair
	var reg *warp.RegistrationResponse
	var err error
	cachePath := *flagRegCache

	if !*flagFreshReg {
		kp, reg, err = warp.LoadRegistrationCache(cachePath)
		if err == nil {
			log.Printf("Loaded cached registration (account_type=%s, ipv4=%s)",
				reg.Account.AccountType, reg.Config.Addresses.V4)
			if cfg.LicenseKey != "" {
				log.Println("Info: license key ignored when using cached registration (use --fresh-reg to re-register)")
			}
		}
	}

	if kp == nil {
		log.Println("Generating Curve25519 key pair...")
		kp, err = warp.GenerateKeyPair()
		if err != nil {
			log.Fatalf("Failed to generate key pair: %v", err)
		}

		log.Println("Registering with Cloudflare WARP API...")
		reg, err = warp.Register(kp)
		if err != nil {
			log.Fatalf("Failed to register with WARP: %v", err)
		}
		log.Printf("Registered: account_type=%s, ipv4=%s",
			reg.Account.AccountType, reg.Config.Addresses.V4)

		if cfg.LicenseKey != "" {
			log.Println("Applying WARP+ license key...")
			if err := warp.ApplyLicenseKey(reg.ID, cfg.LicenseKey, reg.Token); err != nil {
				log.Printf("Warning: failed to apply license key: %v", err)
			} else {
				log.Println("WARP+ activated successfully")
			}
		}

		if err := warp.SaveRegistrationCache(cachePath, kp, reg); err != nil {
			log.Printf("Warning: failed to save registration cache: %v", err)
		}
	}

	awgParams := config.ConfigToAWGParams(cfg)
	cpsPackets := config.ConfigToCPS(cfg)

	wc := &awg.WireConfig{
		WarpConfig: reg,
		PrivateKey: kp.PrivateKey,
		Params:     awgParams,
		CPS:        cpsPackets,
		DNS:        cfg.DNS,
		Endpoint:   cfg.Endpoint,
		Keepalive:  cfg.Keepalive,
		MTU:        cfg.MTU,
	}

	conf, err := wc.Build()
	if err != nil {
		log.Fatalf("Failed to build config: %v", err)
	}

	outputPath := cfg.Output
	if outputPath == "" {
		outputPath = "warp-awg.conf"
	}

	if err := os.WriteFile(outputPath, []byte(conf), 0644); err != nil {
		log.Fatalf("Failed to write config: %v", err)
	}

	if !strings.HasSuffix(conf, "\n") {
	}

	fmt.Printf("Configuration saved to: %s\n", outputPath)
}

var (
	flagRegCache = flag.String("reg-cache", "warp-reg.json", "Registration cache file path")
	flagFreshReg = flag.Bool("fresh-reg", false, "Force fresh registration (ignore cache)")
)

func parseFlags() *config.CLIFlags {
	f := &config.CLIFlags{}

	flag.StringVar(&f.LicenseKey, "l", "", "WARP+ license key")
	flag.StringVar(&f.LicenseKey, "license-key", "", "WARP+ license key")
	flag.StringVar(&f.Output, "o", "", "Output file path (default: warp-awg.conf)")
	flag.StringVar(&f.Output, "output", "", "Output file path")
	flag.StringVar(&f.ConfigPath, "c", "", "Path to YAML config file")
	flag.StringVar(&f.ConfigPath, "config", "", "Path to YAML config file")
	flag.StringVar(&f.DNS, "dns", "", "DNS servers (default: 1.1.1.1, 1.0.0.1)")
	flag.StringVar(&f.Endpoint, "endpoint", "", "WARP endpoint (default: engage.cloudflareclient.com:2408)")
	flag.IntVar(&f.Keepalive, "keepalive", 0, "PersistentKeepalive interval (default: 25)")
	flag.IntVar(&f.MTU, "mtu", 0, "MTU (default: 1280)")
	flag.StringVar(&f.Mask, "mask", "", "Protocol mask preset: quic, dns, sip, random (default: random)")
	flag.StringVar(&f.Transport, "transport", "", "Transport protocol to mimic: quic, dns, sip, none")
	flag.StringVar(&f.FromHex, "from-hex", "", "Hex dump from Wireshark -> I1 CPS packet")

	flag.IntVar(&f.Jc, "jc", 0, "Junk packet count (0-10)")
	flag.IntVar(&f.Jmin, "jmin", 0, "Junk packet min size (64-1024)")
	flag.IntVar(&f.Jmax, "jmax", 0, "Junk packet max size (64-1024)")
	flag.IntVar(&f.S1, "s1", -1, "S1 random prefix length (0-64 bytes)")
	flag.IntVar(&f.S2, "s2", -1, "S2 random prefix length (0-64 bytes)")
	flag.IntVar(&f.S3, "s3", -1, "S3 random prefix length (0-64 bytes)")
	flag.IntVar(&f.S4, "s4", -1, "S4 random prefix length (0-32 bytes)")
	flag.StringVar(&f.H1, "h1", "", "H1 range (e.g. 100000-200000)")
	flag.StringVar(&f.H2, "h2", "", "H2 range")
	flag.StringVar(&f.H3, "h3", "", "H3 range")
	flag.StringVar(&f.H4, "h4", "", "H4 range")
	flag.StringVar(&f.I1, "i1", "", "I1 CPS expression")
	flag.StringVar(&f.I2, "i2", "", "I2 CPS expression")
	flag.StringVar(&f.I3, "i3", "", "I3 CPS expression")
	flag.StringVar(&f.I4, "i4", "", "I4 CPS expression")
	flag.StringVar(&f.I5, "i5", "", "I5 CPS expression")

	flag.Parse()

	return f
}

func loadConfig(flags *config.CLIFlags) *config.Config {
	cfg := config.Default()

	if flags.ConfigPath != "" {
		loadedCfg, err := config.Load(flags.ConfigPath)
		if err != nil {
			log.Fatalf("Failed to load config file: %v", err)
		}
		cfg = loadedCfg
	}

	cfg = config.Merge(cfg, flags)
	return cfg
}

func init() {
	log.SetPrefix("[warp-awg] ")
	log.SetFlags(log.Ltime | log.Lmsgprefix)
}
