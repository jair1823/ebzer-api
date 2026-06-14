# Guía de Integración Frontend - Ebzer API
**Fecha:** 18 de abril, 2026  
**Última actualización:** 22 de abril, 2026 (Filtro de status acepta múltiples valores separados por comas)  
**Para:** Agente de Frontend / Equipo de UI  
**De:** Backend Review Agent

---

## 📋 Propósito de esta Guía

Esta guía te ayudará a integrar el frontend con la API de Ebzer después de los cambios arquitectónicos recientes. Contiene:

1. ✅ Estado actual del backend
2. 🔌 Endpoints disponibles y contratos
3. ⚠️ Cambios breaking que debes aplicar
4. 💡 Ejemplos de integración
5. 🚫 Endpoints NO disponibles (para evitar confusión)

---

## 1. Estado del Backend (Resumen Ejecutivo)

### ✅ Lo que SÍ está disponible

| Módulo | Estado | Endpoints | Notas |
|--------|--------|-----------|-------|
| **Orders** | ✅ Funcional | 7 endpoints | CRUD completo + PaymentStatus |
| **Incomes** | ✅ Funcional | 5 endpoints | CRUD completo |
| **Expenses** | ✅ Funcional | 11 endpoints | CRUD completo + filtros |
| **Expense Categories** | ✅ Funcional | 5 endpoints | CRUD completo |
| **SQLite** | ✅ Activo | - | Migración completada |
| **CORS** | ✅ Activo | - | Acepta todos los orígenes |

### ❌ Lo que NO está disponible

| Módulo | Estado | Impacto |
|--------|--------|---------|
| **Users / Auth** | ❌ No implementado | No hay autenticación |
| **Financial Reports** | ❌ No implementado | No hay reportes financieros |

---

## 2. Base URL y Configuration

### Backend URL
```
http://localhost:3000
```

### Health Checks

```bash
# Ping básico
GET http://localhost:3000/ping
Response: { "message": "pong" }

# Database health
GET http://localhost:3000/dbping
Response: { "message": "Database connection successful" }
```

### CORS
El backend acepta **todos los orígenes** actualmente:
```go
AllowOrigins: "*"
```
No necesitas configuración especial para desarrollo.

---

## 3. Módulo: Orders

### 3.1 Modelo de Datos

```typescript
// TypeScript types para el frontend
type OrderStatus = 
  | "new"       // Orden confirmada, pendiente de iniciar
  | "active"    // En producción
  | "ready"     // Lista para entrega/retiro
  | "completed" // Entregada al cliente
  | "cancelled"; // Cancelada

type DeliveryType = "pickup" | "shipping" | "delivery";

interface Order {
  id: number;
  description: string;
  amount_charged: number;
  status: OrderStatus;
  entry_date: string;  // ISO8601 timestamp
  estimated_delivery_date: string | null;  // ISO8601 timestamp
  delivery_type: DeliveryType;
  client_name: string;  // Requerido desde v1.1
  client_phone: string | null;
  notes: string | null;
  created_at: string;  // ISO8601 timestamp
  updated_at: string;  // ISO8601 timestamp
}

interface PaymentStatus {
  total_paid: number;
  amount_charged: number;
  remaining: number;
  percentage_paid: number;
  is_fully_paid: boolean;
}
```

### 3.2 Endpoints Disponibles

#### 📌 POST `/api/orders` - Crear Orden

**Request:**
```json
{
  "description": "Botella térmica personalizada con logo",
  "amount_charged": 1000.50,
  "client_name": "Juan Pérez",  // Requerido
  "status": "new",  // Opcional, default: "new"
  "estimated_delivery_date": "2026-04-25T10:00:00Z",  // Opcional
  "delivery_type": "pickup",  // Opcional, default: "pickup"
  "client_phone": "+50671612841",  // Opcional
  "notes": "Logo en dorado"  // Opcional
}
```

**Response (201 Created):**
```json
{
  "id": 123
}
```

**Validaciones:**
- `description`: requerido, no vacío
- `amount_charged`: requerido, >= 0, acepta string o number
- `client_name`: **requerido**, no vacío - una orden siempre está asociada a un cliente
- `status`: debe ser uno de los 6 estados válidos
- `delivery_type`: debe ser uno de los 3 tipos válidos

**Ejemplo en JavaScript:**
```javascript
async function crearOrden(orderData) {
  const response = await fetch('http://localhost:3000/api/orders', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(orderData)
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message);
  }
  
  return response.json();  // { id: 123 }
}
```

---

#### 📌 GET `/api/orders` - Listar Órdenes

**Query Parameters (todos opcionales):**
- `status` - Filtrar por estado(s). Acepta **uno o múltiples** valores separados por comas
  - Un estado: `?status=new`
  - Múltiples estados: `?status=new,active,ready`
- `from` - Fecha desde (formato: `YYYY-MM-DD`, ej: `?from=2026-04-01`)
  - Incluye desde la medianoche (00:00:00) de la fecha especificada
- `to` - Fecha hasta (formato: `YYYY-MM-DD`, ej: `?to=2026-04-30`)
  - Incluye **TODO el día** especificado (hasta las 23:59:59)
  - Ejemplo: `to=2026-04-22` incluye órdenes hasta `2026-04-22 23:59:59`

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "description": "Botella térmica personalizada con logo",
    "amount_charged": 1000.50,
    "status": "new",
    "entry_date": "2026-04-18T10:30:00Z",
    "estimated_delivery_date": "2026-04-25T10:00:00Z",
    "delivery_type": "pickup",
    "client_name": "Juan Pérez",
    "client_phone": "+50671612841",
    "notes": "Logo en dorado",
    "created_at": "2026-04-18T10:30:00Z",
    "updated_at": "2026-04-18T10:30:00Z"
  }
]
```

**Ejemplos:**
```javascript
// Todas las órdenes
const orders = await fetch('http://localhost:3000/api/orders').then(r => r.json());

// Solo órdenes nuevas
const nuevas = await fetch('http://localhost:3000/api/orders?status=new').then(r => r.json());

// ✨ NUEVO: Múltiples status (órdenes activas en producción o listas)
const enProceso = await fetch(
  'http://localhost:3000/api/orders?status=new,active,ready'
).then(r => r.json());

// Órdenes de abril 2026
const abril = await fetch(
  'http://localhost:3000/api/orders?from=2026-04-01&to=2026-04-30'
).then(r => r.json());

// Órdenes activas de abril
const activeAbril = await fetch(
  'http://localhost:3000/api/orders?status=active&from=2026-04-01&to=2026-04-30'
).then(r => r.json());

