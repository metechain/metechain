package transaction

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"
	"sync"

	"metachain/pkg/storage/miscellaneous"

	"github.com/ethereum/go-ethereum/common"
	cbg "github.com/whyrusleeping/cbor-gen"
)

type TransactionType = uint8

// var ZeroAddress = common.Address{}

const (
	TransferTransaction TransactionType = iota
	CoinBaseTransaction
	LockTransaction
	UnlockTransaction
	MortgageTransaction

	BindingAddressTransaction
	EvmContractTransaction
	// PledgeTrasnaction
	EvmMetaTransaction

	IsTokenTransaction
	// PledgeBreakTransaction

	WithdrawToEthTransaction
)

// Transaction
type Transaction struct {
	Version uint64
	Type    TransactionType
	From    *common.Address
	To      *common.Address
	Nonce   uint64

	Amount    *big.Int
	GasLimit  *big.Int
	GasFeeCap *big.Int
	GasPrice  *big.Int

	Input []byte
}

func NewTransaction() *Transaction {
	return &Transaction{
		Amount:    new(big.Int),
		GasLimit:  new(big.Int),
		GasPrice:  new(big.Int),
		GasFeeCap: new(big.Int),
	}
}

// Caller address
func (t *Transaction) Caller() *common.Address {
	return t.From
}

// Receiver address
func (t *Transaction) Receiver() *common.Address {
	return t.To
}

func (t *Transaction) AmountReceived() big.Int {
	return *t.Amount
}

func (t *Transaction) GetFrom() *common.Address {
	return t.From
}

func (t *Transaction) GetTo() *common.Address {
	return t.To
}

func (t *Transaction) GetAmount() big.Int {
	return *t.Amount
}

func (t *Transaction) GetNonce() uint64 {
	return t.Nonce
}

func (t *Transaction) Hash() []byte {
	from := t.From.Bytes()
	to := t.To.Bytes()
	version := miscellaneous.E64func(t.Version)
	//amount := miscellaneous.E64func(t.Amount)
	nonce := miscellaneous.E64func(t.Nonce)
	amount := t.Amount.Bytes()

	gasLimit := t.GasLimit.Bytes()
	gasPrice := t.GasPrice.Bytes()
	gasFeeCap := t.GasFeeCap.Bytes()

	data := bytes.Join([][]byte{from, to, version, amount, nonce, gasLimit, gasPrice, gasFeeCap}, nil)
	hash := sha256.Sum256(data)
	return hash[:]
}

func (t *Transaction) GetInput() []byte {
	return t.Input
}

// SignHash required for signature
func (t *Transaction) SignHash() []byte {
	data, err := t.Serialize()
	if err != nil {
		//TODO: handling errors
		panic(err)
	}
	hash := sha256.Sum256(data)
	return hash[:]
}

// GasCap gas fee upper limit
func (t *Transaction) GasCap() *big.Int {
	return new(big.Int).Mul(t.GasPrice, t.GasLimit)
}

// Serialize transaction in the cbor format
func (t *Transaction) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := t.MarshalCBOR(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DeserializeTransaction deserializes binary data in cbor format into
// transaction, and returns an error if the data format is incorrect
func DeserializeTransaction(data []byte) (*Transaction, error) {
	tx := &Transaction{}
	buf := bytes.NewBuffer(data)
	if err := tx.UnmarshalCBOR(buf); err != nil {
		return nil, err
	}
	return tx, nil
}

func (t *Transaction) String() string {
	//TODO： string
	return fmt.Sprintf("caller:%s , nonce:%d", t.Caller().String(), t.Nonce)
}

func (t *Transaction) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	scratch := make([]byte, 9)
	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, t.Version); err != nil {
		return err
	}

	{
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, 1); err != nil {
			return err
		}

		if _, err := w.Write([]byte{t.Type}); err != nil {
			return err
		}
	}
	/*
		if err := t.From.MarshalCBOR(w); err != nil {
			return err
		}

		if err := t.To.MarshalCBOR(w); err != nil {
			return err
		} */

	//from
	{
		data := t.From.Bytes()
		if err := cbg.WriteMajorTypeHeader(w, cbg.MajByteString, uint64(len(data))); err != nil {
			return err
		}

		if _, err := w.Write(data); err != nil {
			return err
		}
	}

	//to
	{
		data := t.To.Bytes()
		if err := cbg.WriteMajorTypeHeader(w, cbg.MajByteString, uint64(len(data))); err != nil {
			return err
		}

		if _, err := w.Write(data); err != nil {
			return err
		}
	}

	// if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, t.Amount); err != nil {
	// 	return err
	// }
	{
		amountBytes := t.Amount.Bytes()
		if err := cbg.WriteMajorTypeHeader(w, cbg.MajByteString, uint64(len(amountBytes))); err != nil {
			return err
		}

		if _, err := w.Write(amountBytes); err != nil {
			return err
		}
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, t.Nonce); err != nil {
		return err
	}

	// gasfeecap
	{
		data := t.GasFeeCap.Bytes()
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(data))); err != nil {
			return err
		}

		if _, err := w.Write(data); err != nil {
			return err
		}
	}

	// gaslimit
	{
		data := t.GasLimit.Bytes()
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(data))); err != nil {
			return err
		}

		if _, err := w.Write(data); err != nil {
			return err
		}
	}

	// gasPrice
	{
		data := t.GasPrice.Bytes()
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(data))); err != nil {
			return err
		}

		if _, err := w.Write(data); err != nil {
			return err
		}
	}

	// input
	{
		if err := cbg.WriteMajorTypeHeader(w, cbg.MajByteString, uint64(len(t.Input))); err != nil {
			return err
		}

		if _, err := w.Write(t.Input[:]); err != nil {
			return err
		}
	}

	return nil
}

