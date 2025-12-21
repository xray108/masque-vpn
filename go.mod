module masque-test

go 1.25

require (
	github.com/iselt/masque-vpn/common v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.0
)

require (
	github.com/quic-go/quic-go v0.57.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	golang.zx2c4.com/wireguard v0.0.0-20231211153847-12269c276173 // indirect
)

replace github.com/iselt/masque-vpn/common => ./common
