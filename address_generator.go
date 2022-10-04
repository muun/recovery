package main

import (
	"fmt"
	"log"

	"github.com/muun/libwallet"
	"github.com/muun/recovery/utils"
)

type AddressGenerator struct {
	addressCount     int
	userKey          *libwallet.HDPrivateKey
	muunKey          *libwallet.HDPrivateKey
	generateContacts bool
}

func NewAddressGenerator(userKey, muunKey *libwallet.HDPrivateKey, generateContacts bool) *AddressGenerator {
	return &AddressGenerator{
		addressCount:     0,
		userKey:          userKey,
		muunKey:          muunKey,
		generateContacts: generateContacts,
	}
}

// Stream returns a channel that emits all addresses generated.
func (g *AddressGenerator) Stream() chan libwallet.MuunAddress {
	ch := make(chan libwallet.MuunAddress)

	go func() {
		g.generate(ch)
		utils.NewLogger("ADDR").Printf("Addresses %v\n", g.addressCount)

		close(ch)
	}()

	return ch
}

func (g *AddressGenerator) generate(consumer chan libwallet.MuunAddress) {
	g.generateChangeAddrs(consumer)
	g.generateExternalAddrs(consumer)
	if g.generateContacts {
		g.generateContactAddrs(consumer, 100)
	}
}

func (g *AddressGenerator) generateChangeAddrs(consumer chan libwallet.MuunAddress) {
	const changePath = "m/1'/1'/0"
	changeUserKey, _ := g.userKey.DeriveTo(changePath)
	changeMuunKey, _ := g.muunKey.DeriveTo(changePath)

	g.deriveTree(consumer, changeUserKey, changeMuunKey, 2500, "change")
}

func (g *AddressGenerator) generateExternalAddrs(consumer chan libwallet.MuunAddress) {
	const externalPath = "m/1'/1'/1"
	externalUserKey, _ := g.userKey.DeriveTo(externalPath)
	externalMuunKey, _ := g.muunKey.DeriveTo(externalPath)

	g.deriveTree(consumer, externalUserKey, externalMuunKey, 2500, "external")
}

func (g *AddressGenerator) generateContactAddrs(consumer chan libwallet.MuunAddress, numContacts int64) {
	const addressPath = "m/1'/1'/2"
	contactUserKey, _ := g.userKey.DeriveTo(addressPath)
	contactMuunKey, _ := g.muunKey.DeriveTo(addressPath)
	for i := int64(0); i <= numContacts; i++ {
		partialContactUserKey, _ := contactUserKey.DerivedAt(i, false)
		partialMuunUserKey, _ := contactMuunKey.DerivedAt(i, false)

		branch := fmt.Sprintf("contacts-%v", i)
		g.deriveTree(consumer, partialContactUserKey, partialMuunUserKey, 200, branch)
	}
}

func (g *AddressGenerator) deriveTree(
	consumer chan libwallet.MuunAddress,
	rootUserKey, rootMuunKey *libwallet.HDPrivateKey,
	count int64,
	name string,
) {

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
			consumer <- addrV2
			g.addressCount++
		} else {
			log.Printf("failed to generate %v v2 for %v due to %v", name, i, err)
		}

		addrV3, err := libwallet.CreateAddressV3(userKey.PublicKey(), muunKey.PublicKey())
		if err == nil {
			consumer <- addrV3
			g.addressCount++
		} else {
			log.Printf("failed to generate %v v3 for %v due to %v", name, i, err)
		}

		addrV4, err := libwallet.CreateAddressV4(userKey.PublicKey(), muunKey.PublicKey())
		if err == nil {
			consumer <- addrV4
			g.addressCount++
		} else {
			log.Printf("failed to generate %v v4 for %v due to %v", name, i, err)
		}

		addrV5, err := libwallet.CreateAddressV5(userKey.PublicKey(), muunKey.PublicKey())
		if err == nil {
			consumer <- addrV5
			g.addressCount++
		} else {
			log.Printf("failed to generate %v v5 for %v due to %v", name, i, err)
		}
	}
}
