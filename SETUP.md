# ebzer-api - Guía de Configuración y Prueba

## ✅ Configuración Completada

La aplicación ha sido configurada exitosamente con SQLite y está lista para usar.

## 📋 Prerequisitos Instalados

- ✅ Go 1.25.5
- ✅ GCC 13.3.0 (build-essential)
- ✅ SQLite3 CLI
- ✅ Dependencias Go (go.mod)

## 🚀 Inicio Rápido

### 1. Iniciar el Servidor

```bash
go run cmd/server/main.go
```

El servidor:
- Escuchará en `http://localhost:3000`
- Creará automáticamente la base de datos en `./data/ebzer.db`
- Ejecutará las migraciones automáticamente al iniciar

### 2. Verificar que el Servidor Funciona

```bash
curl http://localhost:3000/ping
# Respuesta esperada: {"message":"pong"}

curl http://localhost:3000/dbping
# Respuesta esperada: {"message":"Database connection successful"}
```

### 3. Ejecutar Pruebas Automatizadas

```bash
./test-api.sh
```

Este script probará:
- Health checks (`/ping`, `/dbping`)
- CRUD de Orders
- CRUD de Incomes
- Cálculo de estado de pago
- Filtros por status y fechas

## 📊 Estructura de la Base de Datos

### Tablas Activas

- **orders** - Gestión de pedidos
- **income** - Registro de pagos (múltiples pagos por orden)
- **expenses** - Registro de gastos
- **expense_categories** - Categorización de gastos

### Tablas Reservadas (para futuras funcionalidades)

- **users** - Gestión de usuarios y autenticación
- **financial_percentages** - Reportes financieros

### Migraciones

Las migraciones se ejecutan automáticamente al iniciar el servidor. El sistema rastrea las migraciones aplicadas en la tabla `schema_migrations`.

## 🔧 Endpoints Disponibles

### Health Checks

```bash
GET /ping
GET /dbping
```

### Orders API

```bash
# Crear orden
POST /api/orders
Content-Type: application/json
{
  "description": "Vestido personalizado",
  "amount_charged": "1500.00",
  "status": "confirmed",
  "estimated_delivery_date": "2026-05-15T18:00:00Z",
  "delivery_type": "pickup",
  "client_name": "María González",
  "client_phone": "+52 123 456 7890",
  "notes": "Color: azul cielo"
}

# Listar órdenes
GET /api/orders
GET /api/orders?status=confirmed
GET /api/orders?from=2026-04-01&to=2026-04-30

# Obtener orden específica
GET /api/orders/:id

# Actualizar orden
PUT /api/orders/:id
Content-Type: application/json
{
  "status": "in_progress"
}

# Ver estado de pago
GET /api/orders/:id/payment-status

# Finalizar orden (marca como delivered)
POST /api/orders/:id/finish

# Eliminar orden
DELETE /api/orders/:id
```

### Incomes API

```bash
# Crear ingreso (pago)
POST /api/incomes
Content-Type: application/json
{
  "order_id": "1",
  "amount": "750.00"
}

# Listar ingresos
GET /api/incomes
GET /api/incomes?from=2026-04-01&to=2026-04-30

# Obtener ingreso específico
GET /api/incomes/:id

# Actualizar ingreso
PUT /api/incomes/:id
Content-Type: application/json
{
  "amount": "800.00"
}

# Eliminar ingreso
DELETE /api/incomes/:id
```

## 🗄️ Gestión de la Base de Datos

### Acceder a la Base de Datos SQLite

```bash
sqlite3 ./data/ebzer.db
```

### Comandos Útiles de SQLite

```sql
-- Listar todas las tablas
.tables

-- Ver migraciones aplicadas
SELECT * FROM schema_migrations;

-- Ver órdenes
SELECT id, description, amount_charged, status FROM orders;

-- Ver ingresos
SELECT id, order_id, amount FROM income;

-- Salir
.quit
```

### Resetear la Base de Datos

```bash
# Detener el servidor primero
rm -rf ./data/ebzer.db

# Reiniciar el servidor (recreará la BD y aplicará migraciones)
go run cmd/server/main.go
```

## 🧪 Flujo de Prueba Manual

### Ejemplo: Orden con Pagos Parciales

