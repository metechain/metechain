package address

/* package address

import (
	"fmt"
	"io"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	cbg "github.com/whyrusleeping/cbor-gen"
)

// Address
type Address struct {
	//string // string(public key)
	*common.Address
}

// NetWork
type NetWork = byte

const (
	Mainnet NetWork = iota
	Testnet
)

const (
	ED25519PubKeySize    = 32
	Secp256k1PayloadSize = 21
)

const (
	AddressSize        = 44
	AddresPrefixSize   = 3
	SecpAddressSzie    = 44
	Ed25519AddressSzie = 47
)

const (
	MainnetPrefix = "mtc"
	TestnetPrefix = "ctm"
)

const (
	UndefAddressString = ""
)

var GenesisAddress = "otK00000000000000000000000000000000000000000"

var Undef = Address{}
var CurrentNetWork = Testnet

var prefixSet = map[byte]string{
	Mainnet: MainnetPrefix,
	Testnet: TestnetPrefix,
}

func SetNetWork(networkType string) {
	nwt := strings.ToLower(networkType)
	switch nwt {
	case "mainnet":
		CurrentNetWork = Mainnet
	case "testnet":
		CurrentNetWork = Testnet
	default:
		CurrentNetWork = Testnet
	}

	GenesisAddress = prefixSet[CurrentNetWork] + GenesisAddress[3:]
}

func NewAddr(pub []byte) (Address, error) {
	secpPub, err := ethcrypto.UnmarshalPubkey(pub)
	if err != nil {
		return Undef, err
	}
	ethAddr := ethcrypto.PubkeyToAddress(*secpPub)
	return NewFromEthAddr(ethAddr)
}

func NewAddrFromString(str string) (Address, error) {
	data := common.HexToAddress(str)
	return Address{&data}, nil
}

func NewFromBytes(addr []byte) (Address, error) {
	data := common.BytesToAddress(addr)
	return Address{&data}, nil
}

func NewFromEthAddr(eaddr common.Address) (Address, error) {

	return Address{&eaddr}, nil
}

func (a *Address) MarshalCBOR(w io.Writer) error {
	if a == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	if err := cbg.WriteMajorTypeHeader(w, cbg.MajByteString, uint64(len(a.Bytes()))); err != nil {
		return err
	}

	if _, err := io.WriteString(w, string(a.Bytes())); err != nil {
		return err
	}

	return nil
}

func (a *Address) UnmarshalCBOR(r io.Reader) error {
	br := cbg.GetPeeker(r)

	maj, extra, err := cbg.CborReadHeader(br)
	if err != nil {
		return err
	}

	if maj != cbg.MajByteString {
		return fmt.Errorf("cbor type for address unmarshal was not byte string")
	}

	buf := make([]byte, extra)
	if _, err := io.ReadFull(br, buf); err != nil {
		return err
	}

	addr := common.BytesToAddress(buf)
	a.Address = &addr
	return nil
}
*/
