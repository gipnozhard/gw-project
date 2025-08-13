CREATE TABLE IF NOT EXISTS exchange_rates (
    id SERIAL PRIMARY KEY,
    currency VARCHAR(3) NOT NULL UNIQUE,
    rate DECIMAL(10, 6) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
                             );

INSERT INTO exchange_rates (currency, rate)
SELECT 'EUR', 0.85
    WHERE NOT EXISTS
    (SELECT 1
    FROM exchange_rates
    WHERE currency = 'EUR');

INSERT INTO exchange_rates (currency, rate)
SELECT 'RUB', 75.0
    WHERE NOT EXISTS
    (SELECT 1
    FROM exchange_rates
    WHERE currency = 'RUB');