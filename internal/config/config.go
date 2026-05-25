package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"warp-awg-gen/internal/awg"
)

type Config struct {
	LicenseKey string `yaml:"license_key"`
	Output     string `yaml:"output"`
	DNS        string `yaml:"dns"`
	Endpoint   string `yaml:"endpoint"`
	Keepalive  int    `yaml:"keepalive"`
	MTU        int    `yaml:"mtu"`
	Transport  string `yaml:"transport"`
	Mask       string `yaml:"mask"`
	FromHex    string `yaml:"from_hex"`
	RegCache   string `yaml:"reg_cache"`
	FreshReg   bool   `yaml:"fresh_reg"`

	AWG struct {
		Jc   *int    `yaml:"jc"`
		Jmin *int    `yaml:"jmin"`
		Jmax *int    `yaml:"jmax"`
		S1   *int    `yaml:"s1"`
		S2   *int    `yaml:"s2"`
		S3   *int    `yaml:"s3"`
		S4   *int    `yaml:"s4"`
		H1   *string `yaml:"h1"`
		H2   *string `yaml:"h2"`
		H3   *string `yaml:"h3"`
		H4   *string `yaml:"h4"`
	} `yaml:"awg"`

	CPS struct {
		I1 *string `yaml:"i1"`
		I2 *string `yaml:"i2"`
		I3 *string `yaml:"i3"`
		I4 *string `yaml:"i4"`
		I5 *string `yaml:"i5"`
	} `yaml:"cps"`
}

type CLIFlags struct {
	LicenseKey string
	Output     string
	ConfigPath string
	DNS        string
	Endpoint   string
	Keepalive  int
	MTU        int
	Mask       string
	Transport  string
	FromHex    string
	RegCache   string
	FreshReg   bool
	Jc         int
	Jmin       int
	Jmax       int
	S1         int
	S2         int
	S3         int
	S4         int
	H1         string
	H2         string
	H3         string
	H4         string
	I1         string
	I2         string
	I3         string
	I4         string
	I5         string
}

