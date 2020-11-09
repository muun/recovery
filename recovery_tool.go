package main

import (
	"bytes"
	"fmt"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"

	"github.com/muun/libwallet"
)

func buildSweepTx(utxos []*RelevantTx, sweepAddress btcutil.Address, fee int64) []byte {

	tx := wire.NewMsgTx(2)
	value := int64(0)

	for _, utxo := range utxos {
		tx.AddTxIn(wire.NewTxIn(&utxo.Outpoint, []byte{}, [][]byte{}))
		value += utxo.Satoshis
	}

	fmt.Println()
	fmt.Printf("Total balance in satoshis: %v", value)
	fmt.Println()

	value -= fee

	script, err := txscript.PayToAddrScript(sweepAddress)
	if err != nil {
		printError(err)
	}
	tx.AddTxOut(wire.NewTxOut(value, script))

	writer := &bytes.Buffer{}
	err = tx.Serialize(writer)
	if err != nil {
		panic(err)
	}

	if fee != 0 {
		readConfirmation(value, fee, sweepAddress.String())
	}

	return writer.Bytes()
}

func buildSignedTx(utxos []*RelevantTx, sweepTx []byte, userKey *libwallet.HDPrivateKey,
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
		printError(err)
	}

	signedTx, err := pstx.FullySign(userKey, muunKey)
	if err != nil {
		return nil, err
	}

	wireTx := wire.NewMsgTx(0)
	wireTx.BtcDecode(bytes.NewReader(signedTx.Bytes), 0, wire.WitnessEncoding)
	return wireTx, nil
}

type input struct {
	tx            *RelevantTx
	muunSignature []byte
}

func (i *input) OutPoint() libwallet.Outpoint {
	return &outpoint{tx: i.tx}
}

func (i *input) Address() libwallet.MuunAddress {
	return i.tx.SigningDetails.Address
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

type outpoint struct {
	tx *RelevantTx
}

func (o *outpoint) TxId() []byte {
	return o.tx.Outpoint.Hash.CloneBytes()
}

func (o *outpoint) Index() int {
	return int(o.tx.Outpoint.Index)
}

func (o *outpoint) Amount() int64 {
	return o.tx.Satoshis
}
