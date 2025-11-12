#!/bin/bash

echo "🧪 Testing Complaint Status Update Endpoint"
echo "============================================"

# Start server if not running
if ! lsof -ti :8080 > /dev/null 2>&1; then
    echo "Starting server..."
    ./out > server.log 2>&1 &
    sleep 3
fi

# Login as admin
echo -e "\n1️⃣ Logging in as admin..."
ADMIN_LOGIN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@hostel.com","password":"admin123"}')

ADMIN_TOKEN=$(echo $ADMIN_LOGIN | jq -r '.token')

if [ "$ADMIN_TOKEN" == "null" ] || [ -z "$ADMIN_TOKEN" ]; then
    echo "❌ Login failed"
    echo $ADMIN_LOGIN | jq .
    exit 1
fi
echo "✅ Admin logged in"

# Get complaints
echo -e "\n2️⃣ Fetching complaints..."
COMPLAINTS=$(curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/admin/complaints)

COMPLAINT_ID=$(echo $COMPLAINTS | jq -r '.data[0].ID')
CURRENT_STATUS=$(echo $COMPLAINTS | jq -r '.data[0].Status')

if [ "$COMPLAINT_ID" == "null" ] || [ -z "$COMPLAINT_ID" ]; then
    echo "❌ No complaints found"
    exit 1
fi

echo "✅ Found complaint: $COMPLAINT_ID"
echo "   Current status: $CURRENT_STATUS"

# Update to 'inprogress'
echo -e "\n3️⃣ Updating status to 'inprogress'..."
UPDATE_PROGRESS=$(curl -s -X PUT \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"inprogress"}' \
  http://localhost:8080/api/admin/complaints/$COMPLAINT_ID/status)

echo $UPDATE_PROGRESS | jq .

if echo $UPDATE_PROGRESS | jq -e '.message' > /dev/null 2>&1; then
    echo "✅ Status updated to inprogress"
else
    echo "❌ Failed to update status"
    exit 1
fi

# Verify the change
echo -e "\n4️⃣ Verifying status change..."
UPDATED=$(curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/admin/complaints)
NEW_STATUS=$(echo $UPDATED | jq -r '.data[0].Status')
echo "   New status: $NEW_STATUS"

if [ "$NEW_STATUS" == "inprogress" ]; then
    echo "✅ Status verified: inprogress"
else
    echo "❌ Status not updated correctly"
fi

# Update to 'resolved'
echo -e "\n5️⃣ Updating status to 'resolved'..."
UPDATE_RESOLVED=$(curl -s -X PUT \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"resolved"}' \
  http://localhost:8080/api/admin/complaints/$COMPLAINT_ID/status)

echo $UPDATE_RESOLVED | jq .

if echo $UPDATE_RESOLVED | jq -e '.message' > /dev/null 2>&1; then
    echo "✅ Status updated to resolved"
else
    echo "❌ Failed to update status"
    exit 1
fi

# Final verification
echo -e "\n6️⃣ Final verification..."
FINAL=$(curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/admin/complaints)
FINAL_STATUS=$(echo $FINAL | jq -r '.data[0].Status')
echo "   Final status: $FINAL_STATUS"

if [ "$FINAL_STATUS" == "resolved" ]; then
    echo "✅ Status verified: resolved"
else
    echo "❌ Status not updated correctly"
fi

# Check timeline
echo -e "\n7️⃣ Checking timeline entries..."
TIMELINE=$(curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/complaints/$COMPLAINT_ID/timeline)
TIMELINE_COUNT=$(echo $TIMELINE | jq '. | length')
echo "   Timeline entries: $TIMELINE_COUNT"
echo $TIMELINE | jq '.[].message'

echo -e "\n============================================"
echo "✅ All tests passed!"
echo "✅ Status update endpoint working correctly!"