```bash
# 1. Crear orden de $1500
curl -X POST http://localhost:3000/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Vestido de quinceañera",
    "amount_charged": "1500.00",
    "status": "confirmed",
    "delivery_type": "pickup",
    "client_name": "Cliente Test"
  }'
# Respuesta: {"id":1}

# 2. Ver estado de pago (0% pagado)
curl http://localhost:3000/api/orders/1/payment-status
# Respuesta: {"total_paid":0,"amount_charged":1500,"remaining":1500,"percentage_paid":0,"is_fully_paid":false}

# 3. Registrar primer pago ($750 - 50%)
curl -X POST http://localhost:3000/api/incomes \
  -H "Content-Type: application/json" \
  -d '{"order_id": "1", "amount": "750.00"}'
# Respuesta: {"id":1}

# 4. Ver estado de pago (50% pagado)
curl http://localhost:3000/api/orders/1/payment-status
# Respuesta: {"total_paid":750,"amount_charged":1500,"remaining":750,"percentage_paid":50,"is_fully_paid":false}

# 5. Registrar segundo pago ($750 - completar)
curl -X POST http://localhost:3000/api/incomes \
  -H "Content-Type: application/json" \
  -d '{"order_id": "1", "amount": "750.00"}'
# Respuesta: {"id":2}

# 6. Ver estado de pago (100% pagado)
curl http://localhost:3000/api/orders/1/payment-status
# Respuesta: {"total_paid":1500,"amount_charged":1500,"remaining":0,"percentage_paid":100,"is_fully_paid":true}

# 7. Actualizar estado de la orden
curl -X PUT http://localhost:3000/api/orders/1 \
  -H "Content-Type: application/json" \
  -d '{"status": "in_progress"}'
# Respuesta: {"updated":true}

# 8. Finalizar orden
curl -X POST http://localhost:3000/api/orders/1/finish
# Respuesta: {"finished":true}
```

## ⚙️ Configuración

### Variables de Entorno

```bash
# Ruta de la base de datos (opcional, default: ./data/ebzer.db)
export SQLITE_DB_PATH=/custom/path/ebzer.db

# Ejemplo con ruta personalizada
SQLITE_DB_PATH=/tmp/test.db go run cmd/server/main.go
```

### Configuración de SQLite

La conexión SQLite está configurada con:
- **Foreign Keys**: Habilitados
- **Journal Mode**: WAL (Write-Ahead Logging)
- **Max Connections**: 1 (óptimo para SQLite)
- **Location Parsing**: Automático (`_loc=auto`)

## 📝 Notas Importantes

### Status Válidos de Órdenes

- `confirmed` - Orden confirmada
- `in_progress` - En proceso
- `ready` - Lista para entrega
- `shipped` - Enviada
- `delivered` - Entregada
- `cancelled` - Cancelada

### Tipos de Entrega Válidos

- `pickup` - Recoger en tienda
- `shipping` - Envío
- `delivery` - Entrega a domicilio

### Formatos de Fecha

Las fechas se manejan en formato ISO 8601:
```
2026-04-18T10:30:00Z
```

Los filtros de fecha aceptan formato simplificado:
```
2026-04-01
```

## ❌ Solución de Problemas

### Puerto 3000 en Uso

```bash
# Encontrar el proceso
lsof -i :3000

# Matar el proceso
kill -9 <PID>
```

### Errores de Compilación

```bash
# Limpiar y reinstalar dependencias
go clean -modcache
go mod download
go mod tidy
```

### Base de Datos Corrupta

```bash
# Eliminar y recrear
rm -rf ./data/ebzer.db
go run cmd/server/main.go
```

## 🎯 Próximos Pasos

La aplicación está lista para:
- ✅ Integración con frontend
- ✅ Agregar más dominios (expenses, categories, etc.)
- ✅ Implementar autenticación (tabla users ya existe)
- ✅ Agregar reportes financieros (tabla financial_percentages está preparada)

## 📚 Arquitectura

El proyecto sigue **Clean Architecture**:

```
cmd/server/          - Punto de entrada
internal/
  db/                - Conexión y migraciones
    migrations/      - Scripts SQL de migraciones
    scanner.go       - Tipos personalizados para fechas SQLite
  orders/            - Dominio de órdenes
    models.go        - Estructuras de datos
    dto.go           - Data Transfer Objects
    repository.go    - Acceso a datos
    service.go       - Lógica de negocio
    handler.go       - Endpoints HTTP
  incomes/           - Dominio de ingresos
    [misma estructura]
```

---

**Estado**: ✅ Aplicación configurada y probada exitosamente en WSL

**Fecha**: 2026-04-18

**Base de Datos**: SQLite 3 (`./data/ebzer.db`)

**Servidor**: http://localhost:3000