func Default() *Config {
	cfg := &Config{
		DNS:       "1.1.1.1, 1.0.0.1, 2606:4700:4700::1111, 2606:4700:4700::1001",
		Endpoint:  "162.159.192.1:943",
		Keepalive: 25,
		MTU:       1280,
		Transport: "none",
		Mask:      "random",
	}
	return cfg
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

func Merge(cfg *Config, flags *CLIFlags) *Config {
	result := Default()

	result.DNS = cfg.DNS
	result.Endpoint = cfg.Endpoint
	result.Keepalive = cfg.Keepalive
	result.MTU = cfg.MTU
	result.Transport = cfg.Transport
	result.Mask = cfg.Mask
	result.LicenseKey = cfg.LicenseKey
	result.Output = cfg.Output
	result.FromHex = cfg.FromHex

	if cfg.AWG.S1 != nil {
		result.AWG.S1 = cfg.AWG.S1
	}
	if cfg.AWG.S2 != nil {
		result.AWG.S2 = cfg.AWG.S2
	}
	if cfg.AWG.S3 != nil {
		result.AWG.S3 = cfg.AWG.S3
	}
	if cfg.AWG.S4 != nil {
		result.AWG.S4 = cfg.AWG.S4
	}
	if cfg.AWG.H1 != nil {
		result.AWG.H1 = cfg.AWG.H1
	}
	if cfg.AWG.H2 != nil {
		result.AWG.H2 = cfg.AWG.H2
	}
	if cfg.AWG.H3 != nil {
		result.AWG.H3 = cfg.AWG.H3
	}
	if cfg.AWG.H4 != nil {
		result.AWG.H4 = cfg.AWG.H4
	}
	if cfg.AWG.Jc != nil {
		result.AWG.Jc = cfg.AWG.Jc
	}
	if cfg.AWG.Jmin != nil {
		result.AWG.Jmin = cfg.AWG.Jmin
	}
	if cfg.AWG.Jmax != nil {
		result.AWG.Jmax = cfg.AWG.Jmax
	}

	if cfg.CPS.I1 != nil {
		result.CPS.I1 = cfg.CPS.I1
	}
	if cfg.CPS.I2 != nil {
		result.CPS.I2 = cfg.CPS.I2
	}
	if cfg.CPS.I3 != nil {
		result.CPS.I3 = cfg.CPS.I3
	}
	if cfg.CPS.I4 != nil {
		result.CPS.I4 = cfg.CPS.I4
	}
	if cfg.CPS.I5 != nil {
		result.CPS.I5 = cfg.CPS.I5
	}

	if flags.LicenseKey != "" {
		result.LicenseKey = flags.LicenseKey
	}
	if flags.Output != "" {
		result.Output = flags.Output
	}
	if flags.DNS != "" {
		result.DNS = flags.DNS
	}
	if flags.Endpoint != "" {
		result.Endpoint = flags.Endpoint
	}
	if flags.Keepalive > 0 {
		result.Keepalive = flags.Keepalive
	}
	if flags.MTU > 0 {
		result.MTU = flags.MTU
	}
	if flags.Mask != "" {
		result.Mask = flags.Mask
	}
	if flags.Transport != "" {
		result.Transport = flags.Transport
	}
	if flags.FromHex != "" {
		result.FromHex = flags.FromHex
	}

	if flags.Jc > 0 {
		v := flags.Jc
		result.AWG.Jc = &v
	}
	if flags.Jmin > 0 {
		v := flags.Jmin
		result.AWG.Jmin = &v
	}
	if flags.Jmax > 0 {
		v := flags.Jmax
		result.AWG.Jmax = &v
	}
	if flags.S1 >= 0 {
		v := flags.S1
		result.AWG.S1 = &v
	}
	if flags.S2 >= 0 {
		v := flags.S2
		result.AWG.S2 = &v
	}
	if flags.S3 >= 0 {
		v := flags.S3
		result.AWG.S3 = &v
	}
	if flags.S4 >= 0 {
		v := flags.S4
		result.AWG.S4 = &v
	}
	if flags.H1 != "" {
		result.AWG.H1 = &flags.H1
	}
	if flags.H2 != "" {
		result.AWG.H2 = &flags.H2
	}
	if flags.H3 != "" {
		result.AWG.H3 = &flags.H3
	}
	if flags.H4 != "" {
		result.AWG.H4 = &flags.H4
	}
	if flags.I1 != "" {
		result.CPS.I1 = &flags.I1
	}
	if flags.I2 != "" {
		result.CPS.I2 = &flags.I2
	}
	if flags.I3 != "" {
		result.CPS.I3 = &flags.I3
	}
	if flags.I4 != "" {
		result.CPS.I4 = &flags.I4
	}
	if flags.I5 != "" {
		result.CPS.I5 = &flags.I5
	}

	return result
}

func ConfigToAWGParams(cfg *Config) *awg.AWGParams {
	p := awg.GenerateDefaultParams()

	if cfg.AWG.Jc != nil {
		p.Jc = *cfg.AWG.Jc
	}
	if cfg.AWG.Jmin != nil {
		p.Jmin = *cfg.AWG.Jmin
	}
	if cfg.AWG.Jmax != nil {
		p.Jmax = *cfg.AWG.Jmax
	}
	if cfg.AWG.S1 != nil {
		p.S1 = *cfg.AWG.S1
	}
	if cfg.AWG.S2 != nil {
		p.S2 = *cfg.AWG.S2
	}
	if cfg.AWG.S3 != nil {
		p.S3 = *cfg.AWG.S3
	}
	if cfg.AWG.S4 != nil {
		p.S4 = *cfg.AWG.S4
	}
	if cfg.AWG.H1 != nil {
		p.H1 = *cfg.AWG.H1
	}
	if cfg.AWG.H2 != nil {
		p.H2 = *cfg.AWG.H2
	}
	if cfg.AWG.H3 != nil {
		p.H3 = *cfg.AWG.H3
	}
	if cfg.AWG.H4 != nil {
		p.H4 = *cfg.AWG.H4
	}

	return p
}

func ConfigToCPS(cfg *Config) *awg.CPSPackets {
	mask := cfg.Mask
	if mask == "" {
		mask = "none"
	}

	cps := awg.GenerateCPS(mask)

	if cfg.CPS.I1 != nil {
		cps.I1 = *cfg.CPS.I1
	}
	if cfg.CPS.I2 != nil {
		cps.I2 = *cfg.CPS.I2
	}
	if cfg.CPS.I3 != nil {
		cps.I3 = *cfg.CPS.I3
	}
	if cfg.CPS.I4 != nil {
		cps.I4 = *cfg.CPS.I4
	}
	if cfg.CPS.I5 != nil {
		cps.I5 = *cfg.CPS.I5
	}

	return cps
}
