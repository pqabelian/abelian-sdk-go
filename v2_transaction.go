package core

import (
	"fmt"
	api "github.com/pqabelian/abec/sdkapi/v2"
)

// GenerateSignedRawTxV2 signs the unsigned transaction
func GenerateSignedRawTxV2(unsignedRawTx *UnsignedRawTx, signerAccounts []Account) (*SignedRawTx, error) {
	if len(signerAccounts) == 0 {
		return nil, fmt.Errorf("no singer specified")
	}
	firstAccountType := signerAccounts[0].accountType()
	for i := 1; i < len(signerAccounts); i++ {
		if signerAccounts[i].accountType() != firstAccountType {
			return nil, fmt.Errorf("all specified account must be the same")
		}
	}

	var serializedTxFull Bytes
	var txid *api.TxId
	var err error
	switch firstAccountType {
	case ACCOUNT_TYPE_ROOT_SEED:
		seeds := make([]*api.CryptoRootSeed, 0, len(signerAccounts))
		for i := 0; i < len(signerAccounts); i++ {
			coinSerialNumberKeyMaterial, coinValueKeyMaterial, coinDetectorKeyMaterial := signerAccounts[i].ViewKeyMaterial()
			coinSpendSecretKeyMaterial := signerAccounts[i].SpendKeyMaterial()
			signerViewAccount := signerAccounts[i].(*RootSeedAccount)
			seeds = append(seeds, api.NewRootSeed(
				signerViewAccount.cryptoScheme,
				signerViewAccount.privacyLevel,
				coinSpendSecretKeyMaterial,
				coinSerialNumberKeyMaterial,
				coinValueKeyMaterial,
				coinDetectorKeyMaterial,
			))
		}
		serializedTxFull, txid, err = api.CreateTransferTxByRootSeed(unsignedRawTx.Bytes, seeds)
		if err != nil {
			return nil, err
		}
	case ACCOUNT_TYPE_CRYPTO_KEYS:
		// Prepare cryptoKeys.
		cryptoKeys := make([]*api.CryptoKey, 0, len(signerAccounts))
		for i := 0; i < len(signerAccounts); i++ {
			coinSerialNumberKeyMaterial, coinValueKeyMaterial, coinDetectorKeyMaterial := signerAccounts[i].ViewKeyMaterial()
			coinSpendSecretKeyMaterial := signerAccounts[i].SpendKeyMaterial()
			signerViewAccount := signerAccounts[i].(*CryptoKeysAccount)
			cryptoKeys = append(cryptoKeys, api.NewCryptoKey(
				signerViewAccount.cryptoAddress.data,
				coinSpendSecretKeyMaterial,
				coinSerialNumberKeyMaterial,
				coinValueKeyMaterial,
				coinDetectorKeyMaterial,
			))
		}

		// Call API to create the signed raw tx.
		serializedTxFull, txid, err = api.CreateTransferTxByCryptoKeys(unsignedRawTx.Bytes, cryptoKeys)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrInvalidAccountType
	}

	// Create a signed raw tx and return it.
	// NOTE: The txid used by the RPC/SDK/UI is a reversed version of the txid used by the API.
	reversedTxid := make([]byte, len(txid))
	for i := 0; i < len(txid); i++ {
		reversedTxid[i] = txid[len(txid)-i-1]
	}

	return NewSignedRawTx(serializedTxFull, AsBytes(reversedTxid)), nil
}
