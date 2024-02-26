package service

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
	"math/big"
)

type BlockTransactionService struct {
	db *bun.DB
}

func NewBlockTransactionsService(db *bun.DB) *BlockTransactionService {
	return &BlockTransactionService{
		db: db,
	}
}

type BlockTransaction struct {
	BlockNumber            uint64
	Hash                   string
	Status                 string
	EffectiveGasTip        uint64
	TransactionGasUsed     uint64
	TransactionFeeEstimate uint64
	BlockExtraData         string
	BlockGasUsed           uint64
	BlockGasLimit          uint64
	BlockMinPriorityFee    uint64
}

type BlockViewModel struct {
	BlockNumber         uint64
	ExtraData           string
	GasUsed             uint64
	GasLimit            uint64
	MinPriorityFee      float64
	BlockSpaceRemaining int64
	PercentageUsed      float64
	MissedTransactions  []TransactionsViewModel
	MissedGasTotal      uint64
	MissedPriorityFees  float64
	MaxPriorityFee      float64
}

type TransactionsViewModel struct {
	Hash                   string  `json:"hash"`
	EffectiveGasTip        float64 `json:"effectiveGasTip"`
	TransactionFeeEstimate float64 `json:"transactionFeeEstimate"`
	TransactionGasUsed     uint64  `json:"transactionGasUsed"`
}

func (s BlockTransactionService) InsertNextTransactions(_ context.Context, block *ethTypes.Block) (int64, error) {

	nextBaseFee := CalcNextBaseFee(
		new(big.Int).SetUint64(block.GasUsed()),
		block.BaseFee(),
		new(big.Int).SetUint64(block.GasLimit()),
	)

	res, err := s.db.Exec(""+
		"INSERT INTO block_transactions (block_number, hash, status, effective_gas_tip, transaction_fee_estimate, transaction_gas_used) "+
		"SELECT "+
		"?0 as block_number, "+
		"hash, "+
		"'PENDING' as status, "+
		"CASE "+
		"	WHEN max_fee_per_gas - ?1 < max_priority_fee_per_gas "+
		"	THEN max_fee_per_gas - ?1 "+
		"	ELSE max_priority_fee_per_gas "+
		"END AS effective_gas_tip, "+
		"CASE "+
		"	WHEN max_fee_per_gas - ?1 < max_priority_fee_per_gas "+
		"	THEN (max_fee_per_gas - ?1) * t.gas_used "+
		"	ELSE max_priority_fee_per_gas * t.gas_used "+
		"END AS transaction_fee_estimate, "+
		"gas_used as transaction_gas_used "+
		"FROM transactions t "+
		"JOIN transaction_counts tc on tc.address = t.\"from\" AND tc.count = t.nonce "+
		"WHERE "+
		"status = 'PENDING' "+
		"and "+
		"max_fee_per_gas > ?1 "+
		"and "+
		"created_at > NOW() - INTERVAL '1 hour' "+
		"ON CONFLICT (block_number, hash) DO NOTHING", block.NumberU64()+1, nextBaseFee.Uint64())

	if err != nil {
		return 0, err
	}

	rows, _ := res.RowsAffected()
	return rows, nil
}

func (s BlockTransactionService) UpdateCurrent(ctx context.Context, block *ethTypes.Block, validatedHashes []string) (int64, error) {

	insert, err := s.db.NewUpdate().
		Model(&BlockTransaction{
			Status: "VALIDATED",
		}).
		Column("status").
		Where("block_number = ? AND hash in (?)", block.NumberU64(), bun.In(validatedHashes)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	positivePriorityFee := lo.Filter(block.Transactions(), func(item *ethTypes.Transaction, index int) bool {
		return item.GasTipCap() != nil && item.GasTipCap().Uint64() > 0
	})

	minPriorityFee := lo.MinBy(positivePriorityFee, func(a *ethTypes.Transaction, b *ethTypes.Transaction) bool {
		return a.GasTipCap().Uint64() < b.GasTipCap().Uint64()
	})

	insert, err = s.db.NewUpdate().
		Model(&BlockTransaction{
			BlockExtraData:      string(block.Extra()),
			BlockGasUsed:        block.GasUsed(),
			BlockGasLimit:       block.GasLimit(),
			BlockMinPriorityFee: minPriorityFee.GasTipCap().Uint64(),
		}).
		Column("block_extra_data", "block_gas_used", "block_gas_limit", "block_min_priority_fee").
		Where("block_number = ?", block.NumberU64()).
		Exec(ctx)

	if err != nil {
		return 0, err
	}

	return insert.RowsAffected()
}

func (s BlockTransactionService) Get(ctx context.Context) ([]*BlockViewModel, error) {
	var blocks []BlockTransaction
	err := s.db.
		NewSelect().
		Model(&blocks).
		Where("status = 'PENDING' AND block_gas_used IS NOT NULL").
		Limit(1000).
		Order("block_number DESC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	part := lo.PartitionBy(blocks, func(b BlockTransaction) uint64 {
		return b.BlockNumber
	})

	blocksViewModel := lo.Map(part, func(block []BlockTransaction, _ int) *BlockViewModel {
		return MapBlockTransaction(block)
	})

	return lo.Filter(blocksViewModel, func(item *BlockViewModel, index int) bool {
		return item.BlockSpaceRemaining > 0
	}), nil
}
func (s BlockTransactionService) GetByNumber(ctx context.Context, blockNumber uint64) (*BlockViewModel, error) {
	var blocks []BlockTransaction
	err := s.db.
		NewSelect().
		Model(&blocks).
		Where("block_number = ? AND status = 'PENDING' AND block_gas_used IS NOT NULL", blockNumber).
		Order("block_number DESC").
		Scan(ctx)

	if blocks == nil {
		return nil, fmt.Errorf("not found or pending")
	}
	if err != nil {
		return nil, err
	}

	return MapBlockTransaction(blocks), nil
}

