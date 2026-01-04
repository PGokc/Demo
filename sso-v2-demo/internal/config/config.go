package config

import (
	"os"
	"strings"
)

type Config struct {
	KA                    string
	DeliveryBase          string
	OpsHost               string
	RedirectPath          string
	SessionSecret         string
	DeliveryPublicKeyPath string
	TlsCertFile           string
	TlsKeyFile            string
}

type SafeConfig struct {
	KA                    string
	DeliveryBase          string
	OpsHost               string
	RedirectPath          string
	SessionSecretSet      bool
	DeliveryPublicKeyPath string
	TlsCertFile           string
	TlsKeyFile            string
}

func LoadFromEnv() Config {
	cfg := Config{
		KA:                    os.Getenv("KA"),
		DeliveryBase:          os.Getenv("DELIVERY_BASE"),
		OpsHost:               os.Getenv("OPS_HOST"),
		RedirectPath:          os.Getenv("REDIRECT_PATH"),
		SessionSecret:         os.Getenv("SESSION_SECRET"),
		DeliveryPublicKeyPath: os.Getenv("DELIVERY_PUBLIC_KEY_PATH"),
		TlsCertFile:           os.Getenv("TLS_CERT_FILE"),
		TlsKeyFile:            os.Getenv("TLS_KEY_FILE"),
	}

	if cfg.KA == "" {
		cfg.KA = "kagaussdb"
	}
	if cfg.SessionSecret == "" {
		cfg.SessionSecret = "default-secret"
	}
	if cfg.TlsCertFile == "" || cfg.TlsKeyFile == "" {
		cfg.TlsCertFile = "tls.cert"
		cfg.TlsKeyFile = "tls.key"
	}

	if cfg.DeliveryBase == "" {
		cfg.DeliveryBase = "https://delivery.feishu.cn"
	}
	cfg.DeliveryBase = strings.TrimRight(cfg.DeliveryBase, "/")

	if cfg.RedirectPath == "" {
		cfg.RedirectPath = "/"
	} else if !strings.HasPrefix(cfg.RedirectPath, "/") {
		cfg.RedirectPath = "/" + cfg.RedirectPath
	}

	return cfg
}

func (c Config) Safe() SafeConfig {
	return SafeConfig{
		KA:                    c.KA,
		DeliveryBase:          c.DeliveryBase,
		OpsHost:               c.OpsHost,
		RedirectPath:          c.RedirectPath,
		DeliveryPublicKeyPath: c.DeliveryPublicKeyPath,
		SessionSecretSet:      c.SessionSecret != "",
		TlsCertFile:           c.TlsCertFile,
		TlsKeyFile:            c.TlsKeyFile,
	}
}
