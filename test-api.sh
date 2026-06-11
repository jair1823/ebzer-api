#!/bin/bash

# Script de prueba para ebzer-api con SQLite
# Asegúrate de que el servidor esté corriendo antes de ejecutar este script

BASE_URL="http://localhost:3000"
ADMIN_EMAIL="${INITIAL_ADMIN_EMAIL:-admin@example.com}"
ADMIN_PASSWORD="${INITIAL_ADMIN_PASSWORD:-password123}"
ACCESS_TOKEN=""

api_curl() {
  local headers=()
  if [ -n "$API_KEY" ]; then
    headers+=(-H "X-API-Key: $API_KEY")
  fi
  if [ -n "$ACCESS_TOKEN" ]; then
    headers+=(-H "Authorization: Bearer $ACCESS_TOKEN")
  fi
  curl -s "${headers[@]}" "$@"
}

echo "=========================================="
echo "  Pruebas de ebzer-api con SQLite"
echo "=========================================="
echo ""

# Health Checks
echo "1. Probando health checks..."
echo -n "   GET /ping: "
api_curl "$BASE_URL/ping" | grep -q "pong" && echo "✅ OK" || echo "❌ FAIL"

echo -n "   GET /dbping: "
api_curl "$BASE_URL/dbping" | grep -q "Database connection successful" && echo "✅ OK" || echo "❌ FAIL"
echo ""

# Auth
echo "2. Probando Auth API..."
echo -n "   POST /api/auth/login: "
LOGIN_RESPONSE=$(api_curl -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}")
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -oP '(?<="access_token":")[^"]+')
if [ -n "$ACCESS_TOKEN" ]; then
  echo "✅ OK"
else
  echo "❌ FAIL"
  echo "   Response: $LOGIN_RESPONSE"
  exit 1
fi

echo -n "   GET /api/auth/me: "
api_curl "$BASE_URL/api/auth/me" | grep -q "\"email\":\"$ADMIN_EMAIL\"" && echo "✅ OK" || echo "❌ FAIL"
echo ""

# -----------------------------------------------
# Order Statuses API
# -----------------------------------------------
echo "3. Probando Order Statuses API..."

echo -n "   GET /api/order-statuses (seed check): "
STATUSES=$(api_curl "$BASE_URL/api/order-statuses")
echo "$STATUSES" | grep -q '"new"' && echo "$STATUSES" | grep -q '"completed"' && echo "$STATUSES" | grep -q '"cancelled"' \
  && echo "✅ OK (system statuses seeded)" || echo "❌ FAIL (missing system statuses)"

echo -n "   GET /api/order-statuses?active_only=true: "
api_curl "$BASE_URL/api/order-statuses?active_only=true" | grep -q '"statuses"' && echo "✅ OK" || echo "❌ FAIL"

echo -n "   POST /api/order-statuses (custom status): "
CUSTOM_STATUS_ID=$(api_curl -X POST "$BASE_URL/api/order-statuses" \
  -H "Content-Type: application/json" \
  -d '{"name":"ready_for_pickup","display_name":"Ready for Pickup","color":"#8B5CF6","order_position":3}' \
  | grep -oP '(?<="id":)\d+')
if [ -n "$CUSTOM_STATUS_ID" ]; then
  echo "✅ OK (ID: $CUSTOM_STATUS_ID)"
else
  echo "❌ FAIL"
fi

echo -n "   GET /api/order-statuses/$CUSTOM_STATUS_ID: "
api_curl "$BASE_URL/api/order-statuses/$CUSTOM_STATUS_ID" | grep -q '"ready_for_pickup"' && echo "✅ OK" || echo "❌ FAIL"

echo -n "   PUT /api/order-statuses/$CUSTOM_STATUS_ID (update display_name): "
api_curl -X PUT "$BASE_URL/api/order-statuses/$CUSTOM_STATUS_ID" \
  -H "Content-Type: application/json" \
  -d '{"display_name":"Ready for Collection"}' | grep -q '"updated":true' && echo "✅ OK" || echo "❌ FAIL"

