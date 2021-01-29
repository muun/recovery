package scanner

import (
	"sync"
	"time"

	"github.com/muun/libwallet"
	"github.com/muun/recovery/electrum"
	"github.com/muun/recovery/utils"
)

const electrumPoolSize = 3
const taskTimeout = 2 * time.Minute
const batchSize = 100

// Scanner finds unspent outputs and their transactions when given a map of addresses.
//
// It implements multi-server support, batching feature detection and use, concurrency control,
// timeouts and cancelations, and provides a channel-based interface.
//
// Servers are provided by a ServerProvider instance, and rotated when unreachable or faulty. We
// trust ServerProvider to prioritize good targets.
//
// Batching is leveraged when supported by a particular server, falling back to sequential requests
// for single addresses (which is much slower, but can get us out of trouble when better servers are
// not available).
//
// Timeouts and cancellations are an internal affair, not configurable by callers. See taskTimeout
// declared above.
//
// Concurrency control works by using an electrum.Pool, limiting access to clients, and not an
// internal worker pool. This is the Go way (limiting access to resources rather than having a fixed
// number of parallel goroutines), and (more to the point) semantically correct. We don't care
// about the number of concurrent workers, what we want to avoid is too many connections to
// Electrum servers.
type Scanner struct {
	pool    *electrum.Pool
	servers *ServerProvider
	log     *utils.Logger
}

// Utxo references a transaction output, plus the associated MuunAddress and script.
type Utxo struct {
	TxID        string
	OutputIndex int
	Amount      int
	Address     libwallet.MuunAddress
	Script      []byte
}

// scanContext contains the synchronization objects for a single Scanner round, to manage Tasks.
type scanContext struct {
	addresses chan libwallet.MuunAddress
	results   chan Utxo
	errors    chan error
	done      chan struct{}
	wg        *sync.WaitGroup
}

// NewScanner creates an initialized Scanner.
func NewScanner() *Scanner {
	return &Scanner{
		pool:    electrum.NewPool(electrumPoolSize),
		servers: NewServerProvider(),
		log:     utils.NewLogger("Scanner"),
	}
}

// Scan an address space and return all relevant transactions for a sweep.
func (s *Scanner) Scan(addresses chan libwallet.MuunAddress) ([]Utxo, error) {
	var results []Utxo
	var waitGroup sync.WaitGroup

	// Create the Context that goroutines will share:
	ctx := &scanContext{
		addresses: addresses,
		results:   make(chan Utxo),
		errors:    make(chan error),
		done:      make(chan struct{}),
		wg:        &waitGroup,
	}

	// Start the scan in background:
	go s.startScan(ctx)

	// Collect all results until the done signal, or abort on the first error:
	for {
		select {
		case err := <-ctx.errors:
			close(ctx.done) // send the done signal ourselves
			return nil, err

		case result := <-ctx.results:
			results = append(results, result)

		case <-ctx.done:
			return results, nil
		}
	}
}

func (s *Scanner) startScan(ctx *scanContext) {
	s.log.Printf("Scan started")

	batches := streamBatches(ctx.addresses)

	var client *electrum.Client

	for batch := range batches {
		// Stop the loop until a client becomes available, or the scan is canceled:
		select {
		case <-ctx.done:
			return

		case client = <-s.pool.Acquire():
		}

		// Start scanning this address in background:
		ctx.wg.Add(1)

		go func(batch []libwallet.MuunAddress) {
			defer s.pool.Release(client)
			defer ctx.wg.Done()

			s.scanBatch(ctx, client, batch)
		}(batch)
	}

	// Wait for all tasks that are still executing to complete:
	ctx.wg.Wait()
	s.log.Printf("Scan complete")

	// Signal to the Scanner that this Context has no more pending work:
	close(ctx.done)
}

func (s *Scanner) scanBatch(ctx *scanContext, client *electrum.Client, batch []libwallet.MuunAddress) {
	// NOTE:
	// We begin by building the task, passing our selected Client. Since we're choosing the instance,
	// it's our job to control acquisition and release of Clients to prevent sharing (remember,
	// clients are single-user). The task won't enforce this safety measure (it can't), it's fully
	// up to us.
	task := &scanTask{
		servers:   s.servers,
		client:    client,
		addresses: batch,
		timeout:   taskTimeout,
		exit:      ctx.done,
	}

	// Do the thing:
	addressResults, err := task.Execute()

	if err != nil {
		ctx.errors <- s.log.Errorf("Scan failed: %w", err)
		return
	}

	// Send back all results:
	for _, result := range addressResults {
		ctx.results <- result
	}
}

func streamBatches(addresses chan libwallet.MuunAddress) chan []libwallet.MuunAddress {
	batches := make(chan []libwallet.MuunAddress)

	go func() {
		var nextBatch []libwallet.MuunAddress

		for address := range addresses {
			// Add items to the batch until we reach the limit:
			nextBatch = append(nextBatch, address)

			if len(nextBatch) < batchSize {
				continue
			}

			// Send back the batch and start over:
			batches <- nextBatch
			nextBatch = []libwallet.MuunAddress{}
		}

		// Send back an incomplete batch with any remaining addresses:
		if len(nextBatch) > 0 {
			batches <- nextBatch
		}

		close(batches)
	}()

	return batches
}
