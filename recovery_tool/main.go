package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/btcsuite/btcutil"
	"github.com/gookit/color"
	"github.com/muun/libwallet"
	"github.com/muun/libwallet/btcsuitew/btcutilw"
	"github.com/muun/libwallet/emergencykit"
	"github.com/muun/recovery/electrum"
	"github.com/muun/recovery/scanner"
	"github.com/muun/recovery/utils"
)

const electrumPoolSize = 6

var debugOutputStream = bytes.NewBuffer(nil)

type config struct {
	generateContacts     bool
	providedElectrum     string
	usesProvidedElectrum bool
	onlyScan             bool
}

func main() {
	utils.SetOutputStream(debugOutputStream)

	var config config

	// Pick up command-line arguments:
	flag.BoolVar(&config.generateContacts, "generate-contacts", false, "Generate contact addresses")
	flag.StringVar(&config.providedElectrum, "electrum-server", "", "Connect to this electrum server to find funds")
	flag.BoolVar(&config.onlyScan, "only-scan", false, "Only scan for UTXOs without generating a transaction")
	flag.Usage = printUsage
	flag.Parse()
	args := flag.Args()

	// Ensure correct form:
	if len(args) > 1 {
		printUsage()
		os.Exit(0)
	}

	// Welcome!
	printWelcomeMessage()

	config.usesProvidedElectrum = len(strings.TrimSpace(config.providedElectrum)) > 0
	if config.usesProvidedElectrum {
		validateProvidedElectrum(config.providedElectrum)
	}

	// We're going to need a few things to move forward with the recovery process. Let's make a list
	// so we keep them in mind:
	var recoveryCode string
	var encryptedKeys []*libwallet.EncryptedPrivateKeyInfo
	var destinationAddress btcutil.Address

	// First on our list is the Recovery Code. This is the time to go looking for that piece of paper:
	recoveryCode = readRecoveryCode()

	// Good! Now, on to those keys. We need to read them and decrypt them:
	encryptedKeys, err := readBackupFromInputOrPDF(flag.Arg(0))
	if err != nil {
		exitWithError(err)
	}

	decryptedKeys, err := decryptKeys(encryptedKeys, recoveryCode)
	if err != nil {
		exitWithError(err)
	}

	decryptedKeys[0].Key.Path = "m/1'/1'" // a little adjustment for legacy users.

	if !config.onlyScan {
		// Finally, we need the destination address to sweep the funds:
		destinationAddress = readAddress()
	}

	sayBlock(`
		Starting scan of all possible addresses. This will take a few minutes.
	`)

	doRecovery(decryptedKeys, destinationAddress, config)

	sayBlock("We appreciate all kinds of feedback. If you have any, send it to {blue contact@muun.com}\n")
}

// doRecovery runs the scan & sweep process, and returns the ID of the broadcasted transaction.
func doRecovery(
	decryptedKeys []*libwallet.DecryptedPrivateKey,
	destinationAddress btcutil.Address,
	config config,
) {

	addrGen := NewAddressGenerator(decryptedKeys[0].Key, decryptedKeys[1].Key, config.generateContacts)

	var electrumProvider *electrum.ServerProvider
	if config.usesProvidedElectrum {
		electrumProvider = electrum.NewServerProvider([]string{
			config.providedElectrum,
		})
	} else {
		electrumProvider = electrum.NewServerProvider(electrum.PublicServers)
	}

	connectionPool := electrum.NewPool(electrumPoolSize, !config.usesProvidedElectrum)

	utxoScanner := scanner.NewScanner(connectionPool, electrumProvider)

	addresses := addrGen.Stream()

	sweeper := Sweeper{
		UserKey:      decryptedKeys[0].Key,
		MuunKey:      decryptedKeys[1].Key,
		Birthday:     decryptedKeys[1].Birthday,
		SweepAddress: destinationAddress,
	}

	reports := utxoScanner.Scan(addresses)

	say("► {white Finding servers...}")

	var lastReport *scanner.Report
	for lastReport = range reports {
		printReport(lastReport)
	}

	fmt.Println()
	fmt.Println()

	if lastReport.Err != nil {
		exitWithError(fmt.Errorf("error while scanning addresses: %w", lastReport.Err))
	}

	say("{green ✓ Scan complete}\n")
	utxos := lastReport.UtxosFound

	if len(utxos) == 0 {
		sayBlock("No funds were discovered\n\n")
		return
	}

	var total int64
	for _, utxo := range utxos {
		total += utxo.Amount
		say("• {white %d} sats in %s\n", utxo.Amount, utxo.Address.Address())
	}

	say("\n— {white %d} sats total\n", total)
	if config.onlyScan {
		return
	}

	txOutputAmount, txWeightInBytes, err := sweeper.GetSweepTxAmountAndWeightInBytes(utxos)
	if err != nil {
		exitWithError(err)
	}

	fee := readFee(txOutputAmount, txWeightInBytes)

	// Then we re-build the sweep tx with the actual fee
	sweepTx, err := sweeper.BuildSweepTx(utxos, fee)
	if err != nil {
		exitWithError(err)
	}

	sayBlock("Sending transaction...")

	err = sweeper.BroadcastTx(sweepTx)
	if err != nil {
		exitWithError(err)
	}

	sayBlock(`
		Transaction sent! You can check the status here: https://mempool.space/tx/%v
		(it will appear in mempool.space after a short delay)

	`, sweepTx.TxHash().String())
}

