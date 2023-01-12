package transaction

import (
	"bytes"
	"errors"
	"fmt"

	_ "metachain/pkg/crypto/sigs/ed25519"
	_ "metachain/pkg/crypto/sigs/secp"

	"github.com/ethereum/go-ethereum/crypto"
)

func (st *SignedTransaction) VerifySign() error {
	switch st.Type {
	// case TransferTransaction, PledgeTrasnaction, PledgeBreakTransaction, WithdrawToEthTransaction:
	case TransferTransaction, WithdrawToEthTransaction:

		// if err := sigs.Verify(st.Signature, st.Caller().Bytes(), st.SignHash()); err != nil {
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		sigPub, err := crypto.SigToPub(st.SignHash(), st.Signature)
		if err != nil {
			return err
		}

		sigAdde := crypto.PubkeyToAddress(*sigPub)
		if !bytes.Equal(sigAdde.Bytes(), st.From.Bytes()) {
			return fmt.Errorf("signature verification failed")
		}
	case EvmContractTransaction, EvmMetaTransaction:
		evm, err := DecodeEvmData(st.Input)
		if err != nil {
			return err
		}
		if !VerifyEthSign(evm.EthData) {
			return errors.New("verify eth transaction failed!")
		}
	}

	return nil
}
