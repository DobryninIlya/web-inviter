-- +goose Up
CREATE TABLE IF NOT EXISTS payment_transactions
(
    uid TEXT NOT NULL,
    ext_id TEXT,
    client_id TEXT,
    date TEXT DEFAULT CURRENT_DATE,
    type TEXT,
    ended INTEGER DEFAULT 0,
    PRIMARY KEY (ext_id)
);

CREATE TABLE IF NOT EXISTS premium_subscribe
(
    uid TEXT NOT NULL,
    date_of_last_payment TEXT,
    total_payed INTEGER,
    link	TEXT,
    CONSTRAINT premium_subscribe_pkey PRIMARY KEY (uid)
);

-- +goose Down

DROP TABLE IF EXISTS payment_transactions;

DROP TABLE IF EXISTS premium_subscribe;
