module github.com/muun/recovery_tool

go 1.12

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f
	github.com/btcsuite/btcutil v1.0.2
	github.com/btcsuite/btcwallet v0.11.1-0.20200612012534-48addcd5591a
	github.com/btcsuite/btcwallet/walletdb v1.3.3
	github.com/lightninglabs/neutrino v0.11.1-0.20200316235139-bffc52e8f200
	github.com/muun/libwallet v0.5.0
	github.com/pkg/errors v0.9.1 // indirect
)

replace github.com/lightninglabs/neutrino => github.com/muun/neutrino v0.0.0-20190914162326-7082af0fa257
