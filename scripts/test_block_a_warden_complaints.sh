#!/bin/bash
# Script: test_block_a_warden_complaints.sh
# Description: Login as Block A warden and fetch complaints

API_URL="http://localhost:8000"
EMAIL="wardend@hostel.com"
PASSWORD="wardenD123"

echo "Logging in as Block A warden..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"'$EMAIL'","password":"'$PASSWORD'"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d '"' -f4)

if [ -z "$TOKEN" ]; then
  echo "❌ Login failed. Response: $LOGIN_RESPONSE"
  exit 1
fi

echo "✅ Login successful. Token: $TOKEN"
echo "\nFetching complaints as Block A warden..."

COMPLAINTS_RESPONSE=$(curl -s -X GET "$API_URL/api/complaints" \
  -H "Authorization: Bearer $TOKEN")

echo "$COMPLAINTS_RESPONSE" | jq . 2>/dev/null || echo "$COMPLAINTS_RESPONSE"
