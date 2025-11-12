-- Fix notifications table by removing old admin_id column
-- Run this script to manually fix the notifications table

DO $$ BEGIN
    -- Check if admin_id exists and drop it
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='notifications' AND column_name='admin_id') THEN
        -- First remove NOT NULL constraint
        ALTER TABLE notifications ALTER COLUMN admin_id DROP NOT NULL;
        -- Then drop the column
        ALTER TABLE notifications DROP COLUMN admin_id;
        RAISE NOTICE 'admin_id column dropped successfully';
    ELSE
        RAISE NOTICE 'admin_id column does not exist';
    END IF;

    -- Check if body exists and drop it
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='notifications' AND column_name='body') THEN
        ALTER TABLE notifications DROP COLUMN body;
        RAISE NOTICE 'body column dropped successfully';
    ELSE
        RAISE NOTICE 'body column does not exist';
    END IF;
END $$;
