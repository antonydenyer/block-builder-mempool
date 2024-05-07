package service

import (
	"context"
	"github.com/uptrace/bun"
	"time"
)

type TransactionService struct {
	db *bun.DB
}

func NewTransactionService(db *bun.DB) TransactionService {
	return TransactionService{
		db: db,
	}
}

type Transaction struct {
	Hash                 string    `json:"hash"`
	Raw                  string    `json:"raw"`
	Status               string    `json:"status"`
	CreatedAt            time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"createdAt"`
	MaxFeePerGas         uint64    `json:"maxFeePerGas"`
	MaxPriorityFeePerGas uint64    `json:"maxPriorityFeePerGas"`
	GasUsed              uint64    `json:"gasUsed"`
	Nonce                uint64    `json:"nonce"`
	From                 string    `json:"from"`
	To                   *string   `json:"to"`
	Input                string    `json:"input"`
	Type                 *uint8    `json:"type"`
}

type TransactionCount struct {
	Address string
	Count   uint64
}

func (s TransactionService) Insert(ctx context.Context, transaction Transaction) (int64, error) {
	insert, err := s.db.NewInsert().
		Model(&transaction).
		On("CONFLICT (hash) DO NOTHING").
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	return insert.RowsAffected()
}

func (s TransactionService) Validated(ctx context.Context, validatedHashes []string) (int64, error) {

	insert, err := s.db.NewUpdate().
		Model(&Transaction{Status: "VALIDATED"}).
		Column("status").
		Where("hash in (?)", bun.In(validatedHashes)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	return insert.RowsAffected()
}

func (s TransactionService) UpdateTransactionCount(ctx context.Context, headerSenderNonce map[string]uint64) (int64, error) {
	if len(headerSenderNonce) == 0 {
		return 0, nil
	}

	var transactionCounts []TransactionCount

	for from, nonce := range headerSenderNonce {
		var tc = TransactionCount{
			Address: from,
			Count:   nonce,
		}
		transactionCounts = append(transactionCounts, tc)
	}

	result, err := s.
		db.
		NewInsert().
		Model(&transactionCounts).
		On("CONFLICT (address) DO UPDATE").
		Set("count = EXCLUDED.count").
		Exec(ctx)

	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rows, nil
}
