# Decisiones de Schema - Pre-migración SQLite

**Fecha:** 18 de abril, 2026  
**Objetivo:** Definir qué cambios hacer ANTES de migrar a SQLite

---

## Resumen Ejecutivo

Hemos identificado **3 problemas críticos** y **3 tablas huérfanas** en el schema actual.

### 🔴 Bloqueadores Críticos

| # | Problema | Impacto | Tiempo Fix | Urgencia |
|---|----------|---------|------------|----------|
| 1 | `income.order_id` UNIQUE | Imposibilita pagos parciales | 15 min | INMEDIATA |
| 2 | `orders.paid_50_percent` | Anti-pattern, duplica datos | 1 hora | ALTA |
| 3 | Estados de `order_status` | Confunde operacional/financiero | 3-4 horas | ALTA |

### 🟡 Tablas Sin Uso

| Tabla | Estado | Decisión Necesaria |
|-------|--------|-------------------|
| `users` | Existe pero sin implementación auth | Eliminar o implementar auth |
| `financial_percentages` | Existe pero sin reportes | Eliminar o implementar reportes |
| `history_order_status` | Existe pero sin triggers/código | Eliminar o implementar auditoría |

---

## Decisión 1: Fix Income Table (OBLIGATORIO)

### Problema

```sql
-- Actual (INCORRECTO):
order_id INTEGER UNIQUE NOT NULL
```

Esto permite **solo 1 ingreso por orden**, violando la regla:  
> "Una orden puede tener 0, 1, o MUCHOS ingresos"

### Escenario Roto

```
Cliente hace pedido de $1000
↓
Paga anticipo $500 → ✅ Se registra income #1 (order_id: 123)
↓
2 semanas después paga $500 restantes → ❌ ERROR: duplicate key
```

### Fix

```sql
ALTER TABLE income DROP CONSTRAINT income_order_id_key;
```

### Consecuencias

- ✅ Permite múltiples pagos por orden
- ✅ Soporta cuotas
- ✅ Alineado con reglas de negocio
- ⚠️ Necesitas agregar índice no-único:
  ```sql
  CREATE INDEX idx_income_order_id ON income(order_id);
  ```

### ¿Hacer esto?

**Respuesta:** ✅ **SÍ, OBLIGATORIO**

**Respuesta real:** En el fix puedes no agregar un ALTER, ya que este sistema aun no se encuentra en prod, entonces puede ser mejor modificar la tabla directamente.

---

## Decisión 2: Eliminar Campo `paid_50_percent`

### Problema

```sql
-- En tabla orders:
paid_50_percent BOOLEAN NOT NULL DEFAULT FALSE
```

**Por qué es malo:**
- Hardcodea asunción del 50%
- ¿Qué si pagan 30%? ¿70%? ¿100% de una vez?
- Información duplicada (ya tienes tabla `income`)

### Fix

```sql
-- Eliminar campo:
ALTER TABLE orders DROP COLUMN paid_50_percent;
```

```go
// En su lugar, calcular en runtime:
type PaymentStatus struct {
    TotalPaid      float64 `json:"total_paid"`
    AmountCharged  float64 `json:"amount_charged"`
    Remaining      float64 `json:"remaining"`
    PercentagePaid float64 `json:"percentage_paid"`
    IsFullyPaid    bool    `json:"is_fully_paid"`
}

func GetPaymentStatus(ctx context.Context, orderID int) (*PaymentStatus, error) {
    order := GetOrderByID(orderID)
    incomes := GetIncomesByOrderID(orderID)
    
    totalPaid := 0.0
    for _, income := range incomes {
        totalPaid += income.Amount
    }
    
    return &PaymentStatus{
        TotalPaid:      totalPaid,
        AmountCharged:  order.AmountCharged,
        Remaining:      order.AmountCharged - totalPaid,
        PercentagePaid: (totalPaid / order.AmountCharged) * 100,
        IsFullyPaid:    totalPaid >= order.AmountCharged,
    }, nil
}
```

### Consecuencias

- ✅ Más flexible
- ✅ Elimina duplicación
- ✅ Soporta cualquier monto/porcentaje
- ⚠️ Frontend necesita cambio si usaba este campo

### Preguntas a Responder

1. **¿El frontend usa actualmente `paid_50_percent`?**
   - Si NO → Eliminar sin problemas
   - Si SÍ → Actualizar frontend para usar nuevo endpoint

