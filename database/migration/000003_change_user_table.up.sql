Alter table if  exists users
    ADD COLUMN if not exists updated_at TIMESTAMP WITH TIME ZONE;


CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique_idx
    ON users (LOWER(email));

ALTER TYPE status ADD VALUE 'Pending';