package core

import (
	"crypto/rand"
	"fmt"
	api "github.com/pqabelian/abec/sdkapi/v2"
)

func RandomBytes(length int) Bytes {
	res := make([]byte, 0, length)

	neededLen := length
	var tmp []byte
	for neededLen > 0 {
		tmp = make([]byte, neededLen)
		// n == len(b) if and only if err == nil.
		n, err := rand.Read(tmp)
		if err != nil {
			continue
		}
		res = append(res, tmp[:n]...)
		neededLen -= n
	}
	return AsBytes(res)
}

type CryptoScheme = api.CryptoScheme

const (
	CryptoSchemePQRingCT  = api.CryptoSchemePQRingCT
	CryptoSchemePQRingCTX = api.CryptoSchemePQRingCTX
)

func CryptoSchemeSerializeSize() int {
	return api.CryptoSchemeSerializeSize()
}
func SerializeCryptoScheme(cryptoScheme CryptoScheme) []byte {
	return api.SerializeCryptoScheme(cryptoScheme)
}
func DeserializeCryptoScheme(serializedCryptoScheme []byte) (CryptoScheme, error) {
	return api.DeserializeCryptoScheme(serializedCryptoScheme)
}

// PrivacyLevel alias type for api.PrivacyLevel to identifier different privacy level
// Defined by a cryptographic layer to differentiate for different target
//
// currently, two full-privacy addresses and one pseudonym address are supported
type PrivacyLevel = api.PrivacyLevel

const (
	PRIVACY_LEVEL_FULL_PRIVACY_PRE  PrivacyLevel = api.PrivacyLevelRINGCTPre // for back-compatibility
	PRIVACY_LEVEL_FULL_PRIVACY_RAND              = api.PrivacyLevelRINGCT
	PRIVACY_LEVEL_PSEUDONYM                      = api.PrivacyLevelPSEUDONYM
)

type seedsType int

const (
	RAND_SEEDS_TYPE seedsType = 0
	ROOT_SEEDS_TYPE           = 1
)

func (seedType seedsType) String() string {
	switch seedType {
	case RAND_SEEDS_TYPE:
		return "RandSeeds"
	case ROOT_SEEDS_TYPE:
		return "RootSeeds"
	default:
		return "UnknownSeeds"
	}
}

// CryptoSeeds encapsulate cryptography seeds for different purposes, details as follow:
// - ROOT_SEEDS_TYPE: allow to generate multiple address, and supports scanning coins of all generated addresses through seed
// (cryptoScheme,privacyLevel)                              coinSpendKeySeed coinSerialNumberKeySeed  coinValueKeySeed coinDetectorKey publicRand
// (CryptoSchemePQRingCT, PRIVACY_LEVEL_FULL_PRIVACY_PRE)           -                  -                      -               -            -
// (CryptoSchemePQRingCTX, PRIVACY_LEVEL_FULL_PRIVACY_RAND)        yes              yes                     yes               yes          no
// (CryptoSchemePQRingCTX, PRIVACY_LEVEL_PSEUDONYM)                yes               no                      no               yes          no
//
// - RAND_SEEDS_TYPE: only one address can be generated, and supports scanning coins through seed or generated keys
// (cryptoScheme,privacyLevel)                              coinSpendKeySeed coinSerialNumberKeySeed  coinValueKeySeed coinDetectorKey publicRand
// (CryptoSchemePQRingCT, PRIVACY_LEVEL_FULL_PRIVACY_PRE)           yes               no                     yes                no           no   // for back-compatibility
// (CryptoSchemePQRingCTX, PRIVACY_LEVEL_FULL_PRIVACY_RAND)         yes              yes                     yes               yes          yes
// (CryptoSchemePQRingCTX, PRIVACY_LEVEL_PSEUDONYM)                 yes               no                      no               yes          yes
type CryptoSeeds struct {
	seedsType               seedsType
	cryptoScheme            CryptoScheme
	privacyLevel            PrivacyLevel
	coinSpendKeySeed        Bytes
	coinSerialNumberKeySeed Bytes
	coinValueKeySeed        Bytes
	coinDetectorKey         *CryptoKey
	publicRand              Bytes
}