// ✨ Combinando múltiples status con filtros de fecha
const pendientesAbril = await fetch(
  'http://localhost:3000/api/orders?status=new,active&from=2026-04-01&to=2026-04-30'
).then(r => r.json());
```

---

#### 📌 GET `/api/orders/:id` - Obtener Detalle de Orden

**Response (200 OK):**
```json
{
  "id": 123,
  "description": "Botella térmica personalizada con logo",
  "amount_charged": 1000.50,
  "status": "active",
  "entry_date": "2026-04-18T10:30:00Z",
  "estimated_delivery_date": null,
  "delivery_type": "pickup",
  "client_name": "Juan Pérez",
  "client_phone": null,
  "notes": null,
  "created_at": "2026-04-18T10:30:00Z",
  "updated_at": "2026-04-18T15:20:00Z"
}
```

**Response (404 Not Found):**
```json
{
  "message": "order not found"
}
```

---

#### 📌 GET `/api/orders/:id/payment-status` - **NUEVO** Estado de Pago

⚠️ **CAMBIO IMPORTANTE:** Este endpoint reemplaza el campo `paid_50_percent` que ya **NO existe** en la tabla orders.

**Response (200 OK):**
```json
{
  "total_paid": 800.00,
  "amount_charged": 1000.50,
  "remaining": 200.50,
  "percentage_paid": 79.95,
  "is_fully_paid": false
}
```

**Uso en el Frontend:**

```javascript
// ❌ INCORRECTO (antiguo):
if (order.paid_50_percent) {
  mostrarBadge("Anticipo Pagado");
}

// ✅ CORRECTO (nuevo):
const paymentStatus = await fetch(`http://localhost:3000/api/orders/${orderId}/payment-status`)
  .then(r => r.json());

if (paymentStatus.is_fully_paid) {
  mostrarBadge("Pagado Completo");
} else if (paymentStatus.percentage_paid >= 50) {
  mostrarBadge(`${paymentStatus.percentage_paid.toFixed(0)}% Pagado`);
} else if (paymentStatus.percentage_paid > 0) {
  mostrarBadge(`Anticipo: $${paymentStatus.total_paid}`);
} else {
  mostrarBadge("Pendiente de Pago");
}

// Mostrar información detallada
console.log(`Pagado: $${paymentStatus.total_paid} de $${paymentStatus.amount_charged}`);
console.log(`Falta: $${paymentStatus.remaining}`);
```

**Ejemplo de componente React:**
```jsx
function OrderPaymentBadge({ orderId }) {
  const [paymentStatus, setPaymentStatus] = useState(null);
  
  useEffect(() => {
    fetch(`http://localhost:3000/api/orders/${orderId}/payment-status`)
      .then(r => r.json())
      .then(setPaymentStatus);
  }, [orderId]);
  
  if (!paymentStatus) return null;
  
  return (
    <div className={`badge ${paymentStatus.is_fully_paid ? 'badge-success' : 'badge-warning'}`}>
      {paymentStatus.is_fully_paid ? (
        '✓ Pagado'
      ) : (
        `${paymentStatus.percentage_paid.toFixed(0)}% - $${paymentStatus.remaining} restante`
      )}
    </div>
  );
}
```

---

#### 📌 PUT `/api/orders/:id` - Actualizar Orden

⚠️ **CRÍTICO - Comportamiento de Actualización:**
- **PUT es un reemplazo total, NO parcial**
- **TODOS los campos son requeridos** en el request
- **Omitir un campo resultará en pérdida de datos**
- Si quieres mantener un valor sin cambios, envía el valor actual
- Para **limpiar campos opcionales de tipo string** (`client_phone`, `notes`): envía `""` (string vacío)
- Para **limpiar campos opcionales de tipo fecha** (`estimated_delivery_date`): envía `null`

**Request (ejemplo con todos los campos):**
```json
{
  "description": "Set de 6 tazas personalizadas",
  "amount_charged": 1200.00,
  "status": "active",
  "estimated_delivery_date": "2026-04-26T14:00:00Z",
  "delivery_type": "shipping",
  "client_name": "Juan Pérez García",
  "client_phone": "+56912345678",
  "notes": "Actualizado: nombre y fecha de boda"
}
```

**Request (limpiando campos opcionales):**
```json
{
  "description": "Set de 6 tazas personalizadas",
  "amount_charged": 1200.00,
  "status": "ready",
  "estimated_delivery_date": null,
  "delivery_type": "shipping",
  "client_name": "Juan Pérez García",
  "client_phone": "",
  "notes": ""
}
```

**Response (200 OK):**
```json
{
  "updated": true
}
```

**Response (404 Not Found):**
```json
{
  "message": "order not found"
}
```

**Ejemplo correcto:**
```javascript
async function actualizarOrden(orderId, orderData) {
  // Siempre enviar TODOS los campos
  const payload = {
    description: orderData.description,
    amount_charged: orderData.amount_charged,
    status: orderData.status,
    estimated_delivery_date: orderData.estimated_delivery_date || null,
    delivery_type: orderData.delivery_type,
    client_name: orderData.client_name,
    client_phone: orderData.client_phone || "",  // "" para limpiar
    notes: orderData.notes || ""  // "" para limpiar
  };
  
  const response = await fetch(`http://localhost:3000/api/orders/${orderId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  
  if (!response.ok) throw new Error('Error al actualizar');
  return response.json();
}

// Uso
await actualizarEstado(123, 'active');
```

---

#### 📌 POST `/api/orders/:id/finish` - Finalizar Orden

Marca la orden como `completed` (entregada al cliente).

**Request:** Sin body

**Response (200 OK):**
```json
{
  "finished": true
}
```

**Ejemplo:**
```javascript
async function finalizarOrden(orderId) {
  const response = await fetch(`http://localhost:3000/api/orders/${orderId}/finish`, {
    method: 'POST'
  });
  
  if (!response.ok) throw new Error('Error al finalizar orden');
  return response.json();
}
```

---

#### 📌 DELETE `/api/orders/:id` - Eliminar Orden

⚠️ **IMPORTANTE:** Esto también eliminará **todos los incomes asociados** (CASCADE DELETE).

**Response (200 OK):**
```json
{
  "deleted": true
}
```

**Response (404 Not Found):**
```json
{
  "message": "order not found"
}
```

---

## 4. Módulo: Incomes

### 4.1 Modelo de Datos

```typescript
interface Income {
  id: number;
  order_id: number;
  amount: number;
  date: string;  // ISO8601 timestamp
  created_at: string;
  updated_at: string;
}
```

### 4.2 Reglas de Negocio Importantes

✅ **UNA ORDEN PUEDE TENER MÚLTIPLES INGRESOS** (pagos parciales)  
✅ Un ingreso SIEMPRE está vinculado a una orden  
✅ Si eliminas una orden, se eliminan sus ingresos (CASCADE)  
❌ NO puedes crear un ingreso sin `order_id`  

**📝 Nota sobre tipos de datos:**
- `order_id`: debe enviarse como **number** (es un ID de base de datos)
- `amount`: acepta **number o string** (ambos son válidos para mantener precisión decimal)

### 4.3 Endpoints Disponibles

#### 📌 POST `/api/incomes` - Registrar Ingreso

**Request:**
```json
{
  "order_id": 123,  // number (ID de la orden)
  "amount": 500.00  // number o string (precisión decimal)
}
```

**Response (201 Created):**
```json
{
  "id": 456
}
```

**Ejemplo - Pago en 3 cuotas:**
```javascript
// Orden por $1000
const orderId = 123;

