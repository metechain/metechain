package transaction

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"

	_ "metachain/pkg/crypto/sigs/ed25519"

	"github.com/stretchr/testify/assert"
)

func TestCodec(t *testing.T) {
	assert := assert.New(t)
	var max = ^uint64(0)

	fromKey, _ := crypto.GenerateKey()
	toKey, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(fromKey.PublicKey)
	to := crypto.PubkeyToAddress(toKey.PublicKey)
	t.Logf("from priv:%s", hex.EncodeToString(crypto.FromECDSA(fromKey)))
	t.Logf("from addr:%s\n", from.String())
	iput := make([]byte, 20)
	amount := big.NewInt(100)
	amount.Mul(amount, UnitPrecision)
	tx := &Transaction{
		Version: 1,
		From:    &from,
		To:      &to,
		Amount:  amount,
		Nonce:   max,
		Type:    TransferTransaction,

		GasLimit:  new(big.Int).SetInt64(1100),
		GasFeeCap: new(big.Int).SetInt64(1101),
		GasPrice:  new(big.Int).SetInt64(1102),
		Input:     iput,
	}

	// crypto.Sign(tx.SignHash(), formKey)
	// siganature, err := sigs.Sign(crypto.TypeSecp256k1, fromPriv, tx.SignHash())

	siganature, err := crypto.Sign(tx.SignHash(), fromKey)
	assert.NoError(err)

	st := &SignedTransaction{
		Transaction: *tx,
		Signature:   siganature,
	}

	assert.NoError(st.VerifySign())

	buf, err := st.Serialize()
	assert.NoError(err)

	maybeTx, err := DeserializeSignaturedTransaction(buf)
	assert.NoError(err)

	//assert.Equal(st, maybeTx)
	assert.Equal(st.Amount.String(), maybeTx.Amount.String())
	assert.Equal(st.GasFeeCap.String(), maybeTx.GasFeeCap.String())
	assert.Equal(st.GasPrice.String(), maybeTx.GasPrice.String())
	assert.Equal(st.GasLimit.String(), maybeTx.GasLimit.String())

	// ft := &FinishedTransaction{*st, 10, 0}
	ft := NewFinishedTransaction(st, big.NewInt(10), 0)

	ftBytes, err := ft.Serialize()
	assert.NoError(err)

	mayFt, err := DeserializeFinishedTransaction(ftBytes)
	assert.NoError(err)

	t.Log("ft length:", len(ftBytes))

	// assert.Equal(mayFt, ft)
	assert.Equal(mayFt.Amount.String(), ft.Amount.String())
	assert.Equal(mayFt.GasFeeCap.String(), ft.GasFeeCap.String())
	assert.Equal(mayFt.GasPrice.String(), ft.GasPrice.String())
	assert.Equal(mayFt.GasLimit.String(), ft.GasLimit.String())
}