2. **¿Necesitas endpoint específico para estado de pago?**
   - Sugerencia: `GET /api/orders/:id/payment-status`
    - de acuerdo con la sugerencia de implementación en Go arriba

### ¿Hacer esto?

**Respuesta:** ✅ **SÍ, RECOMENDADO**
**Respuesta real:** El frontend usa este campo pero vamos a aplicar el cambio sugerido el de PaymentStatus y remover el paid_50_percent de una vez para evitar deuda técnica.

---

## Decisión 3: Estados Operacionales Simplificados (Implementado ✅)

### Problema Actual

```sql
CREATE TYPE order_status AS ENUM (
    'pending',      -- ⚠️ ¿Pendiente de qué?
    'completed',    -- ⚠️ ¿Completada producción? ¿Entregada?
    'paid'          -- ⚠️ Estado financiero, NO operacional
);
```

**Problema conceptual:**
```go
// En repository:
func FinishOrder(id int) {
    // ❌ ASUME: orden terminada = orden pagada
    UPDATE orders SET status = 'paid' WHERE id = $1
}
```

Esto NO permite: **"Orden entregada pero parcialmente pagada"**

### Estados Propuestos

#### Opción A: Estados Operacionales Simplificados (Implementado ✅)

```sql
CREATE TYPE order_status AS ENUM (
    'new',        -- Orden confirmada, pendiente de iniciar
    'active',     -- En producción
    'ready',      -- Terminada, lista para entregar
    'completed',  -- Entregada al cliente
    'cancelled'   -- Cancelada
);
```

**El estado financiero se calcula aparte:**
```go
// No es un campo, es una función:
func (o *Order) GetFinancialStatus() string {
    totalPaid := SumIncomes(o.ID)
    
    switch {
    case totalPaid == 0:
        return "not_paid"
    case totalPaid < o.AmountCharged:
        return "partially_paid"
    case totalPaid >= o.AmountCharged:
        return "fully_paid"
    }
}
```

**Respuestas del API incluirían ambos:**
```json
{
    "id": 123,
    "status": "completed",
    "financial_status": "partially_paid",
    "payment_details": {
        "total_paid": 500,
        "amount_charged": 1000,
        "remaining": 500
    }
}
```

#### Opción B: Mantener Simple (solo 3 estados)

```sql
CREATE TYPE order_status AS ENUM (
    'active',      -- Orden activa (confirmada o en progreso)
    'completed',   -- Orden completada y entregada
    'cancelled'    -- Cancelada
);
```

**Pros:**
- ✅ Más simple
- ✅ Menos complejidad

**Contras:**
- ❌ Pierde granularidad (¿está en producción? ¿lista? ¿enviada?)
- ❌ Menos útil para tracking operacional

### Comparación

| Aspecto | Opción A (6 estados) | Opción B (3 estados) |
|---------|---------------------|---------------------|
| Tracking operacional | ✅✅✅ Excelente | ⚠️ Básico |
| Complejidad | ⚠️ Media | ✅ Baja |
| Separación financiero/operacional | ✅ Clara | ✅ Clara |
| Útil para reportes | ✅ Muy útil | ⚠️ Limitado |

### ¿Qué decidir?

**Pregunta clave:** ¿Necesitas trackear el progreso de producción/entrega?

- Si **SÍ** (sabes en qué etapa está cada orden) → **Opción A**
- Si **NO** (solo te importa: activa, terminada, cancelada) → **Opción B**

### ¿Hacer esto?

**Respuesta:** Tu decisión depende de necesidad de tracking operacional
- Si necesitas granularidad → Opción A (6 estados) y calculo financiero aparte, revisar si GetPaymentStatus y GetFinancialStatus pueden unificarse o si conviene mantener separados.
---

## Decisión 4: Tabla `users`

### Estado Actual

- ✅ Tabla existe con estructura correcta
- ❌ NO hay endpoints de login/registro
- ❌ NO hay middleware de autenticación
- ❌ NO hay autorización (roles no se usan)

### Opciones

#### A) Eliminar Ahora

**Cuándo:** Si NO vas a implementar auth en < 2 semanas

```sql
DROP TABLE users;
```

**Pros:**
- ✅ Limpia schema
- ✅ No mantener código muerto

**Contras:**
- ⚠️ Necesitarás crearla de nuevo después