func MapBlockTransaction(block []BlockTransaction) *BlockViewModel {
	missedTransactions := lo.Map(block, func(tx BlockTransaction, _ int) TransactionsViewModel {
		return TransactionsViewModel{
			Hash:                   tx.Hash,
			EffectiveGasTip:        float64(tx.EffectiveGasTip) / 1000 / 1000 / 1000,
			TransactionFeeEstimate: float64(tx.TransactionFeeEstimate) / 1000 / 1000 / 1000,
			TransactionGasUsed:     tx.TransactionGasUsed,
		}
	})

	missedGasTotal := lo.SumBy(missedTransactions, func(a TransactionsViewModel) uint64 {
		return a.TransactionGasUsed
	})
	maxPriorityFee := lo.MaxBy(missedTransactions, func(a TransactionsViewModel, b TransactionsViewModel) bool {
		return a.EffectiveGasTip > b.EffectiveGasTip
	}).EffectiveGasTip

	missedPriorityFees := lo.SumBy(missedTransactions, func(a TransactionsViewModel) float64 {
		return a.TransactionFeeEstimate
	})
	return &BlockViewModel{
		BlockNumber:         block[0].BlockNumber,
		ExtraData:           block[0].BlockExtraData,
		GasUsed:             block[0].BlockGasUsed,
		GasLimit:            block[0].BlockGasLimit,
		MinPriorityFee:      float64(block[0].BlockMinPriorityFee) / 1000 / 1000 / 1000,
		PercentageUsed:      float64(block[0].BlockGasUsed) / float64(block[0].BlockGasLimit) * 100,
		BlockSpaceRemaining: int64(block[0].BlockGasLimit) - int64(block[0].BlockGasUsed+missedGasTotal),
		MissedTransactions:  missedTransactions,
		MissedPriorityFees:  missedPriorityFees,
		MissedGasTotal:      missedGasTotal,
		MaxPriorityFee:      maxPriorityFee,
	}
}

func CalcNextBaseFee(parentGasUsed, parentBaseFee, parentGasLimit *big.Int) *big.Int {
	var (
		denom = new(big.Int).SetUint64(params.DefaultBaseFeeChangeDenominator)
	)

	parentGasTarget := new(big.Int).Div(parentGasLimit, big.NewInt(2))
	// If the parent gasUsed is the same as the target, the baseFee remains unchanged.
	if parentGasUsed == parentGasTarget {
		return parentBaseFee
	}

	if parentGasUsed.Cmp(parentGasTarget) >= 1 {
		// If the parent block used more gas than its target, the baseFee should increase.
		// max(1, parentBaseFee * gasUsedDelta / parentGasTarget / baseFeeChangeDenominator)
		num := new(big.Int).Sub(parentGasUsed, parentGasTarget)
		num.Mul(num, parentBaseFee)
		num.Div(num, parentGasTarget)
		num.Div(num, denom)
		baseFeeDelta := math.BigMax(num, common.Big1)

		num.Add(parentBaseFee, baseFeeDelta)

		return num
	}

	// Otherwise if the parent block used less gas than its target, the baseFee should decrease.
	// max(0, parentBaseFee * gasUsedDelta / parentGasTarget / baseFeeChangeDenominator)
	num := new(big.Int).Sub(parentGasTarget, parentGasUsed)
	num.Mul(num, parentBaseFee)
	num.Div(num, parentGasTarget)
	num.Div(num, denom)
	baseFee := num.Sub(parentBaseFee, num)

	return math.BigMax(baseFee, common.Big0)
}
