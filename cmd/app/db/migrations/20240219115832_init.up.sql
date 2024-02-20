DROP TABLE IF EXISTS transactions;
CREATE TABLE transactions
(
    hash                     varchar     NOT NULL,
    raw                      varchar     NOT NULL,
    status                   varchar     NOT NULL,
    created_at               timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    max_fee_per_gas          numeric     NOT NULL,
    max_priority_fee_per_gas numeric     NOT NULL,
    gas_used                 numeric     NOT NULL,
    nonce                    numeric     NOT NULL,
    "from"                   varchar     NOT NULL,
    "to"                     varchar     NULL,
    input                    varchar     NOT NULL,
    CONSTRAINT transactions_pkey PRIMARY KEY (hash)
);
CREATE INDEX status_idx ON transactions (status);


DROP TABLE IF EXISTS transaction_counts;
CREATE TABLE transaction_counts
(
    address varchar NOT NULL,
    count   numeric NULL,
    CONSTRAINT transaction_counts_pkey PRIMARY KEY (address)
);


DROP TABLE IF EXISTS block_transactions;
CREATE TABLE block_transactions
(
    block_number             numeric NOT NULL,
    hash                     varchar NOT NULL,
    status                   varchar NOT NULL,
    effective_gas_tip        numeric NOT NULL,
    transaction_fee_estimate numeric NOT NULL,
    transaction_gas_used     numeric NOT NULL,
    block_extra_data         varchar NULL,
    block_gas_used           numeric NULL,
    block_gas_limit          numeric NULL,

    CONSTRAINT block_transactions_pkey PRIMARY KEY (block_number, hash)
);
CREATE INDEX block_number_idx ON block_transactions (block_number);
