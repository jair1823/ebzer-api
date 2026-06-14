# Estado Actual - Ebzer API

**Fecha:** 18 de abril, 2026  
**Última actualización:** 19 de abril, 2026 (Campo `client_name` ahora requerido)  
**Estado:** ✅ **READY FOR TESTING & UI INTEGRATION**

---

## 📊 Resumen Ejecutivo

### ✅ Cambios Implementados (Completados)

| Fase | Descripción | Estado |
|------|-------------|--------|
| **1** | Fix constraint UNIQUE en `income.order_id` | ✅ Completado |
| **2** | Eliminado campo `paid_50_percent` | ✅ Completado |
| **3** | Rediseño de estados operacionales | ✅ Completado |
| **3b** | Implementado `PaymentStatus` calculado | ✅ Completado |
| **4** | Limpieza tablas huérfanas | 🟡 Parcial (ver abajo) |
| **5** | **Migración a SQLite** | ✅ **Completado** |
| **6** | **Campo `client_name` requerido** | ✅ **Completado** (migración original actualizada) |

### 🎯 Estado de las Tablas

| Tabla | Estado | Uso Actual |
|-------|--------|------------|
| `orders` | ✅ Activa | Gestión de órdenes |
| `income` | ✅ Activa | Registro de pagos (múltiples por orden) |
| `expenses` | ✅ Activa | Registro de gastos |
| `expense_categories` | ✅ Activa | Categorías de gastos |
| `users` | 🟡 Inactiva | Mantenida para futuro auth |
| `financial_percentages` | 🟡 Inactiva | Mantenida para futuros reportes |
| `history_order_status` | ❌ Eliminada | Será re-agregada post-estabilización |

---

## 🗄️ Estructura de Datos Actual

### Diagrama de Relaciones

```
┌─────────────────────────┐
│      orders             │
│─────────────────────────│
│ id (PK)                 │
│ description             │
│ amount_charged          │◄──────┐
│ status (enum)           │       │
│ entry_date              │       │
│ estimated_delivery_date │       │
│ delivery_type (enum)    │       │
│ client_name             │       │
│ client_phone            │       │
│ notes                   │       │
│ created_at              │       │
│ updated_at              │       │
└─────────────────────────┘       │
           △                       │
           │                       │
           │ FK (order_id)         │
           │                       │
┌─────────────────────────┐       │
│       income            │       │
│─────────────────────────│       │
│ id (PK)                 │       │
│ order_id (FK)           │───────┘
│ amount                  │
│ date                    │
│ created_at              │
│ updated_at              │
└─────────────────────────┘
    ▲
    │ 0..N incomes por order
    │ ✅ Soporta pagos parciales


┌─────────────────────────┐
│  expense_categories     │
│─────────────────────────│
│ id (PK)                 │◄──────┐
│ name                    │       │
│ description             │       │
│ created_at              │       │
└─────────────────────────┘       │
                                   │
                                   │ FK (category_id)
                                   │
                        ┌─────────────────────────┐
                        │      expenses           │
                        │─────────────────────────│
                        │ id (PK)                 │
                        │ description             │
                        │ amount                  │
                        │ category_id (FK)        │
                        │ date                    │
                        │ created_at              │
                        │ updated_at              │
                        └─────────────────────────┘
```

---

## 📐 Schema SQL Actualizado

### Tabla: `orders`

```sql
CREATE TYPE order_status AS ENUM (
    'new',        -- Orden confirmada, pendiente de iniciar
    'active',     -- En producción
    'ready',      -- Terminada, lista para entregar
    'completed',  -- Entregada al cliente
    'cancelled'   -- Cancelada
);

CREATE TYPE delivery_type AS ENUM (
    'pickup',
    'shipping',
    'delivery'
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    amount_charged NUMERIC(10,2) NOT NULL,
    status order_status NOT NULL DEFAULT 'confirmed',
    entry_date TIMESTAMP NOT NULL DEFAULT NOW(),
    estimated_delivery_date TIMESTAMP NULL,
    delivery_type delivery_type NOT NULL DEFAULT 'pickup',
    client_name VARCHAR(255) NULL,
    client_phone VARCHAR(20) NULL,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_orders_entry_date ON orders (entry_date);
CREATE INDEX idx_orders_estimated_delivery_date ON orders (estimated_delivery_date);
```

**✅ Cambios respecto al schema anterior:**
- ❌ Eliminado campo `paid_50_percent`
- ✅ Estados redefinidos como puramente operacionales
- ✅ Estado por defecto: `'confirmed'` (antes era `'pending'`)

---

### Tabla: `income`

```sql
CREATE TABLE income (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    amount NUMERIC(10,2) NOT NULL,
    date TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_income_order_id ON income (order_id);
CREATE INDEX idx_income_date ON income (date);
```

**✅ Cambios respecto al schema anterior:**
- ❌ Eliminado constraint `UNIQUE` en `order_id`
- ✅ Agregado índice no-único `idx_income_order_id`
- ✅ Ahora permite **múltiples income records por orden**

---

### Tablas: `expenses` y `expense_categories`

