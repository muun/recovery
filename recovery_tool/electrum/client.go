package electrum

import (
	"bufio"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/muun/recovery/utils"
)

const defaultLoggerTag = "Electrum/?"
const connectionTimeout = time.Second * 30
const callTimeout = time.Second * 30
const messageDelim = byte('\n')
const noTimeout = 0

var implsWithBatching = []string{"ElectrumX"}

// Client is a TLS client that implements a subset of the Electrum protocol.
//
// It includes a minimal implementation of a JSON-RPC client, since the one provided by the
// standard library doesn't support features such as batching.
//
// It is absolutely not thread-safe. Every Client should have a single owner.
type Client struct {
	Server        string
	ServerImpl    string
	ProtoVersion  string
	nextRequestID int
	conn          net.Conn
	log           *utils.Logger
	requireTls    bool
}

// Request models the structure of all Electrum protocol requests.
type Request struct {
	ID     int     `json:"id"`
	Method string  `json:"method"`
	Params []Param `json:"params"`
}

// ErrorResponse models the structure of a generic error response.
type ErrorResponse struct {
	ID    int         `json:"id"`
	Error interface{} `json:"error"` // type varies among Electrum implementations.
}

// ServerVersionResponse models the structure of a `server.version` response.
type ServerVersionResponse struct {
	ID     int      `json:"id"`
	Result []string `json:"result"`
}

// ServerFeaturesResponse models the structure of a `server.features` response.
type ServerFeaturesResponse struct {
	ID     int            `json:"id"`
	Result ServerFeatures `json:"result"`
}

// ServerPeersResponse models the structure (or lack thereof) of a `server.peers.subscribe` response
type ServerPeersResponse struct {
	ID     int           `json:"id"`
	Result []interface{} `json:"result"`
}

// ListUnspentResponse models a `blockchain.scripthash.listunspent` response.
type ListUnspentResponse struct {
	ID     int          `json:"id"`
	Result []UnspentRef `json:"result"`
}

// GetTransactionResponse models the structure of a `blockchain.transaction.get` response.
type GetTransactionResponse struct {
	ID     int    `json:"id"`
	Result string `json:"result"`
}

// BroadcastResponse models the structure of a `blockchain.transaction.broadcast` response.
type BroadcastResponse struct {
	ID     int    `json:"id"`
	Result string `json:"result"`
}

// UnspentRef models an item in the `ListUnspentResponse` results.
type UnspentRef struct {
	TxHash string `json:"tx_hash"`
	TxPos  int    `json:"tx_pos"`
	Value  int64  `json:"value"`
	Height int    `json:"height"`
}

// ServerFeatures contains the relevant information from `ServerFeatures` results.
type ServerFeatures struct {
	ID            int    `json:"id"`
	GenesisHash   string `json:"genesis_hash"`
	HashFunction  string `json:"hash_function"`
	ServerVersion string `json:"server_version"`
	ProcotolMin   string `json:"protocol_min"`
	ProtocolMax   string `json:"protocol_max"`
	Pruning       int    `json:"pruning"`
}

// Param is a convenience type that models an item in the `Params` array of an Request.
type Param = interface{}

// NewClient creates an initialized Client instance.
func NewClient(requireTls bool) *Client {
	return &Client{
		log:        utils.NewLogger(defaultLoggerTag),
		requireTls: requireTls,
	}
}

// Connect establishes a TLS connection to an Electrum server.
func (c *Client) Connect(server string) error {
	c.Disconnect()

	c.log.SetTag("Electrum/" + server)
	c.Server = server

	c.log.Printf("Connecting")

	err := c.establishConnection()
	if err != nil {
		c.Disconnect()
		return c.log.Errorf("Connect failed: %w", err)
	}

	// Before calling it a day send a test request (trust me), and as we do identify the server:
	err = c.identifyServer()
	if err != nil {
		c.Disconnect()
		return c.log.Errorf("Identifying server failed: %w", err)
	}

	c.log.Printf("Identified as %s (%s)", c.ServerImpl, c.ProtoVersion)

	return nil
}

// Disconnect cuts the connection (if connected) to the Electrum server.
func (c *Client) Disconnect() error {
	if c.conn == nil {
		return nil
	}

	c.log.Printf("Disconnecting")

	err := c.conn.Close()
	if err != nil {
		return c.log.Errorf("Disconnect failed: %w", err)
	}

	c.conn = nil
	return nil
}

// SupportsBatching returns whether this client can process batch requests.
func (c *Client) SupportsBatching() bool {
	for _, implName := range implsWithBatching {
		if strings.HasPrefix(c.ServerImpl, implName) {
			return true
		}
	}

	return false
}

// ServerVersion calls the `server.version` method and returns the [impl, protocol version] tuple.
func (c *Client) ServerVersion() ([]string, error) {
	request := Request{
		Method: "server.version",
		Params: []Param{},
	}

	var response ServerVersionResponse

	err := c.call(&request, &response, callTimeout)
	if err != nil {
		return nil, c.log.Errorf("ServerVersion failed: %w", err)
	}

	return response.Result, nil
}

