#!/usr/bin/env zsh
set -euo pipefail

# Robust notification verification script
# - captures HTTP status codes and response bodies
# - prints pretty JSON when output is valid JSON
# - avoids jq failures aborting the whole script

print_json_or_raw() {
  local body="$1"
  # Try to pretty-print JSON, otherwise print raw
  if echo "$body" | jq . >/dev/null 2>&1; then
    echo "$body" | jq .
  else
    echo "$body"
  fi
}

http_post() {
  # $1 = URL, $2 = JSON body, $3.. = extra curl args
  local url="$1"; shift
  local data="$1"; shift
  local resp
  resp=$(curl -s -w "\n%{http_code}" -X POST "$url" -H "Content-Type: application/json" -d "$data" "$@")
  local code
  code=$(echo "$resp" | tail -n1)
  local body
  body=$(echo "$resp" | sed '$d')
  echo "$code"$'|'"$body"
}

http_get() {
  # $1 = URL, $2.. = extra curl args
  local url="$1"; shift
  local resp
  resp=$(curl -s -w "\n%{http_code}" -X GET "$url" "$@")
  local code
  code=$(echo "$resp" | tail -n1)
  local body
  body=$(echo "$resp" | sed '$d')
  echo "$code"$'|'"$body"
}

echo "1) Login student (studentverify_one@uni.com)"
ST_R=$(http_post "http://localhost:8080/api/login" '{"email":"studentverify_one@uni.com","password":"studentPass123"}')
ST_CODE=${ST_R%%|*}
ST_BODY=${ST_R#*|}
echo "Student login HTTP code: $ST_CODE"
print_json_or_raw "$ST_BODY"
ST_TOKEN=$(echo "$ST_BODY" | jq -r .token 2>/dev/null || true)
if [ -z "$ST_TOKEN" ] || [ "$ST_TOKEN" = "null" ]; then
  echo "ERROR: student token empty; aborting" >&2
  exit 2
fi

echo "\n2) Submit complaint as student (multipart/form-data)"
# Using multipart form (-F) because the server expects form values and may accept attachments
COMPL_RAW=$(curl -s -w "\n%{http_code}" -X POST "http://localhost:8080/api/student/complaints" -H "Authorization: Bearer $ST_TOKEN" \
  -F "title=VerifyNotify-SV1-automated" \
  -F "type=plumbing" \
  -F "description=Trigger notification via script" \
  -F "priority=medium")
COMPL_CODE=$(echo "$COMPL_RAW" | tail -n1)
COMPL_BODY=$(echo "$COMPL_RAW" | sed '$d')
echo "Complaint HTTP code: $COMPL_CODE"
print_json_or_raw "$COMPL_BODY"

echo "\n3) Login block admin (blockadmin_verify@hostel.com)"
BLK_R=$(http_post "http://localhost:8080/api/login" '{"email":"blockadmin_verify@hostel.com","password":"adminPass123"}')
BLK_CODE=${BLK_R%%|*}
BLK_BODY=${BLK_R#*|}
echo "Block admin HTTP code: $BLK_CODE"
print_json_or_raw "$BLK_BODY"
BLK_TOKEN=$(echo "$BLK_BODY" | jq -r .token 2>/dev/null || true)
if [ -z "$BLK_TOKEN" ] || [ "$BLK_TOKEN" = "null" ]; then
  echo "WARNING: block admin token empty" >&2
fi

echo "\n4) Login chief admin (chiefadmin_verify@hostel.com)"
CH_R=$(http_post "http://localhost:8080/api/login" '{"email":"chiefadmin_verify@hostel.com","password":"chiefPass123"}')
CH_CODE=${CH_R%%|*}
CH_BODY=${CH_R#*|}
echo "Chief admin HTTP code: $CH_CODE"
print_json_or_raw "$CH_BODY"
CH_TOKEN=$(echo "$CH_BODY" | jq -r .token 2>/dev/null || true)
if [ -z "$CH_TOKEN" ] || [ "$CH_TOKEN" = "null" ]; then
  echo "WARNING: chief admin token empty" >&2
fi

# wait briefly for async insertion
sleep 1

echo "\n5) Fetch notifications for block admin"
if [ -n "${BLK_TOKEN-}" ] && [ "$BLK_TOKEN" != "null" ]; then
  BLK_NOT_R=$(http_get "http://localhost:8080/api/admin/notifications" -H "Authorization: Bearer $BLK_TOKEN")
  BLK_NOT_CODE=${BLK_NOT_R%%|*}
  BLK_NOT_BODY=${BLK_NOT_R#*|}
  echo "Block admin notifications HTTP code: $BLK_NOT_CODE"
  print_json_or_raw "$BLK_NOT_BODY"
else
  echo "No block admin token; skipping fetch"
fi

echo "\n6) Fetch notifications for chief admin"
if [ -n "${CH_TOKEN-}" ] && [ "$CH_TOKEN" != "null" ]; then
  CH_NOT_R=$(http_get "http://localhost:8080/api/admin/notifications" -H "Authorization: Bearer $CH_TOKEN")
  CH_NOT_CODE=${CH_NOT_R%%|*}
  CH_NOT_BODY=${CH_NOT_R#*|}
  echo "Chief admin notifications HTTP code: $CH_NOT_CODE"
  print_json_or_raw "$CH_NOT_BODY"
else
  echo "No chief admin token; skipping fetch"
fi

echo "\n7) Recent INSERT INTO \"notifications\" lines in server.log (last 20)"
grep "INSERT INTO \"notifications\"" server.log | tail -n 20 || true

echo "\nDone."
