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
	// Fast servers with batching
	"blackie.c3-soft.com:57002",             // impl: Fulcrum 1.6.0, batching: true, ttc: 0.72, speed: 97, from: fortress.qtornado.com:443
	"fullnode.titanconnect.ca:50002",        // impl: Fulcrum 1.6.0, batching: true, ttc: 0.66, speed: 86, from: fortress.qtornado.com:443
	"de.poiuty.com:50002",                   // impl: Fulcrum 1.6.0, batching: true, ttc: 0.76, speed: 78, from: fortress.qtornado.com:443
	"2ex.digitaleveryware.com:50002",        // impl: ElectrumX 1.16.0, batching: true, ttc: 0.68, speed: 68, from: fortress.qtornado.com:443
	"fortress.qtornado.com:443",             // impl: ElectrumX 1.16.0, batching: true, ttc: 0.77, speed: 67, from:
	"hodlers.beer:50002",                    // impl: ElectrumX 1.16.0, batching: true, ttc: 0.74, speed: 64, from: fortress.qtornado.com:443
	"f.keff.org:50002",                      // impl: Fulcrum 1.6.0, batching: true, ttc: 0.76, speed: 60, from: fortress.qtornado.com:443
	"e2.keff.org:50002",                     // impl: Fulcrum 1.6.0, batching: true, ttc: 0.75, speed: 58, from:
	"electrum.stippy.com:50002",             // impl: ElectrumX 1.16.0, batching: true, ttc: 0.75, speed: 56, from: fortress.qtornado.com:443
	"electrum.privateservers.network:50002", // impl: ElectrumX 1.15.0, batching: true, ttc: 0.83, speed: 52, from: fortress.qtornado.com:443
	"fulcrum.grey.pw:51002",                 // impl: Fulcrum 1.6.0, batching: true, ttc: 0.79, speed: 49, from: fortress.qtornado.com:443
	"btc.electroncash.dk:60002",             // impl: Fulcrum 1.6.0, batching: true, ttc: 0.79, speed: 49, from: fortress.qtornado.com:443
	"node.degga.net:50002",                  // impl: ElectrumX 1.16.0, batching: true, ttc: 0.59, speed: 46, from: fortress.qtornado.com:443
	"e.keff.org:50002",                      // impl: ElectrumX 1.10.0, batching: true, ttc: 0.76, speed: 45, from:
	"bitcoin.aranguren.org:50002",           // impl: Fulcrum 1.6.0, batching: true, ttc: 1.25, speed: 41, from:
	"electrum.helali.me:50002",              // impl: Fulcrum 1.6.0, batching: true, ttc: 1.20, speed: 38, from: fortress.qtornado.com:443
	"node1.btccuracao.com:50002",            // impl: ElectrumX 1.16.0, batching: true, ttc: 0.79, speed: 33, from: fortress.qtornado.com:443
	"ASSUREDLY.not.fyi:50002",               // impl: ElectrumX 1.16.0, batching: true, ttc: 0.66, speed: 20, from: electrumx.erbium.eu:50002
	"assuredly.not.fyi:50002",               // impl: ElectrumX 1.16.0, batching: true, ttc: 0.74, speed: 20, from: fortress.qtornado.com:443

	// Other servers
	"smmalis37.ddns.net:50002",            // impl: ElectrumX 1.16.0, batching: true, ttc: 0.56, speed: 19, from: fortress.qtornado.com:443
	"electrumx.papabyte.com:50002",        // impl: ElectrumX 1.16.0, batching: true, ttc: 0.95, speed: 19, from: fortress.qtornado.com:443
	"electrum.bitaroo.net:50002",          // impl: ElectrumX 1.16.0, batching: true, ttc: 0.98, speed: 19, from: fortress.qtornado.com:443
	"electrum.exan.tech:443",              // impl: ElectrumX 1.16.0, batching: true, ttc: 1.08, speed: 19, from: fortress.qtornado.com:443
	"electrum-fulcrum.toggen.net:50002",   // impl: Fulcrum 1.6.0, batching: true, ttc: 1.47, speed: 16, from: fortress.qtornado.com:443
	"vmd84592.contaboserver.net:50002",    // impl: ElectrumX 1.16.0, batching: true, ttc: 0.79, speed: 14, from: fortress.qtornado.com:443
	"gd42.org:50002",                      // impl: ElectrumX 1.16.0, batching: true, ttc: 0.50, speed: 12, from: fortress.qtornado.com:443
	"vmd63185.contaboserver.net:50002",    // impl: ElectrumX 1.16.0, batching: true, ttc: 0.75, speed: 12, from: fortress.qtornado.com:443
	"alfa.cryptotrek.online:50002",        // impl: ElectrumX 1.16.0, batching: true, ttc: 0.75, speed: 12, from: fortress.qtornado.com:443
	"electrumx.ultracloud.tk:50002",       // impl: ElectrumX 1.16.0, batching: true, ttc: 0.76, speed: 12, from: fortress.qtornado.com:443
	"electrum.brainshome.de:50002",        // impl: ElectrumX 1.16.0, batching: true, ttc: 0.89, speed: 12, from: fortress.qtornado.com:443
	"ragtor.duckdns.org:50002",            // impl: ElectrumX 1.16.0, batching: true, ttc: 0.91, speed: 12, from: bitcoin.aranguren.org:50002
	"94.23.247.135:50002",                 // impl: ElectrumX 1.16.0, batching: true, ttc: 6.80, speed: 12, from: fortress.qtornado.com:443
	"eai.coincited.net:50002",             // impl: ElectrumX 1.16.0, batching: true, ttc: 0.12, speed: 11, from: electrumx.erbium.eu:50002
	"142.93.6.38:50002",                   // impl: ElectrumX 1.16.0, batching: true, ttc: 0.56, speed: 11, from: fortress.qtornado.com:443
	"157.245.172.236:50002",               // impl: ElectrumX 1.16.0, batching: true, ttc: 0.62, speed: 11, from: fortress.qtornado.com:443
	"electrum.kendigisland.xyz:50002",     // impl: ElectrumX 1.16.0, batching: true, ttc: 0.68, speed: 11, from: fortress.qtornado.com:443
	"btc.lastingcoin.net:50002",           // impl: ElectrumX 1.16.0, batching: true, ttc: 0.74, speed: 11, from: fortress.qtornado.com:443
	"xtrum.com:50002",                     // impl: ElectrumX 1.16.0, batching: true, ttc: 0.75, speed: 11, from: electrumx.erbium.eu:50002
	"blkhub.net:50002",                    // impl: ElectrumX 1.16.0, batching: true, ttc: 0.76, speed: 11, from: fortress.qtornado.com:443
	"electrumx.erbium.eu:50002",           // impl: ElectrumX 1.16.0, batching: true, ttc: 0.78, speed: 11, from:
	"185.64.116.15:50002",                 // impl: ElectrumX 1.16.0, batching: true, ttc: 0.79, speed: 11, from: bitcoin.aranguren.org:50002
	"188.165.206.215:50002",               // impl: ElectrumX 1.16.0, batching: true, ttc: 0.81, speed: 11, from: fortress.qtornado.com:443
	"ex03.axalgo.com:50002",               // impl: ElectrumX 1.16.0, batching: true, ttc: 0.92, speed: 11, from: fortress.qtornado.com:443
	"electrum.bitcoinlizard.net:50002",    // impl: ElectrumX 1.16.0, batching: true, ttc: 0.98, speed: 11, from: fortress.qtornado.com:443
	"btce.iiiiiii.biz:50002",              // impl: ElectrumX 1.16.0, batching: true, ttc: 1.02, speed: 11, from: fortress.qtornado.com:443
	"btc.ocf.sh:50002",                    // impl: ElectrumX 1.16.0, batching: true, ttc: 1.15, speed: 11, from: fortress.qtornado.com:443
	"68.183.188.105:50002",                // impl: ElectrumX 1.16.0, batching: true, ttc: 1.36, speed: 11, from: fortress.qtornado.com:443
	"vmd71287.contaboserver.net:50002",    // impl: ElectrumX 1.16.0, batching: true, ttc: 0.60, speed: 10, from: fortress.qtornado.com:443
	"2electrumx.hopto.me:56022",           // impl: ElectrumX 1.16.0, batching: true, ttc: 0.89, speed: 10, from: fortress.qtornado.com:443
	"electrum.emzy.de:50002",              // impl: ElectrumX 1.16.0, batching: true, ttc: 1.15, speed: 10, from:
	"electrumx.electricnewyear.net:50002", // impl: ElectrumX 1.15.0, batching: true, ttc: 0.66, speed: 9, from:
	"walle.dedyn.io:50002",                // impl: ElectrumX 1.16.0, batching: true, ttc: 0.76, speed: 8, from: fortress.qtornado.com:443
	"caleb.vegas:50002",                   // impl: ElectrumX 1.16.0, batching: true, ttc: 0.80, speed: 8, from: fortress.qtornado.com:443
	"electrum.neocrypto.io:50002",         // impl: ElectrumX 1.16.0, batching: true, ttc: 0.69, speed: 7, from: fortress.qtornado.com:443
	"guichet.centure.cc:50002",            // impl: ElectrumX 1.16.0, batching: true, ttc: 0.71, speed: 7, from: fortress.qtornado.com:443
	"167.172.42.31:50002",                 // impl: ElectrumX 1.16.0, batching: true, ttc: 0.75, speed: 7, from: fortress.qtornado.com:443
	"2AZZARITA.hopto.org:50002",           // impl: ElectrumX 1.16.0, batching: true, ttc: 0.77, speed: 7, from: fortress.qtornado.com:443
	"jonas.reptiles.se:50002",             // impl: ElectrumX 1.16.0, batching: true, ttc: 0.95, speed: 7, from: fortress.qtornado.com:443
	"167.172.226.175:50002",               // impl: ElectrumX 1.16.0, batching: true, ttc: 0.61, speed: 6, from: fortress.qtornado.com:443
	"electrum-btc.leblancnet.us:50002",    // impl: ElectrumX 1.15.0, batching: true, ttc: 0.64, speed: 6, from: fortress.qtornado.com:443
	"104.248.139.211:50002",               // impl: ElectrumX 1.16.0, batching: true, ttc: 0.90, speed: 6, from: fortress.qtornado.com:443
	"elx.bitske.com:50002",                // impl: ElectrumX 1.16.0, batching: true, ttc: 1.01, speed: 6, from: fortress.qtornado.com:443
	"kareoke.qoppa.org:50002",             // impl: ElectrumX 1.16.0, batching: true, ttc: 1.18, speed: 6, from: electrumx.erbium.eu:50002
	"ex.btcmp.com:50002",                  // impl: ElectrumX 1.16.0, batching: true, ttc: 1.12, speed: 5, from: fortress.qtornado.com:443
	"bitcoins.sk:56002",                   // impl: ElectrumX 1.14.0, batching: true, ttc: 0.77, speed: 3, from: fortress.qtornado.com:443
	"73.92.198.54:50002",                  // impl: ElectrumX 1.15.0, batching: true, ttc: 5.57, speed: 0, from: fortress.qtornado.com:443
}
