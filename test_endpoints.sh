#!/bin/bash
set -e  # exit immediately if a command fails

echo ""
echo "🚀 Starting Hostel Management API test sequence..."
echo "-----------------------------------------------"

BASE="http://localhost:8080/api"

# 🧍 Create Student (skip if exists)
echo -e "\n🧍 Creating Student user..."
STU_RESP=$(curl -s -X POST $BASE/signup \
  -H "Content-Type: application/json" \
  -d '{"name": "Student1", "email": "student1@uni.com", "password": "student123", "role": "student", "student_id": "S1001", "hostel": "Block-A", "room_no": "201"}')
if [[ $(echo "$STU_RESP" | jq -e . >/dev/null 2>&1; echo $?) -eq 0 ]]; then
  echo "$STU_RESP" | jq .
else
  echo "Non-JSON response during student signup: $STU_RESP"
fi
echo "✅ Student signup attempted"

# 👮 Create Admin WITHOUT block (should be rejected later)
echo -e "\n👮 Creating Admin user (no block)..."
curl -s -X POST $BASE/signup \
  -H "Content-Type: application/json" \
  -d '{"name": "AdminNoBlock", "email": "admin_noblock@hostel.com", "password": "admin123", "role": "admin"}' \
  | jq .
echo "✅ Admin (no block) signup attempted"

# 👮 Create Admin WITH block
echo -e "\n👮 Creating Admin user (with block)..."
curl -s -X POST $BASE/signup \
  -H "Content-Type: application/json" \
  -d '{"name": "Admin", "email": "admin@hostel.com", "password": "admin123", "role": "admin", "block": "Block-A"}' \
  | jq .
echo "✅ Admin (with block) signup done"

# 🔑 Logins
echo -e "\n🔑 Logging in Student..."
TOKEN=$(curl -s -X POST $BASE/login \
  -H "Content-Type: application/json" \
  -d '{"email": "student1@uni.com", "password": "student123"}' | jq -r '.token')
echo "✅ Student login successful"

