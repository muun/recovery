package survey

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/muun/recovery/electrum"
)

type Survey struct {
	config  *Config
	tasks   chan *surveyTask
	taskWg  sync.WaitGroup
	results chan *Result
	visited map[string]bool
}

type Config struct {
	InitialServers     []string
	Workers            int
	SpeedTestDuration  time.Duration
	SpeedTestBatchSize int
}

type Result struct {
	Server        string
	FromPeer      string
	IsWorthy      bool
	Err           error
	Impl          string
	Version       string
	TimeToConnect time.Duration
	Speed         int
	BatchSupport  bool
	peers         []string
}

type surveyTask struct {
	server   string
	fromPeer string
}

// Values to check whether we're in the same chain (in a previous version, SV servers snuck in)
var mainnetSomeTx = "985eb411473fa1bbd73efa5e3685edc00366c86b8d4d3f5b969ad59c23f4d959"
var mainnetGenesisHash = "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"

func NewSurvey(config *Config) *Survey {
	return &Survey{
		config:  config,
		tasks:   make(chan *surveyTask),
		results: make(chan *Result),
		visited: make(map[string]bool),
	}
}

func (s *Survey) Run() []*Result {
	// Add initial tasks:
	for _, server := range s.config.InitialServers {
		s.addTask(server, "")
	}

	// Start collecting results in background:
	results := []*Result{}
	go s.startCollect(&results)

	// Launch workers to process tasks and send back results:
	for i := 0; i < s.config.Workers; i++ {
		go s.startWorker()
	}

	// Wait until there's no tasks left, and signal everyone to stop:
	s.taskWg.Wait()
	close(s.tasks)
	close(s.results)

	sort.Slice(results, func(i, j int) bool {
		return results[i].IsBetterThan(results[j])
	})

	return results
}

func (s *Survey) addTask(server string, fromPeer string) {
	task := &surveyTask{server, fromPeer}

	if _, ok := s.visited[task.server]; ok {
		return
	}
	s.visited[task.server] = true

	s.taskWg.Add(1)
	go func() { s.tasks <- task }() // scheduling tasks is non-blocking for users of the type
}

func (s *Survey) notifyResult(result *Result) {
	s.results <- result
	s.taskWg.Done()
}

func (s *Survey) startCollect(resultsRef *[]*Result) {
	for result := range s.results {
		*resultsRef = append(*resultsRef, result)
	}
}

func (s *Survey) startWorker() {
	for task := range s.tasks {
		log("• %s\n", task.server)

		result := s.processTask(task)

		if result.Err != nil {
			log("✕ %s\n", task.server)
		} else {
			log("✓ %s\n", task.server)
		}

		s.notifyResult(result)
	}
}

func (s *Survey) processTask(task *surveyTask) *Result {

	// We're going to perform a number of tests an measurements:
	//
	// 1. How much time does it take to establish a connection?
	// 2. Does the server support batching?
	// 3. Is the server willing to share its peers? If so, crawl.
	// 4. How many requests can the server handle in a given time interval?
	// 5. Did the server fail at any point during testing?
	//
	// Each test can result in a closed socket (since Electrum communicates errors by slapping you
	// in the face with no explanation), so we'll be connecting separately for each attempt.
	//
	// When a testing method returns an error, it means the server failed completely and we couldn't
	// obtain meaningful results (while some internal errors in a test are expected and handled).

	impl, version, timeToConnect, err := testConnection(task)
	if err != nil {
		return &Result{Server: task.server, Err: err}
	}

	isBitcoinMainnet, err := testBitcoinMainnet(task)
	if err != nil || !isBitcoinMainnet {
		return &Result{Server: task.server, Err: fmt.Errorf("not on Bitcoin mainnet: %w", err)}
	}

	batchSupport, err := testBatchSupport(task)
	if err != nil {
		return &Result{Server: task.server, Err: err}
	}

	speed, err := s.measureSpeed(task)
	if err != nil {
		return &Result{Server: task.server, Err: err}
	}

	peers, err := getPeers(task)
	if err != nil {
		return &Result{Server: task.server, Err: err}
	}

	for _, peer := range peers {
		if strings.Contains(peer, ".onion:") {
			continue
		}

		s.addTask(peer, task.server)
	}

	isWorthy := err == nil &&
		batchSupport &&
		timeToConnect.Seconds() < 5.0 &&
		speed >= int(s.config.SpeedTestDuration.Seconds())

	return &Result{
		IsWorthy:      isWorthy,
		Server:        task.server,
		FromPeer:      task.fromPeer,
		Impl:          impl,
		Version:       version,
		TimeToConnect: timeToConnect,
		BatchSupport:  batchSupport,
		Speed:         speed,
		peers:         peers,
	}
}

