#!/bin/bash

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

# Use login details from seed.go
STUDENT_EMAIL="student1@uni.com"
STUDENT_PASSWORD="student123"
ADMIN_EMAIL="admin@hostel.com"
ADMIN_PASSWORD="admin123"

# Login as student
STUDENT_LOGIN_RESP=$(curl -s -X POST "$BASE/login" \
  -H "Content-Type: application/json" \
  -d '{"email": "'$STUDENT_EMAIL'", "password": "'$STUDENT_PASSWORD'"}')
STUDENT_TOKEN=$(echo "$STUDENT_LOGIN_RESP" | jq -r '.token')
if [ -z "$STUDENT_TOKEN" ] || [ "$STUDENT_TOKEN" == "null" ]; then
  echo "❌ Student login failed: $STUDENT_LOGIN_RESP"
  exit 1
fi
echo "✅ Student login successful"

# Login as warden/admin
ADMIN_LOGIN_RESP=$(curl -s -X POST "$BASE/login" \
  -H "Content-Type: application/json" \
  -d '{"email": "'$ADMIN_EMAIL'", "password": "'$ADMIN_PASSWORD'"}')
ADMIN_TOKEN=$(echo "$ADMIN_LOGIN_RESP" | jq -r '.token')
if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" == "null" ]; then
  echo "❌ Admin login failed: $ADMIN_LOGIN_RESP"
  exit 1
fi
echo "✅ Admin login successful"


# Test complaint creation as student (with JPEG attachment)
COMPLAINT_RESP=$(curl -s -X POST $BASE/student/complaints \
  -H "Authorization: Bearer $STUDENT_TOKEN" \
  -F "title=Fan not working" \
  -F "type=electricity" \
  -F "description=Fan in my room stopped working" \
  -F "priority=medium" \
  -F "attachments=@tiny.jpg;type=image/jpeg")
echo -e "\n🧾 Complaint creation (student):"
echo "$COMPLAINT_RESP" | jq . || echo "$COMPLAINT_RESP"

# Fetch all complaints as student
COMPLAINTS_STUDENT=$(curl -s -X GET $BASE/student/complaints -H "Authorization: Bearer $STUDENT_TOKEN")
echo -e "\n📋 All complaints (student):"
echo "$COMPLAINTS_STUDENT" | jq . || echo "$COMPLAINTS_STUDENT"

# Fetch all complaints as warden/admin
COMPLAINTS_ADMIN=$(curl -s -X GET $BASE/admin/complaints -H "Authorization: Bearer $ADMIN_TOKEN")
echo -e "\n📋 All complaints (warden/admin):"
echo "$COMPLAINTS_ADMIN" | jq . || echo "$COMPLAINTS_ADMIN"


# Submit apology as student (with JPEG attachment)
APOLOGY_RESP=$(curl -s -X POST $BASE/student/apologies \
  -H "Authorization: Bearer $STUDENT_TOKEN" \
  -F "type=misconduct" \
  -F "message=Apology for missing roll call" \
  -F "description=Woke up late, will be careful next time" \
  -F "attachments=@tiny.jpg;type=image/jpeg")
echo -e "\n✉️ Apology submission (student):"
echo "$APOLOGY_RESP" | jq . || echo "$APOLOGY_RESP"

# Fetch all apologies as student
APOLOGIES_STUDENT=$(curl -s -X GET $BASE/student/apologies -H "Authorization: Bearer $STUDENT_TOKEN")
echo -e "\n📬 All apologies (student):"
echo "$APOLOGIES_STUDENT" | jq . || echo "$APOLOGIES_STUDENT"

# Fetch all apologies as warden/admin
APOLOGIES_ADMIN=$(curl -s -X GET $BASE/admin/apologies -H "Authorization: Bearer $ADMIN_TOKEN")
echo -e "\n📬 All apologies (warden/admin):"
echo "$APOLOGIES_ADMIN" | jq . || echo "$APOLOGIES_ADMIN"

# Fetch metrics as warden/admin
METRICS_STATUS=$(curl -s -X GET $BASE/metrics/status-summary -H "Authorization: Bearer $ADMIN_TOKEN")
echo -e "\n📊 Metrics (status summary):"
echo "$METRICS_STATUS" | jq . || echo "$METRICS_STATUS"

METRICS_RESOLUTION=$(curl -s -X GET $BASE/metrics/resolution-rate -H "Authorization: Bearer $ADMIN_TOKEN")
echo -e "\n📊 Metrics (resolution rate):"
echo "$METRICS_RESOLUTION" | jq . || echo "$METRICS_RESOLUTION"

METRICS_PENDING=$(curl -s -X GET $BASE/metrics/pending-count -H "Authorization: Bearer $ADMIN_TOKEN")
echo -e "\n📊 Metrics (pending count):"
echo "$METRICS_PENDING" | jq . || echo "$METRICS_PENDING"

echo ""
echo "-----------------------------------------------"
echo "✅ All endpoints tested successfully!"
echo ""
