package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/btcsuite/btcutil"
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

	sweepAddress := readSweepAddress()

	fmt.Println("")
	fmt.Println("Preparing to scan the blockchain from your wallet creation block")
	fmt.Println("Note that only confirmed transactions can be detected")
	fmt.Println("\nThis may take a while")

	sweeper := Sweeper{
		ChainService: chainService,
		UserKey:      userKey.Key,
		MuunKey:      muunKey.Key,
		Birthday:     muunKey.Birthday,
		SweepAddress: sweepAddress,
	}

	utxos := sweeper.GetUTXOs()

	fmt.Println("")

	if len(utxos) > 0 {
		fmt.Printf("The recovery tool has found the following confirmed UTXOs:\n%v", utxos)
	} else {
		fmt.Printf("No confirmed UTXOs found")
		fmt.Println()
		return
	}
	fmt.Println()

	txOutputAmount, txWeightInBytes, err := sweeper.GetSweepTxAmountAndWeightInBytes(utxos)
	if err != nil {
		printError(err)
	}

	fee := readFee(txOutputAmount, txWeightInBytes)

	// Then we re-build the sweep tx with the actual fee
	sweepTx, err := sweeper.BuildSweepTx(utxos, fee)
	if err != nil {
		printError(err)
	}

	fmt.Println("Transaction ready to be sent")

	err = sweeper.BroadcastTx(sweepTx)
	if err != nil {
		printError(err)
	}

	fmt.Printf("Transaction sent! You can check the status here: https://blockstream.info/tx/%v", sweepTx.TxHash().String())
	fmt.Println("")
	fmt.Printf("We appreciate all kinds of feedback. If you have any, send it to contact@muun.com")
	fmt.Println("")
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
	fmt.Println("You can use this tool to transfer all funds from your Muun account to an")
	fmt.Println("address of your choosing.")
	fmt.Println("")
	fmt.Println("To do this you will need:")
	fmt.Println("1. Your Recovery Code, which you wrote down during your security setup")
	fmt.Println("2. Your two encrypted private keys, which you exported from your wallet")
	fmt.Println("3. A destination bitcoin address where all your funds will be sent")
	fmt.Println("")
	fmt.Println("If you have any questions, we'll be happy to answer them. Contact us at support@muun.com")
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
		fmt.Printf("Invalid recovery code. Did you add the '-' separator between each 4-characters segment?")
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

	userInput := scanMultiline(characters)

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

func scanMultiline(minChars int) string {
	var result strings.Builder

	for result.Len() < minChars {
		var line string
		fmt.Scan(&line)

		result.WriteString(strings.TrimSpace(line))
	}

	return result.String()
}
