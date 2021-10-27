package main

import (
	"fmt"
	"log"

	"github.com/muun/libwallet"
)

type signingDetails struct {
	Address libwallet.MuunAddress
}

type AddressGenerator struct {
	addrs   map[string]signingDetails
	userKey *libwallet.HDPrivateKey
	muunKey *libwallet.HDPrivateKey
}

func NewAddressGenerator(userKey, muunKey *libwallet.HDPrivateKey) *AddressGenerator {
	return &AddressGenerator{
		addrs:   make(map[string]signingDetails),
		userKey: userKey,
		muunKey: muunKey,
	}
}

func (g *AddressGenerator) Addresses() map[string]signingDetails {
	return g.addrs
}

// Stream returns a channel that emits all addresses generated.
func (g *AddressGenerator) Stream() chan libwallet.MuunAddress {
	ch := make(chan libwallet.MuunAddress)

	go func() {
		g.generate()

		for _, details := range g.Addresses() {
			ch <- details.Address
		}

		close(ch)
	}()

	return ch
}

func (g *AddressGenerator) generate() {
	g.generateChangeAddrs()
	g.generateExternalAddrs()
	g.generateContactAddrs(100)
}

func (g *AddressGenerator) generateChangeAddrs() {
	const changePath = "m/1'/1'/0"
	changeUserKey, _ := g.userKey.DeriveTo(changePath)
	changeMuunKey, _ := g.muunKey.DeriveTo(changePath)

	g.deriveTree(changeUserKey, changeMuunKey, 2500, "change")
}

func (g *AddressGenerator) generateExternalAddrs() {
	const externalPath = "m/1'/1'/1"
	externalUserKey, _ := g.userKey.DeriveTo(externalPath)
	externalMuunKey, _ := g.muunKey.DeriveTo(externalPath)

	g.deriveTree(externalUserKey, externalMuunKey, 2500, "external")
}

func (g *AddressGenerator) generateContactAddrs(numContacts int64) {
	const addressPath = "m/1'/1'/2"
	contactUserKey, _ := g.userKey.DeriveTo(addressPath)
	contactMuunKey, _ := g.muunKey.DeriveTo(addressPath)
	for i := int64(0); i <= numContacts; i++ {
		partialContactUserKey, _ := contactUserKey.DerivedAt(i, false)
		partialMuunUserKey, _ := contactMuunKey.DerivedAt(i, false)

		branch := fmt.Sprintf("contacts-%v", i)
		g.deriveTree(partialContactUserKey, partialMuunUserKey, 200, branch)
	}
}

func (g *AddressGenerator) deriveTree(rootUserKey, rootMuunKey *libwallet.HDPrivateKey, count int64, name string) {

	for i := int64(0); i <= count; i++ {
		userKey, err := rootUserKey.DerivedAt(i, false)
		if err != nil {
			log.Printf("skipping child %v for %v due to %v", i, name, err)
			continue
		}
		muunKey, err := rootMuunKey.DerivedAt(i, false)
		if err != nil {
			log.Printf("skipping child %v for %v due to %v", i, name, err)
			continue
		}

		addrV2, err := libwallet.CreateAddressV2(userKey.PublicKey(), muunKey.PublicKey())
		if err == nil {
			g.addrs[addrV2.Address()] = signingDetails{
				Address: addrV2,
			}
		} else {
			log.Printf("failed to generate %v v2 for %v due to %v", name, i, err)
		}

		addrV3, err := libwallet.CreateAddressV3(userKey.PublicKey(), muunKey.PublicKey())
		if err == nil {
			g.addrs[addrV3.Address()] = signingDetails{
				Address: addrV3,
			}
		} else {
			log.Printf("failed to generate %v v3 for %v due to %v", name, i, err)
		}

		addrV4, err := libwallet.CreateAddressV4(userKey.PublicKey(), muunKey.PublicKey())
		if err == nil {
			g.addrs[addrV4.Address()] = signingDetails{
				Address: addrV4,
			}
		} else {
			log.Printf("failed to generate %v v4 for %v due to %v", name, i, err)
		}

	}
}
