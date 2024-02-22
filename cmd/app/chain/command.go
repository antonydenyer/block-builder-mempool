package chain

import (
	"context"
	"fmt"
	"github.com/antonydenyer/block-builder-mempool/app"
	"github.com/antonydenyer/block-builder-mempool/service"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
	"log"
	"math/big"
	"strings"
	"time"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "chain",
		Description: "monitors the blockchain, updates our internal state and sends txs to builders",
		Action: func(c *cli.Context) error {
			_, application, err := app.StartCLI(c)
			if err != nil {
				return err
			}
			defer application.Stop()

			if err != nil {
				log.Fatal("Failed to start chain monitor")
				return err
			}

			ethClient, err := ethclient.DialContext(application.Context(), application.Config().Client)
			if err != nil {
				log.Fatal(err)
				return err
			}
			signer := ethTypes.NewLondonSigner(application.Config().ChainID)
			transactionService := service.NewTransactionService(application.DB())
			blockTransactionsService := service.NewBlockTransactionsService(application.DB())

			block, err := ethClient.BlockByNumber(application.Context(), nil)

			if err != nil {
				fmt.Print(err)
				return err
			}

			for {
				fmt.Printf("Inserting block %d\n", block.NumberU64())

				txHashes := make([]string, len(block.Transactions()))
				var headerSenderNonce = map[string]uint64{}

				for i := 0; i < len(block.Transactions()); i++ {
					tx := block.Transactions()[i]
					txHashes[i] = strings.ToLower(tx.Hash().Hex())
					sender, _ := signer.Sender(tx)
					headerSenderNonce[strings.ToLower(sender.Hex())] = tx.Nonce() + 1
				}

				_, err = transactionService.UpdateTransactionCount(context.Background(), headerSenderNonce)
				_, err = transactionService.Validated(context.Background(), txHashes)
				if err != nil {
					fmt.Println(err)
				}

				_, err = blockTransactionsService.InsertNextTransactions(context.Background(), block)
				if err != nil {
					fmt.Println(err)
				}

				_, err = blockTransactionsService.UpdateCurrent(context.Background(), block, txHashes)
				if err != nil {
					fmt.Println(err)
				}

				nextBlockNumber := block.NumberU64() + 1
				time.Sleep(time.Until(time.UnixMilli(int64(block.Time() * 1000)).Add(12 * time.Second).Add(6 * time.Second)))

				for block == nil || block.NumberU64() != nextBlockNumber {
					block, err = ethClient.BlockByNumber(application.Context(), new(big.Int).SetUint64(nextBlockNumber))
					if err != nil || block == nil {
						time.Sleep(250 * time.Millisecond)
					}
				}
			}
			return nil
		},
	}
}
