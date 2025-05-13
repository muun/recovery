module github.com/muun/recovery

go 1.12

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/gookit/color v1.4.2
	github.com/muun/libwallet v0.11.0
)

replace github.com/muun/libwallet => ../libwallet

replace github.com/lightninglabs/neutrino => github.com/muun/neutrino v0.0.0-20190914162326-7082af0fa257
