# Modelo de Datos - Ebzer API

**Fecha de revisión:** 18 de abril, 2026  
**Estado:** ✅ **SCHEMA ACTUALIZADO Y CORREGIDO**

> **Nota:** Para ver el historial de problemas identificados y soluciones aplicadas, consulta [Decisiones de Schema](./decisiones-schema.md)

---

## Resumen del Dominio

**Ebzer** es un sistema de control operacional y financiero para un negocio de productos personalizados.

### Conceptos Clave del Negocio

1. **Order (Orden)** = Venta confirmada, NO es cotización
2. **Income (Ingreso)** = Movimiento financiero de entrada vinculado a una orden
3. **Expense (Gasto)** = Movimiento financiero de salida (puede o no estar vinculado a una orden)
4. **Expense Category** = Clasificación de gastos para reporting

### Reglas de Negocio Críticas

✅ **Una orden puede tener 0, 1, o MUCHOS ingresos** (pagos parciales)  
✅ **La entrega de la orden NO equivale a pago completo**  
✅ **Los gastos pueden ser generales o vinculados a órdenes específicas**  
✅ **Solo se registran gastos PAGADOS (no obligaciones pendientes)**  
✅ **El estado operacional y financiero están SEPARADOS**  

---

## Tablas del Sistema

### 1. `orders` - Órdenes de Trabajo

**Propósito:** Registrar órdenes confirmadas de productos personalizados.

```sql
CREATE TYPE order_status AS ENUM (
    'new',        -- Orden confirmada, pendiente de iniciar
    'active',     -- En producción
    'ready',      -- Terminada, lista para entregar
    'completed',  -- Entregada al cliente
    'cancelled'   -- Cancelada
);

CREATE TYPE delivery_type AS ENUM (
    'pickup',       -- Retiro en local
    'shipping',     -- Envío
    'delivery'      -- Entrega a domicilio
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    amount_charged NUMERIC(10,2) NOT NULL,
    status order_status NOT NULL DEFAULT 'new',
    entry_date TIMESTAMP NOT NULL DEFAULT NOW(),
    estimated_delivery_date TIMESTAMP NULL,
    delivery_type delivery_type NOT NULL DEFAULT 'pickup',
    client_name VARCHAR(255) NOT NULL,
    client_phone VARCHAR(20) NULL,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_orders_entry_date ON orders (entry_date);
CREATE INDEX idx_orders_estimated_delivery_date ON orders (estimated_delivery_date);
```

**Estados Operacionales:**
- `new` - Orden confirmada, pendiente de iniciar producción
- `active` - Orden en proceso de producción/trabajo
- `ready` - Orden completada, lista para ser entregada/retirada
- `completed` - Orden entregada al cliente (estado final)
- `cancelled` - Orden cancelada (puede ocurrir en cualquier etapa)

**Tipos de entrega:**
- `pickup` - Cliente retira en local
- `shipping` - Envío por courier
- `delivery` - Entrega a domicilio

**Índices:**
- `idx_orders_status` en `status` - Para filtrar por estado
- `idx_orders_entry_date` en `entry_date` - Para ordenar cronológicamente
- `idx_orders_estimated_delivery_date` en `estimated_delivery_date` - Para buscar por fecha estimada

#### ✅ Mejoras Aplicadas

- ✅ **Estados operacionales simplificados** - 4 estados reflejan el flujo real de un emprendimiento pequeño: new, active, ready, completed, cancelled
- ✅ **Separación financiera** - El estado de pago se calcula separadamente, no está en el status de la orden
- ✅ **Eliminado `paid_50_percent`** - El porcentaje de pago se calcula en runtime agregando los income records

---

### 2. `income` - Ingresos

**Propósito:** Registrar movimientos financieros de entrada vinculados a órdenes.

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

**Índices:**
- `idx_income_order_id` en `order_id` - Para buscar todos los pagos de una orden
- `idx_income_date` en `date` - Para filtrar por rango de fechas

