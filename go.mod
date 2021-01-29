module github.com/muun/recovery

go 1.12

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f // indirect
	github.com/btcsuite/btcutil v1.0.2
	github.com/btcsuite/btcwallet v0.11.1-0.20200612012534-48addcd5591a // indirect
	github.com/btcsuite/btcwallet/walletdb v1.3.3 // indirect
	github.com/lightninglabs/neutrino v0.11.1-0.20200316235139-bffc52e8f200 // indirect
	github.com/muun/libwallet v0.7.0
)

replace github.com/lightninglabs/neutrino => github.com/muun/neutrino v0.0.0-20190914162326-7082af0fa257