func validateProvidedElectrum(providedElectrum string) {
	client := electrum.NewClient(false)
	err := client.Connect(providedElectrum)
	defer func(client *electrum.Client) {
		_ = client.Disconnect()
	}(client)

	if err != nil {
		sayBlock(`
				{red Error!}
				The Recovery Tool couldn't connect to the provided Electrum server %v.

				If the problem persists, contact {blue support@muun.com}.
				
				――― {white error report} ―――
				%v
				――――――――――――――――――――

				We're always there to help.
			`, providedElectrum, err)

		os.Exit(2)
	}
}

func exitWithError(err error) {
	sayBlock(`
		{red Error!}
		The Recovery Tool encountered a problem. Please, try again.

		If the problem persists, contact {blue support@muun.com} and include the file
		called error_log you can find in the same folder as this tool.
		
		――― {white error report} ―――
		%v
		――――――――――――――――――――

		We're always there to help.
	`, err)

	// Ensure we always log the error in the file
	_ = utils.NewLogger("").Errorf("exited with error: %s", err.Error())
	_ = os.WriteFile("error_log", debugOutputStream.Bytes(), 0600)

	os.Exit(1)
}

func printWelcomeMessage() {
	say(`
		{blue Muun Recovery Tool v%s}

		To recover your funds, you will need:
		
		1. {yellow Your Recovery Code}, which you wrote down during your security setup
		2. {yellow Your Emergency Kit PDF}, which you exported from the app
		3. {yellow Your destination bitcoin address}, where all your funds will be sent
		
		If you have any questions, we'll be happy to answer them. Contact us at {blue support@muun.com}
	`, version)
}

func printUsage() {
	fmt.Println("Usage: recovery-tool [optional: path to Emergency Kit PDF]")
	flag.PrintDefaults()
}

func printReport(report *scanner.Report) {
	if utils.DebugMode {
		return // don't print reports while debugging, there's richer information in the logs
	}

	var total int64
	for _, utxo := range report.UtxosFound {
		total += utxo.Amount
	}

	say("\r► {white Scanned addresses}: %d | {white Sats found}: %d", report.ScannedAddresses, total)
}

func readRecoveryCode() string {
	sayBlock(`
		{yellow Enter your Recovery Code}
		(it looks like this: 'ABCD-1234-POW2-R561-P120-JK26-12RW-45TT')
	`)

	var userInput string
	ask(&userInput)

	userInput = strings.TrimSpace(userInput)
	finalRC := strings.ToUpper(userInput)

	if strings.Count(finalRC, "-") != 7 {
		say(`
			Invalid recovery code. Did you add the '-' separator between each 4-characters segment?
			Please, try again
		`)

		return readRecoveryCode()
	}

	if len(finalRC) != 39 {
		say(`
			Your recovery code must have 39 characters
			Please, try again
		`)

		return readRecoveryCode()
	}

	return finalRC
}

func readBackupFromInputOrPDF(optionalPDF string) ([]*libwallet.EncryptedPrivateKeyInfo, error) {
	// Here we have two possible flows, depending on whether the PDF was provided (pick up the
	// encrypted backup automatically) or not (manual input). If we try for the automatic flow and fail,
	// we can fall back to the manual one.

	// Read metadata from the PDF, if given:
	if optionalPDF != "" {
		encryptedKeys, err := readBackupFromPDF(optionalPDF)

		if err == nil {
			return encryptedKeys, nil
		}

		// Hmm. Okay, we'll confess and fall back to manual input.
		say(`
			Couldn't read the PDF automatically: %v
			Please, enter your data manually
		`, err)
	}

	// Ask for manual input, if we have no PDF or couldn't read it:
	encryptedKeys, err := readBackupFromInput()
	if err != nil {
		return nil, err
	}

	return encryptedKeys, nil
}

func readBackupFromInput() ([]*libwallet.EncryptedPrivateKeyInfo, error) {
	firstRawKey := readKey("first encrypted private key")
	secondRawKey := readKey("second encrypted private key")

	decodedKeys, err := decodeKeysFromInput(firstRawKey, secondRawKey)
	if err != nil {
		return nil, err
	}

	return decodedKeys, nil
}

func readBackupFromPDF(path string) ([]*libwallet.EncryptedPrivateKeyInfo, error) {
	reader := &emergencykit.MetadataReader{SrcFile: path}

	metadata, err := reader.ReadMetadata()
	if err != nil {
		return nil, err
	}

	decodedKeys, err := decodeKeysFromMetadata(metadata)
	if err != nil {
		return nil, err
	}

	return decodedKeys, nil
}

