ALTER TABLE block_transactions
    ADD COLUMN block_min_priority_fee numeric NOT NULL DEFAULT 0;