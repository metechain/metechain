package transaction

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"

	cbg "github.com/whyrusleeping/cbor-gen"
)

type SignedTransaction struct {
	Transaction
	Signature []byte
}

func NewSignedTransaction(tx Transaction, sign []byte) *SignedTransaction {
	return &SignedTransaction{
		Transaction: tx,
		Signature:   sign,
	}
}

func (st *SignedTransaction) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := st.MarshalCBOR(buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DeserializeSignaturedTransaction(data []byte) (*SignedTransaction, error) {
	st := &SignedTransaction{
		Transaction: Transaction{
			Amount: new(big.Int),
		},
	}

	if err := st.UnmarshalCBOR(bytes.NewReader(data)); err != nil {
		return nil, err
	}

	return st, nil
}

func (st *SignedTransaction) MarshalCBOR(w io.Writer) error {
	if st == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	if _, err := w.Write([]byte{130}); err != nil {
		return err
	}

	if err := st.Transaction.MarshalCBOR(w); err != nil {
		return err
	}

	/* 	if err := st.Signature.MarshalCBOR(w); err != nil {
		return err
	} */

	{
		if err := cbg.WriteMajorTypeHeader(w, cbg.MajByteString, uint64(len(st.Signature))); err != nil {
			return err
		}

		if _, err := w.Write(st.Signature); err != nil {
			return err
		}
	}

	return nil
}

func (st *SignedTransaction) UnmarshalCBOR(r io.Reader) error {
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

	// st.Transaction
	if err := st.Transaction.UnmarshalCBOR(r); err != nil {
		return err
	}

	/* 	// st.Signature
	   	if err := st.Signature.UnmarshalCBOR(r); err != nil {
	   		return err
	   	} */

	{
		data, err := UnmarshalByteString(br)
		if err != nil {
			return fmt.Errorf("unmarshal signature:%w", err)
		}

		st.Signature = data
	}

	return nil
}

func (st *SignedTransaction) String() string {
	return fmt.Sprintf("Transaction:{%s} , Hash:%s", st.Transaction.String(), st.HashToString())
}

func (st *SignedTransaction) GetTransaction() Transaction {
	return st.Transaction
}

func (st *SignedTransaction) HashToString() string {
	return hex.EncodeToString(st.Hash())
}
