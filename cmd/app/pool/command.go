package pool

import (
	"context"
	"fmt"
	"github.com/antonydenyer/block-builder-mempool/app"
	"github.com/antonydenyer/block-builder-mempool/service"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/urfave/cli/v2"
	"log"
	"strings"
	"time"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "pool",
		Description: "monitors the public pool for new transactions and fires them into our private db",
		Action: func(c *cli.Context) error {
			_, application, err := app.StartCLI(c)
			if err != nil {
				return err
			}
			defer application.Stop()

			if err != nil {
				log.Fatal("Failed to start Mempool")
				return err
			}

			newPendingTransactionChan := make(chan common.Hash)
			rpcClient, err := rpc.DialContext(application.Context(), application.Config().Client)
			if err != nil {
				return err
			}

			ethClient := ethclient.NewClient(rpcClient)

			signer := ethTypes.NewLondonSigner(application.Config().ChainID)

			transactionService := service.NewTransactionService(application.DB())

			// Subscribe to new blocks.
			sub, err := rpcClient.EthSubscribe(application.Context(), newPendingTransactionChan, "newPendingTransactions")
			defer sub.Unsubscribe()

			if err != nil {
				fmt.Println(err)
				return err
			}
			for {
				select {
				case err := <-sub.Err():
					log.Fatal(err)
				case tChan := <-newPendingTransactionChan:

					tx, isPending, err := ethClient.TransactionByHash(context.Background(), tChan)

					if err != nil || !isPending {
						continue
					}

					sender, err := ethTypes.Sender(signer, tx)
					if err != nil {
						fmt.Println(err)
					}
					raw, _ := tx.MarshalBinary()

					gasUsed, err := ethClient.EstimateGas(application.Context(),
						ethereum.CallMsg{
							From:       sender,
							To:         tx.To(),
							Value:      tx.Value(),
							Data:       tx.Data(),
							AccessList: tx.AccessList(),
						})

					if err != nil {
						fmt.Println(err)
						continue
					}

					pendingTx := &service.Transaction{
						Hash:                 tx.Hash().Hex(),
						Raw:                  "0x" + common.Bytes2Hex(raw),
						Status:               "PENDING",
						CreatedAt:            time.Now(),
						MaxFeePerGas:         tx.GasFeeCap().Uint64(),
						MaxPriorityFeePerGas: tx.GasTipCap().Uint64(),
						GasUsed:              gasUsed,
						Nonce:                tx.Nonce(),
						From:                 strings.ToLower(sender.Hex()),
						Input:                strings.ToLower(hexutils.BytesToHex(tx.Data())),
					}

					if tx.To() != nil {
						to := strings.ToLower(tx.To().Hex())
						pendingTx.To = &to
					}

					nonce, err := ethClient.NonceAt(application.Context(), sender, nil)

					if err != nil {
						fmt.Println(err)
					}
					_, err = transactionService.UpdateTransactionCount(application.Context(), map[string]uint64{strings.ToLower(sender.Hex()): nonce})
					if err != nil {
						fmt.Println(err)
					}

					insert, err := transactionService.Insert(application.Context(), *pendingTx)
					if err != nil {
						fmt.Println(err)
					}
					if insert == 1 {
						fmt.Printf("%s inserted\n", tx.Hash().Hex())
					}
				}
			}
			return nil
		},
	}
}
