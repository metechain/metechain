package block

import (
	"bytes"
	"fmt"
	"io"
	"math/big"

	"metachain/pkg/storage/miscellaneous"
	"metachain/pkg/transaction"

	"github.com/ethereum/go-ethereum/common"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/crypto/sha3"
)

const (
	Precision = int64(18)
)

// var MinerRewardNumber = big.NewInt(50).

// Block Struct
type Block struct {
	Height       uint64                             `json:"height"`       //当前块号
	PrevHash     []byte                             `json:"prevHash"`     //上一块的hash json:"prevBlockHash --> json:"prevHash
	Hash         []byte                             `json:"hash"`         //当前块hash
	Transactions []*transaction.FinishedTransaction `json:"transactions"` //交易数据
	Root         []byte                             `json:"root"`         //Txhash 默克根
	SnapRoot     []byte                             `json:"snaproot"`     //默克根
	Version      uint64                             `json:"version"`      //版本号
	Timestamp    uint64                             `json:"timestamp"`    //时间戳
	UsedTime     uint64                             `json:"usedtime"`
	Miner        *common.Address                    `json:"miner"` //矿工地址
	//Difficulty       *big.Int                           `json:"difficulty"`       //难度
	GlobalDifficulty *big.Int `json:"globaldifficulty"` // 全网难度
	Nonce            uint64   `json:"nonce"`            //区块nonce
	GasLimit         uint64   `json:"gasLimit"`
	GasUsed          *big.Int `json:"gasUsed"`
}

type BlockHead struct {
	Height      uint64
	Hash        []byte
	Host        string
	Port        string
	GenesisHash string
}

