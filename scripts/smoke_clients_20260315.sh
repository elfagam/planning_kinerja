#!/usr/bin/env bash
set -euo pipefail

mysql -uroot -e "USE \`e-plan-ai\`; UPDATE users SET role='OPERATOR', password_hash='\$2a\$10\$A13.hD0DxyCqvXHfUZ3bte.itOLmOYqHeXQTn5Rng7C./11A2YYaK' WHERE id=12; UPDATE users SET password_hash='\$2a\$10\$A13.hD0DxyCqvXHfUZ3bte.itOLmOYqHeXQTn5Rng7C./11A2YYaK' WHERE id=1;"

ADMIN_LOGIN=$(curl -sS -X POST http://127.0.0.1:8081/api/v1/auth/login -H 'Content-Type: application/json' -d '{"email":"superadmin@rsudcontoh.go.id","password":"Admin123!"}')
ADMIN_TOKEN=$(echo "$ADMIN_LOGIN" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')

OP_LOGIN=$(curl -sS -X POST http://127.0.0.1:8081/api/v1/auth/login -H 'Content-Type: application/json' -d '{"email":"operator.uji@rsud.local","password":"Admin123!"}')
OP_TOKEN=$(echo "$OP_LOGIN" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')

echo "ADMIN_TOKEN_LEN=${#ADMIN_TOKEN}"
echo "OP_TOKEN_LEN=${#OP_TOKEN}"

SMOKE_KODE="CL-SMOKE-$(date +%s)"
CREATE_RESP=$(curl -sS -X POST http://127.0.0.1:8081/api/v1/clients -H "Authorization: Bearer $OP_TOKEN" -H 'Content-Type: application/json' -d '{"kode":"'"$SMOKE_KODE"'","nama":"Smoke Test Client 20260315"}')
CLIENT_ID=$(echo "$CREATE_RESP" | sed -n 's/.*"ID":\([0-9][0-9]*\).*/\1/p' | head -n1)

echo "CREATE=$CREATE_RESP"
echo "CLIENT_ID=$CLIENT_ID"

if [ -z "$CLIENT_ID" ]; then
	echo "CLIENT_ID is empty, aborting transition smoke test"
	exit 1
fi

SUBMIT1=$(curl -sS -X POST "http://127.0.0.1:8081/api/v1/clients/$CLIENT_ID/submit" -H "Authorization: Bearer $OP_TOKEN" -H 'Content-Type: application/json' -d '{"note":"submit smoke"}')
UNSUBMIT=$(curl -sS -X POST "http://127.0.0.1:8081/api/v1/clients/$CLIENT_ID/unsubmit" -H "Authorization: Bearer $OP_TOKEN" -H 'Content-Type: application/json' -d '{"reason":"fix data smoke"}')
SUBMIT2=$(curl -sS -X POST "http://127.0.0.1:8081/api/v1/clients/$CLIENT_ID/submit" -H "Authorization: Bearer $OP_TOKEN" -H 'Content-Type: application/json' -d '{"note":"resubmit smoke"}')
APPROVE=$(curl -sS -X POST "http://127.0.0.1:8081/api/v1/clients/$CLIENT_ID/approve" -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' -d '{"note":"approve smoke"}')
HISTORY=$(curl -sS "http://127.0.0.1:8081/api/v1/clients/$CLIENT_ID/status-history" -H "Authorization: Bearer $ADMIN_TOKEN")

echo "SUBMIT1=$SUBMIT1"
echo "UNSUBMIT=$UNSUBMIT"
echo "SUBMIT2=$SUBMIT2"
echo "APPROVE=$APPROVE"
echo "HISTORY=$HISTORY"

mysql -uroot -e "USE \`e-plan-ai\`; SELECT id,client_id,action,from_status,to_status,actor_id,actor_name FROM client_status_histories WHERE client_id=$CLIENT_ID ORDER BY id;"
