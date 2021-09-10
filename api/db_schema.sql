CREATE TABLE accounts_balances
(
    -- for simplification of API demonstration id has type INT
    id      INT    NOT NULL PRIMARY KEY,
    balance BIGINT NOT NULL CHECK ( balance >= 0 )
)