func (s *CryptoSeeds) CryptoScheme() CryptoScheme {
	return s.cryptoScheme
}

func (s *CryptoSeeds) PrivacyLevel() PrivacyLevel {
	return s.privacyLevel
}

func (s *CryptoSeeds) CoinSpendKeySeed() Bytes {
	return s.coinSpendKeySeed
}

func (s *CryptoSeeds) CoinSerialNumberKeySeed() Bytes {
	return s.coinSerialNumberKeySeed
}

func (s *CryptoSeeds) CoinValueKeySeed() Bytes {
	return s.coinValueKeySeed
}

func (s *CryptoSeeds) CoinDetectorKey() *CryptoKey {
	return s.coinDetectorKey
}

func (s *CryptoSeeds) PublicRand() Bytes {
	return s.publicRand
}

func NewRootSeeds(cryptoScheme CryptoScheme, privacyLevel PrivacyLevel,
	coinSpendKeySeed Bytes, coinSerialNumberKeySeed Bytes,
	coinValueKeySeed Bytes, coinDetectorKey *CryptoKey) (*CryptoSeeds, error) {
	if cryptoScheme != CryptoSchemePQRingCTX {
		return nil, ErrInvalidCryptoScheme
	}
	if privacyLevel != PRIVACY_LEVEL_FULL_PRIVACY_RAND && privacyLevel != PRIVACY_LEVEL_PSEUDONYM {
		return nil, ErrInvalidPrivacyLevel
	}

	seed := &CryptoSeeds{
		seedsType:               ROOT_SEEDS_TYPE,
		cryptoScheme:            cryptoScheme,
		privacyLevel:            privacyLevel,
		coinSpendKeySeed:        coinSpendKeySeed,
		coinSerialNumberKeySeed: coinSerialNumberKeySeed,
		coinValueKeySeed:        coinValueKeySeed,
		coinDetectorKey:         coinDetectorKey,
		publicRand:              nil,
	}

	if privacyLevel == PRIVACY_LEVEL_PSEUDONYM {
		seed.coinSerialNumberKeySeed = nil
		seed.coinValueKeySeed = nil
	}
	return seed, nil
}
func NewRandSeeds(cryptoScheme CryptoScheme, privacyLevel PrivacyLevel,
	coinSpendKeySeed Bytes, coinSerialNumberKeySeed Bytes, coinValueKeySeed Bytes,
	coinDetectorKey Bytes, publicRand Bytes) (*CryptoSeeds, error) {
	seed := &CryptoSeeds{
		seedsType:               RAND_SEEDS_TYPE,
		cryptoScheme:            cryptoScheme,
		privacyLevel:            privacyLevel,
		coinSpendKeySeed:        coinSpendKeySeed,
		coinSerialNumberKeySeed: nil,
		coinValueKeySeed:        nil,
		coinDetectorKey:         nil,
		publicRand:              nil,
	}
	switch cryptoScheme {
	case CryptoSchemePQRingCT:
		if privacyLevel != PRIVACY_LEVEL_FULL_PRIVACY_PRE {
			return nil, ErrInvalidPrivacyLevel
		}
		seed.coinValueKeySeed = coinValueKeySeed
	case CryptoSchemePQRingCTX:
		if privacyLevel != PRIVACY_LEVEL_FULL_PRIVACY_RAND && privacyLevel != PRIVACY_LEVEL_PSEUDONYM {
			return nil, ErrInvalidPrivacyLevel
		}

		seed.coinSerialNumberKeySeed = coinSerialNumberKeySeed
		seed.coinValueKeySeed = coinValueKeySeed
		seed.coinDetectorKey = NewCryptoKey(coinDetectorKey)
		seed.publicRand = publicRand

		if privacyLevel == PRIVACY_LEVEL_PSEUDONYM {
			seed.coinSerialNumberKeySeed = nil
			seed.coinValueKeySeed = nil
		}
	default:
		return nil, ErrInvalidCryptoScheme

	}

	return seed, nil
}

