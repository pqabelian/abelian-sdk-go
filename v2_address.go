package core

import (
	"bytes"
	"fmt"
	api "github.com/pqabelian/abec/sdkapi/v2"
	abeAddr "github.com/pqabelian/abeutil/address/instanceaddress"
)

// Define the length of different type of coin address
const (
	COIN_ADDRESS_LENGTH_FULL_PRIVACY_PRE  = 9504
	COIN_ADDRESS_LENGTH_FULL_PRIVACY_RAND = 9633
	COIN_ADDRESS_LENGTH_PSEUDONYM         = 193
)

// CoinAddress abstract application layer’s functional requirements for
// underlying addresses
//
// Currently, method CoinAddressType is used to distinguish different address type
// Other methods are mainly added for convenience of use, display, etc.
type CoinAddress interface {
	PrivacyLevel() PrivacyLevel
	Validate() error

	Data() Bytes
	Fingerprint() Bytes
	String() string
	HexString() string
}

func NewCoinAddress(data Bytes) CoinAddress {
	var coinAddress CoinAddress
	address := NewAddress(data, COIN_ADDRESS_TYPE, data.Sha256())

	switch len(data) {
	case COIN_ADDRESS_LENGTH_FULL_PRIVACY_PRE:
		coinAddress = &CoinAddressFullPrivacyPre{
			Address: address,
		}
	case COIN_ADDRESS_LENGTH_FULL_PRIVACY_RAND:
		coinAddress = &CoinAddressFullPrivacy{
			Address: address,
		}
	case COIN_ADDRESS_LENGTH_PSEUDONYM:
		coinAddress = &CoinAddressPseudonym{
			Address: address,
		}
	default:
		LOG.Panicf("Invalid coin address")
	}

	return coinAddress
}

var _ CoinAddress = &CoinAddressFullPrivacyPre{}
var _ CoinAddress = &CoinAddressFullPrivacy{}
var _ CoinAddress = &CoinAddressPseudonym{}

type CoinAddressFullPrivacyPre struct {
	Address
}

func (a *CoinAddressFullPrivacyPre) PrivacyLevel() PrivacyLevel {
	return PRIVACY_LEVEL_FULL_PRIVACY_PRE
}
func (a *CoinAddressFullPrivacyPre) Validate() error {
	err := a.Address.Validate()
	if err != nil {
		return err
	}

	if a.data.Len() != COIN_ADDRESS_LENGTH_FULL_PRIVACY_PRE {
		return fmt.Errorf("coin address data length is not %d", COIN_ADDRESS_LENGTH_FULL_PRIVACY_PRE)
	}

	return nil
}

type CoinAddressFullPrivacy struct {
	Address
}

func (a *CoinAddressFullPrivacy) PrivacyLevel() PrivacyLevel {
	return PRIVACY_LEVEL_FULL_PRIVACY_RAND
}
func (a *CoinAddressFullPrivacy) Validate() error {
	err := a.Address.Validate()
	if err != nil {
		return err
	}

	if a.data.Len() != COIN_ADDRESS_LENGTH_FULL_PRIVACY_RAND {
		return fmt.Errorf("coin address data length is not %d", COIN_ADDRESS_LENGTH_FULL_PRIVACY_RAND)
	}

	return nil
}

type CoinAddressPseudonym struct {
	Address
}

func (a *CoinAddressPseudonym) PrivacyLevel() PrivacyLevel {
	return PRIVACY_LEVEL_PSEUDONYM
}
func (a *CoinAddressPseudonym) Validate() error {
	err := a.Address.Validate()
	if err != nil {
		return err
	}

	if a.data.Len() != COIN_ADDRESS_LENGTH_PSEUDONYM {
		return fmt.Errorf("coin address data length is not %d", COIN_ADDRESS_LENGTH_PSEUDONYM)
	}

	return nil
}

func NewCoinAddressFullPrivacyPre(data Bytes) *CoinAddressFullPrivacyPre {
	return &CoinAddressFullPrivacyPre{
		NewAddress(data, COIN_ADDRESS_TYPE, data.Sha256()),
	}
}

// Define the length of different type of crypto address
const (
	CRYPTO_ADDRESS_LENGTH_FULL_PRIVACT_PRE  = 10696
	CRYPTO_ADDRESS_LENGTH_FULL_PRIVACY_RAND = 10826
	CRYPTO_ADDRESS_LENGTH_PSEUDONYM         = 198
)

// CryptoAddress encapsulated coin address for upper layer
type CryptoAddress struct {
	Address
	cryptoScheme CryptoScheme
	privacyLevel PrivacyLevel
	coinAddress  CoinAddress
}

func NewCryptoAddress(data Bytes) *CryptoAddress {
	var cryptoAddress CryptoAddress

	if len(data) < 4 {
		LOG.Panicf("Invalid crypto address")
	}
	cryptoScheme, err := api.DeserializeCryptoScheme(data[:4])
	if err != nil {
		LOG.Panicf("Can not parse crypto scheme from input data")
	}

	privacyLevel, coinAddrData, err := api.ExtractCoinAddressFromCryptoAddress(data)
	if err != nil {
		LOG.Panicf("fail to parse coin address in crypto address: %v", err)
	}
	cryptoAddress.cryptoScheme = cryptoScheme
	cryptoAddress.privacyLevel = privacyLevel
	cryptoAddress.coinAddress = NewCoinAddress(coinAddrData)
	cryptoAddress.Address = NewAddress(data, CRYPTO_ADDRESS_TYPE, cryptoAddress.coinAddress.Fingerprint())

	return &cryptoAddress
}

func (a *CryptoAddress) GetCoinAddress() CoinAddress {
	return a.coinAddress
}