// Anticipo 50%
await fetch('http://localhost:3000/api/incomes', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ order_id: orderId, amount: 500 })
});

// Segunda cuota 30%
await fetch('http://localhost:3000/api/incomes', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ order_id: orderId, amount: 300 })
});

// Pago final 20%
await fetch('http://localhost:3000/api/incomes', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ order_id: orderId, amount: 200 })
});

// Verificar estado de pago
const paymentStatus = await fetch(`http://localhost:3000/api/orders/${orderId}/payment-status`)
  .then(r => r.json());

console.log(paymentStatus);
// {
//   total_paid: 1000,
//   amount_charged: 1000,
//   remaining: 0,
//   percentage_paid: 100,
//   is_fully_paid: true
// }
```

---

#### 📌 GET `/api/incomes` - Listar Ingresos

**Query Parameters (todos opcionales):**
- `from` - Fecha desde (formato: `YYYY-MM-DD`, ej: `?from=2026-04-01`)
  - Incluye desde la medianoche (00:00:00) de la fecha especificada
- `to` - Fecha hasta (formato: `YYYY-MM-DD`, ej: `?to=2026-04-30`)
  - Incluye **TODO el día** especificado (hasta las 23:59:59)

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "order_id": 123,
    "amount": 500.00,
    "date": "2026-04-18T10:30:00Z",
    "created_at": "2026-04-18T10:30:00Z",
    "updated_at": "2026-04-18T10:30:00Z"
  },
  {
    "id": 2,
    "order_id": 123,
    "amount": 300.00,
    "date": "2026-04-20T14:15:00Z",
    "created_at": "2026-04-20T14:15:00Z",
    "updated_at": "2026-04-20T14:15:00Z"
  }
]
```

**Ejemplo:**
```javascript
// Ingresos de abril 2026
const ingresos = await fetch(
  'http://localhost:3000/api/incomes?from=2026-04-01&to=2026-04-30'
).then(r => r.json());

// Calcular total ingresado en el mes
const totalMes = ingresos.reduce((sum, income) => sum + income.amount, 0);
console.log(`Total ingresado en abril: $${totalMes}`);
```

---

#### 📌 GET `/api/incomes/:id` - Obtener Detalle de Ingreso

**Response (200 OK):**
```json
{
  "id": 456,
  "order_id": 123,
  "amount": 500.00,
  "date": "2026-04-18T10:30:00Z",
  "created_at": "2026-04-18T10:30:00Z",
  "updated_at": "2026-04-18T10:30:00Z"
}
```

---

#### 📌 PUT `/api/incomes/:id` - Actualizar Ingreso

⚠️ **Nota:** A diferencia de orders, este endpoint **SÍ acepta updates parciales**. Solo envía los campos que quieres cambiar.

**Request (campos opcionales - UPDATE parcial real):**
```json
{
  "order_id": 124,
  "amount": 550.00
}
```

**Response (200 OK):**
```json
{
  "updated": true
}
```

---

#### 📌 DELETE `/api/incomes/:id` - Eliminar Ingreso

**Response (200 OK):**
```json
{
  "deleted": true
}
```

**Uso típico:**
```javascript
async function corregirPago(incomeId, nuevoMonto) {
  // Opción 1: Actualizar
  await fetch(`http://localhost:3000/api/incomes/${incomeId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ amount: nuevoMonto })
  });
  
  // Opción 2: Eliminar y crear nuevo
  await fetch(`http://localhost:3000/api/incomes/${incomeId}`, { method: 'DELETE' });
  await fetch('http://localhost:3000/api/incomes', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ order_id: orderId, amount: nuevoMonto })
  });
}
```

---

## 5. Módulo: Expenses

### 5.1 Modelo de Datos

```typescript
// TypeScript types para el frontend
type ExpenseType = "general" | "order_linked";

interface Expense {
  id: number;
  description: string;
  amount: number;
  date: string;  // ISO8601 timestamp
  order_id: number | null;  // null para gastos generales
  category_id: number | null;  // null si no tiene categoría
  type: ExpenseType;
  created_at: string;  // ISO8601 timestamp
}

interface ExpenseCategory {
  id: number;
  name: string;
  description: string | null;
  created_at: string;
}
```

### 5.2 Reglas de Negocio Importantes

✅ **Dos tipos de gastos:**
- `"general"` - Gastos no vinculados a ninguna orden (ej: servicios, alquiler) → `order_id` debe ser null
- `"order_linked"` - Gastos específicos de una orden (ej: materiales para pedido #123) → `order_id` requerido

✅ **Categorías opcionales** - Un gasto puede tener categoría (más organizado) o no tenerla  
✅ **Delete behavior** - Si eliminas una orden o categoría, los gastos asociados NO se eliminan (SET NULL)  
❌ **No se permiten** montos negativos  

### 5.3 Endpoints de Categorías

#### 📌 POST `/api/expenses/categories` - Crear Categoría

**Request:**
```json
{
  "name": "Materias primas",
  "description": "Materiales y suministros para producción"  // Opcional
}
```

**Response (201 Created):**
```json
{
  "id": 1
}
```

**Validaciones:**
- `name`: requerido, único
- `description`: opcional

**Ejemplo:**
```javascript
async function crearCategoria(nombre, descripcion = null) {
  const response = await fetch('http://localhost:3000/api/expenses/categories', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: nombre, description: descripcion })
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message);
  }
  
  return response.json();  // { id: 1 }
}