// ServerFeatures calls the `server.features` method and returns the relevant part of the result.
func (c *Client) ServerFeatures() (*ServerFeatures, error) {
	request := Request{
		Method: "server.features",
		Params: []Param{},
	}

	var response ServerFeaturesResponse

	err := c.call(&request, &response, callTimeout)
	if err != nil {
		return nil, c.log.Errorf("ServerFeatures failed: %w", err)
	}

	return &response.Result, nil
}

// ServerPeers calls the `server.peers.subscribe` method and returns a list of server addresses.
func (c *Client) ServerPeers() ([]string, error) {
	res, err := c.rawServerPeers()
	if err != nil {
		return nil, err // note that, besides I/O errors, some servers close the socket on this request
	}

	var peers []string

	for _, entry := range res {
		// Get ready for some hot casting action. Not for the faint of heart.
		addr := entry.([]interface{})[1].(string)
		port := entry.([]interface{})[2].([]interface{})[1].(string)[1:]

		peers = append(peers, addr+":"+port)
	}

	return peers, nil
}

// rawServerPeers calls the `server.peers.subscribe` method and returns this monstrosity:
//
//	[ "<ip>", "<domain>", ["<version>", "s<SSL port>", "t<TLS port>"] ]
//
// Ports can be in any order, or absent if the protocol is not supported
func (c *Client) rawServerPeers() ([]interface{}, error) {
	request := Request{
		Method: "server.peers.subscribe",
		Params: []Param{},
	}

	var response ServerPeersResponse

	err := c.call(&request, &response, callTimeout)
	if err != nil {
		return nil, c.log.Errorf("rawServerPeers failed: %w", err)
	}

	return response.Result, nil
}

// Broadcast calls the `blockchain.transaction.broadcast` endpoint and returns the transaction hash.
func (c *Client) Broadcast(rawTx string) (string, error) {
	request := Request{
		Method: "blockchain.transaction.broadcast",
		Params: []Param{rawTx},
	}

	var response BroadcastResponse

	err := c.call(&request, &response, callTimeout)
	if err != nil {
		return "", c.log.Errorf("Broadcast failed: %w", err)
	}

	return response.Result, nil
}

// GetTransaction calls the `blockchain.transaction.get` endpoint and returns the transaction hex.
func (c *Client) GetTransaction(txID string) (string, error) {
	request := Request{
		Method: "blockchain.transaction.get",
		Params: []Param{txID},
	}

	var response GetTransactionResponse

	err := c.call(&request, &response, callTimeout)
	if err != nil {
		return "", c.log.Errorf("GetTransaction failed: %w", err)
	}

	return response.Result, nil
}

// ListUnspent calls `blockchain.scripthash.listunspent` and returns the UTXO results.
func (c *Client) ListUnspent(indexHash string) ([]UnspentRef, error) {
	request := Request{
		Method: "blockchain.scripthash.listunspent",
		Params: []Param{indexHash},
	}
	var response ListUnspentResponse

	err := c.call(&request, &response, callTimeout)
	if err != nil {
		return nil, c.log.Errorf("ListUnspent failed: %w", err)
	}

	return response.Result, nil
}

// ListUnspentBatch is like `ListUnspent`, but using batching.
func (c *Client) ListUnspentBatch(indexHashes []string) ([][]UnspentRef, error) {
	requests := make([]*Request, len(indexHashes))
	method := "blockchain.scripthash.listunspent"

	for i, indexHash := range indexHashes {
		requests[i] = &Request{
			Method: method,
			Params: []Param{indexHash},
		}
	}

	var responses []ListUnspentResponse

	// Give it a little more time than non-batch calls
	timeout := callTimeout * 2

	err := c.callBatch(method, requests, &responses, timeout)
	if err != nil {
		return nil, fmt.Errorf("ListUnspentBatch failed: %w", err)
	}

	// Don't forget to sort responses:
	sort.Slice(responses, func(i, j int) bool {
		return responses[i].ID < responses[j].ID
	})

	// Now we can collect all results:
	var unspentRefs [][]UnspentRef

	for _, response := range responses {
		unspentRefs = append(unspentRefs, response.Result)
	}

	return unspentRefs, nil
}

func (c *Client) establishConnection() error {
	// We first try to connect over TCP+TLS
	// If we fail and requireTls is false, we try over TCP

	// TODO: check if insecure is necessary
	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	dialer := &net.Dialer{
		Timeout: connectionTimeout,
	}

	tlsConn, err := tls.DialWithDialer(dialer, "tcp", c.Server, config)
	if err == nil {
		c.conn = tlsConn
		return nil
	}
	if c.requireTls {
		return err
	}

	conn, err := net.DialTimeout("tcp", c.Server, connectionTimeout)
	if err != nil {
		return err
	}

	c.conn = conn

	return nil
}

func (c *Client) identifyServer() error {
	serverVersion, err := c.ServerVersion()
	if err != nil {
		return err
	}

	c.ServerImpl = serverVersion[0]
	c.ProtoVersion = serverVersion[1]

	c.log.Printf("Identified %s %s", c.ServerImpl, c.ProtoVersion)

	return nil
}

