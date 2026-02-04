-- +migrate Up

CREATE TABLE state (
    key VARCHAR(255) PRIMARY KEY,
    value VARCHAR(255) NOT NULL
);

CREATE TABLE transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    tx_hash BYTEA NOT NULL,
    block_number BIGINT NOT NULL,
    log_index INTEGER NOT NULL,
    block_hash BYTEA NOT NULL,
    block_timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL, 

    from_addr BYTEA NOT NULL,
    to_addr BYTEA NOT NULL,
    amount NUMERIC(78, 0) NOT NULL,

    CONSTRAINT unique_tx_log UNIQUE (tx_hash, log_index)
);

CREATE INDEX idx_transfers_block ON transfers(block_number);
CREATE INDEX idx_transfers_from_time ON transfers(from_addr, block_timestamp DESC);
CREATE INDEX idx_transfers_to_time ON transfers(to_addr, block_timestamp DESC);
CREATE INDEX idx_transfers_time ON transfers(block_timestamp DESC);

-- +migrate Down
DROP INDEX IF EXISTS idx_transfers_time;
DROP INDEX IF EXISTS idx_transfers_to_time;
DROP INDEX IF EXISTS idx_transfers_from_time;
DROP INDEX IF EXISTS idx_transfers_block;
DROP TABLE IF EXISTS transfers;
DROP TABLE IF EXISTS state;