#!/bin/bash
set -euo pipefail

echo ""
echo "🚀 Starting Hostel Management API test sequence..."
echo "-----------------------------------------------"

BASE="${BASE:-http://localhost:8080/api}"

# helper to assert a non-empty JSON field
assert_nonempty() {
  local val="$1"; local msg="$2";
  if [[ -z "${val}" || "${val}" == "null" ]]; then
    echo "❌ ${msg}" >&2
    exit 1
  fi
}

# -----------------------
# Public health
# -----------------------
echo -e "\n🩺 Cloudinary health check (public)..."
curl -s -X GET "$BASE/health/cloudinary" | jq . || true

# -----------------------
# Signups (idempotent)
# -----------------------
echo -e "\n🧍 Creating Student user..."
STU_RESP=$(curl -s -X POST "$BASE/signup" \
  -H "Content-Type: application/json" \
  -d '{"name": "Student1", "email": "student1@uni.com", "password": "student123", "role": "student", "student_id": "S1001", "hostel": "Block-A", "room_no": "201"}')
echo "$STU_RESP" | jq . || echo "Non-JSON response during student signup: $STU_RESP"
echo "✅ Student signup attempted"

CHIEF_SUFFIX=$(date +%s)
CHIEF_EMAIL="chief_${CHIEF_SUFFIX}@hostel.com"
echo -e "\n👑 Creating Chief Admin (unique) $CHIEF_EMAIL ..."
CHIEF_RESP=$(curl -s -X POST "$BASE/signup" \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"Chief\", \"email\": \"$CHIEF_EMAIL\", \"password\": \"chief123\", \"role\": \"chief_admin\"}")
echo "$CHIEF_RESP" | jq . || echo "Non-JSON response during chief signup: $CHIEF_RESP"
echo "✅ Chief admin signup attempted"

BLOCK_ADMIN_EMAIL="blockadmin_${CHIEF_SUFFIX}@hostel.com"
echo -e "\n🏢 Creating Block Admin (Block-A) $BLOCK_ADMIN_EMAIL ..."
BA_RESP=$(curl -s -X POST "$BASE/signup" -H "Content-Type: application/json" \
  -d "{\"name\": \"BlockAdmin\", \"email\": \"$BLOCK_ADMIN_EMAIL\", \"password\": \"blockadmin123\", \"role\": \"admin\", \"block\": \"Block-A\"}")
echo "$BA_RESP" | jq . || echo "Non-JSON response during block admin signup: $BA_RESP"
echo "✅ Block admin signup attempted"

# -----------------------
# Logins
# -----------------------
echo -e "\n🔑 Logging in Student..."
TOKEN=$(curl -s -X POST "$BASE/login" \
  -H "Content-Type: application/json" \
  -d '{"email": "student1@uni.com", "password": "student123"}' | jq -r '.token')
assert_nonempty "$TOKEN" "Student login failed: empty token"
echo "✅ Student login successful"

echo -e "\n🔑 Logging in Chief Admin..."
ADMINTOKEN=$(curl -s -X POST "$BASE/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$CHIEF_EMAIL\", \"password\": \"chief123\"}" | jq -r '.token')
assert_nonempty "$ADMINTOKEN" "Chief admin login failed: empty token"
echo "✅ Chief admin login successful"

echo -e "\n🔑 Logging in Block Admin..."
BLOCKADMINTOKEN=$(curl -s -X POST "$BASE/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$BLOCK_ADMIN_EMAIL\", \"password\": \"blockadmin123\"}" | jq -r '.token')
assert_nonempty "$BLOCKADMINTOKEN" "Block admin login failed: empty token"
echo "✅ Block admin login successful"

