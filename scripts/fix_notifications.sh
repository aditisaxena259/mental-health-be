#!/bin/bash

# Get database connection details from your .env or config
# Modify these values according to your setup

echo "🔧 Fixing notifications table..."
echo "=================================="

# You can either:
# 1. Use psql if you have local postgres
# 2. Connect to your Supabase instance

# For now, let's just create a Go script to do this
cat > /tmp/fix_notifications.go << 'EOF'
package main

import (
	"fmt"
	"log"

	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/initializers"
)

func main() {
	initializers.LoadEnv()
	config.ConnectDB()

	sql := `
	DO $$ BEGIN
		IF EXISTS (SELECT 1 FROM information_schema.columns 
				   WHERE table_name='notifications' AND column_name='admin_id') THEN
			ALTER TABLE notifications ALTER COLUMN admin_id DROP NOT NULL;
			ALTER TABLE notifications DROP COLUMN admin_id;
			RAISE NOTICE 'admin_id column dropped successfully';
		ELSE
			RAISE NOTICE 'admin_id column does not exist';
		END IF;

		IF EXISTS (SELECT 1 FROM information_schema.columns 
				   WHERE table_name='notifications' AND column_name='body') THEN
			ALTER TABLE notifications DROP COLUMN body;
			RAISE NOTICE 'body column dropped successfully';
		ELSE
			RAISE NOTICE 'body column does not exist';
		END IF;
	END $$;
	`

	if err := config.DB.Exec(sql).Error; err != nil {
		log.Fatal("Failed to fix notifications table:", err)
	}

	fmt.Println("✅ Notifications table fixed successfully!")
}
EOF

cd /Users/aditisaxena/mental-health-be
go run /tmp/fix_notifications.go

rm /tmp/fix_notifications.go

echo "=================================="
echo "✅ Done!"