func (s *CryptoSeeds) String() string {
	return fmt.Sprintf("%s{%d|%d}{%x|%x|%x|%x|%x}",
		s.seedsType.String(), s.cryptoScheme, s.privacyLevel,
		s.coinSpendKeySeed, s.coinValueKeySeed,
		s.coinValueKeySeed, s.coinDetectorKey,
		s.publicRand)
}

func (s *CryptoSeeds) Validate() error {
	switch s.cryptoScheme {
	case CryptoSchemePQRingCT:
		if s.seedsType != RAND_SEEDS_TYPE {
			return ErrMismatchedSeedType
		}
		if s.privacyLevel != PRIVACY_LEVEL_FULL_PRIVACY_PRE {
			return ErrMismatchedCryptoSchemePrivacyLevel
		}
		if s.coinSpendKeySeed == nil && s.coinSpendKeySeed.Len() == 0 {
			return ErrCorruptedSeed
		}
		if s.coinValueKeySeed == nil && s.coinValueKeySeed.Len() == 0 {
			return ErrCorruptedSeed
		}
	case CryptoSchemePQRingCTX:
		if s.privacyLevel != PRIVACY_LEVEL_FULL_PRIVACY_RAND && s.privacyLevel != PRIVACY_LEVEL_PSEUDONYM {
			return ErrMismatchedCryptoSchemePrivacyLevel
		}
		if s.seedsType != RAND_SEEDS_TYPE && s.seedsType != ROOT_SEEDS_TYPE {
			return ErrMismatchedCryptoSchemePrivacyLevel
		}
		if s.coinSpendKeySeed == nil && s.coinSpendKeySeed.Len() == 0 {
			return ErrCorruptedSeed
		}
		if s.privacyLevel != PRIVACY_LEVEL_PSEUDONYM {
			if s.coinSerialNumberKeySeed == nil && s.coinSerialNumberKeySeed.Len() == 0 {
				return ErrCorruptedSeed
			}
			if s.coinValueKeySeed == nil && s.coinValueKeySeed.Len() == 0 {
				return ErrCorruptedSeed
			}
		}
		if s.coinDetectorKey == nil && s.coinDetectorKey.Len() == 0 {
			return ErrCorruptedSeed
		}
		if s.seedsType == RAND_SEEDS_TYPE {
			if s.publicRand == nil || s.publicRand.Len() == 0 {
				return ErrCorruptedSeed
			}
		}
	default:
		return ErrInvalidCryptoScheme
	}

	return nil
}

