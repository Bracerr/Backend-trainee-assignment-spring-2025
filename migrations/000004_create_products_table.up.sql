CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    date_time TIMESTAMP NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь')),
    reception_id UUID NOT NULL REFERENCES receptions(id)
); 