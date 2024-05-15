package core

// Define the TxInDesc data type and methods.
type TxInDesc struct {
	TxOutData        Bytes
	CoinValue        int64
	Owner            *ShortAbelAddress
	Height           int64
	TxHash           Bytes
	TxOutIndex       uint8
	CoinSerialNumber Bytes
}

func NewTxInDesc(data Bytes, coinValue ...int64) *TxInDesc {
	if len(coinValue) == 0 {
		coinValue = append(coinValue, -1)
	}

	return &TxInDesc{
		TxOutData: data,
		CoinValue: coinValue[0],
	}
}

func (d *TxInDesc) GetCoinAddress() (*CoinAddress, error) {
	coinAddress, err := DecodeCoinAddressFromTxOutData(d.TxOutData)
	if err != nil {
		return nil, err
	}

	return coinAddress, nil
}

func (d *TxInDesc) GetFingerprint() Bytes {
	coinAddress, err := d.GetCoinAddress()
	if err != nil {
		return nil
	}

	return coinAddress.Fingerprint()
}

// Define the TxOutDesc data type and methods.
type TxOutDesc struct {
	AbelAddress *AbelAddress
	CoinValue   int64
}

func NewTxOutDesc(abelAddress *AbelAddress, coinValue int64) *TxOutDesc {
	return &TxOutDesc{
		AbelAddress: abelAddress,
		CoinValue:   coinValue,
	}
}

// Define the TxBlockDesc data type and methods.
type TxBlockDesc struct {
	BinData Bytes
	Height  int64
}

func NewTxBlockDesc(binData Bytes, height int64) *TxBlockDesc {
	return &TxBlockDesc{
		BinData: binData,
		Height:  height,
	}
}

// Define the TxDesc data type and methods.
type TxDesc struct {
	TxInDescs        []*TxInDesc
	TxOutDescs       []*TxOutDesc
	TxFee            int64
	TxMemo           Bytes
	TxRingBlockDescs map[int64]*TxBlockDesc
}

func NewTxDesc(txInDescs []*TxInDesc, txOutDescs []*TxOutDesc, txFee int64, txRingBlockDescs map[int64]*TxBlockDesc) *TxDesc {
	return &TxDesc{
		TxInDescs:        txInDescs,
		TxOutDescs:       txOutDescs,
		TxFee:            txFee,
		TxRingBlockDescs: txRingBlockDescs,
	}
}
func NewTxDescWithMemo(txInDescs []*TxInDesc, txOutDescs []*TxOutDesc, txFee int64, txRingBlockDescs map[int64]*TxBlockDesc, txMemo Bytes) *TxDesc {
	return &TxDesc{
		TxInDescs:        txInDescs,
		TxOutDescs:       txOutDescs,
		TxFee:            txFee,
		TxMemo:           txMemo,
		TxRingBlockDescs: txRingBlockDescs,
	}
}

// Define the UnsignedRawTx data type and methods.
type UnsignedRawTx struct {
	Bytes
	Signers []*ShortAbelAddress
}

func NewUnsignedRawTx(data Bytes, signers ...[]*ShortAbelAddress) *UnsignedRawTx {
	if len(signers) == 0 {
		signers = append(signers, []*ShortAbelAddress{})
	}

	return &UnsignedRawTx{
		Bytes:   data,
		Signers: signers[0],
	}
}

// Define the SignedRawTx data type and methods.
type SignedRawTx struct {
	Bytes
	Txid Bytes
}

func NewSignedRawTx(data Bytes, txid Bytes) *SignedRawTx {
	return &SignedRawTx{
		Bytes: data,
		Txid:  txid,
	}
}

// Define the TxSubmissionResult data type and methods.
type TxSubmissionResult struct {
	SignedRawTx    *SignedRawTx
	SubmissionTime int64
	Success        bool
	Error          string
}
