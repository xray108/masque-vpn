package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"vpn-server/internal/server"

	"github.com/BurntSushi/toml"
	common "github.com/iselt/masque-vpn/common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	serverConfig common.ServerConfig
	logger       *zap.Logger
)

// initLogger initializes structured logging with zap
func initLogger(logLevel string) error {
	var config zap.Config
	
	// Determine environment and configure accordingly
	if os.Getenv("ENVIRONMENT") == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	
	// Set log level from configuration
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		level = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(level)
	
	// Configure time encoding
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	
	// Build logger
	logger, err = config.Build()
	if err != nil {
		return err
	}
	
	// Replace global logger
	zap.ReplaceGlobals(logger)
	
	return nil
}

func main() {
	if os.Getenv("PERF_PROFILE") != "" {
		f, _ := os.OpenFile("cpu.pprof", os.O_CREATE|os.O_RDWR, 0666)
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Parse command line flags
	configFile := flag.String("c", "config.server.toml", "Config file path")
	flag.Parse()
	
	// Load configuration
	if _, err := toml.DecodeFile(*configFile, &serverConfig); err != nil {
		// Use standard log for initial errors before logger is initialized
		panic("Error loading config file " + *configFile + ": " + err.Error())
	}

	// Initialize structured logging
	if err := initLogger(serverConfig.LogLevel); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("Starting MASQUE VPN Server",
		zap.String("config_file", *configFile),
		zap.String("log_level", serverConfig.LogLevel),
		zap.String("listen_addr", serverConfig.ListenAddr),
		zap.String("server_name", serverConfig.ServerName),
	)

	// Validate required configuration
	if serverConfig.ListenAddr == "" || serverConfig.CertFile == "" || serverConfig.KeyFile == "" ||
		serverConfig.AssignCIDR == "" || serverConfig.ServerName == "" {
		logger.Fatal("Missing required configuration values",
			zap.String("listen_addr", serverConfig.ListenAddr),
			zap.String("cert_file", serverConfig.CertFile),
			zap.String("key_file", serverConfig.KeyFile),
			zap.String("assign_cidr", serverConfig.AssignCIDR),
			zap.String("server_name", serverConfig.ServerName),
		)
	}

	// Initialize Server
	srv, err := server.New(serverConfig)
	if err != nil {
		logger.Fatal("Failed to initialize server", zap.Error(err))
	}

	// Create context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("Server initialized successfully, starting...")

	// Run Server
	if err := srv.Run(ctx); err != nil {
		logger.Error("Server error", zap.Error(err))
	}
}

