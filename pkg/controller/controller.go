package controller

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"

	"metachain/pkg/block"
	"metachain/pkg/blockchain"
	"metachain/pkg/logger"
	"metachain/pkg/miner"
	"metachain/pkg/storage/miscellaneous"
	"metachain/pkg/transaction"
	"metachain/pkg/txpool"

	"go.uber.org/zap"
)

const (
	TypeTransaction byte = iota
	TypeBlock
	TypeBlockHead
	TypeIPs
)

var txPool *txpool.Pool

func InitController(pool *txpool.Pool) error {
	if txPool != nil {
		return nil
	}

	if pool == nil {
		return fmt.Errorf("pool cannot be nil")
	}
	txPool = pool

	return nil
}

type Controller struct {
	Pool          *txpool.Pool
	Miner         *miner.Miner
	BlockChain    *blockchain.Blockchain
	logger        *zap.Logger
	AdvertiseAddr string
	initBits      uint32
	clientPool    *pool
}

func New(pool *txpool.Pool, blockChain *blockchain.Blockchain, miner *miner.Miner, logger *zap.Logger, AdvertiseAddr string, initBits uint32) (*Controller, error) {

	if pool == nil || blockChain == nil {
		return nil, fmt.Errorf("invalid pool and blockchain")
	}
	con := &Controller{
		Pool:          pool,
		Miner:         miner,
		BlockChain:    blockChain,
		logger:        logger,
		AdvertiseAddr: AdvertiseAddr,
		initBits:      initBits,
		clientPool:    newPool(),
	}
	return con, nil
}

func (c *Controller) tcpServer() {
	lis, err := net.Listen("tcp", ":9998")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			c.logger.Error("tcp accept", zap.Error(err))
		}

		go c.handleConn(conn)
	}

}

func (c *Controller) handleConn(conn net.Conn) {
	defer conn.Close()

	head := make([]byte, 4)
	for {
		_, err := conn.Read(head)
		if err != nil {
			c.logger.Error("read head from conn", zap.String("remote address", conn.RemoteAddr().String()), zap.Error(err))
			return
		}

		l, err := miscellaneous.D32func(head)
		if err != nil {
			c.logger.Error("decode head", zap.Error(err))
		}

		msg := make([]byte, l)

		_, err = conn.Read(msg)
		if err != nil {
			c.logger.Error("read msg from conn", zap.String("remote address", conn.RemoteAddr().String()), zap.Error(err))
			return
		}

		c.handleMessage(msg)
	}
}

func (c *Controller) handleMessage(msg []byte) {
	switch msg[0] {
	case TypeBlock:
		block, err := block.Deserialize(msg[1:])
		if err != nil {
			c.logger.Error("handleMessage", zap.Error(err))
			return
		}

		c.Miner.AcceptBlockFromP2P(block)
	default:
		c.logger.Error("handleMessage", zap.Error(fmt.Errorf("unkonwn message type")))
	}
}

func (c *Controller) RegisterHandleFunc() func([]byte) error {
	f := func(msg []byte) error {
		if len(msg) < 1 {
			return fmt.Errorf("invalid message")
		}

		switch msg[0] {
		case TypeTransaction:
			if err := c.HandleTransactionMessage(msg[1:]); err != nil {
				return err
			}
		case TypeBlockHead:
			if err := c.HandleBlockHeadMessage(msg[1:]); err != nil {
				return err
			}
		case TypeIPs:
			if err := c.HandleIPsMessage(msg[1:]); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown msg type:%v", msg[0])
		}

		return nil
	}

	return f
}

func (c *Controller) HandleTransactionMessage(msg []byte) error {
	st, err := transaction.DeserializeSignaturedTransaction(msg)
	if err != nil {
		return err
	}
	return c.Pool.Add(st)
}

func (c *Controller) HandleBlockHeadMessage(msg []byte) error {
	//TODO:
	var err error

	blockHead := &block.BlockHead{}
	if err = json.Unmarshal(msg, blockHead); err != nil {
		return err
	}

	if blockHead.Host == c.AdvertiseAddr {
		return nil
	}

	h, err := c.BlockChain.GetMaxBlockHeight()
	if err != nil {
		return err
	}

	if blockHead.Height+100 < h {
		return nil
	}

	// if blockHead.InitBits != c.initBits {
	// 	logger.Error("init bits not equal", zap.Uint32("local bits", c.initBits), zap.Uint32("remote bits", blockHead.InitBits))
	// 	return nil
	// }

	if blockHead.GenesisHash != c.Miner.GenesisHash {
		/* logger.Error("GenesisHash not equal", zap.String("remote ip", blockHead.Host), zap.String("hash", hex.EncodeToString(blockHead.Hash)),
		zap.String("local GenesisHash", c.Miner.GenesisHash), zap.String("remote GenesisHash", blockHead.GenesisHash)) */
		return nil
	}

	hash := blockHead.Hash
	url := net.JoinHostPort(blockHead.Host, blockHead.Port)

	logger.Debug("start get block", zap.String("remote ip", blockHead.Host), zap.String("hash", hex.EncodeToString(hash)))

	//var client pb.InsideGreeterClient
	var client *miner.InsideClient
	for {
		if _, err := c.BlockChain.GetBlockByHash(hash); err == nil {
			break
		}

		if b, ok := c.Miner.OrphanBlockIsExist(hash); ok {
			hash = b.PrevHash
			continue
		}

		if client == nil {
			// cc, err := c.clientPool.GetClient(url)
			// if err != nil {
			// 	c.logger.Error("HandleBlockHeadMessage", zap.Error(err))
			// 	return err
			// }

			// client = pb.NewInsideGreeterClient(cc)
			// defer c.clientPool.PutClient(url, cc)

			client, err = miner.NewInsideClient(url, c.BlockChain, c.Pool, c.Miner)
			if err != nil {
				return err
			}

			defer client.Close()
		}

		// resp, err := client.GetBlock(context.Background(), &pb.ReqBlock{Hash: hash})
		// if err != nil {
		// 	c.logger.Error("HandleBlockHeadMessage", zap.String("remote ip", blockHead.Host), zap.String("GetBlockByHash", hex.EncodeToString(hash)), zap.Error(err))
		// 	break
		// }

		// b, err := block.Deserialize(resp.Data)
		// if err != nil {
		// 	c.logger.Error("HandleBlockHeadMessage", zap.Error(err))
		// 	break
		// }

		b, err := client.GetBlockByHash(hash)
		if err != nil {
			return err
		}

		c.Miner.AcceptBlockFromP2P(b)
		hash = b.PrevHash
	}

	return nil
}

func (c *Controller) HandleIPsMessage(msg []byte) error {
	//TODO:
	return nil
}
