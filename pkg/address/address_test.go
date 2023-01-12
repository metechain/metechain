package address

/*
import (
	"bytes"
	"testing"

	"metachain/pkg/crypto"
	"metachain/pkg/crypto/sigs"
	_ "metachain/pkg/crypto/sigs/ed25519"
	_ "metachain/pkg/crypto/sigs/secp"

	"github.com/stretchr/testify/assert"
)

func TestTransactionCBOR(t *testing.T) {
	assert := assert.New(t)
	a, _ := NewFromBytes([]byte{224, 116, 53, 72, 47, 178, 42, 166, 231, 150, 128, 178, 181, 240, 198, 37, 204, 23, 29, 220, 79, 134, 85, 155, 225, 181, 80, 76, 255, 153, 249, 54})
	buf := bytes.NewBuffer(nil)

	err := a.MarshalCBOR(buf)
	assert.NoError(err)

	maybeAddr := new(Address)
	err = maybeAddr.UnmarshalCBOR(buf)
	assert.NoError(err)

	assert.Equal(a, *maybeAddr)
}

func TestGenesisAddress(t *testing.T) {
	assert := assert.New(t)
	priv, err := sigs.Generate(crypto.TypeSecp256k1)
	assert.NoError(err)

	pub, err := sigs.ToPublic(crypto.TypeSecp256k1, priv)
	assert.NoError(err)

	addr, err := NewAddr(pub)
	assert.NoError(err)

	t.Log("addr:", addr.String())
	t.Log("length:", len(addr.String()))
	genAddr := "otK00000000000000000000000000000000000000000"
	// otK00000000000000000000000000000000000000000
	// otK00000000000000000000000000000000000000000
	addr, err = NewAddrFromString(genAddr)
	assert.NoError(err)

	t.Log(addr.String())

}

func TestZreoAddress(t *testing.T) {
	assert := assert.New(t)

	str := ZeroAddress.String()

	maybeZero, err := NewAddrFromString(str)
	assert.NoError(err)

	assert.Equal(ZeroAddress, maybeZero)
	t.Log(ZeroAddress.Hex())

}
*/