// Uso
await crearCategoria("Materias primas", "Materiales y suministros");
await crearCategoria("Transporte");
await crearCategoria("Servicios", "Electricidad, agua, internet");
```

---

#### 📌 GET `/api/expenses/categories` - Listar Categorías

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "name": "Materias primas",
    "description": "Materiales y suministros para producción",
    "created_at": "2026-04-18T10:00:00Z"
  },
  {
    "id": 2,
    "name": "Transporte",
    "description": null,
    "created_at": "2026-04-18T10:05:00Z"
  }
]
```

**Ejemplo:**
```javascript
const categories = await fetch('http://localhost:3000/api/expenses/categories')
  .then(r => r.json());

// Usar en un dropdown
const selectHTML = categories.map(cat => 
  `<option value="${cat.id}">${cat.name}</option>`
).join('');
```

---

#### 📌 GET `/api/expenses/categories/:id` - Obtener Categoría

**Response (200 OK):**
```json
{
  "id": 1,
  "name": "Materias primas",
  "description": "Materiales y suministros",
  "created_at": "2026-04-18T10:00:00Z"
}
```

---

#### 📌 PUT `/api/expenses/categories/:id` - Actualizar Categoría

**Request (campos opcionales):**
```json
{
  "name": "Materias primas e insumos",
  "description": "Materiales, suministros y herramientas"
}
```

**Response (200 OK):**
```json
{
  "updated": true
}
```

---

#### 📌 DELETE `/api/expenses/categories/:id` - Eliminar Categoría

**Response (200 OK):**
```json
{
  "deleted": true
}
```

**Nota:** Los gastos asociados NO se eliminan, solo pierden la referencia (category_id → NULL)

---

### 5.4 Endpoints de Gastos

#### 📌 POST `/api/expenses` - Crear Gasto

**Request (Gasto General):**
```json
{
  "description": "Pago de electricidad marzo",
  "amount": 150.50,
  "type": "general",
  "category_id": 3,  // Opcional
  "date": "2026-04-18T10:00:00Z"  // Opcional, default: now
}
```

**Request (Gasto Vinculado a Orden):**
```json
{
  "description": "Vinilo adhesivo premium importado",
  "amount": 85.00,
  "type": "order_linked",
  "order_id": 123,  // Requerido si type="order_linked"
  "category_id": 1
}
```

**Response (201 Created):**
```json
{
  "id": 456
}
```

**Validaciones:**
- `description`: requerido, no vacío
- `amount`: requerido, >= 0, acepta string o number
- `type`: requerido, debe ser "general" o "order_linked"
- `order_id`: requerido si type="order_linked", debe ser null si type="general"
- `category_id`: opcional
- `date`: opcional, ISO8601

**Ejemplo:**
```javascript
async function registrarGastoGeneral(descripcion, monto, categoryId = null) {
  return fetch('http://localhost:3000/api/expenses', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      description: descripcion,
      amount: monto,
      type: "general",
      category_id: categoryId
    })
  }).then(r => r.json());
}

async function registrarGastoDeOrden(descripcion, monto, orderId, categoryId = null) {
  return fetch('http://localhost:3000/api/expenses', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      description: descripcion,
      amount: monto,
      type: "order_linked",
      order_id: orderId,
      category_id: categoryId
    })
  }).then(r => r.json());
}

// Uso
await registrarGastoGeneral("Alquiler local", 500, 4);
await registrarGastoDeOrden("Tinta para serigrafía", 45, 123, 1);
```

---

#### 📌 GET `/api/expenses` - Listar Gastos

**Query Parameters (todos opcionales):**
- `from` - Fecha desde (**formato requerido: `YYYY-MM-DD`**, ej: `?from=2026-04-01`)
- `to` - Fecha hasta (**formato requerido: `YYYY-MM-DD`**, ej: `?to=2026-04-30`)
- `category` - Filtrar por ID de categoría (**debe ser un entero válido**, retorna error si es inválido)
- `type` - Filtrar por tipo ("general" o "order_linked")

⚠️ **Nota sobre filtros de fecha:**  
El backend NO valida estrictamente el formato de fecha en expenses. Se recomienda usar `YYYY-MM-DD` para consistencia con otros endpoints (orders, incomes). Otros formatos pueden funcionar pero no están garantizados.

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "description": "Vinilo adhesivo premium",
    "amount": 50.00,
    "date": "2026-04-18T09:00:00Z",
    "order_id": 123,
    "category_id": 1,
    "type": "order_linked",
    "created_at": "2026-04-18T09:00:00Z"
  },
  {
    "id": 2,
    "description": "Electricidad",
    "amount": 150.00,
    "date": "2026-04-18T10:00:00Z",
    "order_id": null,
    "category_id": 3,
    "type": "general",
    "created_at": "2026-04-18T10:00:00Z"
  }
]
```

**Ejemplos:**
```javascript
// Todos los gastos
const allExpenses = await fetch('http://localhost:3000/api/expenses').then(r => r.json());

// Solo gastos generales
const generalExpenses = await fetch(
  'http://localhost:3000/api/expenses?type=general'
).then(r => r.json());

// Gastos de abril 2026
const aprilExpenses = await fetch(
  'http://localhost:3000/api/expenses?from=2026-04-01&to=2026-04-30'
).then(r => r.json());

// Gastos de categoría "Materias primas" en abril
const materiaPrimaAbril = await fetch(
  'http://localhost:3000/api/expenses?category=1&from=2026-04-01&to=2026-04-30'
).then(r => r.json());

// Solo gastos vinculados a órdenes
const orderExpenses = await fetch(
  'http://localhost:3000/api/expenses?type=order_linked'
).then(r => r.json());

// Calcular total de gastos del mes
const total = aprilExpenses.reduce((sum, exp) => sum + exp.amount, 0);
console.log(`Total gastado en abril: $${total}`);

