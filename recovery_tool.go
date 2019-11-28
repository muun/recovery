package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"

	"github.com/muun/libwallet"
)

func main() {
	chainService, close, _ := startChainService()
	defer close()

	printWelcomeMessage()

	recoveryCode := readRecoveryCode()

	userRawKey := readKey("first encrypted private key", 147)
	userKey := buildExtendedKey(userRawKey, recoveryCode)
	userKey.Key.Path = "m/1'/1'"

	muunRawKey := readKey("second encrypted private key", 147)
	muunKey := buildExtendedKey(muunRawKey, recoveryCode)
	derivedMuunKey, err := muunKey.Key.DeriveTo("m/1'/1'")
	if err != nil {
		printError(err)
	}

	sweepAddress := readSweepAddress()

	fmt.Println("")
	fmt.Println("Starting to scan the blockchain. This may take a while.")

	g := NewAddressGenerator(userKey.Key, muunKey.Key)
	g.Generate()

	birthday := muunKey.Birthday
	if birthday == 0xFFFF {
		birthday = 0
	}

	utxos := startRescan(chainService, g.Addresses(), birthday)
	fmt.Println("")

	if len(utxos) > 0 {
		fmt.Printf("The recovery tool has found the following utxos: %v", utxos)
	} else {
		fmt.Printf("No utxos found")
		fmt.Println()
		return
	}
	fmt.Println()

	// This is fun:
	// First we build a sweep tx with 0 fee with the only purpouse of seeing its signed size
	zeroFeehexSweepTx := buildSweepTx(utxos, sweepAddress, 0)
	zeroFeeSweepTx, err := buildSignedTx(utxos, zeroFeehexSweepTx, userKey.Key, derivedMuunKey)
	if err != nil {
		printError(err)
	}
	weightInBytes := int64(zeroFeeSweepTx.SerializeSize())
	fee := readFee(zeroFeeSweepTx.TxOut[0].Value, weightInBytes)
	// Then we re-build the sweep tx with the actual fee
	hexSweepTx := buildSweepTx(utxos, sweepAddress, fee)
	tx, err := buildSignedTx(utxos, hexSweepTx, userKey.Key, derivedMuunKey)
	if err != nil {
		printError(err)
	}
	fmt.Println("Transaction ready to be sent")

	err = chainService.SendTransaction(tx)
	if err != nil {
		printError(err)
	}

	fmt.Printf("Transaction sent! You can check the status here: https://blockstream.info/tx/%v", tx.TxHash().String())
	fmt.Println("")
	fmt.Printf("If you have any feedback, feel free to share it with us. Our email is contact@muun.com")
	fmt.Println("")

}

func buildSweepTx(utxos []*RelevantTx, sweepAddress btcutil.Address, fee int64) string {

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

	return hex.EncodeToString(writer.Bytes())
}

func buildSignedTx(utxos []*RelevantTx, hexSweepTx string, userKey *libwallet.HDPrivateKey,
	muunKey *libwallet.HDPrivateKey) (*wire.MsgTx, error) {

	pstx, err := libwallet.NewPartiallySignedTransaction(hexSweepTx)
	if err != nil {
		printError(err)
	}

	for index, utxo := range utxos {
		input := &input{
			utxo,
			[]byte{},
		}

		pstx.AddInput(input)
		sig, err := pstx.MuunSignatureForInput(index, userKey.PublicKey(), muunKey)
		if err != nil {
			panic(err)
		}
		input.muunSignature = sig
	}

	signedTx, err := pstx.Sign(userKey, muunKey.PublicKey())
	if err != nil {
		return nil, err
	}

	wireTx := wire.NewMsgTx(0)
	wireTx.BtcDecode(bytes.NewReader(signedTx.Bytes), 0, wire.WitnessEncoding)
	return wireTx, nil
}

func printError(err error) {
	log.Printf("The recovery tool failed with the following error: %v", err.Error())
	log.Printf("")
	log.Printf("You can try again or contact us at support@muun.com")
	panic(err)
}

