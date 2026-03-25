#!/usr/bin/env bash
set -u

BASE_URL="${BASE_URL:-http://192.168.20.1:7002}"
EMAIL="${EMAIL:-sim.perencana@rsud.local}"
PASSWORD="${PASSWORD:-Admin123!}"
OUT_DIR="${OUT_DIR:-./test-results}"
TS="$(date +%Y%m%d-%H%M%S)"
TXT_REPORT="$OUT_DIR/api-validation-$TS.txt"
JSON_REPORT="$OUT_DIR/api-validation-$TS.json"

pass=0
fail=0
CLIENT_ID=""

mkdir -p "$OUT_DIR"

log() {
  local line="$1"
  echo "$line" | tee -a "$TXT_REPORT"
}

json_escape() {
  local s="$1"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  s="${s//$'\n'/\\n}"
  s="${s//$'\r'/}"
  printf '%s' "$s"
}

json_items=""
add_json_item() {
  local name="$1"
  local ok="$2"
  local code="$3"
  local detail="$4"

  local status="FAIL"
  if [[ "$ok" == "1" ]]; then
    status="PASS"
  fi

  local item
  item="{\"name\":\"$(json_escape "$name")\",\"status\":\"$status\",\"http_code\":$code,\"detail\":\"$(json_escape "$detail")\"}"
  if [[ -z "$json_items" ]]; then
    json_items="$item"
  else
    json_items="$json_items,$item"
  fi
}

print_result() {
  local ok="$1"
  local name="$2"
  local code="$3"
  local detail="${4:-}"

  if [[ "$ok" == "1" ]]; then
    log "[PASS] $name (code=$code)"
    pass=$((pass+1))
  else
    log "[FAIL] $name (code=$code) :: $detail"
    fail=$((fail+1))
  fi

  add_json_item "$name" "$ok" "$code" "$detail"
}

extract_status() {
  awk 'match($0, /^HTTP\/[0-9.]+ ([0-9]{3})/, a){code=a[1]} END{print code+0}'
}

contains_text() {
  local body="$1"
  local expected="$2"
  echo "$body" | grep -qi "$expected"
}

request_i() {
  local method="$1"
  local url="$2"
  local body="${3:-}"
  local auth="${4:-}"

  if [[ -n "$body" ]]; then
    if [[ -n "$auth" ]]; then
      curl -sS -i -X "$method" "$url" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $auth" \
        --data-raw "$body"
    else
      curl -sS -i -X "$method" "$url" \
        -H "Content-Type: application/json" \
        --data-raw "$body"
    fi
  else
    if [[ -n "$auth" ]]; then
      curl -sS -i -X "$method" "$url" \
        -H "Authorization: Bearer $auth"
    else
      curl -sS -i -X "$method" "$url"
    fi
  fi
}

write_json_report() {
  cat > "$JSON_REPORT" <<EOF
{
  "base_url": "$(json_escape "$BASE_URL")",
  "timestamp": "$(date -Iseconds)",
  "summary": {
    "pass": $pass,
    "fail": $fail,
    "total": $((pass+fail))
  },
  "results": [$json_items]
}
EOF
}

: > "$TXT_REPORT"

log "== API Validation Run =="
log "Base URL: $BASE_URL"
log "Report TXT: $TXT_REPORT"
log "Report JSON: $JSON_REPORT"
log ""

log "== Health check =="
health="$(curl -sS "$BASE_URL/health" || true)"
if [[ -z "$health" ]]; then
  log "Server tidak merespons di $BASE_URL"
  add_json_item "health_check" "0" "0" "server not reachable"
  write_json_report
  exit 1
fi
log "$health"
log ""

log "== Login valid (ambil token) =="
login_res="$(curl -sS -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  --data-raw "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")"

TOKEN="$(echo "$login_res" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')"
if [[ -z "$TOKEN" ]]; then
  log "Gagal login setup test."
  log "$login_res"
  add_json_item "login_setup" "0" "0" "failed to get access token"
  write_json_report
  exit 1
fi
log "TOKEN_LEN=${#TOKEN}"
log ""

log "== 1) login tanpa password => 400 + password is required =="
r="$(request_i POST "$BASE_URL/api/v1/auth/login" "{\"email\":\"$EMAIL\",\"password\":\"\"}")"
code="$(echo "$r" | extract_status)"
body="$(echo "$r" | sed -n '/^\r$/,$p')"
ok=0
[[ "$code" -eq 400 ]] && contains_text "$body" "password is required" && ok=1
print_result "$ok" "login_tanpa_password" "$code" "expected 400 + password is required"

log "== 2) login email invalid => 400 + valid email =="
r="$(request_i POST "$BASE_URL/api/v1/auth/login" '{"email":"abc","password":"Admin123!"}')"
code="$(echo "$r" | extract_status)"
body="$(echo "$r" | sed -n '/^\r$/,$p')"
ok=0
[[ "$code" -eq 400 ]] && contains_text "$body" "email must be a valid email" && ok=1
print_result "$ok" "login_email_invalid" "$code" "expected 400 + valid email"