// ⚠️ Manejo de error para category inválido
try {
  const response = await fetch('http://localhost:3000/api/expenses?category=invalido');
  if (!response.ok) {
    const error = await response.json();
    console.error(error.message); // "invalid category: must be a valid integer"
  }
} catch (err) {
  console.error('Error al obtener gastos:', err);
}
```

---

#### 📌 GET `/api/expenses/:id` - Obtener Detalle de Gasto

**Response (200 OK):**
```json
{
  "id": 456,
  "description": "Vinilo adhesivo premium",
  "amount": 85.00,
  "date": "2026-04-18T11:30:00Z",
  "order_id": 123,
  "category_id": 1,
  "type": "order_linked",
  "created_at": "2026-04-18T11:30:00Z"
}
```

**Response (404 Not Found):**
```json
{
  "message": "expense not found"
}
```

---

#### 📌 GET `/api/expenses/order/:orderId` - Gastos de una Orden

**Endpoint especial** para obtener todos los gastos asociados a una orden específica.

**Response (200 OK):**
```json
[
  {
    "id": 10,
    "description": "Vinilo adhesivo especial",
    "amount": 35.00,
    "date": "2026-04-15T10:00:00Z",
    "order_id": 123,
    "category_id": 1,
    "type": "order_linked",
    "created_at": "2026-04-15T10:00:00Z"
  },
  {
    "id": 12,
    "description": "Esmalte para personalización",
    "amount": 25.50,
    "date": "2026-04-16T14:00:00Z",
    "order_id": 123,
    "category_id": 1,
    "type": "order_linked",
    "created_at": "2026-04-16T14:00:00Z"
  }
]
```

**Uso típico:**
```javascript
async function obtenerGastosDeOrden(orderId) {
  const expenses = await fetch(`http://localhost:3000/api/expenses/order/${orderId}`)
    .then(r => r.json());
  
  const totalGastado = expenses.reduce((sum, exp) => sum + exp.amount, 0);
  
  return { expenses, totalGastado };
}

// Uso
const { expenses, totalGastado } = await obtenerGastosDeOrden(123);
console.log(`Gastos de orden #123: $${totalGastado}`);
```

**Ejemplo de componente React:**
```jsx
function OrderExpensesSummary({ orderId }) {
  const [data, setData] = useState({ expenses: [], total: 0 });
  
  useEffect(() => {
    fetch(`http://localhost:3000/api/expenses/order/${orderId}`)
      .then(r => r.json())
      .then(expenses => {
        const total = expenses.reduce((sum, exp) => sum + exp.amount, 0);
        setData({ expenses, total });
      });
  }, [orderId]);
  
  return (
    <div className="expense-summary">
      <h3>Gastos de la Orden</h3>
      <ul>
        {data.expenses.map(exp => (
          <li key={exp.id}>
            {exp.description}: ${exp.amount}
          </li>
        ))}
      </ul>
      <p className="total">Total: ${data.total}</p>
    </div>
  );
}
```

---

#### 📌 PUT `/api/expenses/:id` - Actualizar Gasto

⚠️ **CRÍTICO - Comportamiento de Actualización:**
- **PUT es un reemplazo total, NO parcial** (igual que orders)
- **TODOS los campos son requeridos** en el request
- **Omitir un campo resultará en pérdida de datos**
- Si quieres mantener un valor sin cambios, envía el valor actual
- Para **limpiar campos opcionales** (`order_id`, `category_id`, `date`): envía `null`

**Request (ejemplo con todos los campos):**
```json
{
  "description": "Vinilo adhesivo premium (actualizado)",
  "amount": 90.00,
  "type": "order_linked",
  "order_id": 124,
  "category_id": 2,
  "date": "2026-04-18T12:00:00Z"
}
```

**Request (limpiando campos opcionales):**
```json
{
  "description": "Vinilo adhesivo premium",
  "amount": 90.00,
  "type": "general",
  "order_id": null,
  "category_id": null,
  "date": null
}
```

**Response (200 OK):**
```json
{
  "updated": true
}
```

**Response (404 Not Found):**
```json
{
  "message": "expense not found"
}
```

**Ejemplo correcto:**
```javascript
async function actualizarGasto(expenseId, expenseData) {
  // Siempre enviar TODOS los campos
  const payload = {
    description: expenseData.description,
    amount: expenseData.amount,
    type: expenseData.type,
    order_id: expenseData.order_id || null,
    category_id: expenseData.category_id || null,
    date: expenseData.date || null
  };
  
  return fetch(`http://localhost:3000/api/expenses/${expenseId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  }).then(r => r.json());
}

---

#### 📌 DELETE `/api/expenses/:id` - Eliminar Gasto

**Response (200 OK):**
```json
{
  "deleted": true
}
```

**Response (404 Not Found):**
```json
{
  "message": "expense not found"
}
```

**Ejemplo:**
```javascript
async function eliminarGasto(expenseId) {
  const response = await fetch(`http://localhost:3000/api/expenses/${expenseId}`, {
    method: 'DELETE'
  });
  
  if (!response.ok) {
    throw new Error('No se pudo eliminar el gasto');
  }
  
  return response.json();
}
```

---

### 5.5 Flujos de Integración para Expenses

#### Flujo: Crear Orden con Gastos Asociados

```javascript
async function crearOrdenConGastos() {
  // 1. Crear orden
  const { id: orderId } = await fetch('http://localhost:3000/api/orders', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      description: "Set 12 llaveros personalizados",
      amount_charged: 1000,
      client_name: "María López"
    })
  }).then(r => r.json());
  
  console.log(`✅ Orden ${orderId} creada`);
  
  // 2. Registrar gastos asociados a la orden
  const gastos = [
    { description: "Llaveros base metal", amount: 35 },
    { description: "Grabado láser", amount: 85 },
    { description: "Cadenas y argollas", amount: 45 }
  ];
  
  for (const gasto of gastos) {
    await fetch('http://localhost:3000/api/expenses', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        ...gasto,
        type: "order_linked",
        order_id: orderId,
        category_id: 1  // Materias primas
      })
    });
  }
  
  console.log(`✅ ${gastos.length} gastos registrados`);
  
  // 3. Calcular margen
  const totalGastos = gastos.reduce((sum, g) => sum + g.amount, 0);
  const margen = 1000 - totalGastos;
  const porcentajeMargen = (margen / 1000) * 100;
  
  console.log(`
    ---- Resumen Orden #${orderId} ----
    Precio venta: $1000
    Gastos totales: $${totalGastos}
    Margen bruto: $${margen} (${porcentajeMargen.toFixed(1)}%)
  `);
}
```

---

#### Flujo: Dashboard de Gastos del Mes

```javascript
async function dashboardGastosMes() {
  const inicio = '2026-04-01';
  const fin = '2026-04-30';
  
  // 1. Obtener todos los gastos del mes
  const expenses = await fetch(
    `http://localhost:3000/api/expenses?from=${inicio}&to=${fin}`
  ).then(r => r.json());
  
  // 2. Separar por tipo
  const generales = expenses.filter(e => e.type === 'general');
  const vinculados = expenses.filter(e => e.type === 'order_linked');
  
  // 3. Calcular totales
  const totalGeneral = generales.reduce((sum, e) => sum + e.amount, 0);
  const totalVinculado = vinculados.reduce((sum, e) => sum + e.amount, 0);
  const totalMes = totalGeneral + totalVinculado;
  
  // 4. Agrupar por categoría
  const porCategoria = {};
  expenses.forEach(exp => {
    if (exp.category_id) {
      if (!porCategoria[exp.category_id]) {
        porCategoria[exp.category_id] = 0;
      }
      porCategoria[exp.category_id] += exp.amount;
    }
  });
  
  console.log(`
    ---- Dashboard Gastos Abril 2026 ----
    Total gastos: $${totalMes}
    
    Por tipo:
    - Generales (fijos): $${totalGeneral} (${generales.length} gastos)
    - Vinculados a órdenes: $${totalVinculado} (${vinculados.length} gastos)
    
    Por categoría:
  `);
  
  for (const [categoryId, total] of Object.entries(porCategoria)) {
    console.log(`    - Categoría #${categoryId}: $${total}`);
  }
  
  return {
    totalMes,
    totalGeneral,
    totalVinculado,
    porCategoria,
    cantidadGastos: expenses.length
  };
}
```

---

#### Flujo: Calcular Rentabilidad de una Orden

```javascript
async function calcularRentabilidadOrden(orderId) {
  // 1. Obtener orden
  const order = await fetch(`http://localhost:3000/api/orders/${orderId}`)
    .then(r => r.json());
  
  // 2. Obtener ingresos (pagos)
  const paymentStatus = await fetch(
    `http://localhost:3000/api/orders/${orderId}/payment-status`
  ).then(r => r.json());
  
  // 3. Obtener gastos de la orden
  const expenses = await fetch(
    `http://localhost:3000/api/expenses/order/${orderId}`
  ).then(r => r.json());
  
  const totalGastos = expenses.reduce((sum, exp) => sum + exp.amount, 0);
  
  // 4. Calcular rentabilidad
  const ingresoReal = paymentStatus.total_paid;
  const ganancia = ingresoReal - totalGastos;
  const margen = (ganancia / ingresoReal) * 100;
  
  return {
    orden: order.description,
    precioVenta: order.amount_charged,
    ingresado: ingresoReal,
    pendienteCobro: paymentStatus.remaining,
    gastosDirectos: totalGastos,
    detalleGastos: expenses,
    gananciaActual: ganancia,
    margenActual: margen.toFixed(1) + '%',
    estaCompleto: paymentStatus.is_fully_paid
  };
}