```sql
CREATE TABLE expense_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE expenses (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    category_id INTEGER NOT NULL REFERENCES expense_categories(id),
    date TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_expenses_date ON expenses (date);
CREATE INDEX idx_expenses_category_id ON expenses (category_id);
```

**Sin cambios** respecto al schema anterior.

---

## 🔄 Flujo de Datos: Estado Operacional vs Financiero

### Separación de Conceptos

```
┌─────────────────────────────────────────────────────────────┐
│                    Order Lifecycle                           │
│                  (ESTADO OPERACIONAL)                         │
└─────────────────────────────────────────────────────────────┘
                         │
                         ▼
        confirmed → in_progress → ready → shipped → delivered
                         │
                         ├─→ cancelled (en cualquier momento)
                         │
                         ▼
            ✅ Estados NO indican pago


┌─────────────────────────────────────────────────────────────┐
│                  Payment Lifecycle                           │
│                  (ESTADO FINANCIERO)                         │
└─────────────────────────────────────────────────────────────┘
                         │
                         ▼
                Se calcula en runtime:
                         │
            ┌────────────┼────────────┐
            │            │            │
        not_paid   partially_paid  fully_paid
         (0%)        (1-99%)        (100%)
            │            │            │
            └────────────┴────────────┘
                         │
            ✅ Basado en SUM(income.amount)
               vs orders.amount_charged
```

### Ejemplo de Escenario Real

```
Orden #123:
- Description: "Set de 12 tazas personalizadas con logo"
- Amount Charged: $1000
- Status: in_progress  (OPERACIONAL)

Incomes:
- Income #1: $500  (anticipo, 18/04/2026)
- Income #2: $300  (segundo pago, 20/04/2026)
- Total Paid: $800

Payment Status (CALCULADO):
- Total Paid: $800
- Amount Charged: $1000
- Remaining: $200
- Percentage Paid: 80%
- Is Fully Paid: false

Estado Final:
✅ Order.Status = "in_progress" (aún en producción)
✅ Payment Status = "partially_paid" (80% pagado)
```

---

## 🎯 API Endpoints Actuales

### Orders Module

| Método | Endpoint | Descripción | Body/Params |
|--------|----------|-------------|-------------|
| `POST` | `/api/orders` | Crear nueva orden | `CreateOrderDTO` |
| `GET` | `/api/orders` | Listar órdenes | Query: `status`, `from`, `to` |
| `GET` | `/api/orders/:id` | Obtener detalle de orden | Param: `id` |
| `GET` | `/api/orders/:id/payment-status` | **NUEVO**: Estado de pago | Param: `id` |
| `PUT` | `/api/orders/:id` | Actualizar orden | `UpdateOrderDTO` |
| `POST` | `/api/orders/:id/finish` | Finalizar orden (→ `delivered`) | Param: `id` |
| `DELETE` | `/api/orders/:id` | Eliminar orden | Param: `id` |

### Incomes Module

| Método | Endpoint | Descripción | Body/Params |
|--------|----------|-------------|-------------|
| `POST` | `/api/incomes` | Registrar ingreso | `CreateIncomeDTO` |
| `GET` | `/api/incomes` | Listar ingresos | Query: `from`, `to` |
| `GET` | `/api/incomes/:id` | Obtener detalle de ingreso | Param: `id` |
| `PUT` | `/api/incomes/:id` | Actualizar ingreso | `UpdateIncomeDTO` |
| `DELETE` | `/api/incomes/:id` | Eliminar ingreso | Param: `id` |

---

## 📦 Contratos de API (DTOs)

### Orders

#### CreateOrderDTO

```json
{
  "description": "string (required)",
  "amount_charged": "number|string (required)",
  "status": "order_status (optional, default: confirmed)",
  "estimated_delivery_date": "timestamp (optional)",
  "delivery_type": "delivery_type (optional, default: pickup)",
  "client_name": "string (optional)",
  "client_phone": "string (optional)",
  "notes": "string (optional)"
}
```

**Cambios:**
- ❌ Eliminado: `paid_50_percent`

#### UpdateOrderDTO

```json
{
  "description": "string (optional)",
  "amount_charged": "number|string (optional)",
  "status": "order_status (optional)",
  "estimated_delivery_date": "timestamp (optional)",
  "delivery_type": "delivery_type (optional)",
  "client_name": "string (optional)",
  "client_phone": "string (optional)",
  "notes": "string (optional)"
}
```

**Cambios:**
- ❌ Eliminado: `paid_50_percent`

#### PaymentStatus (Response) **NUEVO**

```json
{
  "total_paid": 800.00,
  "amount_charged": 1000.00,
  "remaining": 200.00,
  "percentage_paid": 80.00,
  "is_fully_paid": false
}
```

Endpoint: `GET /api/orders/:id/payment-status`

---

### Incomes

#### CreateIncomeDTO

```json
{
  "order_id": "number (required)",
  "amount": "number (required)"
}
```

**Sin cambios**.

#### UpdateIncomeDTO