func printWelcomeMessage() {
	fmt.Println("Welcome to Muun's Recovery Tool")
	fmt.Println("")
	fmt.Println("You can use this tool to swipe all the balance in your muun account to an")
	fmt.Println("address of your choosing.")
	fmt.Println("")
	fmt.Println("To do this you will need:")
	fmt.Println("* The recovery code, that you set up when you created your muun account")
	fmt.Println("* The two encrypted private keys that you exported from your muun wallet")
	fmt.Println("* A destination bitcoin address where all your funds will be sent")
	fmt.Println("")
	fmt.Println("If you have any questions, contact us at contact@muun.com")
	fmt.Println("")
}

func readRecoveryCode() string {
	fmt.Println("")
	fmt.Printf("Enter your Recovery Code")
	fmt.Println()
	fmt.Println("(it looks like this: 'ABCD-1234-POW2-R561-P120-JK26-12RW-45TT')")
	fmt.Print("> ")
	var userInput string
	fmt.Scan(&userInput)
	userInput = strings.TrimSpace(userInput)

	finalRC := strings.ToUpper(userInput)

	if strings.Count(finalRC, "-") != 7 {
		fmt.Printf("Wrong recovery code, remember to add the '-' separator between the 4 characters chunks")
		fmt.Println()
		fmt.Println("Please, try again")

		return readRecoveryCode()
	}

	if len(finalRC) != 39 {
		fmt.Println("Your recovery code must have 39 characters")
		fmt.Println("Please, try again")

		return readRecoveryCode()
	}

	return finalRC
}

func readKey(keyType string, characters int) string {
	fmt.Println("")
	fmt.Printf("Enter your %v", keyType)
	fmt.Println()
	fmt.Println("(it looks like this: '9xzpc7y6sNtRvh8Fh...')")
	fmt.Print("> ")
	var userInput string
	fmt.Scan(&userInput)
	userInput = strings.TrimSpace(userInput)

	if len(userInput) != characters {
		fmt.Printf("Your %v must have %v characters", keyType, characters)
		fmt.Println("")
		fmt.Println("Please, try again")

		return readKey(keyType, characters)
	}

	return userInput
}

func readSweepAddress() btcutil.Address {
	fmt.Println("")
	fmt.Println("Enter your destination bitcoin address")
	fmt.Print("> ")
	var userInput string
	fmt.Scan(&userInput)
	userInput = strings.TrimSpace(userInput)

	addr, err := btcutil.DecodeAddress(userInput, &chainParams)
	if err != nil {
		fmt.Println("This is not a valid bitcoin address")
		fmt.Println("")
		fmt.Println("Please, try again")

		return readSweepAddress()
	}

	return addr
}

func readFee(totalBalance, weight int64) int64 {
	fmt.Println("")
	fmt.Printf("Enter the fee in satoshis per byte. Tx weight: %v bytes. You can check the status of the mempool here: https://bitcoinfees.earn.com/#fees", weight)
	fmt.Println()
	fmt.Println("(Example: 5)")
	fmt.Print("> ")
	var userInput string
	fmt.Scan(&userInput)
	feeInSatsPerByte, err := strconv.ParseInt(userInput, 10, 64)
	if err != nil || feeInSatsPerByte <= 0 {
		fmt.Printf("The fee must be a number")
		fmt.Println("")
		fmt.Println("Please, try again")

		return readFee(totalBalance, weight)
	}

	totalFee := feeInSatsPerByte * weight

	if totalBalance-totalFee < 546 {
		fmt.Printf("The fee is too high. The amount left must be higher than dust")
		fmt.Println("")
		fmt.Println("Please, try again")

		return readFee(totalBalance, weight)
	}

	return totalFee
}

func readConfirmation(value, fee int64, address string) {
	fmt.Println("")
	fmt.Printf("About to send %v satoshis with fee: %v satoshis to %v", value, fee, address)
	fmt.Println()
	fmt.Println("Confirm? (y/n)")
	fmt.Print("> ")
	var userInput string
	fmt.Scan(&userInput)

	if userInput == "y" || userInput == "Y" {
		return
	}

	if userInput == "n" || userInput == "N" {
		log.Println()
		log.Printf("Recovery tool stopped")
		log.Println()
		log.Printf("You can try again or contact us at support@muun.com")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("You can only enter 'y' to accept or 'n' to cancel")
	readConfirmation(value, fee, address)
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
