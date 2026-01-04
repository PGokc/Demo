package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"ops_local_demo/internal/config"
	"ops_local_demo/internal/server"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	cfg := config.LoadFromEnv()
	safe := cfg.Safe()
	if safe.OpsHost == "" {
		safe.OpsHost = "https://ops.kabeta4.statusfeishu.cn"
	}
	log.Printf("step=startup delivery_base=%s ka=%s redirect_path=%s ops_host=%s public_key_path=%s session_secret_set=%t tls_cert=%s tls_key=%s", safe.DeliveryBase, safe.KA, safe.RedirectPath, safe.OpsHost, safe.DeliveryPublicKeyPath, safe.SessionSecretSet, safe.TlsCertFile, safe.TlsKeyFile)

	srvImpl, err := server.New(cfg)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	addr := os.Getenv("LISTEN_ADDR")
	certFile := cfg.TlsCertFile
	keyFile := cfg.TlsKeyFile

	if addr == "" {
		if certFile != "" && keyFile != "" {
			addr = ":443"
		} else {
			addr = ":8080"
		}
	}

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      srvImpl.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("step=listen addr=%s tls_cert=%s tls_key=%s", addr, certFile, keyFile)

	if certFile != "" && keyFile != "" {
		err = httpServer.ListenAndServeTLS(certFile, keyFile)
	} else {
		err = httpServer.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("server exited with error: %v", err)
	}
}
