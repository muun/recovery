package main

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/muun/recovery/electrum"
	"github.com/muun/recovery/scanner"

	"github.com/btcsuite/btcd/chaincfg"

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

func (s *Sweeper) GetSweepTxAmountAndWeightInBytes(utxos []*scanner.Utxo) (outputAmount int64, weightInBytes int64, err error) {
	// we build a sweep tx with 0 fee with the only purpose of checking its signed size
	zeroFeeSweepTx, err := s.BuildSweepTx(utxos, 0)
	if err != nil {
		return 0, 0, err
	}

	outputAmount = zeroFeeSweepTx.TxOut[0].Value
	weightInBytes = int64(zeroFeeSweepTx.SerializeSize())

	return outputAmount, weightInBytes, nil
}

func (s *Sweeper) BuildSweepTx(utxos []*scanner.Utxo, fee int64) (*wire.MsgTx, error) {
	derivedMuunKey, err := s.MuunKey.DeriveTo("m/1'/1'")
	if err != nil {
		return nil, err
	}
	sweepTx, err := buildSweepTx(utxos, s.SweepAddress, fee)
	if err != nil {
		return nil, err
	}

	return buildSignedTx(utxos, sweepTx, s.UserKey, derivedMuunKey)
}

func (s *Sweeper) BroadcastTx(tx *wire.MsgTx) error {
	// Connect to an Electurm server using a fresh client and provider pair:
	sp := electrum.NewServerProvider(electrum.PublicServers) // TODO create servers module, for provider and pool
	client := electrum.NewClient(true)

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