#### B) Mantener e Implementar

**Cuándo:** Si SÍ vas a implementar auth pronto

**Requiere implementar:**
1. Endpoint `POST /api/auth/register`
2. Endpoint `POST /api/auth/login` (retorna JWT)
3. Middleware JWT en rutas protegidas
4. RBAC (Role Based Access Control)

**Estimación:** 1-2 días

#### C) Mantener Sin Usar (No Recomendado)

**Cuándo:** Nunca

### ¿Qué decidir?

**Pregunta:** ¿Vas a implementar login/auth en las próximas 2 semanas?

- Si **SÍ** → Opción B (mantener e implementar)
- Si **NO** → Opción A (eliminar, recrear después)

** respuesta real:** Vamos a mantener la tabla pero no implementar auth por ahora, ya que no es prioridad inmediata y eliminarla ahora solo para recrearla después generaría trabajo innecesario.

---

## Decisión 5: Tabla `financial_percentages`

### Estado Actual

- ✅ Tabla existe
- ❌ NO hay endpoints para gestionarla
- ❌ NO hay lógica que la use
- ❌ NO hay reportes financieros

### Propósito (según diseño)

Registrar cómo distribuir los ingresos:
- X% Reinversión
- Y% Insumos
- Z% Ganancia

### Opciones

#### A) Eliminar Ahora

```sql
DROP TABLE financial_percentages;
```

**Cuándo:** Si no vas a usar en corto plazo

#### B) Mantener e Implementar

**Requiere:**
1. CRUD de financial_percentages
2. Lógica de "qué registro está activo"
3. Reportes que calculen distribución
4. Constraint: suma = 100%

**Estimación:** 1-2 días

### ¿Qué decidir?

**Pregunta:** ¿Necesitas reportes de distribución financiera ahora?

- Si **SÍ** → Opción B
- Si **NO** → Opción A (eliminar)

** respuesta real:** Vamos a mantener la tabla pero no implementar su lógica/reportes por ahora, ya que no es prioridad inmediata y eliminarla ahora solo para recrearla después generaría trabajo innecesario.

---

## Decisión 6: Tabla `history_order_status`

### Estado Actual

- ✅ Tabla existe con estructura correcta
- ❌ NO hay código que inserte registros
- ❌ NO hay triggers
- ❌ NO hay endpoints que la consulten

### Propósito

Auditoría de cambios de estado:
```
Order #123:
- 2026-04-01 10:00 → new
- 2026-04-05 14:30 → active
- 2026-04-10 09:15 → ready
- 2026-04-12 16:00 → completed
```

### Opciones

#### A) Eliminar Ahora

```sql
DROP TABLE history_order_status;
```

**Cuándo:** Si no necesitas auditoría

#### B) Implementar Auditoría

**Opción B1: Trigger en PostgreSQL**
```sql
CREATE OR REPLACE FUNCTION log_order_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status != NEW.status THEN
        INSERT INTO history_order_status (order_id, previous_status, new_status)
        VALUES (NEW.id, OLD.status, NEW.status);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER order_status_audit
AFTER UPDATE ON orders
FOR EACH ROW
EXECUTE FUNCTION log_order_status_change();
```

**Opción B2: En código Go (recomendado para SQLite)**
```go
func (r *repository) UpdateStatus(ctx context.Context, orderID int, newStatus OrderStatus) error {
    tx, _ := r.db.BeginTx(ctx, nil)
    defer tx.Rollback()
    
    // 1. Get current status
    var currentStatus OrderStatus
    tx.QueryRowContext(ctx, "SELECT status FROM orders WHERE id = $1", orderID).Scan(&currentStatus)
    
    // 2. Update order
    tx.ExecContext(ctx, "UPDATE orders SET status = $1 WHERE id = $2", newStatus, orderID)
    
    // 3. Log change
    tx.ExecContext(ctx, `
        INSERT INTO history_order_status (order_id, previous_status, new_status)
        VALUES ($1, $2, $3)
    `, orderID, currentStatus, newStatus)
    
    return tx.Commit()
}
```

**Estimación:** 2-3 horas

### ¿Qué decidir?

**Pregunta:** ¿Necesitas ver historial de cambios de estado?

- Si **SÍ** → Opción B (implementar)
- Si **NO** (al menos al principio) → Opción A (eliminar, agregar después)