**Relaciones:**
- `income.order_id` → `orders.id` (FK con CASCADE DELETE)

#### ✅ Corrección Aplicada

- ✅ **Eliminado constraint UNIQUE en `order_id`** - Ahora una orden puede tener múltiples income records
- ✅ **Agregado índice no-único** - Permite búsquedas eficientes manteniendo la capacidad de múltiples pagos
- ✅ **Soporta pagos parciales** - El comportamiento correcto según las reglas de negocio

**Ejemplos de uso:**

```sql
-- Orden con pago en 3 cuotas
INSERT INTO income (order_id, amount, date) VALUES (123, 500.00, '2026-04-01');  -- Anticipo
INSERT INTO income (order_id, amount, date) VALUES (123, 300.00, '2026-04-10');  -- Segunda cuota
INSERT INTO income (order_id, amount, date) VALUES (123, 200.00, '2026-04-20');  -- Cuota final

-- Calcular total pagado de una orden
SELECT SUM(amount) as total_paid FROM income WHERE order_id = 123;
-- → 1000.00
```

---

### 3. `expenses` - Gastos

**Propósito:** Registrar gastos pagados del negocio.

```sql
CREATE TABLE expenses (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    date TIMESTAMP NOT NULL DEFAULT NOW(),
    order_id INTEGER REFERENCES orders(id) ON DELETE SET NULL,
    category_id INTEGER REFERENCES expense_categories(id) ON DELETE SET NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('general', 'order_linked')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Índices:**
- `idx_expenses_date` en `date`
- `idx_expenses_order_id` en `order_id`
- `idx_expenses_category_id` en `category_id`

**Relaciones:**
- `expenses.order_id` → `orders.id` (FK con SET NULL, **opcional**)
- `expenses.category_id` → `expense_categories.id` (FK con SET NULL, **opcional**)

**Tipos de gasto:**
- `general` - Gasto no vinculado a orden específica (ej: alquiler, servicios)
- `order_linked` - Gasto vinculado a orden específica (ej: materiales para pedido X)

#### ✅ Bien Diseñado

- ✅ Soporta gastos generales y específicos de órdenes
- ✅ `ON DELETE SET NULL` correcto (si se borra orden, gasto permanece pero sin vínculo)
- ✅ Categorización opcional permite reporting
- ✅ Solo registra gastos PAGADOS (según fase actual)

#### 💡 Consideración Futura

El diseño actual cierra algunas puertas para futuras features:
- ⚠️ No hay campo para **fecha de vencimiento** (si quieres soportar facturas pendientes)
- ⚠️ No hay campo de **estado** (paid/pending/overdue)
- ⚠️ No hay soporte para **gastos recurrentes**

**Recomendación:** El diseño es adecuado para la fase actual. Si en el futuro necesitas:
- Facturas por pagar
- Gastos recurrentes
- Control de vencimientos

Entonces deberás agregar:
```sql
-- Futura evolución (NO implementar ahora)
ALTER TABLE expenses ADD COLUMN due_date TIMESTAMP NULL;
ALTER TABLE expenses ADD COLUMN status VARCHAR(20) DEFAULT 'paid';
ALTER TABLE expenses ADD COLUMN is_recurring BOOLEAN DEFAULT FALSE;
```

---

### 4. `expense_categories` - Categorías de Gastos

**Propósito:** Clasificar gastos para análisis y reporting.

```sql
CREATE TABLE expense_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Relaciones:**
- `expense_categories.id` ← `expenses.category_id` (1:N)

#### ✅ Bien Diseñado

- ✅ Catálogo simple y efectivo
- ✅ Sin constraints innecesarias
- ✅ Permite evolución fácil (agregar categorías dinámicamente)

**Ejemplos de categorías sugeridas:**
- Materiales
- Herramientas
- Servicios (luz, agua, internet)
- Alquiler
- Transporte
- Marketing
- Administrativos

---

### 5. `users` - Usuarios del Sistema

**Propósito:** Control de acceso y autenticación.