log "== 3) refresh tanpa token => 400 + refresh_token is required =="
r="$(request_i POST "$BASE_URL/api/v1/auth/refresh" '{}')"
code="$(echo "$r" | extract_status)"
body="$(echo "$r" | sed -n '/^\r$/,$p')"
ok=0
[[ "$code" -eq 400 ]] && contains_text "$body" "refresh_token is required" && ok=1
print_result "$ok" "refresh_tanpa_token" "$code" "expected 400 + refresh_token is required"

log "== 4) create client tanpa kode => 400 + kode is required =="
r="$(request_i POST "$BASE_URL/api/v1/clients" '{"nama":"Client Uji","unit_pengusul_id":3}' "$TOKEN")"
code="$(echo "$r" | extract_status)"
body="$(echo "$r" | sed -n '/^\r$/,$p')"
ok=0
[[ "$code" -eq 400 ]] && contains_text "$body" "kode is required" && ok=1
print_result "$ok" "create_tanpa_kode" "$code" "expected 400 + kode is required"

log "== 5) create client nama pendek => 400 + nama minimum length is 3 =="
r="$(request_i POST "$BASE_URL/api/v1/clients" '{"kode":"CL-VAL-001","nama":"AB","unit_pengusul_id":3}' "$TOKEN")"
code="$(echo "$r" | extract_status)"
body="$(echo "$r" | sed -n '/^\r$/,$p')"
ok=0
[[ "$code" -eq 400 ]] && contains_text "$body" "nama minimum length is 3" && ok=1
print_result "$ok" "create_nama_pendek" "$code" "expected 400 + nama minimum length is 3"

log "== 6) create client unit_pengusul_id=0 => 400 + must be greater than 0 =="
r="$(request_i POST "$BASE_URL/api/v1/clients" '{"kode":"CL-VAL-002","nama":"Client Uji","unit_pengusul_id":0}' "$TOKEN")"
code="$(echo "$r" | extract_status)"
body="$(echo "$r" | sed -n '/^\r$/,$p')"
ok=0
[[ "$code" -eq 400 ]] && contains_text "$body" "unit_pengusul_id must be greater than 0" && ok=1
print_result "$ok" "create_unit_0" "$code" "expected 400 + unit_pengusul_id must be greater than 0"

log "== 7) create client valid (setup) =="
codeval="CL-VAL-$(date +%s)"
r="$(request_i POST "$BASE_URL/api/v1/clients" "{\"kode\":\"$codeval\",\"nama\":\"Client Valid\",\"unit_pengusul_id\":3}" "$TOKEN")"
code="$(echo "$r" | extract_status)"
body="$(echo "$r" | sed -n '/^\r$/,$p')"
CLIENT_ID="$(echo "$body" | sed -n 's/.*"id":\([0-9]*\).*/\1/p' | head -n1)"
ok=0
[[ "$code" -eq 200 ]] && [[ -n "$CLIENT_ID" ]] && ok=1
print_result "$ok" "create_valid" "$code" "expected 200 and client_id"

if [[ -z "$CLIENT_ID" ]]; then
  log "Setup client gagal, skenario 8-10 dilewati."
  write_json_report
  log ""
  log "Summary: PASS=$pass FAIL=$fail TOTAL=$((pass+fail))"
  exit 0
fi

log "== 8) submit client valid => 200 =="
r="$(request_i POST "$BASE_URL/api/v1/clients/$CLIENT_ID/submit" '{"note":"submit uji"}' "$TOKEN")"
code="$(echo "$r" | extract_status)"
ok=0
[[ "$code" -eq 200 ]] && ok=1
print_result "$ok" "submit_valid" "$code" "expected 200"

log "== 9) reject reason >1000 => 400 + maximum length =="
LONG_REASON="$(printf 'R%.0s' $(seq 1 1101))"
r="$(request_i POST "$BASE_URL/api/v1/clients/$CLIENT_ID/reject" "{\"reason\":\"$LONG_REASON\"}" "$TOKEN")"
code="$(echo "$r" | extract_status)"
body="$(echo "$r" | sed -n '/^\r$/,$p')"
ok=0
[[ "$code" -eq 400 ]] && contains_text "$body" "reason maximum length is 1000" && ok=1
print_result "$ok" "reject_reason_panjang" "$code" "expected 400 + reason maximum length is 1000"

log "== 10) approve note >1000 => 400 + maximum length =="
LONG_NOTE="$(printf 'N%.0s' $(seq 1 1101))"
r="$(request_i POST "$BASE_URL/api/v1/clients/$CLIENT_ID/approve" "{\"note\":\"$LONG_NOTE\"}" "$TOKEN")"
code="$(echo "$r" | extract_status)"
body="$(echo "$r" | sed -n '/^\r$/,$p')"
ok=0
[[ "$code" -eq 400 ]] && contains_text "$body" "note maximum length is 1000" && ok=1
print_result "$ok" "approve_note_panjang" "$code" "expected 400 + note maximum length is 1000"

write_json_report

log ""
log "Summary: PASS=$pass FAIL=$fail TOTAL=$((pass+fail))"
log "JSON report saved: $JSON_REPORT"