** respuesta real:** Si requerimos historial de cambios de estado, pero vamos a remover esto por ahora ya que estamos haciendo otros cambios sobre el estado entonces entonces primero vamos a resolver los otros cambios y luego implementamos esta auditoría.

---

## Resumen  de Decisiones

### ✅ Hacer OBLIGATORIAMENTE

| # | Acción | Tiempo | Archivo |
|---|--------|--------|---------|
| 1 | Eliminar UNIQUE de `income.order_id` | 15 min | Nueva migration |
| 2 | Agregar índice no-único en `income.order_id` | 5 min | Nueva migration |

### 🟡 Hacer RECOMENDADO

| # | Acción | Tiempo | Decisión Necesaria |
|---|--------|--------|--------------------|
| 3 | Eliminar `orders.paid_50_percent` | 1 hora | ¿Frontend lo usa? |
| 4 | Rediseñar `order_status` | 3-4 horas | ¿Opción A o B? |

### ❓ Decidir Caso por Caso

| Tabla | ¿Implementar ahora? | ¿Eliminar? | Tiempo si implementas |
|-------|---------------------|------------|-----------------------|
| `users` | No | No | 1-2 días |
| `financial_percentages` | No | No | 1-2 días |
| `history_order_status` | No | Si | 2-3 horas |

---

## Plan de Acción Sugerido

### Fase 1: Fixes Críticos (HOY - 2 horas)

```bash
# 1. Crear migration para income fix
# Archivo: 000008_fix_income_unique_constraint.up.sql

ALTER TABLE income DROP CONSTRAINT income_order_id_key;
CREATE INDEX idx_income_order_id ON income(order_id);
```

```bash
# 2. Fix down migration
# Archivo: 000008_fix_income_unique_constraint.down.sql

DROP INDEX idx_income_order_id;
ALTER TABLE income ADD CONSTRAINT income_order_id_key UNIQUE (order_id);
```

sobre estos fixes recuerda que esto no esta en prod entonces puedes modificar la tabla directamente sin necesidad de crear una migration.

### Fase 2: Eliminar paid_50_percent (1 hora)

```bash
# 3. Crear migration
# Archivo: 000009_remove_paid_50_percent.up.sql

ALTER TABLE orders DROP COLUMN paid_50_percent;
```

```go
// 4. Crear nuevo service method
func (s *service) GetPaymentStatus(ctx context.Context, orderID int) (*PaymentStatus, error) {
    // Implementación arriba
}

// 5. Crear nuevo handler
func (h *Handler) GetPaymentStatus(c *fiber.Ctx) error {
    // Implementación
}

// 6. Registrar ruta
router.Get("/:id/payment-status", h.GetPaymentStatus)
```

Sobre este cambio, recuerda que no esta en prod entonces puedes eliminar el campo directamente y luego implementar el nuevo método sin necesidad de crear una migration.

### Fase 3: Rediseñar Estados (3-4 horas)

**TÚ DECIDES:** ¿Opción A (6 estados) o Opción B (3 estados)?

Después de decidir, crear migration y actualizar código.

Opción A (6 estados)

### Fase 4: Limpiar Tablas Huérfanas (variable)

**TÚ DECIDES para cada tabla:**
- `users`: ¿Eliminar o implementar?
- `financial_percentages`: ¿Eliminar o implementar?
- `history_order_status`: ¿Eliminar o implementar?

### Fase 5: Migrar a SQLite (después de todo lo anterior)

Solo cuando el schema esté limpio y correcto.

---

## Próximos Pasos INMEDIATOS

1. **LEE** el documento [modelo-de-datos.md](./modelo-de-datos.md) completo
2. **DECIDE** sobre cada tabla en la sección "Decidir Caso por Caso"
3. **EJECUTA** Fase 1 (fixes críticos) - OBLIGATORIO
4. **CONSIDERA** Fase 2 (eliminar paid_50_percent) - RECOMENDADO
5. **DECIDE** Fase 3 (nuevos estados) - Opción A o B
6. **LIMPIA** Fase 4 (tablas huérfanas)
7. **MIGRA** Fase 5 (a SQLite) - Solo cuando todo esté listo

---

**Última actualización:** 18 de abril, 2026  
**Próximo paso:** Tomar decisiones marcadas con ❓  
**Contacto:** Responde con tus decisiones y empezamos implementación
