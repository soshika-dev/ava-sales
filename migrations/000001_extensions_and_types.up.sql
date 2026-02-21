CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ticket_status') THEN
        CREATE TYPE ticket_status AS ENUM (
            'NEW',
            'IN_REVIEW',
            'IN_PROGRESS',
            'NEED_MORE_INFO',
            'RESOLVED',
            'REJECTED'
        );
    END IF;
END$$;