// IsConnected returns whether this client is connected to a server.
// It does not guarantee the next request will succeed.
func (c *Client) IsConnected() bool {
	return c.conn != nil
}

// call executes a request with JSON marshalling, and loads the response into a pointer.
func (c *Client) call(request *Request, response interface{}, timeout time.Duration) error {
	// Assign a fresh request ID:
	request.ID = c.incRequestID()

	// Serialize the request:
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return c.log.Errorf("Marshal failed %v: %w", request, err)
	}

	// Make the call, obtain the serialized response:
	responseBytes, err := c.callRaw(request.Method, requestBytes, timeout)
	if err != nil {
		return c.log.Errorf("Send failed %s: %w", request.Method, err)
	}

	// Deserialize into an error, to see if there's any:
	var maybeErrorResponse ErrorResponse

	err = json.Unmarshal(responseBytes, &maybeErrorResponse)
	if err != nil {
		return c.log.Errorf("Unmarshal of potential error failed: %s %w", request.Method, err)
	}

	if maybeErrorResponse.Error != nil {
		return c.log.Errorf("Electrum error: %v", maybeErrorResponse.Error)
	}

	// Deserialize the response:
	err = json.Unmarshal(responseBytes, response)
	if err != nil {
		return c.log.Errorf("Unmarshal failed %s: %w", string(responseBytes), err)
	}

	return nil
}

// call executes a batch request with JSON marshalling, and loads the response into a pointer.
// Response may not match request order, so callers MUST sort them by ID.
func (c *Client) callBatch(
	method string, requests []*Request, response interface{}, timeout time.Duration,
) error {
	// Assign fresh request IDs:
	for _, request := range requests {
		request.ID = c.incRequestID()
	}

	// Serialize the request:
	requestBytes, err := json.Marshal(requests)
	if err != nil {
		return c.log.Errorf("Marshal failed %v: %w", requests, err)
	}

	// Make the call, obtain the serialized response:
	responseBytes, err := c.callRaw(method, requestBytes, timeout)
	if err != nil {
		return c.log.Errorf("Send failed %s: %w", method, err)
	}

	// Deserialize into an array of errors, to see if there's any:
	var maybeErrorResponses []ErrorResponse

	err = json.Unmarshal(responseBytes, &maybeErrorResponses)
	if err != nil {
		return c.log.Errorf("Unmarshal of potential error failed: %s %w", string(responseBytes), err)
	}

	// Walk the responses, returning the first error found:
	for _, maybeErrorResponse := range maybeErrorResponses {
		if maybeErrorResponse.Error != nil {
			return c.log.Errorf("Electrum error: %v", maybeErrorResponse.Error)
		}
	}

	// Deserialize the response:
	err = json.Unmarshal(responseBytes, response)
	if err != nil {
		return c.log.Errorf("Unmarshal failed %s: %w", string(responseBytes), err)
	}

	return nil
}

// callRaw sends a raw request in bytes, and returns a raw response (or an error).
func (c *Client) callRaw(method string, request []byte, timeout time.Duration) ([]byte, error) {
	c.log.Printf("Sending %s request", method)
	c.log.Tracef("Sending %s body: %s", method, string(request))

	if !c.IsConnected() {
		return nil, c.log.Errorf("Send failed %s: not connected", method)
	}

	request = append(request, messageDelim)

	start := time.Now()

	// SetDeadline is an absolute time based timeout. That is, we set an exact
	// time we want it to fail.
	var deadline time.Time
	if timeout == noTimeout {
		// This means no deadline
		deadline = time.Time{}
	} else {
		deadline = start.Add(timeout)
	}
	err := c.conn.SetDeadline(deadline)
	if err != nil {
		return nil, c.log.Errorf("Send failed %s: SetDeadline failed", method)
	}

	_, err = c.conn.Write(request)

	if err != nil {
		duration := time.Now().Sub(start)
		return nil, c.log.Errorf("Send failed %s after %vms: %w", method, duration.Milliseconds(), err)
	}

	reader := bufio.NewReader(c.conn)

	response, err := reader.ReadBytes(messageDelim)
	duration := time.Now().Sub(start)
	if err != nil {
		return nil, c.log.Errorf("Receive failed %s after %vms: %w", method, duration.Milliseconds(), err)
	}

	c.log.Printf("Received %s after %vms", method, duration.Milliseconds())
	c.log.Tracef("Received %s: %s", method, string(response))

	return response, nil
}

func (c *Client) incRequestID() int {
	c.nextRequestID++
	return c.nextRequestID
}

// GetIndexHash returns the script parameter to use with Electrum, given a Bitcoin address.
func GetIndexHash(script []byte) string {
	indexHash := sha256.Sum256(script)
	reverse(&indexHash)

	return hex.EncodeToString(indexHash[:])
}

// reverse the order of the provided byte array, in place.
func reverse(a *[32]byte) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}
