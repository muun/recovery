module recovery_tool

go 1.12

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/btcsuite/btcwallet v0.10.0
	github.com/btcsuite/btcwallet/walletdb v1.1.0
	github.com/lightninglabs/neutrino v0.10.0
	github.com/muun/libwallet v0.1.4
)

replace github.com/lightninglabs/neutrino => github.com/muun/neutrino v0.0.0-20190914162326-7082af0fa257