// Uso
const rentabilidad = await calcularRentabilidadOrden(123);
console.log(`
  ---- Rentabilidad Orden #123 ----
  ${rentabilidad.orden}
  
  Precio de venta: $${rentabilidad.precioVenta}
  Ya cobrado: $${rentabilidad.ingresado}
  Pendiente: $${rentabilidad.pendienteCobro}
  
  Gastos directos: $${rentabilidad.gastosDirectos}
  Ganancia actual: $${rentabilidad.gananciaActual}
  Margen: ${rentabilidad.margenActual}
`);
```

---

## 6. Flujos de Integración Comunes

### 6.1 Crear Orden y Registrar Anticipo

```javascript
async function crearOrdenConAnticipo(orderData, anticipoAmount) {
  // 1. Crear orden
  const { id: orderId } = await fetch('http://localhost:3000/api/orders', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(orderData)
  }).then(r => r.json());
  
  // 2. Registrar anticipo
  await fetch('http://localhost:3000/api/incomes', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ 
      order_id: orderId, 
      amount: anticipoAmount 
    })
  });
  
  // 3. Obtener estado de pago actualizado
  const paymentStatus = await fetch(`http://localhost:3000/api/orders/${orderId}/payment-status`)
    .then(r => r.json());
  
  return { orderId, paymentStatus };
}

// Uso
const { orderId, paymentStatus } = await crearOrdenConAnticipo({
  description: "Set 12 llaveros personalizados",
  amount_charged: 1000,
  client_name: "Juan Pérez"
}, 500);

console.log(`Orden ${orderId} creada - ${paymentStatus.percentage_paid}% pagado`);
```

---

### 6.2 Listar Órdenes con Estado de Pago

```javascript
async function obtenerOrdenesConPagos(filters = {}) {
  // 1. Obtener órdenes
  const queryParams = new URLSearchParams(filters).toString();
  const orders = await fetch(`http://localhost:3000/api/orders?${queryParams}`)
    .then(r => r.json());
  
  // 2. Obtener estado de pago para cada orden
  const ordersWithPayment = await Promise.all(
    orders.map(async (order) => {
      const paymentStatus = await fetch(`http://localhost:3000/api/orders/${order.id}/payment-status`)
        .then(r => r.json());
      
      return { ...order, payment_status: paymentStatus };
    })
  );
  
  return ordersWithPayment;
}

// Uso
const orders = await obtenerOrdenesConPagos({ status: 'new' });

orders.forEach(order => {
  console.log(
    `${order.description} - ${order.payment_status.percentage_paid}% pagado`
  );
});
```

---

### 6.3 Workflow Completo de Orden

```javascript
async function workflowOrdenCompleto() {
  // 1. Crear orden
  const { id: orderId } = await fetch('http://localhost:3000/api/orders', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      description: "Set de 30 camisas personalizadas",
      amount_charged: 2000,
      client_name: "María González",
      delivery_type: "delivery",
      estimated_delivery_date: "2026-05-01T15:00:00Z"
    })
  }).then(r => r.json());
  
  console.log(`✅ Orden ${orderId} creada (estado: new)`);
  
  // 2. Registrar anticipo 50%
  await fetch('http://localhost:3000/api/incomes', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ order_id: orderId, amount: 1000 })
  });
  
  console.log(`✅ Anticipo $1000 registrado`);
  
  // 3. Iniciar producción
  await fetch(`http://localhost:3000/api/orders/${orderId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status: 'active' })
  });
  
  console.log(`✅ Orden en producción`);
  
  // 4. Marcar como lista
  await fetch(`http://localhost:3000/api/orders/${orderId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status: 'ready' })
  });
  
  console.log(`✅ Orden lista para entregar`);
  
  // 5. Registrar segundo pago (30%)
  await fetch('http://localhost:3000/api/incomes', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ order_id: orderId, amount: 600 })
  });
  
  console.log(`✅ Segundo pago $600 registrado`);
  
  // 6. Marcar como lista para entrega (no hay estado 'shipped', se va directo a completado)
  // Opcional: se puede mantener en 'ready' hasta la entrega
  
  // 7. Registrar pago final (20%)
  await fetch('http://localhost:3000/api/incomes', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ order_id: orderId, amount: 400 })
  });
  
  console.log(`✅ Pago final $400 registrado`);
  
  // 8. Finalizar orden (marca como 'completed')
  await fetch(`http://localhost:3000/api/orders/${orderId}/finish`, {
    method: 'POST'
  });
  
  console.log(`✅ Orden finalizada (estado: completed)`);
  
  // 9. Verificar estado final
  const order = await fetch(`http://localhost:3000/api/orders/${orderId}`)
    .then(r => r.json());
  
  const paymentStatus = await fetch(`http://localhost:3000/api/orders/${orderId}/payment-status`)
    .then(r => r.json());
  
  console.log(`
    Estado final:
    - Estado operacional: ${order.status}
    - Total pagado: $${paymentStatus.total_paid}
    - Pago completo: ${paymentStatus.is_fully_paid ? 'Sí' : 'No'}
  `);
}
```

---

## 7. ✨ Mejoras y Nuevas Características

### ✨ NUEVA CARACTERÍSTICA: Filtro de Múltiples Status

**Fecha de implementación:** 22 de abril, 2026  
**Retrocompatible:** ✅ Sí - El código anterior sigue funcionando

El endpoint `GET /api/orders` ahora acepta **múltiples valores de status** separados por comas, lo que permite filtrar órdenes que tengan cualquiera de los status especificados.

**Antes (aún funciona):**
```javascript
// ✅ Sigue funcionando - filtrar por un solo status
const nuevas = await fetch('http://localhost:3000/api/orders?status=new')
  .then(r => r.json());