```json
{
  "order_id": "number (optional)",
  "amount": "number (optional)"
}
```

**Sin cambios**.

---

## 🏗️ Arquitectura Backend

### Estructura de Capas

```
┌────────────────────────────────────────────────┐
│              HTTP Layer (Fiber)                 │
│  ┌──────────────┐      ┌──────────────┐       │
│  │    Orders    │      │   Incomes    │       │
│  │   Handler    │      │   Handler    │       │
│  └──────────────┘      └──────────────┘       │
└────────────────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────────────┐
│             Business Logic Layer                │
│  ┌──────────────┐      ┌──────────────┐       │
│  │    Orders    │      │   Incomes    │       │
│  │   Service    │◄────►│   Service    │       │
│  └──────────────┘      └──────────────┘       │
└────────────────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────────────┐
│            Data Access Layer                    │
│  ┌──────────────┐      ┌──────────────┐       │
│  │    Orders    │      │   Incomes    │       │
│  │  Repository  │      │  Repository  │       │
│  └──────────────┘      └──────────────┘       │
└────────────────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────────────┐
│              PostgreSQL Database                │
└────────────────────────────────────────────────┘
```

### Flujo de Dependencias

```
main.go
  │
  ├─→ DB Connection (pgx pool)
  │
  ├─→ Orders Module
  │     │
  │     ├─→ OrdersRepository(db)
  │     ├─→ IncomesRepository(db)
  │     ├─→ OrdersService(ordersRepo, incomesRepo)  ← ✅ NUEVO: inyecta income repo
  │     └─→ OrdersHandler(ordersService)
  │
  └─→ Incomes Module
        │
        ├─→ IncomesRepository(db)
        ├─→ IncomesService(incomesRepo)
        └─→ IncomesHandler(incomesService)
```

**✅ Cambio clave:**
- `OrdersService` ahora depende de `IncomesRepository` para calcular `PaymentStatus`

---

## 🧪 Próximos Pasos

### 1. Testing

```
Prioridad:
  ├─→ Unit Tests para Services (lógica de negocio)
  ├─→ Integration Tests para Repositories (queries SQL)
  └─→ E2E Tests para Handlers (endpoints HTTP)

Casos críticos a testear:
  ├─→ Múltiples income records por orden
  ├─→ Cálculo correcto de PaymentStatus
  ├─→ Transiciones de estado operacional
  └─→ Validación de reglas de negocio
```

### 2. Integración con UI

```
Frontend necesita:
  ├─→ Actualizar DTOs (eliminar paid_50_percent)
  ├─→ Consumir nuevo endpoint /payment-status
  ├─→ Mostrar estados operacionales nuevos
  │     confirmed | in_progress | ready | shipped | delivered | cancelled
  │
  └─→ Mostrar estados financieros calculados
        not_paid | partially_paid | fully_paid
```

### 3. Validaciones Pendientes

- [ ] Validar transiciones de estados válidas (ej: no saltar de `confirmed` a `delivered`)
- [ ] Agregar tests de regresión para schema changes
- [ ] Documentar variables de entorno
- [ ] Revisar seguridad de CORS para producción

---

## 🚀 Cómo Ejecutar

### Prerequisitos

```bash
# Go 1.25.5 instalado
# SQLite3 (opcional, solo para inspección manual)
```

### Pasos

```bash
# 1. Clonar repositorio
git clone <repo>

# 2. Instalar dependencias
go mod download

# 3. Iniciar servidor
go run cmd/server/main.go

# La aplicación automáticamente:
# - Crea la base de datos en ./data/ebzer.db
# - Ejecuta todas las migrations pendientes
# - Inicia el servidor HTTP

# 4. Healthcheck
curl http://localhost:3000/ping
# → {"message":"pong"}

curl http://localhost:3000/dbping
# → {"message":"Database connection successful"}
```

### Variables de Entorno (Opcionales)

```bash
# Cambiar ubicación de la base de datos
export SQLITE_DB_PATH="/custom/path/ebzer.db"
```

**Ver:** [Guía de Migración a SQLite](./migracion-sqlite.md) para detalles completos

---

## ✅ Checklist de Preparación

### Backend
- [x] Schema corregido (income sin UNIQUE)
- [x] Estados operacionales puros
- [x] Campo `paid_50_percent` eliminado
- [x] PaymentStatus calculado implementado
- [x] Endpoint `/payment-status` creado
- [x] Repository `GetByOrderID()` para incomes
- [ ] Tests unitarios
- [ ] Tests de integración
- [ ] Validaciones de transición de estados

### Documentación
- [x] Estado actual con diagramas
- [x] Schema SQL actualizado
- [x] Contratos de API documentados
- [x] Flujos de datos explicados
- [ ] Variables de entorno
- [ ] Guía de desarrollo

### Integración UI
- [ ] Actualizar DTOs en frontend
- [ ] Implementar consumo de `/payment-status`
- [ ] Actualizar UI para nuevos estados
- [ ] Testing E2E frontend-backend

---

**Última actualización:** 18 de abril, 2026  
**Estado:** ✅ Ready for Testing & UI Integration