func (s *CryptoSeeds) Bytes() (Bytes, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	cryptoSchemeSize := api.CryptoSchemeSerializeSize()
	serializeCryptoScheme := api.SerializeCryptoScheme(s.cryptoScheme)

	underlyingSeedLen, err := api.GetCryptoSchemeParamSeedBytesLen(s.cryptoScheme)
	if err != nil {
		return nil, err
	}

	switch s.cryptoScheme {
	case CryptoSchemePQRingCT:
		if s.privacyLevel != PRIVACY_LEVEL_FULL_PRIVACY_PRE {
			return nil, ErrMismatchedCryptoSchemePrivacyLevel
		}
		addressKeyCryptoSeed := make([]byte, 0, cryptoSchemeSize+2*underlyingSeedLen)
		addressKeyCryptoSeed = append(addressKeyCryptoSeed, serializeCryptoScheme...)
		addressKeyCryptoSeed = append(addressKeyCryptoSeed, s.coinSpendKeySeed...)
		addressKeyCryptoSeed = append(addressKeyCryptoSeed, s.coinValueKeySeed...)
		return AsBytes(addressKeyCryptoSeed), nil
	case CryptoSchemePQRingCTX:
		if s.privacyLevel != PRIVACY_LEVEL_FULL_PRIVACY_RAND && s.privacyLevel != PRIVACY_LEVEL_PSEUDONYM {
			return nil, ErrMismatchedCryptoSchemePrivacyLevel
		}
		expectedSeedLen := 2 * underlyingSeedLen
		if s.privacyLevel == PRIVACY_LEVEL_FULL_PRIVACY_RAND {
			expectedSeedLen = 4 * underlyingSeedLen
		}

		// compute total length
		length := cryptoSchemeSize + 1 + expectedSeedLen
		if s.seedsType == RAND_SEEDS_TYPE {
			publicRandLen, _ := api.GetParamKeyGenPublicRandBytesLen(s.cryptoScheme)
			length += publicRandLen
		}

		addressKeySeed := make([]byte, 0, length)
		addressKeySeed = append(addressKeySeed, serializeCryptoScheme...)
		addressKeySeed = append(addressKeySeed, byte(s.privacyLevel))

		addressKeySeed = append(addressKeySeed, s.coinSpendKeySeed...)
		if s.privacyLevel == PRIVACY_LEVEL_FULL_PRIVACY_RAND {
			addressKeySeed = append(addressKeySeed, s.coinSerialNumberKeySeed...)
			addressKeySeed = append(addressKeySeed, s.coinValueKeySeed...)
		}
		addressKeySeed = append(addressKeySeed, s.coinDetectorKey.Bytes...)

		if s.seedsType == RAND_SEEDS_TYPE {
			addressKeySeed = append(addressKeySeed, s.publicRand...)
		}

		return AsBytes(addressKeySeed), nil

	default:
		return nil, ErrInvalidCryptoScheme
	}
}

func NewCryptoSeed(bytes Bytes) (*CryptoSeeds, error) {
	return deserializeSeed(bytes)
}
func GetCryptoSchemeParamSeedBytesLen(cryptoScheme CryptoScheme) (int, error) {
	return api.GetCryptoSchemeParamSeedBytesLen(cryptoScheme)
}

// GenerateSeed generate a safe seed
// For back-compatibility, hoisting origin implementation here
//
// for CryptoSchemePQRingCT: to keep back-compatibility
//  1. PRIVACY_LEVEL_FULL_PRIVACY_PRE
//     cryptoScheme || coinSpendKeyRootSeed || coinValueKeyRootSeed
//
// for CryptoSchemePQRingCTX:
//  1. PRIVACY_LEVEL_FULL_PRIVACY_RAND
//     cryptoScheme || privacyLevel || coinSpendKeyRootSeed || coinSerialNumberKeyRootSeed || coinValueKeyRootSeed || coinDetectorRootKey
//  2. PRIVACY_LEVEL_PSEUDONYM
//     cryptoScheme || privacyLevel || coinSpendKeyRootSeed || coinDetectorRootKey
func GenerateSeed(cryptoScheme CryptoScheme, privacyLevel PrivacyLevel) (*CryptoSeeds, error) {
	underlyingSeedLen, err := api.GetCryptoSchemeParamSeedBytesLen(cryptoScheme)
	if err != nil {
		return nil, err
	}
	switch cryptoScheme {
	case CryptoSchemePQRingCT:
		if privacyLevel != PRIVACY_LEVEL_FULL_PRIVACY_PRE {
			return nil, fmt.Errorf("invalid privacy level %d for crypto scheme %d", privacyLevel, cryptoScheme)
		}
		coinAddressKeySeed := RandomBytes(underlyingSeedLen)
		coinValueKeySeed := RandomBytes(underlyingSeedLen)
		return NewRandSeeds(cryptoScheme, privacyLevel, coinAddressKeySeed, nil, coinValueKeySeed, nil, nil)
	case CryptoSchemePQRingCTX:
		if privacyLevel != PRIVACY_LEVEL_FULL_PRIVACY_RAND && privacyLevel != PRIVACY_LEVEL_PSEUDONYM {
			return nil, fmt.Errorf("invalid privacy level %d for crypto scheme %d", privacyLevel, cryptoScheme)
		}
		coinSpendKeyRootSeed := RandomBytes(underlyingSeedLen)
		var coinSerialNumberKeyRootSeed, coinValueKeyRootSeed []byte
		if privacyLevel == PRIVACY_LEVEL_FULL_PRIVACY_RAND {
			coinSerialNumberKeyRootSeed = RandomBytes(underlyingSeedLen)
			coinValueKeyRootSeed = RandomBytes(underlyingSeedLen)
		}
		coinDetectorRootKey := NewCryptoKey(RandomBytes(underlyingSeedLen))
		return NewRootSeeds(cryptoScheme, privacyLevel, coinSpendKeyRootSeed, coinSerialNumberKeyRootSeed, coinValueKeyRootSeed, coinDetectorRootKey)

	default:
		return nil, fmt.Errorf("unsupported crypto scheme %v, expected %v", cryptoScheme, CryptoSchemePQRingCT)
	}
}

