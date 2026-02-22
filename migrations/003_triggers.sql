-- +goose Up
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_tickets_updated_at ON tickets;
CREATE TRIGGER trg_tickets_updated_at
BEFORE UPDATE ON tickets
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_agencies_updated_at ON agencies;
CREATE TRIGGER trg_agencies_updated_at
BEFORE UPDATE ON agencies
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS trg_tickets_updated_at ON tickets;
DROP TRIGGER IF EXISTS trg_agencies_updated_at ON agencies;
DROP FUNCTION IF EXISTS set_updated_at();
