package core

import (
	"encoding/hex"
	"sort"

	api "github.com/abesuite/abec/sdkapi/v1"
)

// Define constants.
const (
	DEFAULT_CHAIN_ID = 0x00
)

// Define util functions.
func GetRingBlockHeights(height int64) []int64 {
	firstRingBlockHeight := height - height%3
	ringBlockHeights := []int64{firstRingBlockHeight, firstRingBlockHeight + 1, firstRingBlockHeight + 2}
	return ringBlockHeights
}

// Define the CryptoKey data type.
type CryptoKey struct {
	Bytes
}

// Define methods for CryptoKey.
func NewCryptoKey(data Bytes) *CryptoKey {
	return &CryptoKey{Bytes: data}
}

// Define the CryptoKeysAndAddress data type.
type CryptoKeysAndAddress struct {
	SpendSecretKey    CryptoKey
	SerialNoSecretKey CryptoKey
	ViewSecretKey     CryptoKey
	CryptoAddress     CryptoAddress
}

// Define wrapper methods for Abec APIs.
func GenerateSafeCryptoSeed() (Bytes, error) {
	return api.CryptoAddressKeySeedGen()
}

func GenerateCryptoKeysAndAddress(cryptoSeed Bytes) (*CryptoKeysAndAddress, error) {
	cryptoAddress, spendSecretKey, serialNoSecretKey, viewSecretKey, err := api.CryptoAddressKeyGen(cryptoSeed)
	if err != nil {
		return nil, err
	}

	cryptoKeysAndAddress := &CryptoKeysAndAddress{
		SpendSecretKey:    *NewCryptoKey(spendSecretKey),
		SerialNoSecretKey: *NewCryptoKey(serialNoSecretKey),
		ViewSecretKey:     *NewCryptoKey(viewSecretKey),
		CryptoAddress:     *NewCryptoAddress(cryptoAddress),
	}

	return cryptoKeysAndAddress, nil
}

func DecodeCoinAddressFromTxOutData(txOutData Bytes) (*CoinAddress, error) {
	coinAddressData, err := api.ExtractCoinAddressFromSerializedTxOut(txOutData)
	if err != nil {
		return nil, err
	}

	return NewCoinAddress(coinAddressData), nil
}

func DecodeValueFromTxOutData(txOutData Bytes, viewSecretKey *CryptoKey) (int64, error) {
	// api.ExtractCoinValueFromSerializedTxOut will clear up the view secret key param.
	// Thus we pass a copy of the view secret key to avoid this side effect.
	viewSecretKeyData := make([]byte, viewSecretKey.Len())
	copy(viewSecretKeyData, viewSecretKey.Bytes)

	value, err := api.ExtractCoinValueFromSerializedTxOut(txOutData, viewSecretKeyData)
	if err != nil {
		return -1, err
	}

	return int64(value), nil
}

func GenerateUnsignedRawTx(txDesc *TxDesc) (*UnsignedRawTx, error) {
	// Prepare outPointsToSpend.
	outPointsToSpend := make([]*api.OutPoint, 0, len(txDesc.TxInDescs))
	for i := 0; i < len(txDesc.TxInDescs); i++ {
		txidStr := hex.EncodeToString(txDesc.TxInDescs[i].TxHash)
		outPoint, err := api.NewOutPointFromTxIdStr(txidStr, txDesc.TxInDescs[i].TxOutIndex)
		if err != nil {
			return nil, err
		}
		outPointsToSpend = append(outPointsToSpend, outPoint)
	}

	// Prepare serializedBlocksForRingGroup.
	serializedBlocksForRingGroup := getSerializedBlocksForRingGroup(txDesc.TxRingBlockDescs)

	// Prepare txRequestOutputDesc.
	txRequestOutputDescs := make([]*api.TxRequestOutputDesc, 0, len(txDesc.TxOutDescs))
	for i := 0; i < len(txDesc.TxOutDescs); i++ {
		cryptoAddressData := txDesc.TxOutDescs[i].AbelAddress.GetCryptoAddress().Data()
		coinValue := uint64(txDesc.TxOutDescs[i].CoinValue)
		txRequestOutputDesc := api.NewTxRequestOutputDesc(cryptoAddressData, coinValue)
		txRequestOutputDescs = append(txRequestOutputDescs, txRequestOutputDesc)
	}

	// Call API to build the serializedTxRequestDesc.
	serializedTxRequestDesc, err := api.BuildTransferTxRequestDescFromBlocks(
		outPointsToSpend,
		serializedBlocksForRingGroup,
		txRequestOutputDescs,
		uint64(txDesc.TxFee),
		txDesc.TxMemo,
	)
	if err != nil {
		return nil, err
	}

	// Create an unsigned raw tx and return it.
	signers := make([]*ShortAbelAddress, 0, len(txDesc.TxInDescs))
	for _, txInDesc := range txDesc.TxInDescs {
		signers = append(signers, txInDesc.Owner)
	}

	return NewUnsignedRawTx(serializedTxRequestDesc, signers), nil
}

