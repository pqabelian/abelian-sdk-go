package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Define data types.
type AbecRPCClient struct {
	httpClient *http.Client
	endpoint   string
	username   string
	password   string
}

type AbecJSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      string        `json:"id"`
}

type AbecJSONRPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  json.RawMessage `json:"error"`
	ID     string          `json:"id"`
}

type AbecChainInfo struct {
	NumBlocks       int64   `json:"blocks"`
	IsTestnet       bool    `json:"testnet"`
	Version         int64   `json:"version"`
	ProtocolVersion int64   `json:"protocolversion"`
	RelayFee        float64 `json:"relayfee"`
	NetID           byte    `json:"netid"`
}

type AbecMempool map[string]struct {
	Size             int64   `json:"size"`
	FullSize         int64   `json:"fullsize"`
	Fee              float64 `json:"fee"`
	Time             int64   `json:"time"`
	Height           int64   `json:"height"`
	StartingPriority float64 `json:"startingpriority"`
	CurrentPriority  float64 `json:"currentpriority"`
}

type AbecBlock struct {
	Height        int64     `json:"height"`
	Confirmations int64     `json:"confirmations"`
	Version       int64     `json:"version"`
	VersionHex    string    `json:"versionHex"`
	Time          int64     `json:"time"`
	Nonce         uint64    `json:"nonce"`
	Size          int64     `json:"size"`
	FullSize      int64     `json:"fullsize"`
	Difficulty    float64   `json:"difficulty"`
	BlockHash     string    `json:"hash"`
	PrevBlockHash string    `json:"previousblockhash"`
	NextBlockHash string    `json:"nextblockhash"`
	ContentHash   string    `json:"contenthash"`
	MerkleRoot    string    `json:"merkleroot"`
	Bits          string    `json:"bits"`
	SealHash      string    `json:"sealhash"`
	Mixdigest     string    `json:"mixdigest"`
	TxHashes      []string  `json:"tx"`
	RawTxs        []*AbecTx `json:"rawTx"`
}

type AbecTx struct {
	Hex           string        `json:"hex"`
	TxID          string        `json:"txid"`
	TxHash        string        `json:"hash"`
	Time          int64         `json:"time"`
	BlockHash     string        `json:"blockhash"`
	BlockTime     int64         `json:"blocktime"`
	Confirmations int64         `bson:"confirmations"`
	Version       int64         `json:"version"`
	Size          int64         `json:"size"`
	FullSize      int64         `json:"fullsize"`
	Memo          []byte        `json:"memo"`
	Fee           float64       `json:"fee"`
	Witness       string        `json:"witness"`
	Vin           []*AbecTxVin  `json:"vin"`
	Vout          []*AbecTxVout `json:"vout"`
}

type AbecTxVin struct {
	UTXORing     AbecUTXORing `json:"prevutxoring"`
	SerialNumber string       `json:"serialnumber"`
}

type AbecUTXORing struct {
	Version     int64    `json:"version"`
	BlockHashes []string `json:"blockhashs"`
	OutPoints   []struct {
		TxHash string `json:"txid"`
		Index  int64  `json:"index"`
	} `json:"outpoints"`
}

type AbecTxVout struct {
	N      int64  `json:"n"`
	Script string `json:"script"`
}

// Define methods for AbecRPCClient.
func NewAbecRPCClient(endpoint string, username string, password string) *AbecRPCClient {
	return &AbecRPCClient{
		httpClient: &http.Client{},
		endpoint:   endpoint,
		username:   username,
		password:   password,
	}
}

