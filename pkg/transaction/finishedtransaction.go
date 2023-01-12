package transaction

import (
	"bytes"
	"fmt"
	"io"
	"math/big"

	cbg "github.com/whyrusleeping/cbor-gen"
)

type FinishedTransaction struct {
	SignedTransaction
	GasUsed  *big.Int
	BlockNum uint64
}

func NewFinishedTransaction(s *SignedTransaction, GasUsed *big.Int, blockNum uint64) *FinishedTransaction {

	ft := &FinishedTransaction{
		SignedTransaction: *s,
		GasUsed:           new(big.Int),
		BlockNum:          blockNum,
	}

	if GasUsed != nil {
		ft.GasUsed.Set(GasUsed)
	}

	return ft
}

func (t *FinishedTransaction) GetGasUsed() *big.Int {
	return t.GasUsed
}

func (ft *FinishedTransaction) MarshalCBOR(w io.Writer) error {
	if ft == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	if _, err := w.Write([]byte{130}); err != nil {
		return err
	}

	if err := ft.SignedTransaction.MarshalCBOR(w); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// GasUsed
	{
		gasUsedBytes := ft.GasUsed.Bytes()
		if err := cbg.WriteMajorTypeHeader(w, cbg.MajByteString, uint64(len(gasUsedBytes))); err != nil {
			return err
		}

		if _, err := w.Write(gasUsedBytes); err != nil {
			return err
		}
	}

	// BlockNum
	{
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, ft.BlockNum); err != nil {
			return err
		}
	}
	return nil
}

func (ft *FinishedTransaction) UnmarshalCBOR(r io.Reader) error {
	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)
	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}

	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 2 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// ft.Transaction
	if err := ft.SignedTransaction.UnmarshalCBOR(r); err != nil {
		return err
	}

	// GasUsed
	{
		// maj, extra, err = cbg.CborReadHeader(r)
		// if err != nil {
		// 	return err
		// }

		// if maj != cbg.MajUnsignedInt {
		// 	return fmt.Errorf("wrong type for uint64 field")
		// }

		// ft.GasUsed = extra

		gasUsedBytes, err := UnmarshalByteString(br)
		if err != nil {
			return fmt.Errorf("unmarshalByteString :%w", err)
		}

		ft.GasUsed = new(big.Int).SetBytes(gasUsedBytes)
	}

	// BlockNum
	{
		maj, extra, err := cbg.CborReadHeader(r)
		if err != nil {
			return err
		}

		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field")
		}

		ft.BlockNum = extra
	}

	return nil
}

func (st *FinishedTransaction) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := st.MarshalCBOR(buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DeserializeFinishedTransaction(data []byte) (*FinishedTransaction, error) {
	st := &FinishedTransaction{}

	if err := st.UnmarshalCBOR(bytes.NewReader(data)); err != nil {
		return nil, err
	}

	return st, nil
}