func deserializeSeed(seed Bytes) (*CryptoSeeds, error) {
	cryptoSchemeSize := api.CryptoSchemeSerializeSize()
	if len(seed) < cryptoSchemeSize {
		return nil, fmt.Errorf("invalid seed seed")
	}
	cryptoScheme, err := api.DeserializeCryptoScheme(seed[:cryptoSchemeSize])
	if err != nil {
		return nil, fmt.Errorf("can not parse the crypto seed")
	}

	underlyingSeedLen, err := api.GetCryptoSchemeParamSeedBytesLen(cryptoScheme)
	if err != nil {
		return nil, err
	}

	var privacyLevel PrivacyLevel
	var coinSpendKeyRootSeed, coinSerialNumberKeyRootSeed, coinValueKeyRootSeed, coinDetectorRootKey []byte
	switch cryptoScheme {
	case CryptoSchemePQRingCT:
		if len(seed) != cryptoSchemeSize+2*underlyingSeedLen {
			return nil, fmt.Errorf("invalid length of crypto seed")
		}
		return NewRandSeeds(
			cryptoScheme, PRIVACY_LEVEL_FULL_PRIVACY_PRE,
			AsBytes(seed[cryptoSchemeSize:cryptoSchemeSize+underlyingSeedLen]),
			nil,
			AsBytes(seed[cryptoSchemeSize+underlyingSeedLen:]),
			nil,
			nil)

	case CryptoSchemePQRingCTX:
		if len(seed) < cryptoSchemeSize+1+2*underlyingSeedLen {
			return nil, fmt.Errorf("invalid length of root seed")
		}
		offset := cryptoSchemeSize

		privacyLevel = PrivacyLevel(seed[offset])
		offset += 1

		if privacyLevel != PRIVACY_LEVEL_FULL_PRIVACY_RAND && privacyLevel != PRIVACY_LEVEL_PSEUDONYM {
			return nil, fmt.Errorf("corrupted crypto seed")
		}

		publicRandLen, _ := api.GetParamKeyGenPublicRandBytesLen(cryptoScheme)
		if privacyLevel == PRIVACY_LEVEL_FULL_PRIVACY_RAND {
			if len(seed) != cryptoSchemeSize+1+4*underlyingSeedLen && len(seed) != cryptoSchemeSize+1+4*underlyingSeedLen+publicRandLen {
				return nil, fmt.Errorf("invalid length of seed")
			}
		}
		if privacyLevel == PRIVACY_LEVEL_PSEUDONYM {
			if len(seed) != cryptoSchemeSize+1+2*underlyingSeedLen && len(seed) != cryptoSchemeSize+1+2*underlyingSeedLen+publicRandLen {
				return nil, fmt.Errorf("invalid length of seed")
			}
		}

		coinSpendKeyRootSeed = seed[offset : offset+underlyingSeedLen]
		offset += underlyingSeedLen
		if privacyLevel == PRIVACY_LEVEL_FULL_PRIVACY_RAND {
			coinSerialNumberKeyRootSeed = seed[offset : offset+underlyingSeedLen]
			offset += underlyingSeedLen
			coinValueKeyRootSeed = seed[offset : offset+underlyingSeedLen]
			offset += underlyingSeedLen
		}
		coinDetectorRootKey = seed[offset : offset+underlyingSeedLen]
		offset += underlyingSeedLen

		if offset < len(seed) {
			if offset+publicRandLen == len(seed) {
				return NewRandSeeds(cryptoScheme, privacyLevel, AsBytes(coinSpendKeyRootSeed), AsBytes(coinSerialNumberKeyRootSeed),
					AsBytes(coinValueKeyRootSeed), AsBytes(coinDetectorRootKey), AsBytes(seed[offset:]))
			}
		}

	default:
		return nil, ErrCorruptedSeed
	}

	return NewRootSeeds(cryptoScheme, privacyLevel, AsBytes(coinSpendKeyRootSeed), AsBytes(coinSerialNumberKeyRootSeed),
		AsBytes(coinValueKeyRootSeed), NewCryptoKey(coinDetectorRootKey))
}

