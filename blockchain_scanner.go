package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/wire"

	"github.com/btcsuite/btclog"

	"github.com/btcsuite/btcd/rpcclient"

	"github.com/btcsuite/btcutil"

	_ "github.com/btcsuite/btcwallet/chain"
	"github.com/btcsuite/btcwallet/walletdb"
	_ "github.com/btcsuite/btcwallet/walletdb/bdb"

	"github.com/btcsuite/btcd/chaincfg"

	"github.com/lightninglabs/neutrino"
	"github.com/lightninglabs/neutrino/headerfs"
)

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

var (
	chainParams        = chaincfg.MainNetParams
	bitcoinGenesisDate = chainParams.GenesisBlock.Header.Timestamp
)

var relevantTxs = make(map[wire.OutPoint]*RelevantTx)
var rescan *neutrino.Rescan

// TODO: Add signing details to the watchAddresses map
var watchAddresses = make(map[string]signingDetails)

func startRescan(chainService *neutrino.ChainService, addrs map[string]signingDetails, birthday int) []*RelevantTx {
	watchAddresses = addrs

	// Wait till we know where the tip is
	for !chainService.IsCurrent() {
	}
	bestBlock, _ := chainService.BestBlock()

	startHeight := findStartHeight(birthday, chainService)
	fmt.Println()
	fmt.Printf("Starting at height %v", startHeight.Height)
	fmt.Println()

	ntfn := rpcclient.NotificationHandlers{
		OnBlockConnected: func(hash *chainhash.Hash, height int32, t time.Time) {
			totalDif := bestBlock.Height - startHeight.Height
			currentDif := height - startHeight.Height
			progress := (float64(currentDif) / float64(totalDif)) * 100.0
			progressBar := ""
			numberOfBars := int(progress / 5)
			for index := 0; index <= 20; index++ {
				if index <= numberOfBars {
					progressBar += "■"
				} else {
					progressBar += "□"
				}
			}

			fmt.Printf("\rProgress: [%v] %.2f%%. Scanning block %v of %v.", progressBar, progress, currentDif, totalDif)
		},
		OnRedeemingTx: func(tx *btcutil.Tx, details *btcjson.BlockDetails) {
			for _, input := range tx.MsgTx().TxIn {
				outpoint := input.PreviousOutPoint
				if _, ok := relevantTxs[outpoint]; ok {
					relevantTxs[outpoint].Spent = true
				}
			}
		},
		OnRecvTx: func(tx *btcutil.Tx, details *btcjson.BlockDetails) {
			checkOutpoints(tx, details.Height)
		},
	}

	rescan = neutrino.NewRescan(
		&neutrino.RescanChainSource{
			ChainService: chainService,
		},
		neutrino.WatchAddrs(buildAddresses()...),
		neutrino.NotificationHandlers(ntfn),
		neutrino.StartBlock(startHeight),
		neutrino.EndBlock(bestBlock),
	)
	errorChan := rescan.Start()
	rescan.WaitForShutdown()
	if err := <-errorChan; err != nil {
		panic(err)
	}

	return buildUtxos()
}

func startChainService() (*neutrino.ChainService, func(), error) {
	setUpLogger()

	dir := os.TempDir()
	dirFolder := filepath.Join(dir, "muunRecoveryTool")
	os.RemoveAll(dirFolder)
	os.MkdirAll(dirFolder, 0700)
	dbPath := filepath.Join(dirFolder, "neutrino.db")

	db, err := walletdb.Open("bdb", dbPath, true)
	if err == walletdb.ErrDbDoesNotExist {
		db, err = walletdb.Create("bdb", dbPath, true)
		if err != nil {
			panic(err)
		}
	}

	peers := make([]string, 1)
	peers[0] = "btcd-mainnet.lightning.computer"
	chainService, err := neutrino.NewChainService(neutrino.Config{
		DataDir:      dirFolder,
		Database:     db,
		ChainParams:  chainParams,
		ConnectPeers: peers,
		AddPeers:     peers,
	})
	if err != nil {
		panic(err)
	}

	err = chainService.Start()
	if err != nil {
		panic(err)
	}

	close := func() {
		db.Close()
		err := chainService.Stop()
		if err != nil {
			panic(err)
		}
		os.Remove(dbPath)
		os.RemoveAll(dirFolder)
	}
	return chainService, close, err
}

