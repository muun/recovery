package electrum

import "sync/atomic"

// ServerProvider manages a rotating server list, from which callers can pull server addresses.
type ServerProvider struct {
	nextIndex int32
}

// NewServerProvider returns an initialized ServerProvider.
func NewServerProvider() *ServerProvider {
	return &ServerProvider{-1}
}

// NextServer returns an address from the rotating list. It's thread-safe.
func (p *ServerProvider) NextServer() string {
	index := int(atomic.AddInt32(&p.nextIndex, 1))
	return PublicServers[index%len(PublicServers)]
}

// PublicServers list.
//
// This list was taken from Electrum repositories, keeping TLS servers and excluding onion URIs.
// It was then sorted into sections using the `cmd/survey` program, to prioritize the more reliable
// servers with batch support.
//
// See https://github.com/spesmilo/electrum/blob/master/electrum/servers.json
// See https://github.com/kyuupichan/electrumx/blob/master/electrumx/lib/coins.py
// See `cmd/survey/main.go`
//
var PublicServers = []string{
	// With batch support:
	"electrum.hsmiths.com:50002",
	"E-X.not.fyi:50002",
	"VPS.hsmiths.com:50002",
	"btc.cihar.com:50002",
	"e.keff.org:50002",
	"electrum.qtornado.com:50002",
	"electrum.emzy.de:50002",
	"tardis.bauerj.eu:50002",
	"electrum.hodlister.co:50002",
	"electrum3.hodlister.co:50002",
	"electrum5.hodlister.co:50002",
	"fortress.qtornado.com:443",
	"electrumx.erbium.eu:50002",
	"bitcoin.lukechilds.co:50002",
	"electrum.bitkoins.nl:50512",

	// Without batch support:
	"electrum.aantonop.com:50002",
	"electrum.blockstream.info:50002",
	"blockstream.info:700",

	// Unclassified:
	"81-7-10-251.blue.kundencontroller.de:50002",
	"b.ooze.cc:50002",
	"bitcoin.corgi.party:50002",
	"bitcoins.sk:50002",
	"btc.xskyx.net:50002",
	"electrum.jochen-hoenicke.de:50005",
	"dragon085.startdedicated.de:50002",
	"e-1.claudioboxx.com:50002",
	"electrum-server.ninja:50002",
	"electrum-unlimited.criptolayer.net:50002",
	"electrum.eff.ro:50002",
	"electrum.festivaldelhumor.org:50002",
	"electrum.leblancnet.us:50002",
	"electrum.mindspot.org:50002",
	"electrum.taborsky.cz:50002",
	"electrum.villocq.com:50002",
	"electrum2.eff.ro:50002",
	"electrum2.villocq.com:50002",
	"electrumx.bot.nu:50002",
	"electrumx.ddns.net:50002",
	"electrumx.ftp.sh:50002",
	"electrumx.soon.it:50002",
	"elx01.knas.systems:50002",
	"fedaykin.goip.de:50002",
	"fn.48.org:50002",
	"ndnd.selfhost.eu:50002",
	"orannis.com:50002",
	"rbx.curalle.ovh:50002",
	"technetium.network:50002",
	"tomscryptos.com:50002",
	"ulrichard.ch:50002",
	"vmd27610.contaboserver.net:50002",
	"vmd30612.contaboserver.net:50002",
	"xray587.startdedicated.de:50002",
	"yuio.top:50002",
	"bitcoin.dragon.zone:50004",
	"ecdsa.net:110",
	"btc.usebsv.com:50006",
	"e2.keff.org:50002",
	"electrumx.electricnewyear.net:50002",
	"green-gold.westeurope.cloudapp.azure.com:56002",
	"electrumx-core.1209k.com:50002",
	"bitcoin.aranguren.org:50002",
}
