package core

import (
	"bytes"
	"encoding/hex"
	api "github.com/pqabelian/abec/sdkapi/v2"
)

type accountType int

const (
	ACCOUNT_TYPE_ROOT_SEED accountType = iota
	ACCOUNT_TYPE_CRYPTO_KEYS
)

// ViewAccount encapsulates the ability to
// - determine whether the coin belongs to the corresponding account
// - generate serial number for specified coins
type ViewAccount interface {
	ReceiveCoin(txOutData Bytes) (success bool, v uint64, err error)
	GenerateSerialNumbers(coinIDs []*CoinID, ringBlockDescs map[int64]*TxBlockDesc) (coinSerialNumbers []Bytes, err error)
	ViewKeyMaterial() (Bytes, Bytes, Bytes)

	accountType() accountType
}

type RootSeedViewAccount struct {
	cryptoScheme            CryptoScheme
	privacyLevel            PrivacyLevel
	coinSerialNumberKeySeed Bytes
	coinValueKeySeed        Bytes
	coinDetectorKey         *CryptoKey
}

func NewRootSeedViewAccount(
	cryptoScheme CryptoScheme,
	privacyLevel PrivacyLevel,
	coinSerialNumberKeySeed Bytes,
	coinValueKeySeed Bytes,
	coinDetectorKey *CryptoKey,
) *RootSeedViewAccount {
	return &RootSeedViewAccount{cryptoScheme: cryptoScheme, privacyLevel: privacyLevel, coinSerialNumberKeySeed: coinSerialNumberKeySeed, coinValueKeySeed: coinValueKeySeed, coinDetectorKey: coinDetectorKey}
}

