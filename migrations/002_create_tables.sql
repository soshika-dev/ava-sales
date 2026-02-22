-- +goose Up
CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_number TEXT NOT NULL UNIQUE,
    customer_id UUID NOT NULL,
    device_serial TEXT NOT NULL,
    subject TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NULL,
    status ticket_status NOT NULL,
    assigned_technician_id UUID NULL,
    close_reason TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ticket_attachments_id UUID NULL,
    feedback_score INTEGER NULL,
    closed_at TIMESTAMPTZ NULL,
    CONSTRAINT tickets_close_reason_required
        CHECK (
            status NOT IN ('RESOLVED', 'REJECTED')
            OR (
                close_reason IS NOT NULL
                AND LENGTH(BTRIM(close_reason)) > 0
            )
        )
);

CREATE INDEX IF NOT EXISTS idx_tickets_customer_id ON tickets(customer_id);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);
CREATE INDEX IF NOT EXISTS idx_tickets_assigned_technician_id ON tickets(assigned_technician_id);
CREATE INDEX IF NOT EXISTS idx_tickets_created_at ON tickets(created_at);

CREATE TABLE IF NOT EXISTS ticket_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL CHECK (event_type IN ('STATUS_CHANGED', 'COMMENT', 'ASSIGN', 'CLOSE')),
    from_status ticket_status NULL,
    to_status ticket_status NULL,
    message TEXT NULL,
    actor_role TEXT NOT NULL CHECK (actor_role IN ('CUSTOMER', 'TECHNICIAN', 'ADMIN')),
    actor_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ticket_events_ticket_created_desc ON ticket_events(ticket_id, created_at DESC);

CREATE TABLE IF NOT EXISTS ticket_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    file_type TEXT NOT NULL CHECK (file_type IN ('IMAGE', 'VIDEO')),
    file_url TEXT NOT NULL,
    uploaded_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ticket_attachments_ticket_created_desc ON ticket_attachments(ticket_id, created_at DESC);

CREATE TABLE IF NOT EXISTS ticket_feedback (
    ticket_id UUID PRIMARY KEY REFERENCES tickets(id) ON DELETE CASCADE,
    score INT NOT NULL CHECK (score BETWEEN 1 AND 5),
    solved BOOLEAN NOT NULL,
    comment TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS agencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    province TEXT NOT NULL,
    city TEXT NOT NULL,
    address TEXT NOT NULL,
    phone TEXT NOT NULL,
    latitude NUMERIC NULL,
    longitude NUMERIC NULL,
    version REAL NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agencies_province ON agencies(province);
CREATE INDEX IF NOT EXISTS idx_agencies_city ON agencies(city);
CREATE INDEX IF NOT EXISTS idx_agencies_name ON agencies(name);

-- +goose Down
DROP TABLE IF EXISTS ticket_feedback;
DROP TABLE IF EXISTS ticket_attachments;
DROP TABLE IF EXISTS ticket_events;
DROP TABLE IF EXISTS agencies;
DROP TABLE IF EXISTS tickets;