// GenerateCryptoKeysAndAddressByRootSeeds generate an address key pair from root
// NOTE:
// 1. Multiple call use will produce DIFFERENT pairs
// 2. SHOULD call ONLY when crypto scheme is CryptoSchemePQRingCTX
//
// Two ways to generate the same address pair:
// 1. use ExtractPublicRandFromCryptoAddress to extract the public rand from the generated address,
// and then call GenerateCryptoKeysAndAddressByRootSeedsFromPublicRand to generate
// 2. use ExtractPublicRandFromCryptoAddress to extract the public rand from the generated address,
// and with that call GenerateRandSeedsByRootSeedsFromPublicRand to generate the rand seed,
// and then call GenerateCryptoKeysAndAddressByRandSeeds to generate
func GenerateCryptoKeysAndAddressByRootSeeds(rootSeedBytes Bytes) (*CryptoKeysAndAddress, error) {
	rootSeeds, err := deserializeSeed(rootSeedBytes)
	if err != nil {
		return nil, err
	}
	if rootSeeds.seedsType != ROOT_SEEDS_TYPE {
		return nil, AssertError("call GenerateCryptoKeysAndAddressByRootSeeds with invalid seeds")
	}
	cryptoAddress, cryptoSpendSecretKey,
		cryptoSerialNoSecretKey, cryptoViewSecretKey, cryptoDetectorKey, err := api.CryptoAddressKeyGenByRootSeeds(
		rootSeeds.cryptoScheme, rootSeeds.privacyLevel,
		rootSeeds.coinSpendKeySeed, rootSeeds.coinSerialNumberKeySeed,
		rootSeeds.coinValueKeySeed, rootSeeds.coinDetectorKey.Bytes)
	if err != nil {
		return nil, err
	}

	cryptoKeysAndAddress := &CryptoKeysAndAddress{
		SpendSecretKey:    *NewCryptoKey(cryptoSpendSecretKey),
		SerialNoSecretKey: *NewCryptoKey(cryptoSerialNoSecretKey),
		ViewSecretKey:     *NewCryptoKey(cryptoViewSecretKey),
		DetectorKey:       *NewCryptoKey(cryptoDetectorKey),
		CryptoAddress:     *NewCryptoAddress(cryptoAddress),
	}

	return cryptoKeysAndAddress, nil
}

