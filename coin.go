package core

import "fmt"

// Define the CoinID and Coin data types.
type CoinID struct {
	TxHash Bytes
	Index  uint8
}

type Coin struct {
	ID                CoinID
	OwnerShortAddress *ShortAbelAddress
	OwnerAddress      *AbelAddress
	Value             int64
	SerialNumber      Bytes
	TxVoutData        Bytes
	BlockHash         Bytes
	BlockHeight       int64
}

// Define methods for CoinID.
func NewCoinID(txHash Bytes, index uint8) *CoinID {
	return &CoinID{
		TxHash: txHash,
		Index:  index,
	}
}

func (id CoinID) String() string {
	return fmt.Sprintf("%s:%d", id.TxHash.HexString(), id.Index)
}

// Define util functions.
func NeutrinoToAbel(neutrinoAmount int64) float64 {
	return float64(neutrinoAmount) / 1e7
}

func AbelToNeutrino(abelAmount float64) int64 {
	return int64(abelAmount * 1e7)
}