func (account *RootSeedViewAccount) accountType() accountType {
	return ACCOUNT_TYPE_ROOT_SEED
}
func (account *RootSeedViewAccount) GenerateSerialNumbers(coinIDs []*CoinID, ringBlockDescs map[int64]*TxBlockDesc) ([]Bytes, error) {
	if len(coinIDs) == 0 {
		return nil, nil
	}
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

	// Call API to generate coin serial numbers.
	serialNumbers, err := api.GenerateCoinSerialNumberByRootSeeds(outPoints, serializedBlocksForRingGroup, account.coinSerialNumberKeySeed)
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

func (account *RootSeedViewAccount) ReceiveCoin(txOutData Bytes) (success bool, v uint64, err error) {
	privacyLevel, err := api.GetTxoPrivacyLevel(api.TxVersion, txOutData)
	if err != nil {
		return false, 0, err
	}
	if account.privacyLevel != privacyLevel {
		return false, 0, nil
	}

	success, err = api.TxoCoinDetectByCoinDetectorRootKey(api.TxVersion, txOutData, account.coinDetectorKey.Bytes)
	if err != nil {
		return false, 0, err
	}
	if !success {
		return false, 0, nil
	}

	success, v, err = api.TxoCoinReceiveByRootSeeds(api.TxVersion, txOutData, account.coinValueKeySeed.Slice(), account.coinDetectorKey.Bytes)
	if err != nil {
		return false, 0, err
	}
	return success, v, nil
}
func (account *RootSeedViewAccount) ViewKeyMaterial() (Bytes, Bytes, Bytes) {
	var coinDetectorKey Bytes
	if account.coinDetectorKey != nil {
		coinDetectorKey = account.coinDetectorKey.Bytes
	}
	return account.coinSerialNumberKeySeed, account.coinValueKeySeed, coinDetectorKey
}

type CryptoKeysViewAccount struct {
	serialNoSecretKey *CryptoKey
	viewSecretKey     *CryptoKey
	detectorKey       *CryptoKey
	cryptoAddress     *CryptoAddress
}

func NewCryptoKeyViewAccount(serialNoSecretKey *CryptoKey, viewSecretKey *CryptoKey, detectorKey *CryptoKey, cryptoAddress *CryptoAddress) *CryptoKeysViewAccount {
	return &CryptoKeysViewAccount{serialNoSecretKey: serialNoSecretKey, viewSecretKey: viewSecretKey, detectorKey: detectorKey, cryptoAddress: cryptoAddress}
}
func (account *CryptoKeysViewAccount) accountType() accountType {
	return ACCOUNT_TYPE_CRYPTO_KEYS
}
func (account *CryptoKeysViewAccount) GenerateSerialNumbers(coinIDs []*CoinID, ringBlockDescs map[int64]*TxBlockDesc) ([]Bytes, error) {
	if len(coinIDs) == 0 {
		return nil, nil
	}
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
	cryptoSerialNumberSecretKeys := make([][]byte, len(coinIDs))
	for i := 0; i < len(coinIDs); i++ {
		cryptoSerialNumberSecretKeys[i] = account.serialNoSecretKey.Bytes
	}

	// Call API to generate coin serial numbers.
	serialNumbers, err := api.GenerateCoinSerialNumberByKeys(outPoints, serializedBlocksForRingGroup, cryptoSerialNumberSecretKeys)
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

func (account *CryptoKeysViewAccount) ReceiveCoin(txOutData Bytes) (success bool, v uint64, err error) {
	coinAddressFromSerializedTxOut, err := api.ExtractCoinAddressFromSerializedTxOut(api.TxVersion, txOutData)
	if err != nil {
		return false, 0, err
	}
	_, coinAddressFromCryptoAddress, err := api.ExtractCoinAddressFromCryptoAddress(account.cryptoAddress.Data())
	if err != nil {
		return false, 0, err
	}
	if !bytes.Equal(coinAddressFromCryptoAddress, coinAddressFromSerializedTxOut) {
		return false, 0, nil
	}

	copiedVskBytes := make([]byte, len(account.viewSecretKey.Bytes))
	copy(copiedVskBytes, account.viewSecretKey.Bytes)
	success, v, err = api.TxoCoinReceiveByKeys(api.TxVersion, txOutData, account.cryptoAddress.Data(), copiedVskBytes)
	if err != nil {
		return false, 0, err
	}
	return success, v, nil
}
func (account *CryptoKeysViewAccount) ViewKeyMaterial() (Bytes, Bytes, Bytes) {
	var coinSerialNoSecretKey Bytes
	if account.serialNoSecretKey != nil {
		coinSerialNoSecretKey = account.serialNoSecretKey.Bytes
	}
	var coinViewSecretKey Bytes
	if account.viewSecretKey != nil {
		coinViewSecretKey = account.viewSecretKey.Bytes
	}

	var coinDetectorKey Bytes
	if account.detectorKey != nil {
		coinDetectorKey = account.detectorKey.Bytes
	}
	return coinSerialNoSecretKey, coinViewSecretKey, coinDetectorKey
}

type Account interface {
	ViewAccount
	SpendKeyMaterial() Bytes
}

type RootSeedAccount struct {
	RootSeedViewAccount
	coinSpendKeySeed Bytes
}

func (account *RootSeedAccount) SpendKeyMaterial() Bytes {
	return account.coinSpendKeySeed
}

func NewRootSeedAccount(cryptoScheme CryptoScheme, privacyLevel PrivacyLevel, coinSpendKeySeed Bytes, coinSerialNumberKeySeed Bytes, coinValueKeySeed Bytes, coinDetectorKey *CryptoKey) *RootSeedAccount {
	return &RootSeedAccount{
		RootSeedViewAccount: RootSeedViewAccount{
			cryptoScheme:            cryptoScheme,
			privacyLevel:            privacyLevel,
			coinSerialNumberKeySeed: coinSerialNumberKeySeed,
			coinValueKeySeed:        coinValueKeySeed,
			coinDetectorKey:         coinDetectorKey,
		},
		coinSpendKeySeed: coinSpendKeySeed,
	}
}
func NewRootSeedAccountFromViewAccount(viewAccount RootSeedViewAccount, coinSpendKeySeed Bytes) *RootSeedAccount {
	return &RootSeedAccount{
		RootSeedViewAccount: viewAccount,
		coinSpendKeySeed:    coinSpendKeySeed,
	}
}

type CryptoKeysAccount struct {
	CryptoKeysViewAccount
	spendSecretKey *CryptoKey
}

func (account *CryptoKeysAccount) SpendKeyMaterial() Bytes {
	return account.spendSecretKey.Bytes
}
func NewCryptoKeysAccount(spendSecretKey *CryptoKey, serialNoSecretKey *CryptoKey, viewSecretKey *CryptoKey, detectorKey *CryptoKey, cryptoAddress *CryptoAddress) *CryptoKeysAccount {
	return &CryptoKeysAccount{
		CryptoKeysViewAccount: CryptoKeysViewAccount{
			serialNoSecretKey: serialNoSecretKey,
			viewSecretKey:     viewSecretKey,
			detectorKey:       detectorKey,
			cryptoAddress:     cryptoAddress,
		},
		spendSecretKey: spendSecretKey,
	}
}
func NewCryptoKeysAccountFromViewAccount(viewAccount CryptoKeysViewAccount, spendSecretKey *CryptoKey) *CryptoKeysAccount {
	return &CryptoKeysAccount{
		CryptoKeysViewAccount: viewAccount,
		spendSecretKey:        spendSecretKey,
	}
}
