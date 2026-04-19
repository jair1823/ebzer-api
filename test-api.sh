#!/bin/bash

# Script de prueba para ebzer-api con SQLite
# Asegúrate de que el servidor esté corriendo antes de ejecutar este script

BASE_URL="http://localhost:3000"

echo "=========================================="
echo "  Pruebas de ebzer-api con SQLite"
echo "=========================================="
echo ""

# Health Checks
echo "1. Probando health checks..."
echo -n "   GET /ping: "
curl -s "$BASE_URL/ping" | grep -q "pong" && echo "✅ OK" || echo "❌ FAIL"

echo -n "   GET /dbping: "
curl -s "$BASE_URL/dbping" | grep -q "Database connection successful" && echo "✅ OK" || echo "❌ FAIL"
echo ""

# Orders API
echo "2. Probando Orders API..."
echo -n "   POST /api/orders: "
ORDER_ID=$(curl -s -X POST "$BASE_URL/api/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Test Order",
    "amount_charged": "100.00",
    "status": "confirmed",
    "delivery_type": "pickup",
    "client_name": "Test Client"
  }' | grep -oP '(?<="id":)\d+')

if [ -n "$ORDER_ID" ]; then
  echo "✅ OK (ID: $ORDER_ID)"
else
  echo "❌ FAIL"
  exit 1
fi

echo -n "   GET /api/orders: "
curl -s "$BASE_URL/api/orders" | grep -q "Test Order" && echo "✅ OK" || echo "❌ FAIL"

echo -n "   GET /api/orders/$ORDER_ID: "
curl -s "$BASE_URL/api/orders/$ORDER_ID" | grep -q "Test Order" && echo "✅ OK" || echo "❌ FAIL"

echo -n "   PUT /api/orders/$ORDER_ID: "
curl -s -X PUT "$BASE_URL/api/orders/$ORDER_ID" \
  -H "Content-Type: application/json" \
  -d '{"status": "in_progress"}' | grep -q "\"updated\":true" && echo "✅ OK" || echo "❌ FAIL"
echo ""

# Incomes API
echo "3. Probando Incomes API..."
echo -n "   POST /api/incomes (50%): "
INCOME_ID=$(curl -s -X POST "$BASE_URL/api/incomes" \
  -H "Content-Type: application/json" \
  -d "{\"order_id\": \"$ORDER_ID\", \"amount\": \"50.00\"}" | grep -oP '(?<="id":)\d+')

if [ -n "$INCOME_ID" ]; then
  echo "✅ OK (ID: $INCOME_ID)"
else
  echo "❌ FAIL"
fi

echo -n "   GET /api/orders/$ORDER_ID/payment-status: "
curl -s "$BASE_URL/api/orders/$ORDER_ID/payment-status" | grep -q "\"percentage_paid\":50" && echo "✅ OK (50% pagado)" || echo "❌ FAIL"

echo -n "   POST /api/incomes (completar pago): "
curl -s -X POST "$BASE_URL/api/incomes" \
  -H "Content-Type: application/json" \
  -d "{\"order_id\": \"$ORDER_ID\", \"amount\": \"50.00\"}" > /dev/null && echo "✅ OK" || echo "❌ FAIL"

echo -n "   Verificar pago completo (100%): "
curl -s "$BASE_URL/api/orders/$ORDER_ID/payment-status" | grep -q "\"percentage_paid\":100" && echo "✅ OK (100% pagado)" || echo "❌ FAIL"

echo -n "   GET /api/incomes: "
curl -s "$BASE_URL/api/incomes" | grep -q "$ORDER_ID" && echo "✅ OK" || echo "❌ FAIL"
echo ""

# Filtros
echo "4. Probando filtros..."
echo -n "   GET /api/orders?status=in_progress: "
curl -s "$BASE_URL/api/orders?status=in_progress" | grep -q "$ORDER_ID" && echo "✅ OK" || echo "❌ FAIL"

echo -n "   GET /api/orders?from=2026-04-01&to=2026-12-31: "
curl -s "$BASE_URL/api/orders?from=2026-04-01&to=2026-12-31" | grep -q "$ORDER_ID" && echo "✅ OK" || echo "❌ FAIL"
echo ""

# Cleanup (opcional)
echo "5. Limpieza (opcional)..."
read -p "   ¿Deseas eliminar los datos de prueba? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  echo -n "   DELETE /api/incomes/$INCOME_ID: "
  curl -s -X DELETE "$BASE_URL/api/incomes/$INCOME_ID" > /dev/null && echo "✅ OK" || echo "❌ FAIL"
  
  echo -n "   DELETE /api/orders/$ORDER_ID: "
  curl -s -X DELETE "$BASE_URL/api/orders/$ORDER_ID" > /dev/null && echo "✅ OK" || echo "❌ FAIL"
fi

echo ""
echo "=========================================="
echo "  Pruebas completadas"
echo "=========================================="
