package core

import (
	"fmt"
)

// Define constants.
type AddressType int

const (
	ANY_ADDRESS_TYPE AddressType = iota
	COIN_ADDRESS_TYPE
	CRYPTO_ADDRESS_TYPE
	ABEL_ADDRESS_TYPE
	SHORT_ABEL_ADDRESS_TYPE
)

const (
	COIN_ADDRESS_LENGTH       = COIN_ADDRESS_LENGTH_FULL_PRIVACY_PRE
	CRYPTO_ADDRESS_LENGTH     = CRYPTO_ADDRESS_LENGTH_FULL_PRIVACT_PRE
	ABEL_ADDRESS_LENGTH       = ABEL_ADDRESS_LENGTH_FULL_PRIVACT_PRE
	SHORT_ABEL_ADDRESS_LENGTH = 66
)

func (addressType AddressType) String() string {
	switch addressType {
	case ANY_ADDRESS_TYPE:
		return "AnyAddress"
	case COIN_ADDRESS_TYPE:
		return "CoinAddress"
	case CRYPTO_ADDRESS_TYPE:
		return "CryptoAddress"
	case ABEL_ADDRESS_TYPE:
		return "AbelAddress"
	case SHORT_ABEL_ADDRESS_TYPE:
		return "ShortAbelAddress"
	default:
		return "UnknownAddress"
	}
}

// Define the Address data type.
type Address struct {
	data        Bytes
	addressType AddressType
	fingerprint Bytes
}

// Define methods for Address.
func NewAddress(data Bytes, addressType AddressType, fingerprint ...Bytes) Address {
	if data == nil {
		data = make([]byte, 0)
	}

	if len(fingerprint) == 0 {
		fingerprint = nil
	}

	return Address{
		data:        data,
		addressType: addressType,
		fingerprint: fingerprint[0],
	}
}

func (a Address) String() string {
	return fmt.Sprintf("%s{%s|fp:%s}", a.addressType.String(), a.data.Summary(1, 8), a.fingerprint.Summary(0, 2))
}

func (a *Address) Type() AddressType {
	return a.addressType
}

func (a *Address) Data() Bytes {
	return a.data
}

func (a *Address) Fingerprint() Bytes {
	return a.fingerprint
}

func (a *Address) Hash() Bytes {
	return a.data.Sha256()
}

func (a *Address) HexString() string {
	return a.data.HexString()
}

func (a *Address) Validate() error {
	if a.data == nil || a.data.Len() == 0 {
		return fmt.Errorf("address data is empty")
	}

	if a.fingerprint == nil || a.fingerprint.Len() == 0 {
		return fmt.Errorf("address fingerprint is empty")
	}

	return nil
}

// Define the ShortAbelAddress data type.
type ShortAbelAddress struct {
	Address
}

// Define methods for ShortAbelAddress.
func NewShortAbelAddress(data Bytes) *ShortAbelAddress {
	shortAddress := &ShortAbelAddress{Address: NewAddress(data, SHORT_ABEL_ADDRESS_TYPE, nil)}
	shortAddress.fingerprint = data.Slice()[2:34]
	return shortAddress
}

func MakeShortAbelAddress(fingerprint Bytes, cryptoAddressHash Bytes, chainID ...int8) *ShortAbelAddress {
	if len(chainID) == 0 {
		chainID = []int8{int8(DEFAULT_CHAIN_ID)}
	}

	saData := make([]byte, 0, 2+fingerprint.Len()+cryptoAddressHash.Len())
	saData = append(saData, 0xab, 0xe1+byte(chainID[0]))
	saData = append(saData, fingerprint.Slice()...)
	saData = append(saData, cryptoAddressHash.Slice()...)

	return NewShortAbelAddress(saData)
}

func (a *ShortAbelAddress) Validate() error {
	err := a.Address.Validate()
	if err != nil {
		return err
	}

	if a.data.Len() != SHORT_ABEL_ADDRESS_LENGTH {
		return fmt.Errorf("short abel address data length is not %d", SHORT_ABEL_ADDRESS_LENGTH)
	}

	if a.data.Slice()[0] != 0xab {
		return fmt.Errorf("short abel address data is not prefixed with 0xab")
	}

	chainID := a.GetChainID()
	if chainID < 0 || chainID > 15 {
		return fmt.Errorf("short abel address chain id is not in range [0, 15]")
	}

	return nil
}

func (a *ShortAbelAddress) GetChainID() int8 {
	return int8(a.data.Slice()[1] - 0xe1)
}