func findStartHeight(birthday int, chain *neutrino.ChainService) *headerfs.BlockStamp {
	if birthday == 0 {
		return &headerfs.BlockStamp{}
	}

	const (
		// birthdayBlockDelta is the maximum time delta allowed between our
		// birthday timestamp and our birthday block's timestamp when searching
		// for a better birthday block candidate (if possible).
		birthdayBlockDelta = 2 * time.Hour
	)

	birthtime := bitcoinGenesisDate.Add(time.Duration(birthday-2) * 24 * time.Hour)

	block, _ := chain.BestBlock()

	startHeight := int32(0)
	bestHeight := block.Height

	left, right := startHeight, bestHeight

	for {
		mid := left + (right-left)/2
		hash, _ := chain.GetBlockHash(int64(mid))
		header, _ := chain.GetBlockHeader(hash)

		// If the search happened to reach either of our range extremes,
		// then we'll just use that as there's nothing left to search.
		if mid == startHeight || mid == bestHeight || mid == left {
			return &headerfs.BlockStamp{
				Hash:      *hash,
				Height:    mid,
				Timestamp: header.Timestamp,
			}
		}

		// The block's timestamp is more than 2 hours after the
		// birthday, so look for a lower block.
		if header.Timestamp.Sub(birthtime) > birthdayBlockDelta {
			right = mid
			continue
		}

		// The birthday is more than 2 hours before the block's
		// timestamp, so look for a higher block.
		if header.Timestamp.Sub(birthtime) < -birthdayBlockDelta {
			left = mid
			continue
		}

		return &headerfs.BlockStamp{
			Hash:      *hash,
			Height:    mid,
			Timestamp: header.Timestamp,
		}
	}
}

func checkOutpoints(tx *btcutil.Tx, height int32) {
	// Loop in the output addresses
	for index, output := range tx.MsgTx().TxOut {
		_, addrs, _, _ := txscript.ExtractPkScriptAddrs(output.PkScript, &chainParams)
		for _, addr := range addrs {
			// If one of the output addresses is in our Watch Addresses map, we try to add it to our relevant tx model
			if _, ok := watchAddresses[addr.EncodeAddress()]; ok {
				hash := tx.Hash()
				relevantTx := &RelevantTx{
					PkScript:       output.PkScript,
					Address:        addr.String(),
					Spent:          false,
					Satoshis:       output.Value,
					SigningDetails: watchAddresses[addr.EncodeAddress()],
					Outpoint:       wire.OutPoint{
						Hash: *hash,
						Index: uint32(index),
					},
				}

				if _, ok := relevantTxs[relevantTx.Outpoint]; ok {
					// If its already there we dont need to do anything
					return
				}

				relevantTxs[relevantTx.Outpoint] = relevantTx
			}
		}
	}
}

func buildUtxos() []*RelevantTx {
	var utxos []*RelevantTx
	for _, output := range relevantTxs {
		if !output.Spent {
			utxos = append(utxos, output)
		}
	}
	return utxos
}

func buildAddresses() []btcutil.Address {
	addresses := make([]btcutil.Address, 0, len(watchAddresses))
	for addr := range watchAddresses {
		address, err := btcutil.DecodeAddress(addr, &chainParams)
		if err != nil {
			panic(err)
		}
		addresses = append(addresses, address)
	}
	return addresses
}

func setUpLogger() {
	logger := btclog.NewBackend(os.Stdout).Logger("MUUN")
	logger.SetLevel(btclog.LevelOff)
	neutrino.UseLogger(logger)
}
