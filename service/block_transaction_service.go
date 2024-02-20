package service

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/uptrace/bun"
	"math/big"
)

type BlockTransactionService struct {
	db *bun.DB
}

func NewBlockTransactionsService(db *bun.DB) BlockTransactionService {
	return BlockTransactionService{
		db: db,
	}
}

type BlockTransaction struct {
	BlockNumber     uint64 `json:"blockNumber"`
	Hash            string `json:"hash"`
	Status          string `json:"status"`
	EffectiveGasTip uint64 `json:"effectiveGasTip"`
	BlockExtraData  string `json:"blockExtraData"`
	BlockGasUsed    uint64 `json:"blockGasUsed"`
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

	insert, err = s.db.NewUpdate().
		Model(&BlockTransaction{
			BlockExtraData: string(block.Extra()),
			BlockGasUsed:   block.GasUsed(),
		}).
		Column("block_extra_data", "block_gas_used").
		Where("block_number = ?", block.NumberU64()).
		Exec(ctx)

	if err != nil {
		return 0, err
	}

	return insert.RowsAffected()
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
