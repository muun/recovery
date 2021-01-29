package main

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/muun/recovery/electrum"
	"github.com/muun/recovery/scanner"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/muun/libwallet"
)

var (
	chainParams = chaincfg.MainNetParams
)

type Sweeper struct {
	UserKey      *libwallet.HDPrivateKey
	MuunKey      *libwallet.HDPrivateKey
	Birthday     int
	SweepAddress btcutil.Address
}

// RelevantTx contains a PKScipt, an Address an a boolean to check if its spent or not
type RelevantTx struct {
	PkScript       []byte
	Address        string
	Spent          bool
	Satoshis       int64
	SigningDetails signingDetails
	Outpoint       wire.OutPoint
}

func (tx *RelevantTx) String() string {
	return fmt.Sprintf("outpoint %v:%v for %v sats on path %v",
		tx.Outpoint.Hash, tx.Outpoint.Index, tx.Satoshis, tx.SigningDetails.Address.DerivationPath())
}

func (s *Sweeper) GetUTXOs() ([]*RelevantTx, error) {
	addresses := s.generateAddresses()

	results, err := scanner.NewScanner().Scan(addresses)
	if err != nil {
		return nil, fmt.Errorf("error while scanning addresses: %w", err)
	}

	txs, err := buildRelevantTxs(results)
	if err != nil {
		return nil, fmt.Errorf("error while crafting transaction: %w", err)
	}

	return txs, nil
}

func (s *Sweeper) generateAddresses() chan libwallet.MuunAddress {
	ch := make(chan libwallet.MuunAddress)

	go func() {
		g := NewAddressGenerator(s.UserKey, s.MuunKey)
		g.Generate()

		for _, details := range g.Addresses() {
			ch <- details.Address
		}

		close(ch)
	}()

	return ch
}

func (s *Sweeper) GetSweepTxAmountAndWeightInBytes(utxos []*RelevantTx) (outputAmount int64, weightInBytes int64, err error) {
	// we build a sweep tx with 0 fee with the only purpose of checking its signed size
	zeroFeeSweepTx, err := s.BuildSweepTx(utxos, 0)
	if err != nil {
		return 0, 0, err
	}

	outputAmount = zeroFeeSweepTx.TxOut[0].Value
	weightInBytes = int64(zeroFeeSweepTx.SerializeSize())

	return outputAmount, weightInBytes, nil
}

func (s *Sweeper) BuildSweepTx(utxos []*RelevantTx, fee int64) (*wire.MsgTx, error) {
	derivedMuunKey, err := s.MuunKey.DeriveTo("m/1'/1'")
	if err != nil {
		return nil, err
	}
	sweepTx := buildSweepTx(utxos, s.SweepAddress, fee)
	return buildSignedTx(utxos, sweepTx, s.UserKey, derivedMuunKey)
}

func (s *Sweeper) BroadcastTx(tx *wire.MsgTx) error {
	// Connect to an Electurm server using a fresh client and provider pair:
	sp := scanner.NewServerProvider() // TODO create servers module, for provider and pool
	client := electrum.NewClient()

	for !client.IsConnected() {
		client.Connect(sp.NextServer())
	}

	// Encode the transaction for broadcast:
	txBytes := new(bytes.Buffer)

	err := tx.BtcEncode(txBytes, wire.ProtocolVersion, wire.WitnessEncoding)
	if err != nil {
		return fmt.Errorf("error while encoding tx: %w", err)
	}

	txHex := hex.EncodeToString(txBytes.Bytes())

	// Do the thing!
	_, err = client.Broadcast(txHex)
	if err != nil {
		return fmt.Errorf("error while broadcasting: %w", err)
	}

	return nil
}

// buildRelevantTxs prepares the output from Scanner for crafting.
func buildRelevantTxs(utxos []scanner.Utxo) ([]*RelevantTx, error) {
	var relevantTxs []*RelevantTx

	for _, utxo := range utxos {
		address := utxo.Address.Address()

		chainHash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			return nil, err
		}

		relevantTx := &RelevantTx{
			PkScript:       utxo.Script,
			Address:        address,
			Spent:          false,
			Satoshis:       int64(utxo.Amount),
			SigningDetails: signingDetails{utxo.Address},
			Outpoint: wire.OutPoint{
				Hash:  *chainHash,
				Index: uint32(utxo.OutputIndex),
			},
		}

		relevantTxs = append(relevantTxs, relevantTx)
	}

	return relevantTxs, nil
}