// ExtractPublicRandFromCryptoAddress extract public rand from crypto address
func ExtractPublicRandFromCryptoAddress(cryptoAddress CryptoAddress) (Bytes, error) {
	publicRand, err := api.ExtractPublicRandFromCryptoAddress(cryptoAddress.Data())
	return AsBytes(publicRand), err
}

// GenerateRandSeedsByRootSeedsFromPublicRand generate rand seed with specified root seed and public rand
// NOTE: SHOULD call ONLY when crypto scheme is CryptoSchemePQRingCTX
func GenerateRandSeedsByRootSeedsFromPublicRand(rootSeedBytes Bytes, publicRand Bytes) (*CryptoSeeds, error) {
	rootSeeds, err := deserializeSeed(rootSeedBytes)
	if err != nil {
		return nil, err
	}
	if rootSeeds.seedsType != ROOT_SEEDS_TYPE {
		return nil, AssertError("call GenerateRandSeedsByRootSeedsFromPublicRand with invalid seeds")
	}

	if rootSeeds.cryptoScheme != CryptoSchemePQRingCTX {
		return nil, fmt.Errorf("expected crypto scheme %d, but got %d ", CryptoSchemePQRingCTX, rootSeeds.cryptoScheme)
	}
	if rootSeeds.privacyLevel != PRIVACY_LEVEL_FULL_PRIVACY_RAND && rootSeeds.privacyLevel != PRIVACY_LEVEL_PSEUDONYM {
		return nil, fmt.Errorf("invalid privacy level %d for crypto scheme %d", rootSeeds.privacyLevel, rootSeeds.cryptoScheme)
	}

	coinSpendKeyRandSeed, coinSerialNumberKeyRandSeed,
		coinValueKeyRandSeed, coinDetectorKey, err := api.RandSeedsGenByRootSeedsFromPublicRand(
		rootSeeds.cryptoScheme, rootSeeds.privacyLevel,
		rootSeeds.coinSpendKeySeed, rootSeeds.coinSerialNumberKeySeed,
		rootSeeds.coinValueKeySeed, rootSeeds.coinDetectorKey.Bytes,
		publicRand)
	if err != nil {
		return nil, fmt.Errorf("fail to generate crypto seed from root seed")
	}

	return NewRandSeeds(rootSeeds.cryptoScheme, rootSeeds.privacyLevel,
		coinSpendKeyRandSeed, coinSerialNumberKeyRandSeed,
		coinValueKeyRandSeed, coinDetectorKey,
		publicRand)
}

// GenerateCryptoKeysAndAddressByRandSeeds generate an address key pair from rand seeds
// Different from GenerateCryptoKeysAndAddressByRootSeeds, multiple call use will produce THE SAME pairs
func GenerateCryptoKeysAndAddressByRandSeeds(randSeedBytes Bytes) (*CryptoKeysAndAddress, error) {
	randSeeds, err := deserializeSeed(randSeedBytes)
	if err != nil {
		return nil, err
	}
	if randSeeds.seedsType != RAND_SEEDS_TYPE {
		return nil, AssertError("call GenerateCryptoKeysAndAddressByRandSeeds invalid seeds")
	}

	var coinDetectorKey []byte
	if randSeeds.coinDetectorKey != nil {
		coinDetectorKey = randSeeds.coinDetectorKey.Bytes
	}

	cryptoAddress, cryptoSpendSecretKey, cryptoSerialNoSecretKey,
		cryptoViewSecretKey, cryptoDetectorKey, err := api.CryptoAddressKeyGenByRandSeeds(
		randSeeds.cryptoScheme, randSeeds.privacyLevel,
		randSeeds.coinSpendKeySeed, randSeeds.coinSerialNumberKeySeed,
		randSeeds.coinValueKeySeed, coinDetectorKey,
		randSeeds.publicRand)
	if err != nil {
		return nil, err
	}

	cryptoKeysAndAddress := &CryptoKeysAndAddress{
		SpendSecretKey:    *NewCryptoKey(cryptoSpendSecretKey),
		SerialNoSecretKey: *NewCryptoKey(cryptoSerialNoSecretKey),
		ViewSecretKey:     *NewCryptoKey(cryptoViewSecretKey),
		DetectorKey:       *NewCryptoKey(cryptoDetectorKey),
		CryptoAddress:     *NewCryptoAddress(cryptoAddress),
	}

	return cryptoKeysAndAddress, nil
}

