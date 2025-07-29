-- +goose Up
CREATE TABLE IF NOT EXISTS users
(
    id         int         GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username   text        NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notes
(
    id         int         GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    int         NOT NULL REFERENCES users(id),
    title      text        NOT NULL,
    content    text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

DROP TRIGGER IF EXISTS notes_before_update_trg ON notes;
DROP FUNCTION IF EXISTS tg_set_updated_at;

-- +goose StatementBegin
CREATE FUNCTION tg_set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER notes_before_update_trg BEFORE UPDATE ON notes
    FOR EACH ROW EXECUTE PROCEDURE tg_set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS notes_before_update_trg ON notes;
DROP FUNCTION IF EXISTS tg_set_updated_at;

DROP TABLE IF EXISTS notes;
DROP TABLE IF EXISTS users;