func (b *Block) UnmarshalCBOR(r io.Reader) error {
	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	// height
	{
		maj, val, err := cbg.CborReadHeaderBuf(r, scratch)
		if err != nil {
			return fmt.Errorf("block height:%w", err)
		}

		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("block height:%w", fmt.Errorf("wrong type for uint64 field for height"))
		}

		b.Height = val
	}

	// PreHash
	{
		maj, val, err := cbg.CborReadHeaderBuf(r, scratch)
		if err != nil {
			return fmt.Errorf("prehash %w", err)
		}

		if maj != cbg.MajByteString {
			return fmt.Errorf("expected byte array")
		}

		b.PrevHash = make([]byte, val)
		if _, err := r.Read(b.PrevHash); err != nil {
			return fmt.Errorf("perhash:%w", err)
		}
	}

	// Hash
	{
		maj, val, err := cbg.CborReadHeaderBuf(r, scratch)
		if err != nil {
			return fmt.Errorf("hash :%w", err)
		}

		if maj != cbg.MajByteString {
			return fmt.Errorf("expected byte array")
		}

		b.Hash = make([]byte, val)
		if _, err := r.Read(b.Hash); err != nil {
			return fmt.Errorf("hash :%w", err)
		}
	}

	// transactions
	{
		maj, val, err := cbg.CborReadHeaderBuf(r, scratch)
		if err != nil {
			return fmt.Errorf("transaction :%w", err)
		}

		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field for transactions(len)")
		}
		b.Transactions = make([]*transaction.FinishedTransaction, 0, val)
		for i := uint64(0); i < val; i++ {
			var ft transaction.FinishedTransaction
			if err := ft.UnmarshalCBOR(r); err != nil {
				return err
			}
			b.Transactions = append(b.Transactions, &ft)
		}
	}

	// root
	{
		maj, val, err := cbg.CborReadHeaderBuf(r, scratch)
		if err != nil {
			return fmt.Errorf("root :%w", err)
		}

		if maj != cbg.MajByteString {
			return fmt.Errorf("expected byte array")
		}

		b.Root = make([]byte, val)
		if _, err := r.Read(b.Root); err != nil {
			return err
		}
	}

	// snaproot
	{
		maj, val, err := cbg.CborReadHeaderBuf(r, scratch)
		if err != nil {
			return err
		}

		if maj != cbg.MajByteString {
			return fmt.Errorf("expected byte array")
		}

		b.SnapRoot = make([]byte, val)
		if _, err := r.Read(b.SnapRoot); err != nil {
			return err
		}
	}

	// version
	{
		maj, val, err := cbg.CborReadHeaderBuf(r, scratch)
		if err != nil {
			return err
		}

		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field for version")
		}
		b.Version = val
	}

	// Timestamp
	{
		maj, val, err := cbg.CborReadHeaderBuf(r, scratch)
		if err != nil {
			return err
		}

		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field for  Timestamp")
		}
		b.Timestamp = val
	}

	// UsedTime
	{
		maj, val, err := cbg.CborReadHeaderBuf(r, scratch)
		if err != nil {
			return err
		}

		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field for  UsedTime")
		}
		b.UsedTime = val
	}

	/* 	//Miner
	   	{
	   		b.Miner = address.Address{}
	   		if err := b.Miner.UnmarshalCBOR(r); err != nil {
	   			logger.SugarLogger.Infof("Miner.UnmarshalCBOR error:%v", err)
	   			return err
	   		}
	   	} */

	//miner
	{
		data, err := transaction.UnmarshalByteString(br)
		if err != nil {
			return fmt.Errorf("unmarshal miner address :%w", err)
		}

		if len(data) != 0 {
			miner := common.BytesToAddress(data)
			b.Miner = &miner
		}

	}

	// // Difficulty
	// {
	// 	maj, val, err := cbg.CborReadHeaderBuf(r, scratch)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if maj != cbg.MajByteString {
	// 		return fmt.Errorf("expected byte array")
	// 	}

	// 	bigBytes := make([]byte, val)
	// 	if _, err := r.Read(bigBytes); err != nil {
	// 		return err
	// 	}

	// 	b.Difficulty = big.NewInt(0).SetBytes(bigBytes)
	// }

	// GlobalDifficulty
	{
		maj, val, err := cbg.CborReadHeaderBuf(r, scratch)
		if err != nil {
			return err
		}

		if maj != cbg.MajByteString {
			return fmt.Errorf("expected byte array")
		}

		bigBytes := make([]byte, val)
		if _, err := r.Read(bigBytes); err != nil {
			return err
		}

		b.GlobalDifficulty = big.NewInt(0).SetBytes(bigBytes)
	}

	// nonce
	{
		maj, val, err := cbg.CborReadHeaderBuf(r, scratch)
		if err != nil {
			return err
		}

		if maj != cbg.MajUnsignedInt {
			fmt.Println(5)
			return fmt.Errorf("wrong type for uint64 field for nonce")
		}
		b.Nonce = val
	}

	// gaslimit
	{
		maj, val, err := cbg.CborReadHeaderBuf(r, scratch)
		if err != nil {
			return err
		}

		if maj != cbg.MajUnsignedInt {
			return fmt.Errorf("wrong type for uint64 field for gaslimit")
		}
		b.GasLimit = val
	}

	// gasused
	{
		// maj, val, err := cbg.CborReadHeaderBuf(r, scratch)
		// if err != nil {
		// 	return err
		// }

		// if maj != cbg.MajUnsignedInt {
		// 	return fmt.Errorf("wrong type for uint64 field for  gaslimit")
		// }
		// b.GasUsed = val
		data, err := transaction.UnmarshalByteString(br)
		if err != nil {
			return fmt.Errorf("UnmarshalByteString:%w", err)
		}

		b.GasUsed = new(big.Int).SetBytes(data)
	}

	return nil
}