```

**Ahora (mejorado):**
```javascript
// ✨ NUEVO - filtrar por múltiples status
const activas = await fetch('http://localhost:3000/api/orders?status=new,active,ready')
  .then(r => r.json());
// Retorna todas las órdenes que estén en estado 'new' O 'active' O 'ready'

// Ejemplo práctico: Ver todas las órdenes "en proceso" (no completadas ni canceladas)
const enProceso = await fetch('http://localhost:3000/api/orders?status=new,active,ready')
  .then(r => r.json());

// Combinar con otros filtros
const pendientesAbril = await fetch(
  'http://localhost:3000/api/orders?status=new,active&from=2026-04-01&to=2026-04-30'
).then(r => r.json());
```

**Casos de uso comunes:**

```javascript
// Dashboard: Ver órdenes que requieren atención
const requierenAtencion = await fetch(
  'http://localhost:3000/api/orders?status=new,active'
).then(r => r.json());

// Vista de "Listas para entrega"
const listasParaEntrega = await fetch(
  'http://localhost:3000/api/orders?status=ready'
).then(r => r.json());

// Vista de "Historial" (solo completadas y canceladas)
const historial = await fetch(
  'http://localhost:3000/api/orders?status=completed,cancelled'
).then(r => r.json());

// Vista de "Activas" (todo excepto completadas/canceladas)
const todasActivas = await fetch(
  'http://localhost:3000/api/orders?status=new,active,ready'
).then(r => r.json());
```

**Cómo funciona internamente:**
- El backend genera una query SQL con cláusula `IN`: `WHERE status IN ('new', 'active', 'ready')`
- Los espacios en blanco son ignorados automáticamente: `new, active` = `new,active`
- No hay límite en la cantidad de status (puedes enviar 1, 2, 3... todos los que necesites)

**Migración recomendada:**
1. **No requiere cambios** - Tu código actual seguirá funcionando
2. Identifica lugares donde filtras múltiples veces por status diferente
3. Reemplaza múltiples llamadas con una sola usando el nuevo formato

**Ejemplo de refactorización:**

```javascript
// ❌ ANTES: Múltiples llamadas
const nuevas = await fetch('http://localhost:3000/api/orders?status=new').then(r => r.json());
const activas = await fetch('http://localhost:3000/api/orders?status=active').then(r => r.json());
const listas = await fetch('http://localhost:3000/api/orders?status=ready').then(r => r.json());
const todasActivas = [...nuevas, ...activas, ...listas];

// ✅ AHORA: Una sola llamada
const todasActivas = await fetch(
  'http://localhost:3000/api/orders?status=new,active,ready'
).then(r => r.json());
```

---

## 8. Cambios Breaking que Debes Aplicar

### ⚠️ BREAKING CHANGE #1: Campo `client_name` ahora requerido

**Fecha de cambio:** 19 de abril, 2026  
**Implementación:** Modificación directa en migración original `000001_create_orders_table`

**Antes:**
```javascript
// ❌ Esto ya NO funciona - client_name era opcional
const order = await fetch('http://localhost:3000/api/orders', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    description: "Set 12 llaveros personalizados",
    amount_charged: 1000
    // client_name era opcional
  })
});
```

**Ahora:**
```javascript
// ✅ client_name es REQUERIDO
const order = await fetch('http://localhost:3000/api/orders', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    description: "Set 12 llaveros personalizados",
    amount_charged: 1000,
    client_name: "Juan Pérez"  // REQUERIDO
  })
});
```

**Interface TypeScript actualizada:**
```typescript
interface Order {
  id: number;
  description: string;
  amount_charged: number;
  status: OrderStatus;
  entry_date: string;
  estimated_delivery_date: string | null;
  delivery_type: DeliveryType;
  client_name: string;  // Ya no es "string | null"
  client_phone: string | null;
  notes: string | null;
  created_at: string;
  updated_at: string;
}
```

**Migración recomendada:**
1. Actualizar todos los formularios de creación de órdenes para hacer `client_name` obligatorio
2. Actualizar validaciones frontend antes de enviar la orden
3. Actualizar interfaces TypeScript para reflejar el cambio (`string` en lugar de `string | null`)

**Justificación:** Una orden siempre está relacionada a un cliente, por lo que el campo debe ser obligatorio desde el modelo de negocio.

**Nota:** Como el proyecto no está en producción, el cambio se aplicó directamente en la migración original de la tabla orders.

---

### ⚠️ BREAKING CHANGE #2: Eliminado campo `paid_50_percent`

**Antes:**
```javascript
// ❌ Esto ya NO funciona
const order = await fetch(`http://localhost:3000/api/orders/${id}`).then(r => r.json());

if (order.paid_50_percent) {
  mostrarBadge("Anticipo Pagado");
}
```

**Ahora:**
```javascript
// ✅ Usar endpoint de PaymentStatus
const paymentStatus = await fetch(`http://localhost:3000/api/orders/${id}/payment-status`)
  .then(r => r.json());