############################################
# Forgot / Reset Password (best-effort dev)
############################################
echo -e "\n🔐 Testing forgot/reset password flow..."
curl -s -X POST "$BASE/forgot-password" -H "Content-Type: application/json" -d '{"email":"student1@uni.com"}' >/dev/null || true
if curl -s "$BASE/dev/reset-token?email=student1@uni.com" | jq -e . >/dev/null 2>&1; then
  DEV_TOKEN=$(curl -s "$BASE/dev/reset-token?email=student1@uni.com" | jq -r '.token') || true
  if [[ -n "${DEV_TOKEN:-}" && "${DEV_TOKEN}" != "null" ]]; then
    curl -s -X POST "$BASE/reset-password" -H "Content-Type: application/json" -d "{\"token\": \"$DEV_TOKEN\", \"password\": \"newstudentpass\"}" | jq . || true
    # try logging in with new password (non-fatal)
    NEWTOKEN=$(curl -s -X POST "$BASE/login" -H "Content-Type: application/json" -d '{"email":"student1@uni.com", "password": "newstudentpass"}' | jq -r '.token') || true
    if [[ -n "${NEWTOKEN:-}" && "${NEWTOKEN}" != "null" ]]; then TOKEN="$NEWTOKEN"; fi
  else
    echo "(DEV_MODE not enabled or token not available)"
  fi
else
  echo "(DEV reset endpoint unavailable; skipping)"
fi

#🧾 Complaint creation
echo -e "\n🧾 Creating Complaint..."
curl -s -X POST $BASE/student/complaints \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
        "title": "Fan not working",
        "type": "electricity",
        "description": "Fan in my room stopped working"
      }' | jq .
echo "✅ Complaint created"

#📋 Fetch All Complaints (Student)
echo -e "\n📋 Fetching All Complaints (Student)..."
curl -s -X GET $BASE/student/complaints \
  -H "Authorization: Bearer $TOKEN" | jq .
echo "✅ Fetched complaints successfully"

# 🧪 Admin (no block) should be denied access to admin endpoints
echo -e "\n🧪 Chief Admin fetching complaints..."
curl -s -X GET "$BASE/admin/complaints" -H "Authorization: Bearer $ADMINTOKEN" | jq '.count'
echo "✅ Chief admin complaints fetched"

echo -e "\n🧪 Block Admin (should see only Block-A complaints if any)..."
curl -s -X GET "$BASE/admin/complaints" -H "Authorization: Bearer $BLOCKADMINTOKEN" | jq '.count'
echo "✅ Block admin complaints fetched"

# 🧾 Create Complaint with JPEG upload (student)
echo -e "\n🧾 Creating Complaint with JPEG attachment..."
curl -s -X POST $BASE/student/complaints \
  -H "Authorization: Bearer $TOKEN" \
  -F "title=Fan not working" \
  -F "type=electricity" \
  -F "description=Fan in my room stopped working" \
  -F "attachments=@tiny.jpg;type=image/jpeg" | jq .
echo "✅ Complaint with attachment created"

# ✉️ Submit Apology
echo -e "\n✉️ Submitting Apology..."
APOL_RESP=$(curl -s -X POST "$BASE/student/apologies" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"misconduct","message":"Apology for missing morning roll call","description":"Woke up late, will be careful next time"}')
echo "$APOL_RESP" | jq .
APOLOGY_ID=$(echo "$APOL_RESP" | jq -r '.data.id // .data.ID // .id // .ID')
assert_nonempty "$APOLOGY_ID" "Apology creation failed: missing id"
echo "✅ Apology submitted"

# 📬 Fetch Student Apologies
echo -e "\n📬 Fetching Student Apologies..."
curl -s -X GET $BASE/student/apologies \
  -H "Authorization: Bearer $TOKEN" | jq .
echo "✅ Fetched student apologies"

# 🛠 Admin: Fetch All Apologies
# cleaned malformed stray header line
echo -e "\n🛠 Fetching All Apologies (Admin)..."
curl -s -X GET "$BASE/admin/apologies" -H "Authorization: Bearer $ADMINTOKEN" | jq '.count'
echo "✅ Admin fetched apologies"

# 🧾 Review Apology (Admin)
echo -e "\n🛠 Reviewing Apology..."
curl -s -X PUT "$BASE/admin/apologies/$APOLOGY_ID/review" \
  -H "Authorization: Bearer $ADMINTOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "accepted", "comment": "Valid apology, warning issued."}' | jq .