func (a *CryptoAddress) Validate() error {
	err := a.Address.Validate()
	if err != nil {
		return err
	}

	switch a.cryptoScheme {
	case CryptoSchemePQRingCT:
		if a.privacyLevel != PRIVACY_LEVEL_FULL_PRIVACY_PRE {
			return fmt.Errorf("mismatched crypto scheme %d and privacy level %d", a.cryptoScheme, a.privacyLevel)
		}
		if a.data.Len() != CRYPTO_ADDRESS_LENGTH_FULL_PRIVACT_PRE {
			return fmt.Errorf("crypto address data length is not %d, but got %d", CRYPTO_ADDRESS_LENGTH_FULL_PRIVACT_PRE, a.Address.Data().Len())
		}
	case CryptoSchemePQRingCTX:
		if a.privacyLevel == PRIVACY_LEVEL_FULL_PRIVACY_RAND {
			if a.data.Len() != CRYPTO_ADDRESS_LENGTH_FULL_PRIVACY_RAND {
				return fmt.Errorf("crypto address data length should be %d, but got %d", CRYPTO_ADDRESS_LENGTH_FULL_PRIVACY_RAND, a.Address.Data().Len())
			}
		} else if a.privacyLevel != PRIVACY_LEVEL_PSEUDONYM {
			if a.data.Len() != CRYPTO_ADDRESS_LENGTH_PSEUDONYM {
				return fmt.Errorf("crypto address data length should be %d, but got %d", CRYPTO_ADDRESS_LENGTH_PSEUDONYM, a.Address.Data().Len())
			}
		} else {
			return fmt.Errorf("mismatched crypto scheme %d and privacy level %d", a.cryptoScheme, a.privacyLevel)
		}

	default:
		LOG.Panicf("unsupported crypto scheme")
	}

	return nil
}

const (
	ABEL_ADDRESS_LENGTH_FULL_PRIVACT_PRE = 10729
	ABEL_ADDRESS_LENGTH_RINGCT           = 10859
	ABEL_ADDRESS_LENGTH_PSEUDONYM        = 231
)

// AbelAddress encapsulated crypto address for application layer
type AbelAddress struct {
	Address
	netID         NetID
	cryptoAddress *CryptoAddress
}

func NewAbelAddress(data Bytes) *AbelAddress {
	var abelAddress AbelAddress
	if len(data) < 1 {
		LOG.Panicf("invalid length")
	}
	if int(data[0]) > len(NetIDToName) {
		LOG.Panicf("invalid network id")
	}
	abelAddress.netID = NetID(data[0])
	abelAddress.cryptoAddress = NewCryptoAddress(data.Slice()[1 : data.Len()-abeAddr.CheckSumLength()])
	abelAddress.Address = NewAddress(data, ABEL_ADDRESS_TYPE, abelAddress.cryptoAddress.Fingerprint())
	return &abelAddress
}

func NewAbelAddressFromCryptoAddress(cryptoAddress *CryptoAddress, chainID ...int8) *AbelAddress {
	if len(chainID) == 0 {
		chainID = []int8{int8(DEFAULT_CHAIN_ID)}
	}
	instanceAddress := abeAddr.NewInstanceAddress(byte(chainID[0]), cryptoAddress.Data())
	serializedInstanceAddress := instanceAddress.Serialize()
	checkSum := abeAddr.CheckSum(serializedInstanceAddress)
	abelAddressData := append(serializedInstanceAddress, checkSum...)

	return &AbelAddress{
		Address:       NewAddress(abelAddressData, ABEL_ADDRESS_TYPE, cryptoAddress.Fingerprint()),
		netID:         NetID(chainID[0]),
		cryptoAddress: cryptoAddress,
	}
}

func (a *AbelAddress) Validate() error {
	err := a.Address.Validate()
	if err != nil {
		return err
	}

	if a.data.Len() != ABEL_ADDRESS_LENGTH_FULL_PRIVACT_PRE &&
		a.data.Len() != ABEL_ADDRESS_LENGTH_RINGCT &&
		a.data.Len() != ABEL_ADDRESS_LENGTH_PSEUDONYM {
		return fmt.Errorf("abel address data length is not one of {%d,%d,%d}",
			ABEL_ADDRESS_LENGTH_FULL_PRIVACT_PRE,
			ABEL_ADDRESS_LENGTH_RINGCT,
			ABEL_ADDRESS_LENGTH_PSEUDONYM,
		)
	}

	chainID := a.GetNetID()
	if chainID < 0 || chainID > 14 {
		return fmt.Errorf("abel address chain id is not in range [0, 14]")
	}

	cryptoAddress := a.GetCryptoAddress()
	bl, _ := api.CheckCryptoAddress(cryptoAddress.Data())
	if !bl {
		return fmt.Errorf("abel address crypto address is not cryptographically valid")
	}

	checksum := a.GetChecksum()
	calculatedChecksum := abeAddr.CheckSum(append([]byte{byte(chainID)}, cryptoAddress.Data()...))
	if !bytes.Equal(checksum, calculatedChecksum) {
		return fmt.Errorf("abel address checksum is not valid")
	}

	return nil
}

func (a *AbelAddress) GetNetID() NetID {
	return a.netID
}

func (a *AbelAddress) GetCryptoAddress() *CryptoAddress {
	return a.cryptoAddress
}

func (a *AbelAddress) GetChecksum() Bytes {
	return a.data.Slice()[a.data.Len()-abeAddr.CheckSumLength():]
}

func (a *AbelAddress) GetShortAbelAddress() *ShortAbelAddress {
	return MakeShortAbelAddress(a.fingerprint, a.Hash(), int8(a.GetNetID()))
}