if (paymentStatus.percentage_paid >= 50) {
  mostrarBadge(`${paymentStatus.percentage_paid.toFixed(0)}% Pagado`);
}
```

**Migración recomendada:**
1. Buscar todos los usos de `order.paid_50_percent` en tu código
2. Reemplazar con llamadas a `/api/orders/:id/payment-status`
3. Actualizar lógica de badges/indicadores para usar `percentage_paid` dinámico

---

## 9. Manejo de Errores

El backend retorna errores en formato Fiber:

```json
{
  "message": "Error message here"
}
```

**Códigos de estado comunes:**

- `200` - OK
- `201` - Created (POST exitoso)
- `400` - Bad Request (validación fallida)
- `404` - Not Found (recurso no existe)
- `500` - Internal Server Error

**Ejemplo de manejo:**
```javascript
async function crearOrdenSafe(orderData) {
  try {
    const response = await fetch('http://localhost:3000/api/orders', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(orderData)
    });
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Error desconocido');
    }
    
    return await response.json();
  } catch (error) {
    console.error('Error al crear orden:', error);
    // Mostrar al usuario
    alert(`No se pudo crear la orden: ${error.message}`);
    throw error;
  }
}
```

---

## 10. Testing de Integración

### Test Manual con cURL

```bash
# 1. Crear orden
curl -X POST http://localhost:3000/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Botella térmica 1L personalizada",
    "amount_charged": 1000,
    "client_name": "Test User"
  }'

# Response: {"id": 1}

# 2. Registrar anticipo
curl -X POST http://localhost:3000/api/incomes \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "1",
    "amount": 500
  }'

# 3. Ver estado de pago
curl http://localhost:3000/api/orders/1/payment-status

# Response:
# {
#   "total_paid": 500,
#   "amount_charged": 1000,
#   "remaining": 500,
#   "percentage_paid": 50,
#   "is_fully_paid": false
# }

# 4. Actualizar estado
curl -X PUT http://localhost:3000/api/orders/1 \
  -H "Content-Type: application/json" \
  -d '{ "status": "in_progress" }'

# 5. Ver orden actualizada
curl http://localhost:3000/api/orders/1

# 6. Finalizar orden
curl -X POST http://localhost:3000/api/orders/1/finish

# 7. Listar todas las órdenes
curl http://localhost:3000/api/orders

# 8. Listar solo órdenes nuevas
curl "http://localhost:3000/api/orders?status=new"

# 9. ✨ Listar órdenes con múltiples status (nuevas, activas, listas)
curl "http://localhost:3000/api/orders?status=new,active,ready"
```

---

## 11. Checklist de Integración

### Antes de Comenzar

- [ ] Backend corriendo en `http://localhost:3000`
- [ ] Healthcheck exitoso: `GET /ping`
- [ ] Database conectado: `GET /dbping`
- [ ] Leer esta guía completa

### Durante Desarrollo

- [ ] Crear tipos TypeScript/interfaces para Order, Income, PaymentStatus
- [ ] Implementar servicio/API client para Orders
- [ ] Implementar servicio/API client para Incomes
- [ ] **Migrar** código que usaba `paid_50_percent` a `PaymentStatus`
- [ ] Implementar UI para múltiples pagos por orden
- [ ] Testear workflow completo: crear orden → pagar → actualizar estado → finalizar
- [ ] Manejar errores 400, 404, 500 apropiadamente
- [ ] Validar que CORS funciona desde tu dominio

### Validaciones Finales

- [ ] Crear orden sin campos opcionales funciona
- [ ] Crear orden con todos los campos funciona
- [ ] Listar órdenes sin filtros funciona
- [ ] Filtrar por status funciona
- [ ] Filtrar por fechas funciona
- [ ] Actualizar orden cambia `updated_at`
- [ ] Eliminar orden elimina sus incomes (CASCADE)
- [ ] Crear múltiples incomes para una orden funciona
- [ ] PaymentStatus calcula correctamente percentage_paid
- [ ] Finalizar orden cambia status a "completed"

### Expenses Module

- [ ] Crear tipos TypeScript para Expense y ExpenseCategory
- [ ] Implementar CRUD de categorías
- [ ] Implementar CRUD de gastos
- [ ] Distinguir gastos generales vs vinculados a órdenes
- [ ] Mostrar gastos por orden en vista de detalle
- [ ] Calcular totales y rentabilidad
- [ ] Filtrar por categoría y tipo
- [ ] Filtrar por rango de fechas

### No Intentes Usar

- [ ] Endpoints de `/api/users` (no implementado)
- [ ] Campo `order.paid_50_percent` (eliminado)

---

## 12. Soporte y Recursos

### Documentación del Backend

Lee estos documentos en `docs/`:
1. [estado-actual.md](docs/estado-actual.md) - Diagramas y schema actualizado
2. [modelo-de-datos.md](docs/modelo-de-datos.md) - Reglas de negocio
3. [decisiones-schema.md](docs/decisiones-schema.md) - Por qué se hicieron cambios
4. [migracion-sqlite.md](docs/migracion-sqlite.md) - Info sobre SQLite

### Recursos Útiles

- **Fiber Docs:** https://docs.gofiber.io
- **SQLite Docs:** https://www.sqlite.org/docs.html
- **ISO8601 Timestamps:** https://en.wikipedia.org/wiki/ISO_8601

### Troubleshooting

**Problema:** 404 en todos los endpoints  
**Solución:** Verifica que el backend esté corriendo en puerto 3000

**Problema:** CORS error  
**Solución:** El backend acepta todos los orígenes, verifica tu configuración de frontend

**Problema:** "order not found" al crear income  
**Solución:** Verifica que el `order_id` existe en la tabla orders

**Problema:** PaymentStatus retorna 404  
**Solución:** Verifica que la orden existe, el endpoint requiere una orden válida

**Problema:** Timestamps en formato incorrecto  
**Solución:** Usa formato ISO8601: `2026-04-18T10:30:00Z`

---

## 13. Próximos Pasos (Roadmap)

Estos módulos están pendientes de implementación en el backend:

### � Prioridad Alta
- [ ] **Users & Auth** - Autenticación y autorización
- [ ] **Financial Reports** - Reportes financieros automatizados

### 🟢 Prioridad Media
- [ ] History tracking de cambios de estado
- [ ] Webhooks para notificaciones
- [ ] Exportación de datos (CSV, Excel)

**Cuando estos módulos se implementen, se actualizará esta guía.**

---

**Generado por:** Backend Review Agent  
**Fecha:** 18 de abril, 2026  
**Última actualización:** 19 de abril, 2026 (Corrección de filtros de fecha y advertencias PUT)  
**Versión de la API:** SQLite - Post-migración  
**Estado:** ✅ Producción Lista para Integración Frontend  
**Módulos disponibles:** Orders (7 endpoints), Incomes (5 endpoints), Expenses (11 endpoints)
