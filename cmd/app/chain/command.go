package chain

import (
	"context"
	"fmt"
	"github.com/antonydenyer/block-builder-mempool/app"
	"github.com/antonydenyer/block-builder-mempool/service"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/samber/lo"
	"github.com/urfave/cli/v2"
	"log"
	"math/big"
	"strings"
	"time"
)

var blockHash map[uint64]string

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

			blockHash = make(map[uint64]string)
			if err != nil {
				log.Fatal("Failed to start chain monitor")
				return err
			}

			ctx := application.Context()

			ethClient, err := ethclient.DialContext(ctx, application.Config().Client)
			if err != nil {
				log.Fatal(err)
				return err
			}
			signer := ethTypes.NewLondonSigner(application.Config().ChainID)
			transactionService := service.NewTransactionService(application.DB())
			blockTransactionsService := service.NewBlockTransactionsService(application.DB())

			nextBlock, err := ethClient.BlockByNumber(ctx, nil)

			if err != nil {
				fmt.Print(err)
				return err
			}

			for {
				nextBlock = InsertBlock(ctx, transactionService, blockTransactionsService, ethClient, signer, nextBlock)
				checkForForksOn(ctx, transactionService, blockTransactionsService, ethClient, signer, nextBlock)
			}
			return nil
		},
	}
}

func InsertBlock(ctx context.Context, transactionService service.TransactionService, blockTransactionsService *service.BlockTransactionService, ethClient *ethclient.Client, signer ethTypes.Signer, block *ethTypes.Block) *ethTypes.Block {
	fmt.Printf("Inserting block %d\n", block.NumberU64())
	blockHash[block.NumberU64()] = block.Hash().Hex()

	txHashes := make([]string, len(block.Transactions()))
	var headerSenderNonce = map[string]uint64{}

	for i := 0; i < len(block.Transactions()); i++ {
		tx := block.Transactions()[i]
		txHashes[i] = strings.ToLower(tx.Hash().Hex())
		sender, _ := signer.Sender(tx)
		headerSenderNonce[strings.ToLower(sender.Hex())] = tx.Nonce() + 1
	}

	_, err := transactionService.UpdateTransactionCount(context.Background(), headerSenderNonce)
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
		block, err = ethClient.BlockByNumber(ctx, new(big.Int).SetUint64(nextBlockNumber))
		if err != nil || block == nil {
			time.Sleep(250 * time.Millisecond)
		}
	}
	return block
}

func checkForForksOn(ctx context.Context, transactionService service.TransactionService, blockTransactionService *service.BlockTransactionService, client *ethclient.Client, signer ethTypes.Signer, block *ethTypes.Block) {
	reorged := walk(ctx, client, block.NumberU64(), block.ParentHash(), []*ethTypes.Block{})

	if len(reorged) > 0 {
		senders := lo.Reduce(reorged, func(acc map[string]uint64, b *ethTypes.Block, index int) map[string]uint64 {
			txCounts := lo.FlatMap(b.Transactions(), func(t *ethTypes.Transaction, index int) []lo.Entry[string, uint64] {
				from, _ := signer.Sender(t)
				nonce, _ := client.NonceAt(ctx, from, nil)
				return []lo.Entry[string, uint64]{
					{Key: strings.ToLower(from.Hex()), Value: nonce},
				}
			})

			for _, entry := range txCounts {
				acc[entry.Key] = entry.Value
			}

			return acc
		}, make(map[string]uint64))

		_, _ = transactionService.UpdateTransactionCount(ctx, senders)

		for i := len(reorged) - 1; i >= 0; i-- {
			fmt.Printf("REORG %d", reorged[i].NumberU64())
			InsertBlock(ctx, transactionService, blockTransactionService, client, signer, reorged[i])
		}
	}
}

func walk(ctx context.Context, client *ethclient.Client, blockNumber uint64, remoteParentHash common.Hash, canonical []*ethTypes.Block) []*ethTypes.Block {
	parentBlockNumber := blockNumber - 1
	localParentBlock, ok := blockHash[parentBlockNumber]

	// we've reached the top or have a fresh instance
	if !ok || localParentBlock == "0x0" {
		return canonical
	}

	if remoteParentHash.Hex() != localParentBlock {
		block, _ := client.BlockByNumber(ctx, new(big.Int).SetUint64(parentBlockNumber))
		return walk(ctx, client, parentBlockNumber, block.ParentHash(), append(canonical, block))
	}

	return canonical
}
