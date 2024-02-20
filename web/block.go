package web

type BlockViewModel struct {
	BlockNumber         uint64                  `json:"blockNumber"`
	ExtraData           string                  `json:"extraData"`
	GasUsed             uint64                  `json:"gasUsed"`
	GasLimit            uint64                  `json:"gasLimit"`
	BlockSpaceRemaining int64                   `json:"gasDiff"`
	MissedTransactions  []TransactionsViewModel `json:"missedTransactions"`
	MissedGasTotal      uint64                  `json:"missedGasTotal"`
	MissedPriorityFees  uint64                  `json:"missedPriorityFees"`
	MaxPriorityFee      uint64                  `json:"maxPriorityFee"`
}

type TransactionsViewModel struct {
	Hash                   string `json:"hash"`
	EffectiveGasTip        uint64 `json:"effectiveGasTip"`
	TransactionFeeEstimate uint64 `json:"transactionFeeEstimate"`
	TransactionGasUsed     uint64 `json:"transactionGasUsed"`
}
