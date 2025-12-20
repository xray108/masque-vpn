package common

import (
	common_fec "github.com/iselt/masque-vpn/common/fec"
)

// ClientConfig структура для хранения конфигурации клиента из TOML файла
type ClientConfig struct {
	ServerAddr         string `toml:"server_addr"`
	ServerName         string `toml:"server_name"`
	CAFile             string `toml:"ca_file"`
	CAPEM              string `toml:"ca_pem"`
	TLSCert            string `toml:"tls_cert"`
	TLSKey             string `toml:"tls_key"`
	CertPEM            string `toml:"cert_pem"`
	KeyPEM             string `toml:"key_pem"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`
	TunName            string `toml:"tun_name"`
	KeyLogFile         string `toml:"key_log_file"`
	LogLevel           string `toml:"log_level"`
	MTU                int    `toml:"mtu"`
	PreferIPv6         bool   `toml:"prefer_ipv6"`
	FEC                common_fec.Config `toml:"fec"`
}

// APIServerConfig 结构体，用于存储 API 服务器的配置信息
type APIServerConfig struct {
	ListenAddr   string `toml:"listen_addr"`
	StaticDir    string `toml:"static_dir"`
	DatabasePath string `toml:"database_path"`
}

// ServerConfig структура для хранения конфигурации сервера из TOML файла
type ServerConfig struct {
	ListenAddr      string   `toml:"listen_addr"`
	CertFile        string   `toml:"cert_file"`
	KeyFile         string   `toml:"key_file"`
	CACertFile      string   `toml:"ca_cert_file"`
	CAKeyFile       string   `toml:"ca_key_file"`
	CertPEM         string   `toml:"cert_pem"`
	KeyPEM          string   `toml:"key_pem"`
	CAKeyPEM        string   `toml:"ca_key_pem"`
	CACertPEM       string   `toml:"ca_cert_pem"`
	AssignCIDR      string   `toml:"assign_cidr"`
	AssignCIDRv6    string   `toml:"assign_cidr_v6"`
	AdvertiseRoutes []string `toml:"advertise_routes"`
	AdvertiseRoutesv6 []string `toml:"advertise_routes_v6"`
	TunName         string   `toml:"tun_name"`
	LogLevel        string   `toml:"log_level"`
	ServerName      string   `toml:"server_name"`
	MTU             int      `toml:"mtu"`
	EnableIPv6      bool     `toml:"enable_ipv6"`

	// API server configuration
	APIServer APIServerConfig `toml:"api_server"`

	// Metrics configuration
	Metrics MetricsConfig `toml:"metrics"`

	// FEC configuration
	FEC common_fec.Config `toml:"fec"`
}

// MetricsConfig holds metrics server configuration
type MetricsConfig struct {
	Enabled    bool   `toml:"enabled"`
	ListenAddr string `toml:"listen_addr"`
}