func readKey(keyType string) string {
	sayBlock(`
		{yellow Enter your %v}
		(it looks like this: '9xzpc7y6sNtRvh8Fh...')
	`, keyType)

	// NOTE:
	// Users will most likely copy and paste their keys from the Emergency Kit PDF. In this case,
	// input will come suddenly in multiple lines, so a simple scan & retry (let's say 3 lines
	// were pasted)  will attempt to parse a key and fail 2 times in a row, with leftover characters
	// until the user presses enter to fail for a 3rd time.

	// Given the line lengths actually found in our Emergency Kits, we have a simple solution for now:
	// scan a minimum length of characters. Pasing from current versions of the Emergency Kit will
	// only go past a minimum length when the key being entered is complete, in all cases.
	userInput := askMultiline(libwallet.EncodedKeyLengthLegacy)

	if len(userInput) < libwallet.EncodedKeyLengthLegacy {
		// This is obviously invalid. Other problems will be detected later on, during the actual
		// decoding and decryption stage.
		say(`
			The key you entered doesn't look valid
			Please, try again
		`)

		return readKey(keyType)
	}

	return userInput
}

func readAddress() btcutil.Address {
	sayBlock(`
		{yellow Enter your destination bitcoin address}
	`)

	var userInput string
	ask(&userInput)

	userInput = strings.TrimSpace(userInput)

	addr, err := btcutilw.DecodeAddress(userInput, &chainParams)
	if err != nil {
		say(`
			This is not a valid bitcoin address
			Please, try again
		`)

		return readAddress()
	}

	return addr
}

func readFee(totalBalance, weight int64) int64 {
	sayBlock(`
		{yellow Enter the fee rate (sats/byte)}
		Your transaction weighs %v bytes. You can get suggestions in https://mempool.space/ under "Transaction fees".
	`, weight)

	var userInput string
	ask(&userInput)

	feeInSatsPerByte, err := strconv.ParseInt(userInput, 10, 64)
	if err != nil || feeInSatsPerByte <= 0 {
		say(`
			The fee must be a whole number
			Please, try again
		`)

		return readFee(totalBalance, weight)
	}

	totalFee := feeInSatsPerByte * weight

	if totalBalance-totalFee < 546 {
		say(`
			The fee is too high. The remaining amount after deducting is too low to send.
			Please, try again
		`)

		return readFee(totalBalance, weight)
	}

	return totalFee
}

func readConfirmation(value, fee int64, address string) {
	sayBlock(`
		{whiteUnderline Summary}
		  {white Amount}: %v sats
		  {white Fee}: %v sats
		  {white Destination}: %v

		{yellow Confirm?} (y/n)
	`, value, fee, address)

	var userInput string
	ask(&userInput)

	if userInput == "y" || userInput == "Y" {
		return
	}

	if userInput == "n" || userInput == "N" {
		sayBlock(`
			Recovery tool stopped
			You can try again or contact us at {blue support@muun.com}
		`)
		os.Exit(1)
	}

	say(`You can only enter 'y' to confirm or 'n' to cancel`)

	fmt.Print("\n\n")
	readConfirmation(value, fee, address)
}

var leadingIndentRe = regexp.MustCompile("^[ \t]+")
var colorRe = regexp.MustCompile(`\{(\w+?) ([^\}]+?)\}`)

func say(message string, v ...interface{}) {
	noEmptyLine := strings.TrimLeft(message, " \n")
	firstIndent := leadingIndentRe.FindString(noEmptyLine)

	noIndent := strings.ReplaceAll(noEmptyLine, firstIndent, "")

	noTrailingSpace := strings.TrimRight(noIndent, " \t")

	withColors := colorRe.ReplaceAllStringFunc(noTrailingSpace, func(match string) string {
		groups := colorRe.FindStringSubmatch(match)
		return applyColor(groups[1], groups[2])
	})

	fmt.Printf(withColors, v...)
}

func sayBlock(message string, v ...interface{}) {
	fmt.Println()
	say(message, v...)
}

func applyColor(colorName string, text string) string {
	switch colorName {
	case "red":
		return color.New(color.FgRed, color.BgDefault, color.OpBold).Sprint(text)
	case "blue":
		return color.New(color.FgBlue, color.BgDefault, color.OpBold).Sprint(text)
	case "yellow":
		return color.New(color.FgYellow, color.BgDefault, color.OpBold).Sprint(text)
	case "green":
		return color.New(color.FgGreen, color.BgDefault, color.OpBold).Sprint(text)
	case "white":
		return color.New(color.FgWhite, color.BgDefault, color.OpBold).Sprint(text)
	case "whiteUnderline":
		return color.New(color.FgWhite, color.BgDefault, color.OpBold, color.OpUnderscore).Sprint(text)
	}

	panic("No such color: " + colorName)
}

func askMultiline(minChars int) string {
	fmt.Print("➜ ")

	var result strings.Builder

	for result.Len() < minChars {
		var line string
		fmt.Scan(&line)

		result.WriteString(strings.TrimSpace(line))
	}

	return result.String()
}

func ask(result *string) {
	fmt.Print("➜ ")
	fmt.Scan(result)
}