// GenerateCryptoKeysAndAddressByRootSeedsFromPublicRand generate the address key pair from root seed
// and public rand which can extracted from crypto address by calling ExtractPublicRandFromCryptoAddress
func GenerateCryptoKeysAndAddressByRootSeedsFromPublicRand(rootSeedBytes Bytes, publicRand Bytes) (*CryptoKeysAndAddress, error) {
	rootSeeds, err := deserializeSeed(rootSeedBytes)
	if err != nil {
		return nil, err
	}
	if rootSeeds.seedsType != ROOT_SEEDS_TYPE {
		return nil, AssertError("call GenerateCryptoKeysAndAddressByRootSeedsFromPublicRand with invalid seeds")
	}

	var cryptoAddress, cryptoSpendSecretKey, cryptoSerialNoSecretKey, cryptoViewSecretKey, cryptoDetectorKey []byte

	cryptoAddress, cryptoSpendSecretKey, cryptoSerialNoSecretKey,
		cryptoViewSecretKey, cryptoDetectorKey, err = api.CryptoAddressKeyReGenByRootSeedsFromPublicRand(
		rootSeeds.cryptoScheme, rootSeeds.privacyLevel,
		rootSeeds.coinSpendKeySeed, rootSeeds.coinSerialNumberKeySeed,
		rootSeeds.coinValueKeySeed, rootSeeds.coinDetectorKey.Bytes,
		publicRand)
	if err != nil {
		return nil, err
	}

	cryptoKeysAndAddress := &CryptoKeysAndAddress{
		SpendSecretKey:    *NewCryptoKey(cryptoSpendSecretKey),
		SerialNoSecretKey: *NewCryptoKey(cryptoSerialNoSecretKey),
		ViewSecretKey:     *NewCryptoKey(cryptoViewSecretKey),
		DetectorKey:       *NewCryptoKey(cryptoDetectorKey),
		CryptoAddress:     *NewCryptoAddress(cryptoAddress),
	}

	return cryptoKeysAndAddress, nil
}

// DecodeCoinAddressFromSerializedTxOutData extract coin address from serialized transaction output
func DecodeCoinAddressFromSerializedTxOutData(txVersion uint32, txOutData Bytes) (CoinAddress, error) {
	// potentially use the latest transaction version default
	coinAddressData, err := api.ExtractCoinAddressFromSerializedTxOut(txVersion, txOutData)
	if err != nil {
		return nil, err
	}

	return NewCoinAddress(coinAddressData), nil
}

// DecodeValueFromTxOutDataByKeys
func DecodeValueFromTxOutDataByKeys(txOutData Bytes, cryptoAddress *CryptoAddress, cryptoViewSecretKey *CryptoKey) (int64, error) {
	// api.ExtractCoinValueFromSerializedTxOut will clear up the view secret key param.
	// Thus we pass a copy of the view secret key to avoid this side effect.
	viewSecretKeyData := make([]byte, cryptoViewSecretKey.Len())
	copy(viewSecretKeyData, cryptoViewSecretKey.Bytes)

	value, err := api.ExtractCoinValueFromSerializedTxOutByKeys(api.TxVersion, txOutData, cryptoAddress.Data(), viewSecretKeyData)
	if err != nil {
		return -1, err
	}

	return int64(value), nil
}
