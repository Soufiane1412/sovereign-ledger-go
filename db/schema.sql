CREATE TABLE accounts (
    id      int PRIMARY KEY,
    balance int NOT NULL
);

CREATE TABLE pending_holds (
    id         serial PRIMARY KEY,
    account_id int NOT NULL REFERENCES accounts(id),
    amount     int NOT NULL
);

INSERT INTO accounts (id, balance) VALUES (1, 100);