// testConnection returns the server implementation, protocol version and time to connect
func testConnection(task *surveyTask) (string, string, time.Duration, error) {
	client := electrum.NewClient(true)

	start := time.Now()
	err := client.Connect(task.server)
	if err != nil {
		return "", "", 0, err
	}

	return client.ServerImpl, client.ProtoVersion, time.Since(start), nil
}

// testsBlockchain returns whether this server is operating on Bitcoin mainnet
func testBitcoinMainnet(task *surveyTask) (bool, error) {
	client := electrum.NewClient(true)

	err := client.Connect(task.server)
	if err != nil {
		return false, err
	}

	features, err := client.ServerFeatures()
	if err != nil || features.GenesisHash != mainnetGenesisHash {
		return false, err
	}

	_, err = client.GetTransaction(mainnetSomeTx)
	if err != nil {
		return false, err
	}

	return true, nil
}

// testBatchSupport returns whether the server successfully responded to a batched request
func testBatchSupport(task *surveyTask) (bool, error) {
	client := electrum.NewClient(true)

	err := client.Connect(task.server)
	if err != nil {
		return false, err
	}

	_, err = client.ListUnspentBatch(createFakeHashes(2))
	if err != nil {
		return false, nil // an error here suggests lack of support for this call
	}

	return true, nil
}

// measureSpeed returns the amount of successful ListUnspentBatch calls in SPEED_TEST_DURATION
// seconds. It assumes batch support was verified beforehand.
func (s *Survey) measureSpeed(task *surveyTask) (int, error) {
	client := electrum.NewClient(true)

	err := client.Connect(task.server)
	if err != nil {
		return 0, err
	}

	start := time.Now()
	responseCount := 0

	for time.Since(start) < s.config.SpeedTestDuration {
		fakeHashes := createFakeHashes(s.config.SpeedTestBatchSize)

		_, err := client.ListUnspentBatch(fakeHashes) // TODO: is the faking affecting the result?
		if err != nil {
			return 0, err
		}

		responseCount++
	}

	return responseCount - 1, nil // the last one was over the time limit
}

// getPeers returns the list of peers from a server, or empty if it doesn't responds to the request
func getPeers(task *surveyTask) ([]string, error) {
	client := electrum.NewClient(true)

	err := client.Connect(task.server)
	if err != nil {
		return nil, err
	}

	peers, err := client.ServerPeers()
	if err != nil {
		return []string{}, nil // an error here suggests lack of support for this call
	}

	return peers, nil
}

func (r *Result) IsBetterThan(other *Result) bool {
	if r.Err != nil {
		return false
	}
	if other.Err != nil {
		return true
	}

	if r.IsWorthy != other.IsWorthy {
		return r.IsWorthy
	}

	if r.BatchSupport != other.BatchSupport {
		return r.BatchSupport
	}

	if r.Speed != other.Speed {
		return (r.Speed > other.Speed)
	}

	return (r.TimeToConnect < other.TimeToConnect)
}

func (r *Result) String() string {
	return fmt.Sprintf(
		"%s, %s, %s, %v, %v, %d, %v",
		r.Server,
		r.Impl,
		r.Version,
		r.BatchSupport,
		r.TimeToConnect.Seconds(),
		r.Speed,
		r.Err,
	)
}

func createFakeHashes(count int) []string {
	randomBuffer := make([]byte, 32)
	fakeHashes := make([]string, count)

	for i := 0; i < count; i++ {
		rand.Read(randomBuffer)
		fakeHashes[i] = hex.EncodeToString(randomBuffer)
	}

	return fakeHashes
}

func log(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, args...)
}