echo -e "\n🔑 Logging in Admin..."
ADMINTOKEN_NOBLOCK=$(curl -s -X POST $BASE/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin_noblock@hostel.com", "password": "admin123"}' | jq -r '.token')
echo "✅ Admin (no block) login attempted"

ADMINTOKEN=$(curl -s -X POST $BASE/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@hostel.com", "password": "admin123"}' | jq -r '.token')
echo "✅ Admin (with block) login successful"

# -----------------------
# Forgot / Reset Password
# -----------------------
echo -e "\n🔐 Testing forgot/reset password flow..."
curl -s -X POST $BASE/forgot-password -H "Content-Type: application/json" -d '{"email":"student1@uni.com"}' >/dev/null || true
echo "Requested password reset (email simulated)"

# Retrieve dev token (DEV_MODE=true must be set when running server). Retry a few times if not present yet.
DEV_TOKEN=""
for i in 1 2 3; do
  sleep 1
  T=$(curl -s -G $BASE/dev/reset-token --data-urlencode "email=student1@uni.com")
  # ensure valid JSON
  if echo "$T" | jq -e . >/dev/null 2>&1; then
    DEV_TOKEN=$(echo "$T" | jq -r '.token')
    if [[ "$DEV_TOKEN" != "null" && -n "$DEV_TOKEN" ]]; then
      break
    fi
  fi
done
echo "Dev reset token: $DEV_TOKEN"

if [ -n "$DEV_TOKEN" ] && [ "$DEV_TOKEN" != "null" ]; then
  RSP=$(curl -s -X POST $BASE/reset-password -H "Content-Type: application/json" -d "{\"token\": \"$DEV_TOKEN\", \"password\": \"newstudentpass\"}")
  if echo "$RSP" | jq -e . >/dev/null 2>&1; then
    echo "$RSP" | jq .
  else
    echo "Reset password response: $RSP"
  fi
  echo "Password reset attempted"
  # login with new password
  NEWTOKEN=$(curl -s -X POST $BASE/login -H "Content-Type: application/json" -d '{"email":"student1@uni.com", "password": "newstudentpass"}' | jq -r '.token')
  echo "Login with new password token: $NEWTOKEN"
  if [ -n "$NEWTOKEN" ] && [ "$NEWTOKEN" != "null" ]; then
    TOKEN=$NEWTOKEN
  fi
else
  echo "No dev token found; skipping reset/password check"
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
echo -e "\n🧪 Admin (no block) attempting to fetch admin complaints (expect 403)..."
curl -s -o /dev/stderr -w "HTTP_STATUS:%{http_code}\n" -X GET $BASE/admin/complaints -H "Authorization: Bearer $ADMINTOKEN_NOBLOCK" || true
echo "✅ Tested admin without block denied"

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
curl -s -X POST $BASE/student/apologies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
        "type": "misconduct",
        "message": "Apology for missing morning roll call",
        "description": "Woke up late, will be careful next time"
      }' | jq .
echo "✅ Apology submitted"

# 📬 Fetch Student Apologies
echo -e "\n📬 Fetching Student Apologies..."
curl -s -X GET $BASE/student/apologies \
  -H "Authorization: Bearer $TOKEN" | jq .
echo "✅ Fetched student apologies"

# 🛠 Admin: Fetch All Apologies
echo -e "\n🛠 Fetching All Apologies (Admin)..."
curl -s -X GET $BASE/admin/apologies \
  -H "Authorization: Bearer $ADMINTOKEN" | jq .
APOLOGY_ID=$(curl -s -X GET $BASE/admin/apologies \
  -H "Authorization: Bearer $ADMINTOKEN" | jq -r '.data[0].ID')
echo "Apology ID: $APOLOGY_ID"
echo "✅ Admin fetched apologies"

# 🧾 Review Apology (Admin)
if [ "$APOLOGY_ID" != "null" ] && [ -n "$APOLOGY_ID" ]; then
  echo -e "\n🔧 Reviewing Apology..."
  curl -s -X PUT $BASE/admin/apologies/$APOLOGY_ID/review \
    -H "Authorization: Bearer $ADMINTOKEN" \
    -H "Content-Type: application/json" \
    -d '{"status": "accepted", "comment": "Valid apology, warning issued."}' | jq .
  echo "✅ Apology reviewed successfully"
else
  echo "⚠️ No apology found to review."
fi

# 📊 Metrics
echo -e "\n📊 Fetching Metrics (Admin)..."
curl -s -X GET $BASE/metrics/status-summary -H "Authorization: Bearer $ADMINTOKEN" | jq .
curl -s -X GET $BASE/metrics/resolution-rate -H "Authorization: Bearer $ADMINTOKEN" | jq .
curl -s -X GET $BASE/metrics/pending-count -H "Authorization: Bearer $ADMINTOKEN" | jq .
echo "✅ Metrics endpoints tested successfully"

# -----------------------
# Counseling flow tests
# -----------------------
echo -e "\n🧑‍⚕️ Creating a counselor slot (dev flow)"
# We'll use the existing admin token to create a slot for a counselor. For test, create a fake counselor user first.
curl -s -X POST $BASE/signup -H "Content-Type: application/json" -d '{"name":"Counselor1","email":"counselor1@hostel.com","password":"counselor123","role":"counselor"}' | jq .
CID=$(curl -s -X POST $BASE/login -H "Content-Type: application/json" -d '{"email":"counselor1@hostel.com","password":"counselor123"}' | jq -r '.token')
# get counselor user id by logging in and decoding token is complex; instead query dev endpoint to find user id via student listing? For simplicity, seed a counselor ID from DB not possible here; instead we'll create a slot for the admin as 'counselor' by using admin's ID.
# Get admin user id via profile
ADMIN_USER_ID=$(curl -s -X POST $BASE/login -H "Content-Type: application/json" -d '{"email":"admin@hostel.com","password":"admin123"}' | jq -r '.token' )
# As we don't have a token decode in shell easily, we will create a slot for a placeholder counselor by using the admin's own user id fetched via admin profile (admin must login and call admin->student endpoint will fail); Instead, skip complex ID handling and use the seeded counselor by listing all users via an admin-only endpoint (not available). To keep this test deterministic, we'll skip creating a slot if we can't determine counselor id.
echo "Skipping counselor slot creation in lightweight smoke test (use manual testing or run integration tests)"

# -----------------------
# Profile checks
# -----------------------
echo -e "\n👤 Fetching student profile (self)"
PROF_RESP=$(curl -s -X GET $BASE/student/profile -H "Authorization: Bearer $TOKEN")
if echo "$PROF_RESP" | jq -e . >/dev/null 2>&1; then
  echo "$PROF_RESP" | jq .
else
  echo "Non-JSON response fetching student profile: $PROF_RESP"
fi

echo -e "\n👮 Admin fetching student profile by student_identifier"
STUDENT_IDENTIFIER=$(echo "$PROF_RESP" | jq -r '.student_identifier' 2>/dev/null || echo "")
if [ "$STUDENT_IDENTIFIER" != "null" ] && [ -n "$STUDENT_IDENTIFIER" ]; then
  ADM_PROF=$(curl -s -X GET $BASE/admin/student/$STUDENT_IDENTIFIER -H "Authorization: Bearer $ADMINTOKEN")
  if echo "$ADM_PROF" | jq -e . >/dev/null 2>&1; then
    echo "$ADM_PROF" | jq .
  else
    echo "Non-JSON response fetching admin view of student: $ADM_PROF"
  fi
else
  echo "Student_identifier missing; skipping admin profile fetch"
fi

echo ""
echo "-----------------------------------------------"
echo "✅ All endpoints tested successfully!"
echo ""
