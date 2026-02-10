ALTER TABLE if exists todos
    ADD COLUMN if not exists description VARCHAR(500),
    ADD COLUMN if not exists deadline    TIMESTAMP;