func (b *Block) MarshalCBOR(w io.Writer) error {
	if b == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	scratch := make([]byte, 9)

	// Height
	{
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, b.Height); err != nil {
			return err
		}
	}

	// PreHash
	{
		if err := cbg.WriteMajorTypeHeader(w, cbg.MajByteString, uint64(len(b.PrevHash))); err != nil {
			return err
		}

		if _, err := w.Write(b.PrevHash); err != nil {
			return err
		}
	}

	// Hash
	{
		if err := cbg.WriteMajorTypeHeader(w, cbg.MajByteString, uint64(len(b.Hash))); err != nil {
			return err
		}

		if _, err := w.Write(b.Hash); err != nil {
			return err
		}
	}

	// transactions
	{
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(len(b.Transactions))); err != nil {
			return err
		}

		for _, tx := range b.Transactions {
			if err := tx.MarshalCBOR(w); err != nil {
				return err
			}
		}
	}

	// root
	{
		if err := cbg.WriteMajorTypeHeader(w, cbg.MajByteString, uint64(len(b.Root))); err != nil {
			return err
		}

		if _, err := w.Write(b.Root); err != nil {
			return err
		}
	}

	// snaproot
	{
		if err := cbg.WriteMajorTypeHeader(w, cbg.MajByteString, uint64(len(b.SnapRoot))); err != nil {
			return err
		}

		if _, err := w.Write(b.SnapRoot); err != nil {
			return err
		}
	}

	// version
	{
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, b.Version); err != nil {
			return err
		}
	}

	// Timestamp
	{
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, b.Timestamp); err != nil {
			return err
		}
	}

	// NewTimestamp
	{
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, b.UsedTime); err != nil {
			return err
		}
	}

	/* 	//Miner
	   	{
	   		if err := b.Miner.MarshalCBOR(w); err != nil {
	   			return err
	   		}
	   	} */

	//miner
	{
		if err := cbg.WriteMajorTypeHeader(w, cbg.MajByteString, uint64(len(b.Miner.Bytes()))); err != nil {
			return err
		}

		if _, err := w.Write(b.Miner.Bytes()); err != nil {
			return err
		}
	}

	// // Difficulty
	// {
	// 	if b.Difficulty == nil {
	// 		return fmt.Errorf("difficulty cannot be nil")
	// 	}

	// 	bigBytes := b.Difficulty.Bytes()
	// 	if err := cbg.WriteMajorTypeHeader(w, cbg.MajByteString, uint64(len(bigBytes))); err != nil {
	// 		return err
	// 	}

	// 	if _, err := w.Write(bigBytes); err != nil {
	// 		return err
	// 	}
	// }

	// GlobalDifficulty
	{
		if b.GlobalDifficulty == nil {
			return fmt.Errorf("global difficulty cannot be nil")
		}

		bigBytes := b.GlobalDifficulty.Bytes()
		if err := cbg.WriteMajorTypeHeader(w, cbg.MajByteString, uint64(len(bigBytes))); err != nil {
			return err
		}

		if _, err := w.Write(bigBytes); err != nil {
			return err
		}
	}

	// nonce
	{
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, b.Nonce); err != nil {
			return err
		}
	}

	// gaslimit
	{
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, b.GasLimit); err != nil {
			return err
		}
	}

	// gasused
	{
		// if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, b.GasUsed); err != nil {
		// 	return err
		// }
		data := b.GasUsed.Bytes()
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(data))); err != nil {
			return err
		}

		if _, err := w.Write(data); err != nil {
			return err
		}
	}

	return nil
}

// Serialize serialization using JSON format
// errUpdated
func (b *Block) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := b.MarshalCBOR(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize deserialization of JSON formatted block data
// errUpdated
func Deserialize(data []byte) (*Block, error) {
	buf := bytes.NewBuffer(data)

	b := &Block{}
	if err := b.UnmarshalCBOR(buf); err != nil {
		fmt.Printf("unmarshal block:%s\n", err)
		return nil, err
	}

	return b, nil
}

//SetHash hash the block data
func (b *Block) SetHash() error {
	data, err := b.Serialize()
	if err != nil {
		return err
	}

	hash := sha3.Sum256(data)
	b.Hash = hash[:]
	return nil
}

//SetHash hash the block data
func (b *Block) MinerHash() []byte {
	heightBytes := miscellaneous.E64func(b.Height)
	var txs [][]byte
	for _, ft := range b.Transactions {
		st := (*ft).SignedTransaction
		st.Input = []byte{}
		data, _ := st.Serialize()
		txs = append(txs, data)
	}

	txsBytes := bytes.Join(txs, []byte{})
	timeBytes := miscellaneous.E64func(b.Timestamp)
	nonce := miscellaneous.E64func(uint64(b.Nonce))
	blockBytes := bytes.Join([][]byte{heightBytes, b.PrevHash, txsBytes, timeBytes, nonce, b.Root}, []byte{})
	hash := sha3.Sum256(blockBytes)

	return hash[:]
}

func (b *Block) Verify() {

}

var GenesisHash = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

func NewGenesisBlock(addr common.Address) *Block {
	return &Block{
		PrevHash:         GenesisHash,
		Hash:             GenesisHash,
		GlobalDifficulty: big.NewInt(0),
		Miner:            &common.Address{},
		// Miner:            &addr,
		// Transactions: []*transaction.FinishedTransaction{
		// 	{
		// 		SignedTransaction: *transaction.GenesisTransaction(&addr),
		// 		GasUsed:           new(big.Int),
		// 	},
		// },
	}
}