func (client *AbecRPCClient) newRequest(id string, method string, params []interface{}) (*http.Request, error) {
	jsonReq := &AbecJSONRPCRequest{
		JSONRPC: "1.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}
	jsonBody, err := json.Marshal(jsonReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, client.endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(client.username, client.password)

	return httpReq, nil
}

func (client *AbecRPCClient) callForBytes(method string, params []interface{}) (Bytes, error) {
	id := fmt.Sprintf("%d", time.Now().UnixMilli())
	req, err := client.newRequest(id, method, params)
	if err != nil {
		return nil, err
	}

	LOG.debug("Request(%s): %s(%+v)\n", id, method, params)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		LOG.debug("Response(%s): ERROR(%s)\n", id, err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		LOG.debug("Response(%s): ERROR(%s)\n", id, err)
		return nil, err
	}
	LOG.debug("Response(%s): %s\n", id, body)

	respObj := &AbecJSONRPCResponse{}
	err = json.Unmarshal(body, respObj)
	if err != nil {
		return nil, err
	}

	errorStr := string(respObj.Error)
	if len(errorStr) > 0 && errorStr != "null" {
		return nil, fmt.Errorf("abec.%s: %s", method, respObj.Error)
	}

	return AsBytes(respObj.Result), nil
}

func AbecRPCClientCallForResult[ResultType any](client *AbecRPCClient, result *ResultType, method string, params []interface{}) (Bytes, *ResultType, error) {
	resultBytes, err := client.callForBytes(method, params)
	if err != nil {
		return nil, nil, err
	}

	if result == nil {
		return resultBytes, nil, nil
	}

	err = resultBytes.JSONUnmarshal(result)
	if err != nil {
		return resultBytes, nil, err
	}

	return resultBytes, result, nil
}

func (client *AbecRPCClient) GetChainInfo() (Bytes, *AbecChainInfo, error) {
	return AbecRPCClientCallForResult(client, &AbecChainInfo{}, "getinfo", nil)
}

func (client *AbecRPCClient) GetMempool() (Bytes, *AbecMempool, error) {
	return AbecRPCClientCallForResult(client, &AbecMempool{}, "getrawmempool", []interface{}{true})
}

func (client *AbecRPCClient) GetBlockHash(height int64) (Bytes, *string, error) {
	return AbecRPCClientCallForResult(client, new(string), "getblockhash", []interface{}{height})
}

func (client *AbecRPCClient) GetBlock(hash string) (Bytes, *AbecBlock, error) {
	return AbecRPCClientCallForResult(client, &AbecBlock{}, "getblockabe", []interface{}{hash, 1})
}

func (client *AbecRPCClient) GetBlockBytes(hash string) (Bytes, error) {
	var data string
	_, result, err := AbecRPCClientCallForResult(client, &data, "getblockabe", []interface{}{hash, 0})
	if err != nil {
		return nil, err
	}

	blockBytes := MakeBytesFromHexString(*result)
	return blockBytes, nil
}

func (client *AbecRPCClient) GetTxBytes(hash string) (Bytes, error) {
	var data string
	_, result, err := AbecRPCClientCallForResult(client, &data, "getrawtransaction", []interface{}{hash, false})
	if err != nil {
		return nil, err
	}

	blockBytes := MakeBytesFromHexString(*result)
	return blockBytes, nil
}

func (client *AbecRPCClient) GetRawTx(hash string) (Bytes, *AbecTx, error) {
	return AbecRPCClientCallForResult(client, &AbecTx{}, "getrawtransaction", []interface{}{hash, true})
}

func (client *AbecRPCClient) GetBlockByHeight(height int64) (Bytes, *AbecBlock, error) {
	_, hash, err := client.GetBlockHash(height)
	if err != nil {
		return nil, nil, err
	}

	return client.GetBlock(*hash)
}

func (client *AbecRPCClient) GetBlockBytesByHeight(height int64) (Bytes, error) {
	_, hash, err := client.GetBlockHash(height)
	if err != nil {
		return nil, err
	}

	return client.GetBlockBytes(*hash)
}

func (client *AbecRPCClient) GetEstimatedTxFee() int64 {
	return AbelToNeutrino(0.1)
}

func (client *AbecRPCClient) SendRawTx(txStr string) (Bytes, *string, error) {
	return AbecRPCClientCallForResult(client, new(string), "sendrawtransactionabe", []interface{}{txStr})
}
