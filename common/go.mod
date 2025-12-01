module github.com/iselt/masque-vpn/common

go 1.25.0

require (
	github.com/iselt/connect-ip-go v0.0.0-20250409071859-bc9a9fcba51d
	github.com/quic-go/quic-go v0.57.1
	github.com/vishvananda/netlink v1.3.0
	golang.zx2c4.com/wireguard v0.0.0-20231211153847-12269c276173
	golang.zx2c4.com/wireguard/windows v0.5.3
)

require (
	github.com/dunglas/httpsfv v1.1.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
)

replace github.com/iselt/connect-ip-go => ../connect-ip-go