func GenerateSignedRawTx(unsignedRawTx *UnsignedRawTx, signerKeys []*CryptoKeysAndAddress) (*SignedRawTx, error) {
	// Prepare cryptoKeys.
	cryptoKeys := make([]*api.CryptoKey, 0, len(signerKeys))
	for i := 0; i < len(signerKeys); i++ {
		cryptoKeys = append(cryptoKeys, api.NewCryptoKey(
			signerKeys[i].CryptoAddress.Data(),
			signerKeys[i].SpendSecretKey.Bytes,
			signerKeys[i].SerialNoSecretKey.Bytes,
			signerKeys[i].ViewSecretKey.Bytes))
	}

	// Call API to create the signed raw tx.
	serializedTxFull, txid, err := api.CreateTransferTx(unsignedRawTx.Bytes, cryptoKeys)
	if err != nil {
		return nil, err
	}

	// Create a signed raw tx and return it.
	// NOTE: The txid used by the RPC/SDK/UI is a reversed version of the txid used by the API.
	reversedTxid := make([]byte, len(txid))
	for i := 0; i < len(txid); i++ {
		reversedTxid[i] = txid[len(txid)-i-1]
	}

	return NewSignedRawTx(serializedTxFull, AsBytes(reversedTxid)), nil
}

func DecodeCoinSerialNumbers(coinIDs []*CoinID, serialNoSecretKeys []*CryptoKey, ringBlockDescs map[int64]*TxBlockDesc) ([]Bytes, error) {
	// Prepare outPoints.
	outPoints := make([]*api.OutPoint, len(coinIDs))
	for i := 0; i < len(coinIDs); i++ {
		txidStr := hex.EncodeToString(coinIDs[i].TxHash)
		outPoint, err := api.NewOutPointFromTxIdStr(txidStr, coinIDs[i].Index)
		if err != nil {
			return nil, err
		}
		outPoints[i] = outPoint
	}

	// Prepare serializedBlocksForRingGroup.
	serializedBlocksForRingGroup := getSerializedBlocksForRingGroup(ringBlockDescs)

	// Prepare cryptoSecretKeys.
	cryptoSecretKeys := make([]*api.CryptoKey, len(serialNoSecretKeys))
	for i := 0; i < len(serialNoSecretKeys); i++ {
		cryptoSecretKeys[i] = api.NewCryptoKey(nil, nil, serialNoSecretKeys[i].Bytes, nil)
	}

	// Call API to generate coin serial numbers.
	serialNumbers, err := api.GenerateCoinSerialNumber(outPoints, serializedBlocksForRingGroup, cryptoSecretKeys)
	if err != nil {
		return nil, err
	}

	// Convert serial numbers to Bytes type and return them.
	coinSerialNumbers := make([]Bytes, len(coinIDs))
	for i := 0; i < len(serialNumbers); i++ {
		coinSerialNumbers[i] = AsBytes(serialNumbers[i])
	}

	return coinSerialNumbers, nil
}

func getSerializedBlocksForRingGroup(ringBlockDescs map[int64]*TxBlockDesc) [][]byte {
	heights := make([]int64, 0, len(ringBlockDescs))
	for height := range ringBlockDescs {
		heights = append(heights, height)
	}

	sort.Slice(heights, func(i, j int) bool {
		return heights[i] < heights[j]
	})

	serializedBlocksForRingGroup := make([][]byte, 0, len(ringBlockDescs))
	for i := 0; i < len(heights); i++ {
		serializedBlocksForRingGroup = append(serializedBlocksForRingGroup, ringBlockDescs[heights[i]].BinData)
	}

	return serializedBlocksForRingGroup
}
