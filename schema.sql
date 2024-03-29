
CREATE TABLE IF NOT EXISTS balances (
    "id" INT PRIMARY KEY,
    "balance" DECIMAL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS id_balances_idx ON balances USING HASH(id);

CREATE TABLE IF NOT EXISTS history (
    "id" BIGSERIAL PRIMARY KEY,
    "wallet_id" INT NOT NULL,
    "date" BIGINT NOT NULL,
    "option" TEXT NOT NULL,
    "amount" DECIMAL NOT NULL,
    "description" TEXT NOT NULL,
    FOREIGN KEY (wallet_id) REFERENCES balances(id)
);

CREATE INDEX IF NOT EXISTS wallet_id_history_idx ON history(wallet_id);