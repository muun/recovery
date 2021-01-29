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
const connectionTimeout = time.Second * 10
const messageDelim = byte('\n')

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

// ListUnspentResponse models a `blockchain.scripthash.listunspent` response.
type ListUnspentResponse struct {
	ID     int          `json:"id"`
	Result []UnspentRef `json:"result"`
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
	Value  int    `json:"value"`
	Height int    `json:"height"`
}

// Param is a convenience type that models an item in the `Params` array of an Request.
type Param = interface{}

// NewClient creates an initialized Client instance.
func NewClient() *Client {
	return &Client{
		log: utils.NewLogger(defaultLoggerTag),
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

	err := c.call(&request, &response)
	if err != nil {
		return nil, c.log.Errorf("ServerVersion failed: %w", err)
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

	err := c.call(&request, &response)
	if err != nil {
		return "", c.log.Errorf("Broadcast failed: %w", err)
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

	err := c.call(&request, &response)
	if err != nil {
		return nil, c.log.Errorf("ListUnspent failed: %w", err)
	}

	return response.Result, nil
}

// ListUnspentBatch is like `ListUnspent`, but using batching.
func (c *Client) ListUnspentBatch(indexHashes []string) ([][]UnspentRef, error) {
	requests := make([]*Request, len(indexHashes))

	for i, indexHash := range indexHashes {
		requests[i] = &Request{
			Method: "blockchain.scripthash.listunspent",
			Params: []Param{indexHash},
		}
	}

	var responses []ListUnspentResponse

	err := c.callBatch(requests, &responses)
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
	// TODO: check if insecure is necessary
	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	dialer := &net.Dialer{
		Timeout: connectionTimeout,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", c.Server, config)
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
func (c *Client) call(request *Request, response interface{}) error {
	// Assign a fresh request ID:
	request.ID = c.incRequestID()

	// Serialize the request:
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return c.log.Errorf("Marshal failed %v: %w", request, err)
	}

	// Make the call, obtain the serialized response:
	responseBytes, err := c.callRaw(requestBytes)
	if err != nil {
		return c.log.Errorf("Send failed %s: %w", string(requestBytes), err)
	}

	// Deserialize into an error, to see if there's any:
	var maybeErrorResponse ErrorResponse

	err = json.Unmarshal(responseBytes, &maybeErrorResponse)
	if err != nil {
		return c.log.Errorf("Unmarshal of potential error failed: %s %w", string(responseBytes), err)
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
func (c *Client) callBatch(requests []*Request, response interface{}) error {
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
	responseBytes, err := c.callRaw(requestBytes)
	if err != nil {
		return c.log.Errorf("Send failed %s: %w", string(requestBytes), err)
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
func (c *Client) callRaw(request []byte) ([]byte, error) {
	c.log.Printf("Sending %s", string(request))

	if !c.IsConnected() {
		return nil, c.log.Errorf("Send failed %s: not connected", string(request))
	}

	request = append(request, messageDelim)

	_, err := c.conn.Write(request)
	if err != nil {
		return nil, c.log.Errorf("Send failed %s: %w", string(request), err)
	}

	reader := bufio.NewReader(c.conn)

	response, err := reader.ReadBytes(messageDelim)
	if err != nil {
		return nil, c.log.Errorf("Receive failed: %w", err)
	}

	c.log.Printf("Received %s", string(response))

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