func (t *Transaction) UnmarshalCBOR(r io.Reader) error {
	br := cbg.GetPeeker(r)

	scratch := make([]byte, 8)

	// Version
	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if maj != cbg.MajUnsignedInt {
		return fmt.Errorf("wrong type for uint64 field for Version ")
	}

	t.Version = extra

	{
		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}

		if maj != cbg.MajByteString {
			return fmt.Errorf("expected byte array")
		}

		if extra != 1 {
			return fmt.Errorf("t.Type: byte array length is wrong(%d)", extra)
		}

		var typeBytes = make([]byte, 1)

		if _, err := io.ReadFull(br, typeBytes[:]); err != nil {
			return err
		}

		t.Type = typeBytes[0]
	}

	/* 	// From
	   	{
	   		if err := t.From.UnmarshalCBOR(br); err != nil {
	   			return err
	   		}
	   	}

	   	// To
	   	{
	   		if err := t.To.UnmarshalCBOR(br); err != nil {
	   			return err
	   		}
	   	} */

	// from
	{
		data, err := UnmarshalByteString(br)
		if err != nil {
			return fmt.Errorf("unmarshal from address data:%w", err)
		}

		from := common.BytesToAddress(data)
		t.From = &from
	}

	// to
	{
		{
			data, err := UnmarshalByteString(br)
			if err != nil {
				return fmt.Errorf("unmarshal from address data:%w", err)
			}

			to := common.BytesToAddress(data)
			t.To = &to
		}
	}

	// Amount
	{
		data, err := UnmarshalByteString(br)
		if err != nil {
			return fmt.Errorf("unmarshalByteString:%w", err)
		}

		t.Amount = new(big.Int).SetBytes(data)
	}

	// Nonce
	{
		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}

		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field for Nonce")
		}

		t.Nonce = uint64(extra)
	}

	// gasfeecap
	{
		data, err := UnmarshalByteString(br)
		if err != nil {
			return fmt.Errorf("unmarshalByteString:%w", err)
		}

		t.GasFeeCap = new(big.Int).SetBytes(data)

	}

	// gaslimit
	{
		data, err := UnmarshalByteString(br)
		if err != nil {
			return fmt.Errorf("unmarshalByteString:%w", err)
		}

		t.GasLimit = new(big.Int).SetBytes(data)
	}

	// GasPrice
	{
		data, err := UnmarshalByteString(br)
		if err != nil {
			return fmt.Errorf("unmarshalByteString:%w", err)
		}
		t.GasPrice = new(big.Int).SetBytes(data)
	}

	// input
	{
		maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}

		if extra > cbg.ByteArrayMaxLen {
			return fmt.Errorf("t.Input: byte array too large (%d)", extra)
		}

		if maj != cbg.MajByteString {
			return fmt.Errorf("expected byte array")
		}

		if extra > 0 {
			t.Input = make([]byte, extra)
		}

		if _, err := io.ReadFull(br, t.Input[:]); err != nil {
			return err
		}
	}

	return nil
}

func UnmarshalByteString(br cbg.BytePeeker) ([]byte, error) {
	scratch := make([]byte, 8)
	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return nil, err
	}

	if maj != cbg.MajByteString {
		return nil, fmt.Errorf("expected byte array")
	}

	if extra > cbg.ByteArrayMaxLen {
		return nil, fmt.Errorf("byte array too large (%d)", extra)
	}

	var data []byte
	if extra > 0 {
		data = make([]byte, extra)
	}

	if _, err := io.ReadFull(br, data[:]); err != nil {
		return nil, err
	}

	return data, nil
}

// errUpdate
func (tx *Transaction) IsTransferTrasnaction() bool {
	return tx.Type == TransferTransaction
}

// errUpdate
func (tx *Transaction) IsLockTransaction() bool {
	return tx.Type == LockTransaction
}

// errUpdate
func (tx *Transaction) IsUnlockTransaction() bool {
	return tx.Type == UnlockTransaction
}

// errUpdate
func (tx *Transaction) IsCoinBaseTransaction() bool {
	return tx.Type == CoinBaseTransaction
}

func (tx *Transaction) IsTokenTransaction() bool {
	return tx.Type == IsTokenTransaction
}

var genesisAmount = big.NewInt(500 * 10000)         // 创世的交易的数量
var UnitPrecision = big.NewInt(1000000000000000000) // 精度单位 == 1 * 精度
var genesisiTransactionOnce sync.Once

func GenesisTransaction(addr *common.Address) *SignedTransaction {
	var genTransction SignedTransaction
	genesisiTransactionOnce.Do(func() {
		//genAddr, _ := address.NewAddrFromString(address.GenesisAddress)
		amount := big.NewInt(0)
		// amount.Mul(amount, genesisAmount).Mul(amount, UnitPrecision)
		genTransction = SignedTransaction{
			Transaction: Transaction{
				From:   &common.Address{},
				To:     addr,
				Nonce:  0,
				Amount: amount,
				Type:   CoinBaseTransaction,

				GasLimit:  new(big.Int),
				GasFeeCap: new(big.Int),
				GasPrice:  new(big.Int),
			},
		}
	})

	return &genTransction
}
