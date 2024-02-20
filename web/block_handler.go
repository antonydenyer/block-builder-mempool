package web

import (
	"github.com/antonydenyer/block-builder-mempool/app"
	"github.com/antonydenyer/block-builder-mempool/service"
	"github.com/samber/lo"
	"net/http"

	"github.com/uptrace/bunrouter"
)

type BlockHandler struct {
	app                     *app.App
	blockTransactionService service.BlockTransactionService
}

func NewBlockHandler(app *app.App) *BlockHandler {
	return &BlockHandler{
		app:                     app,
		blockTransactionService: service.NewBlockTransactionsService(app.DB()),
	}
}

func (h *BlockHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	blocks, err := h.blockTransactionService.Get(ctx)
	if err != nil {
		return err
	}

	part := lo.PartitionBy(blocks, func(b service.BlockTransaction) uint64 {
		return b.BlockNumber
	})

	blocksViewModel := lo.Map(part, func(block []service.BlockTransaction, _ int) BlockViewModel {

		missedTransactions := lo.Map(block, func(tx service.BlockTransaction, _ int) TransactionsViewModel {
			return TransactionsViewModel{
				Hash:                   tx.Hash,
				EffectiveGasTip:        tx.EffectiveGasTip,
				TransactionFeeEstimate: tx.TransactionFeeEstimate,
				TransactionGasUsed:     tx.TransactionGasUsed,
			}
		})

		missedGasTotal := lo.SumBy(missedTransactions, func(a TransactionsViewModel) uint64 {
			return a.TransactionGasUsed
		})

		return BlockViewModel{
			BlockNumber:         block[0].BlockNumber,
			ExtraData:           block[0].BlockExtraData,
			GasUsed:             block[0].BlockGasUsed,
			GasLimit:            block[0].BlockGasLimit,
			BlockSpaceRemaining: int64(block[0].BlockGasLimit) - int64(block[0].BlockGasUsed+missedGasTotal),
			MissedTransactions:  missedTransactions,
			MissedPriorityFees: lo.SumBy(missedTransactions, func(a TransactionsViewModel) uint64 {
				return a.TransactionFeeEstimate
			}),
			MissedGasTotal: missedGasTotal,
			MaxPriorityFee: lo.MaxBy(missedTransactions, func(a TransactionsViewModel, b TransactionsViewModel) bool {
				return a.EffectiveGasTip > b.EffectiveGasTip
			}).EffectiveGasTip,
		}
	})

	return bunrouter.JSON(w, bunrouter.H{
		"blocks": blocksViewModel,
	})
}
