package main

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/lightninglabs/neutrino"
	"github.com/muun/libwallet"
)

type Sweeper struct {
	ChainService *neutrino.ChainService
	UserKey      *libwallet.HDPrivateKey
	MuunKey      *libwallet.HDPrivateKey
	Birthday     int
	SweepAddress btcutil.Address
}

func (s *Sweeper) GetUTXOs() []*RelevantTx {
	g := NewAddressGenerator(s.UserKey, s.MuunKey)
	g.Generate()

	birthday := s.Birthday
	if birthday == 0xFFFF {
		birthday = 0
	}

	return startRescan(s.ChainService, g.Addresses(), birthday)
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
	return s.ChainService.SendTransaction(tx)
}
