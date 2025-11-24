package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"

	"vpn-server/internal/server"

	"github.com/BurntSushi/toml"
	common "github.com/iselt/masque-vpn/common"
)

var serverConfig common.ServerConfig

func main() {
	if os.Getenv("PERF_PROFILE") != "" {
		f, _ := os.OpenFile("cpu.pprof", os.O_CREATE|os.O_RDWR, 0666)
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// --- 配置加载 ---
	configFile := flag.String("c", "config.server.toml", "Config file path")
	flag.Parse()
	if _, err := toml.DecodeFile(*configFile, &serverConfig); err != nil {
		log.Fatalf("Error loading config file %s: %v", *configFile, err)
	}

	// --- 基础验证 ---
	if serverConfig.ListenAddr == "" || serverConfig.CertFile == "" || serverConfig.KeyFile == "" ||
		serverConfig.AssignCIDR == "" || serverConfig.ServerName == "" {
		log.Fatal("Missing required configuration values in config.server.toml")
	}

	// Initialize Server
	srv, err := server.New(serverConfig)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	// Run Server
	srv.Run()
}

