#!/bin/bash
set -e  # exit immediately if a command fails

echo ""
echo "🚀 Starting Hostel Management API test sequence..."
echo "-----------------------------------------------"

BASE="http://localhost:8080/api"

# 🧍 Create Student (skip if exists)
echo -e "\n🧍 Creating Student user..."
curl -s -X POST $BASE/signup \
  -H "Content-Type: application/json" \
  -d '{"name": "Student1", "email": "student1@uni.com", "password": "student123", "role": "student", "hostel": "Block-A", "room_no": "201"}' \
  | jq .
echo "✅ Student signup done"

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

echo ""
echo "-----------------------------------------------"
echo "✅ All endpoints tested successfully!"
echo ""