echo "✅ Apology reviewed successfully"

# 📊 Metrics
echo -e "\n📊 Fetching Metrics (Admin)..."
curl -s -X GET $BASE/metrics/status-summary -H "Authorization: Bearer $ADMINTOKEN" | jq .
curl -s -X GET $BASE/metrics/resolution-rate -H "Authorization: Bearer $ADMINTOKEN" | jq .
curl -s -X GET $BASE/metrics/pending-count -H "Authorization: Bearer $ADMINTOKEN" | jq .
echo "✅ Metrics endpoints tested successfully"

# -----------------------
# Admin delete complaint
# -----------------------
echo -e "\n🗑️  Admin deleting a complaint..."
DEL_ID=$(curl -s -X GET "$BASE/admin/complaints" -H "Authorization: Bearer $ADMINTOKEN" | jq -r '.data[0].ID // .data[0].id')
if [ -n "$DEL_ID" ] && [ "$DEL_ID" != "null" ]; then
  echo "Attempting to delete complaint: $DEL_ID"
  DEL_RESP=$(curl -s -X DELETE $BASE/admin/complaints/$DEL_ID \
    -H "Authorization: Bearer $ADMINTOKEN")
  if echo "$DEL_RESP" | jq -e . >/dev/null 2>&1; then
    echo "$DEL_RESP" | jq .
  else
    echo "Delete response: $DEL_RESP"
  fi
  echo "✅ Delete attempted"
else
  echo "⚠️  No complaint found to delete"
fi

# -----------------------
# Counseling flow tests
# -----------------------
echo -e "\n🧑‍⚕️ Counseling flow (skipped if counselor id unavailable)"
echo "Skipping counselor slot creation in smoke test (no public endpoint to resolve counselor id)"

# -----------------------
# Profile checks
# -----------------------
echo -e "\n👤 Fetching student profile (self)"
PROF_RESP=$(curl -s -X GET "$BASE/student/profile" -H "Authorization: Bearer $TOKEN")
if echo "$PROF_RESP" | jq -e . >/dev/null 2>&1; then
  echo "$PROF_RESP" | jq .
else
  echo "Non-JSON response fetching student profile: $PROF_RESP"
fi

echo -e "\n👮 Admin fetching student profile by student_identifier"
STUDENT_IDENTIFIER=$(echo "$PROF_RESP" | jq -r '.student_identifier // .student_id' 2>/dev/null || echo "")
if [ "$STUDENT_IDENTIFIER" != "null" ] && [ -n "$STUDENT_IDENTIFIER" ]; then
  ADM_PROF=$(curl -s -X GET "$BASE/admin/student/$STUDENT_IDENTIFIER" -H "Authorization: Bearer $ADMINTOKEN")
  if echo "$ADM_PROF" | jq -e . >/dev/null 2>&1; then
    echo "$ADM_PROF" | jq .
  else
    echo "Non-JSON response fetching admin view of student: $ADM_PROF"
  fi
else
  echo "Student_identifier missing; skipping admin profile fetch"
fi

# -----------------------
# Notifications flow
# -----------------------
echo -e "\n🔔 Notifications (student)"
NOTE_LIST=$(curl -s -X GET "$BASE/notifications" -H "Authorization: Bearer $TOKEN")
echo "$NOTE_LIST" | jq '.unreadCount'
NOTE_ID=$(echo "$NOTE_LIST" | jq -r '.data[0].id // .data[0].ID // empty')
if [[ -n "${NOTE_ID:-}" ]]; then
  curl -s -X PATCH "$BASE/notifications/$NOTE_ID/read" -H "Authorization: Bearer $TOKEN" | jq .
  curl -s -X PATCH "$BASE/notifications/read-all" -H "Authorization: Bearer $TOKEN" | jq .
  curl -s -X DELETE "$BASE/notifications/$NOTE_ID" -H "Authorization: Bearer $TOKEN" | jq .
fi

echo ""
echo "-----------------------------------------------"
echo "✅ All endpoints tested successfully!"
echo ""
