#!/usr/bin/env bash
set -euo pipefail

# Smoke test: verify admin block restriction vs chief admin
# Creates unique users, submits complaints from two different blocks,
# and fetches admin complaints as a block-admin and as a chief admin.

BASE=${BASE:-http://localhost:8080}
RAND=$(date +%s)

command -v jq >/dev/null 2>&1 || { echo "This script requires 'jq' (https://stedolan.github.io/jq/)"; exit 1; }

echo "Using base URL: $BASE"

# 1) Create a block-admin (Block-X)
echo "\n1) Creating block admin..."
curl -s -X POST "$BASE/api/signup" \
  -H "Content-Type: application/json" \
  -d @- <<JSON | jq .
{
  "name": "BlockAdmin-$RAND",
  "email": "blockadmin${RAND}@hostel.com",
  "password": "adminPass123",
  "role": "admin",
  "block": "Block-X"
}
JSON

# 2) Create a chief admin
echo "\n2) Creating chief admin..."
curl -s -X POST "$BASE/api/signup" \
  -H "Content-Type: application/json" \
  -d @- <<JSON | jq .
{
  "name": "ChiefAdmin-$RAND",
  "email": "chief${RAND}@hostel.com",
  "password": "chiefPass123",
  "role": "chief_admin"
}
JSON

# 3) Create a student in Block-X (who will submit a complaint)
echo "\n3) Creating student in Block-X..."
curl -s -X POST "$BASE/api/signup" \
  -H "Content-Type: application/json" \
  -d @- <<JSON | jq .
{
  "name": "StudentBlockX-$RAND",
  "email": "studentblockx${RAND}@uni.com",
  "password": "studentPass123",
  "role": "student",
  "student_id": "BX-${RAND}",
  "hostel": "Block-X",
  "room_no": "201"
}
JSON

# 4) Create a student in Block-Y (different block)
echo "\n4) Creating student in Block-Y..."
curl -s -X POST "$BASE/api/signup" \
  -H "Content-Type: application/json" \
  -d @- <<JSON | jq .
{
  "name": "StudentBlockY-$RAND",
  "email": "studentblocky${RAND}@uni.com",
  "password": "studentPass123",
  "role": "student",
  "student_id": "BY-${RAND}",
  "hostel": "Block-Y",
  "room_no": "301"
}
JSON

# 5) Login: block-admin -> get token
echo "\n5) Logging in block admin..."
ADMIN_TOKEN=$(curl -s -X POST "$BASE/api/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"blockadmin${RAND}@hostel.com\",\"password\":\"adminPass123\"}" | jq -r '.token')
echo "Block admin token: ${ADMIN_TOKEN:0:40}..."

# 6) Login: chief-admin -> get token
echo "\n6) Logging in chief admin..."
CHIEF_TOKEN=$(curl -s -X POST "$BASE/api/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"chief${RAND}@hostel.com\",\"password\":\"chiefPass123\"}" | jq -r '.token')
echo "Chief admin token: ${CHIEF_TOKEN:0:40}..."

# 7) Login: student in Block-X -> get token
echo "\n7) Logging in student Block-X..."
STU_X_TOKEN=$(curl -s -X POST "$BASE/api/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"studentblockx${RAND}@uni.com\",\"password\":\"studentPass123\"}" | jq -r '.token')
echo "Student Block-X token: ${STU_X_TOKEN:0:40}..."

# 8) Login: student in Block-Y -> get token
echo "\n8) Logging in student Block-Y..."
STU_Y_TOKEN=$(curl -s -X POST "$BASE/api/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"studentblocky${RAND}@uni.com\",\"password\":\"studentPass123\"}" | jq -r '.token')
echo "Student Block-Y token: ${STU_Y_TOKEN:0:40}..."

# 9) Submit a complaint as Block-X student (multipart form, no attachment)
echo "\n9) Student Block-X submits a complaint..."
curl -s -X POST "$BASE/api/student/complaints" \
  -H "Authorization: Bearer $STU_X_TOKEN" \
  -F "title=Broken Lamp" \
  -F "type=electricity" \
  -F "description=The lamp in my room is broken" \
  -F "priority=high" | jq .

# 10) Submit a complaint as Block-Y student
echo "\n10) Student Block-Y submits a complaint..."
curl -s -X POST "$BASE/api/student/complaints" \
  -H "Authorization: Bearer $STU_Y_TOKEN" \
  -F "title=Clogged Drain" \
  -F "type=plumbing" \
  -F "description=Drain in bathroom is clogged" \
  -F "priority=medium" | jq .

# Give the server a moment to process async notifications
sleep 1

# 11) As Block admin: fetch admin complaints -> should only show Block-X complaint(s)
echo "\n11) Block admin fetching complaints (should only include Block-X)..."
curl -s -X GET "$BASE/api/admin/complaints" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# 12) As Chief admin: fetch admin complaints -> should show both Block-X and Block-Y complaints
echo "\n12) Chief admin fetching complaints (should include all)..."
curl -s -X GET "$BASE/api/admin/complaints" \
  -H "Authorization: Bearer $CHIEF_TOKEN" | jq .

echo "\nDone."
chmod +x scripts/check_admin_complaints.sh
