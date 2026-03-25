DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_type t
        WHERE t.typname = 'parsing_error_type'
    ) THEN
        ALTER TYPE parsing_error_type ADD VALUE IF NOT EXISTS 'empty';
    END IF;
END $$;