```sql
CREATE TYPE user_role AS ENUM ('admin', 'employee');

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role user_role NOT NULL DEFAULT 'employee',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Roles:**
- `admin` - Acceso completo
- `employee` - Acceso limitado (según implementación futura)

#### ⚠️ Problemas y Consideraciones

1. **CRÍTICO:** Esta tabla existe pero NO SE USA en el código actual
   - ❌ No hay endpoints de login
   - ❌ No hay middleware de autenticación
   - ❌ No hay autorización basada en roles
   - **Estado:** Tabla "huérfana"

2. **Seguridad:** La tabla está preparada para auth, pero falta toda la implementación

**Recomendación:**
- Si NO vas a implementar auth en corto plazo (1-2 semanas) → **ELIMINAR** esta tabla
- Si SÍ vas a implementar → Priorizar:
  1. Endpoint de registro/login
  2. JWT middleware
  3. Role-based access control (RBAC)

---

### 6. `financial_percentages` - Porcentajes de Distribución Financiera

**Propósito:** Definir cómo se distribuyen los ingresos (reinversión, insumos, ganancia).

```sql
CREATE TABLE financial_percentages (
    id SERIAL PRIMARY KEY,
    reinvestment_percentage NUMERIC(5,2) NOT NULL,
    supplies_percentage NUMERIC(5,2) NOT NULL,
    profit_percentage NUMERIC(5,2) NOT NULL,
    effective_start_date DATE NOT NULL
);
```

**Diseño:**
- Permite múltiples registros con fechas de vigencia
- Soporta cambios históricos de porcentajes

#### ⚠️ Problemas y Consideraciones

1. **CRÍTICO:** Esta tabla existe pero NO SE USA en el código actual
   - ❌ No hay endpoints para CRUD de financial_percentages
   - ❌ No hay lógica que calcule distribución de ingresos
   - ❌ No hay reportes que usen estos porcentajes
   - **Estado:** Tabla "huérfana"

2. **Validación faltante:** No hay constraint que valide que la suma sea 100%
   ```sql
   -- Falta agregar:
   CHECK (reinvestment_percentage + supplies_percentage + profit_percentage = 100.00)
   ```

3. **Ambigüedad:** ¿Cómo se usa `effective_start_date`?
   - ¿El registro más reciente es el activo?
   - ¿Se puede tener múltiples activos?
   - ¿Hay un proceso de "activación"?

**Recomendación:**
- Si NO se va a usar en corto plazo → **ELIMINAR** tabla
- Si SÍ se va a usar → Documentar reglas de negocio claras y agregar constraint de suma = 100%

---

### 7. `history_order_status` - Historial de Cambios de Estado

**Propósito:** Auditoría de transiciones de estado de órdenes.

```sql
CREATE TABLE history_order_status (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    previous_status order_status,
    new_status order_status NOT NULL,
    change_date TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Índices:**
- `idx_history_order_id` en `order_id`

**Relaciones:**
- `history_order_status.order_id` → `orders.id` (FK con CASCADE DELETE)

#### 🔴 PROBLEMA CRÍTICO: Tabla Huérfana

1. **Esta tabla NO SE USA en el código actual**
   - ❌ No hay triggers que la alimenten
   - ❌ No hay código Go que inserte registros
   - ❌ No hay endpoints que la consulten
   - **Estado:** Tabla inútil actualmente

2. **Diseño es correcto, pero falta implementación:**
   ```go
   // El repository debería hacer esto al actualizar status:
   func (r *repository) UpdateStatus(ctx context.Context, orderID int, newStatus OrderStatus) error {
       tx, _ := r.db.BeginTx(ctx, nil)
       
       // 1. Obtener status actual
       var currentStatus OrderStatus
       tx.QueryRow("SELECT status FROM orders WHERE id = $1", orderID).Scan(&currentStatus)
       
       // 2. Actualizar orden
       tx.Exec("UPDATE orders SET status = $1 WHERE id = $2", newStatus, orderID)
       
       // 3. Registrar en historial
       tx.Exec(`
           INSERT INTO history_order_status (order_id, previous_status, new_status)
           VALUES ($1, $2, $3)
       `, orderID, currentStatus, newStatus)
       
       tx.Commit()
   }
   ```

**Recomendación:**
- **Opción A:** Implementar la lógica de auditoría (2-3 horas)
- **Opción B:** Eliminar tabla hasta que realmente la necesites

---

## Diagrama de Relaciones

```
┌─────────────────────┐
│  expense_categories │
└──────────┬──────────┘
           │
           │ 1:N
           │
           ▼
┌─────────────────────┐         ┌──────────────────┐
│      expenses       │         │      orders      │
│                     │         │                  │
│  - type: general/   │         │  - status        │
│    order_linked     │◄────────│  - amount_charged│
│  - amount           │  0..1:N │  - delivery_type │
│  - category_id      │         │  - paid_50_percent│ ⚠️
└─────────────────────┘         └────────┬─────────┘
                                         │
                                         │ 1:1 ⚠️ (DEBERÍA SER 1:N)
                                         │
                                         ▼
                               ┌──────────────────┐
                               │      income      │
                               │                  │
                               │  - order_id      │ (UNIQUE ⚠️)
                               │  - amount        │
                               │  - date          │
                               └──────────────────┘

                                         │
                                         │ 1:N
                                         │
                                         ▼
                        ┌─────────────────────────────┐
                        │  history_order_status       │
                        │                             │
                        │  - order_id                 │
                        │  - previous_status          │
                        │  - new_status               │
                        │  - change_date              │
                        └─────────────────────────────┘
                                    ⚠️ (NO USADA)

┌─────────────────────────────────┐
│  financial_percentages          │  ⚠️ (NO USADA)
│                                 │
│  - reinvestment_percentage      │
│  - supplies_percentage          │
│  - profit_percentage            │
│  - effective_start_date         │
└─────────────────────────────────┘

┌─────────────────────────────────┐
│          users                  │  ⚠️ (NO USADA)
│                                 │
│  - email (UNIQUE)               │
│  - password_hash                │
│  - role                         │
└─────────────────────────────────┘
```

---

## Problemas Críticos Identificados

### 🔴 Prioridad 1: Bloqueadores de Funcionalidad

#### 1. Income Table Constraint UNIQUE

**Problema:**
```sql
order_id INTEGER UNIQUE NOT NULL
```

**Impacto:** Imposibilita pagos parciales y múltiples ingresos por orden

**Solución:**
```sql
-- Migration nueva (pre-SQLite):
ALTER TABLE income DROP CONSTRAINT income_order_id_key;
```

**Urgencia:** INMEDIATA  
**Estimación:** 10 minutos  
**Riesgo:** NINGUNO (mejora funcionalidad)

---

#### 2. Campo `paid_50_percent` en Orders

**Problema:**
```sql
paid_50_percent BOOLEAN NOT NULL DEFAULT FALSE
```

**Impacto:**
- Hardcodea asunción del 50%
- Duplica información
- No soporta otros escenarios (30%, 70%, pagos múltiples)

**Solución:**
```sql
-- Eliminar campo
ALTER TABLE orders DROP COLUMN paid_50_percent;

-- Calcular en Go:
func GetPaymentStatus(orderID int) PaymentStatus {
    totalPaid := SumIncomes(orderID)
    amountCharged := GetOrder(orderID).AmountCharged
    percentage := (totalPaid / amountCharged) * 100
    
    return PaymentStatus{
        TotalPaid: totalPaid,
        Remaining: amountCharged - totalPaid,
        Percentage: percentage,
        IsFullyPaid: percentage >= 100,
    }
}
```

**Urgencia:** ALTA  
**Estimación:** 1 hora (migration + código + tests)  
**Riesgo:** BAJO (si frontend lo usa, necesita actualización)

---

#### 3. Estados de Order Incorrectos

**Problema:**
```sql
CREATE TYPE order_status AS ENUM (
    'pending',
    'completed',
    'paid'
);
```

**Impacto:**
- Confunde workflow operacional con estado financiero
- `FinishOrder()` pone status en 'paid' → asume orden entregada = orden pagada
- No permite: "orden entregada pero parcialmente pagada"

**Solución:**
```sql
-- Estados operacionales simplificados para emprendimiento pequeño
CREATE TYPE order_status AS ENUM (
    'new',        -- Orden confirmada, pendiente de iniciar
    'active',     -- En producción
    'ready',      -- Terminada, lista para entregar
    'completed',  -- Entregada al cliente
    'cancelled'   -- Cancelada
);
```

**El estado financiero se calcula independientemente:**
```go
type FinancialStatus string

const (
    NotPaid FinancialStatus = "not_paid"
    PartiallyPaid FinancialStatus = "partially_paid"
    FullyPaid FinancialStatus = "fully_paid"
    Overpaid FinancialStatus = "overpaid"
)

func CalculateFinancialStatus(order Order, incomes []Income) FinancialStatus {
    totalPaid := sum(incomes)
    
    if totalPaid == 0 {
        return NotPaid
    }
    if totalPaid < order.AmountCharged {
        return PartiallyPaid
    }
    if totalPaid == order.AmountCharged {
        return FullyPaid
    }
    return Overpaid
}
```

**Urgencia:** ALTA  
**Estimación:** 3-4 horas (migration + actualizar código + tests)  
**Riesgo:** MEDIO (requiere actualizar toda la lógica de estados)

---

### ⚠️ Prioridad 2: Tablas Huérfanas

#### 4. Tabla `users` No Usada

**Problema:** Existe la tabla pero no hay:
- Endpoints de autenticación
- Middleware de autorización
- Lógica de roles

**Solución:**
- **Opción A:** Implementar auth completo (1-2 días)
- **Opción B:** Eliminar tabla hasta fase 2

**Recomendación:** Opción B (eliminar ahora, implementar después)

---

#### 5. Tabla `financial_percentages` No Usada

**Problema:** Tabla sin uso en código

**Solución:**
- **Opción A:** Implementar distribución financiera
- **Opción B:** Eliminar tabla

**Recomendación:** Opción B (eliminar ahora)

---

#### 6. Tabla `history_order_status` No Usada

**Problema:** Tabla de auditoría sin implementación

**Solución:**
- **Opción A:** Implementar triggers o código de auditoría
- **Opción B:** Eliminar tabla temporalmente

**Recomendación:** Depende de tus necesidades de auditoría

---

## Schema Propuesto Antes de Migración a SQLite

### Cambios Mínimos Necesarios

```sql
-- 1. FIX CRÍTICO: Permitir múltiples incomes por order
ALTER TABLE income DROP CONSTRAINT income_order_id_key;

-- 2. Eliminar campo problemático
ALTER TABLE orders DROP COLUMN paid_50_percent;

-- 3. Redefinir estados (requiere migration compleja)
-- Ver sección "Estados de Order Incorrectos"

-- 4. OPCIONAL: Limpiar tablas no usadas
DROP TABLE users;  -- Si no vas a implementar auth pronto
DROP TABLE financial_percentages;  -- Si no vas a usar
DROP TABLE history_order_status;  -- Si no vas a implementar auditoría
```

### Schema Limpio Recomendado

**Tablas Core (imprescindibles):**
1. ✅ `orders` (con fixes)
2. ✅ `income` (sin UNIQUE constraint)
3. ✅ `expenses`
4. ✅ `expense_categories`

**Tablas Secundarias (según necesidad):**
5. ❓ `users` - Solo si implementas auth en < 2 semanas
6. ❓ `financial_percentages` - Solo si implementas reporting financiero
7. ❓ `history_order_status` - Solo si necesitas auditoría

---

## Validaciones de Integridad Recomendadas

### A Nivel de Base de Datos

```sql
-- Montos no negativos
ALTER TABLE orders ADD CONSTRAINT orders_amount_positive 
    CHECK (amount_charged >= 0);

ALTER TABLE income ADD CONSTRAINT income_amount_positive 
    CHECK (amount >= 0);

ALTER TABLE expenses ADD CONSTRAINT expenses_amount_positive 
    CHECK (amount >= 0);

-- Fechas lógicas
ALTER TABLE orders ADD CONSTRAINT estimated_delivery_after_entry 
    CHECK (estimated_delivery_date IS NULL OR estimated_delivery_date >= entry_date);

-- Consistencia de tipo de gasto
ALTER TABLE expenses ADD CONSTRAINT expense_type_consistency
    CHECK (
        (type = 'general' AND order_id IS NULL) OR
        (type = 'order_linked' AND order_id IS NOT NULL)
    );
```

### A Nivel de Aplicación (Go)

```go
// En service layer
func ValidateOrder(dto CreateOrderDTO) error {
    if dto.AmountCharged < 0 {
        return errors.New("amount_charged must be >= 0")
    }
    if dto.Description == "" {
        return errors.New("description cannot be empty")
    }
    if dto.EstimatedDeliveryDate != nil && dto.EstimatedDeliveryDate.Before(time.Now()) {
        return errors.New("estimated delivery date cannot be in the past")
    }
    return nil
}
```

---

## Recomendaciones Pre-Migración SQLite

### Checklist Obligatorio

- [ ] ✅ Eliminar UNIQUE constraint de `income.order_id`
- [ ] ✅ Eliminar campo `orders.paid_50_percent`
- [ ] ✅ Redefinir `order_status` enum
- [ ] ✅ Decidir qué hacer con tablas huérfanas (users, financial_percentages, history_order_status)
- [ ] ✅ Agregar constraints de validación (montos positivos, etc.)
- [ ] ✅ Implementar lógica de cálculo de estado financiero en Go
- [ ] ✅ Actualizar handlers/services para nuevos estados
- [ ] ✅ Crear tests para casos edge (múltiples incomes, pagos parciales)

### Migración a SQLite

Una vez corregidos los problemas arriba:

1. **Actualizar go.mod:**
   ```bash
   go get github.com/mattn/go-sqlite3
   ```

2. **Adaptar connection.go:**
   ```go
   import _ "github.com/mattn/go-sqlite3"
   
   db, err := sql.Open("sqlite3", "./ebzer.db")
   ```

3. **Convertir migraciones PostgreSQL → SQLite:**
   - `SERIAL` → `INTEGER PRIMARY KEY AUTOINCREMENT`
   - `NUMERIC(10,2)` → `REAL`
   - `TIMESTAMP` → `TEXT` o `INTEGER` (Unix timestamp)
   - ENUMs → `TEXT` con CHECK constraints

4. **Ajustar queries incompatibles:**
   - `RETURNING id` funciona en SQLite, pero verificar
   - `ON CONFLICT` sintaxis puede diferir

---

## Conclusión

### Estado Actual: ⚠️ NO LISTO PARA PRODUCCIÓN

**Bloqueadores identificados:**
1. 🔴 Income table impide pagos parciales (CRÍTICO)
2. 🔴 Estados de order confunden operacional con financiero (CRÍTICO)
3. 🟡 Campo paid_50_percent es anti-pattern (ALTO)
4. 🟡 3 tablas huérfanas sin uso (MEDIO)

### Tiempo Estimado de Corrección

- **Mínimo viable:** 4-6 horas (solo fixes críticos)
- **Completo:** 2-3 días (incluye refactor de estados y tests)

### Próximos Pasos

1. **Hoy:** Fix income UNIQUE constraint (15 min)
2. **Esta semana:** Rediseño de order_status (4 horas)
3. **Antes de SQLite migration:** Decidir sobre tablas huérfanas (1 hora)
4. **Después de fixes:** Migrar a SQLite con schema limpio

---

**Última actualización:** 18 de abril, 2026  
**Documento creado por:** Revisión técnica automatizada  
**Próxima revisión:** Después de implementar fixes críticos