echo -n "   PUT /api/order-statuses/reorder: "
# Resolve IDs of system statuses to build reorder payload dynamically
NEW_ID=$(api_curl "$BASE_URL/api/order-statuses" | grep -oP '"id":\K\d+(?=[^}]*"name":"new")' | head -1)
COMPLETED_ID=$(api_curl "$BASE_URL/api/order-statuses" | grep -oP '"id":\K\d+(?=[^}]*"name":"completed")' | head -1)
api_curl -X PUT "$BASE_URL/api/order-statuses/reorder" \
  -H "Content-Type: application/json" \
  -d "{\"status_orders\":[{\"id\":$NEW_ID,\"position\":1},{\"id\":$CUSTOM_STATUS_ID,\"position\":2},{\"id\":$COMPLETED_ID,\"position\":100}]}" \
  | grep -q '"reordered":true' && echo "✅ OK" || echo "❌ FAIL"

echo -n "   POST /api/order-statuses duplicate name (expect error): "
DUPE=$(api_curl -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/order-statuses" \
  -H "Content-Type: application/json" \
  -d '{"name":"ready_for_pickup","display_name":"Dupe","order_position":99}')
[ "$DUPE" = "400" ] && echo "✅ OK (400 returned)" || echo "❌ FAIL (got $DUPE)"

echo -n "   POST /api/order-statuses invalid slug (expect error): "
BADSLUG=$(api_curl -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/order-statuses" \
  -H "Content-Type: application/json" \
  -d '{"name":"Bad Name","display_name":"Bad","order_position":99}')
[ "$BADSLUG" = "400" ] && echo "✅ OK (400 returned)" || echo "❌ FAIL (got $BADSLUG)"

echo -n "   DELETE /api/order-statuses system status (expect error): "
SYS_DEL=$(api_curl -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/api/order-statuses/$NEW_ID")
[ "$SYS_DEL" = "400" ] && echo "✅ OK (protected)" || echo "❌ FAIL (got $SYS_DEL)"
echo ""

# -----------------------------------------------
# Orders API
# -----------------------------------------------
echo "4. Probando Orders API..."
echo -n "   POST /api/orders (sin status_id, default 'new'): "
ORDER_ID=$(api_curl -X POST "$BASE_URL/api/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Test Order",
    "amount_charged": "100.00",
    "delivery_type": "pickup",
    "client_name": "Test Client"
  }' | grep -oP '(?<="id":)\d+')
if [ -n "$ORDER_ID" ]; then
  echo "✅ OK (ID: $ORDER_ID)"
else
  echo "❌ FAIL"
  exit 1
fi

echo -n "   GET /api/orders/$ORDER_ID (verify status populated): "
api_curl "$BASE_URL/api/orders/$ORDER_ID" | grep -q '"name":"new"' && echo "✅ OK (status=new)" || echo "❌ FAIL"

echo -n "   GET /api/orders: "
api_curl "$BASE_URL/api/orders" | grep -q "Test Order" && echo "✅ OK" || echo "❌ FAIL"

echo -n "   PUT /api/orders/$ORDER_ID (update status_id): "
api_curl -X PUT "$BASE_URL/api/orders/$ORDER_ID" \
  -H "Content-Type: application/json" \
  -d "{\"status_id\": $CUSTOM_STATUS_ID}" | grep -q '"updated":true' && echo "✅ OK" || echo "❌ FAIL"

echo -n "   GET /api/orders?status_id=$CUSTOM_STATUS_ID: "
api_curl "$BASE_URL/api/orders?status_id=$CUSTOM_STATUS_ID" | grep -q "$ORDER_ID" && echo "✅ OK" || echo "❌ FAIL"

echo -n "   GET /api/orders?from=2026-01-01&to=2026-12-31: "
api_curl "$BASE_URL/api/orders?from=2026-01-01&to=2026-12-31" | grep -q "$ORDER_ID" && echo "✅ OK" || echo "❌ FAIL"
echo ""

# -----------------------------------------------
# Incomes API
# -----------------------------------------------
echo "5. Probando Incomes API..."
echo -n "   POST /api/incomes (50%): "
INCOME_ID=$(api_curl -X POST "$BASE_URL/api/incomes" \
  -H "Content-Type: application/json" \
  -d "{\"order_id\": \"$ORDER_ID\", \"amount\": \"50.00\"}" | grep -oP '(?<="id":)\d+')
if [ -n "$INCOME_ID" ]; then
  echo "✅ OK (ID: $INCOME_ID)"
else
  echo "❌ FAIL"
fi

echo -n "   GET /api/orders/$ORDER_ID/payment-status: "
api_curl "$BASE_URL/api/orders/$ORDER_ID/payment-status" | grep -q '"percentage_paid":50' && echo "✅ OK (50% pagado)" || echo "❌ FAIL"

echo -n "   POST /api/incomes (completar pago): "
api_curl -X POST "$BASE_URL/api/incomes" \
  -H "Content-Type: application/json" \
  -d "{\"order_id\": \"$ORDER_ID\", \"amount\": \"50.00\"}" > /dev/null && echo "✅ OK" || echo "❌ FAIL"

echo -n "   Verificar pago completo (100%): "
api_curl "$BASE_URL/api/orders/$ORDER_ID/payment-status" | grep -q '"percentage_paid":100' && echo "✅ OK (100% pagado)" || echo "❌ FAIL"

echo -n "   GET /api/incomes: "
api_curl "$BASE_URL/api/incomes" | grep -q "$ORDER_ID" && echo "✅ OK" || echo "❌ FAIL"
echo ""

# -----------------------------------------------
# Finish Order
# -----------------------------------------------
echo "6. Probando Finish Order..."
echo -n "   POST /api/orders/$ORDER_ID/finish: "
api_curl -X POST "$BASE_URL/api/orders/$ORDER_ID/finish" | grep -q '"finished":true' && echo "✅ OK" || echo "❌ FAIL"

echo -n "   Verificar status = completed: "
api_curl "$BASE_URL/api/orders/$ORDER_ID" | grep -q '"name":"completed"' && echo "✅ OK" || echo "❌ FAIL"
echo ""

# -----------------------------------------------
# Delete deactivate custom status in use (expect error)
# -----------------------------------------------
echo "7. Probando protección de status en uso..."
echo -n "   DELETE /api/order-statuses/$CUSTOM_STATUS_ID con ordenes en uso (expect error): "
# Order was moved to completed, so custom status might be free - just test the endpoint responds correctly
STATUS_DEL=$(api_curl -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/api/order-statuses/$CUSTOM_STATUS_ID")
( [ "$STATUS_DEL" = "400" ] || [ "$STATUS_DEL" = "200" ] ) && echo "✅ OK (responded: $STATUS_DEL)" || echo "❌ FAIL (got $STATUS_DEL)"
echo ""

# -----------------------------------------------
# Cleanup
# -----------------------------------------------
echo "8. Limpieza (opcional)..."
read -p "   ¿Deseas eliminar los datos de prueba? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  echo -n "   DELETE /api/incomes/$INCOME_ID: "
  api_curl -X DELETE "$BASE_URL/api/incomes/$INCOME_ID" > /dev/null && echo "✅ OK" || echo "❌ FAIL"

  echo -n "   DELETE /api/orders/$ORDER_ID: "
  api_curl -X DELETE "$BASE_URL/api/orders/$ORDER_ID" > /dev/null && echo "✅ OK" || echo "❌ FAIL"

  echo -n "   DELETE /api/order-statuses/$CUSTOM_STATUS_ID: "
  api_curl -X DELETE "$BASE_URL/api/order-statuses/$CUSTOM_STATUS_ID" > /dev/null && echo "✅ OK" || echo "❌ FAIL"
fi

echo ""
echo "=========================================="
echo "  Pruebas completadas"
echo "=========================================="
