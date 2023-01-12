package contractServer

import (
	"log"
	"net/http"
	"os"

	"metachain/pkg/blockchain"
	"metachain/pkg/config"
	"metachain/pkg/logger"
	"metachain/pkg/server/contractServer/api"
	"metachain/pkg/txpool"
)

func RunMetamaskServer(bc blockchain.Blockchains, tp *txpool.Pool, cfg *config.CfgInfo) {
	s := api.NewMetamaskServer(bc, tp, cfg)
	http.HandleFunc("/", s.HandRequest)

	logger.InfoLogger.Println("Running contractServer...", cfg.MetamaskCfg.ListenPort)
	err := http.ListenAndServe(cfg.MetamaskCfg.ListenPort, nil)
	if err != nil {
		log.Println("start fasthttp fail:", err.Error())
		os.Exit(1)
	}
}
