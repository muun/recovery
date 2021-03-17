package main

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"

	"github.com/muun/libwallet"
	"github.com/muun/recovery/scanner"
)

func buildSweepTx(utxos []*scanner.Utxo, sweepAddress btcutil.Address, fee int64) ([]byte, error) {

	tx := wire.NewMsgTx(2)
	value := int64(0)

	for _, utxo := range utxos {
		chainHash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			return nil, err
		}

		outpoint := wire.OutPoint{
			Hash:  *chainHash,
			Index: uint32(utxo.OutputIndex),
		}

		tx.AddTxIn(wire.NewTxIn(&outpoint, []byte{}, [][]byte{}))
		value += utxo.Amount
	}

	value -= fee

	script, err := txscript.PayToAddrScript(sweepAddress)
	if err != nil {
		return nil, err
	}
	tx.AddTxOut(wire.NewTxOut(value, script))

	writer := &bytes.Buffer{}
	err = tx.Serialize(writer)
	if err != nil {
		return nil, err
	}

	if fee != 0 {
		readConfirmation(value, fee, sweepAddress.String())
	}

	return writer.Bytes(), nil
}

func buildSignedTx(utxos []*scanner.Utxo, sweepTx []byte, userKey *libwallet.HDPrivateKey,
	muunKey *libwallet.HDPrivateKey) (*wire.MsgTx, error) {

	inputList := &libwallet.InputList{}
	for _, utxo := range utxos {
		inputList.Add(&input{
			utxo,
			[]byte{},
		})
	}

	pstx, err := libwallet.NewPartiallySignedTransaction(inputList, sweepTx)
	if err != nil {
		return nil, err
	}

	signedTx, err := pstx.FullySign(userKey, muunKey)
	if err != nil {
		return nil, err
	}

	wireTx := wire.NewMsgTx(0)
	wireTx.BtcDecode(bytes.NewReader(signedTx.Bytes), 0, wire.WitnessEncoding)
	return wireTx, nil
}

// input is a minimal type that implements libwallet.Input
type input struct {
	utxo          *scanner.Utxo
	muunSignature []byte
}

func (i *input) OutPoint() libwallet.Outpoint {
	return &outpoint{utxo: i.utxo}
}

func (i *input) Address() libwallet.MuunAddress {
	return i.utxo.Address
}

func (i *input) UserSignature() []byte {
	return []byte{}
}

func (i *input) MuunSignature() []byte {
	return i.muunSignature
}

func (i *input) SubmarineSwapV1() libwallet.InputSubmarineSwapV1 {
	return nil
}

func (i *input) SubmarineSwapV2() libwallet.InputSubmarineSwapV2 {
	return nil
}

func (i *input) IncomingSwap() libwallet.InputIncomingSwap {
	return nil
}

// outpoint is a minimal type that implements libwallet.Outpoint
type outpoint struct {
	utxo *scanner.Utxo
}

func (o *outpoint) TxId() []byte {
	raw, err := hex.DecodeString(o.utxo.TxID)
	if err != nil {
		panic(err) // we wrote this hex value ourselves, no input from anywhere else
	}

	return raw
}

func (o *outpoint) Index() int {
	return o.utxo.OutputIndex
}

func (o *outpoint) Amount() int64 {
	return o.utxo.Amount
}
