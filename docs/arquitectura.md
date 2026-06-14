# Arquitectura - Ebzer API

## Patrón Arquitectónico: Clean Architecture

El proyecto sigue una arquitectura en capas con separación clara de responsabilidades.

## Flujo de Dependencias

```
main.go
  ├─→ DB Connection
  ├─→ Module Setup (Orders, Incomes, etc.)
  │     ├─→ Handler (HTTP Layer)
  │     │     └─→ Service (Business Logic)
  │     │           └─→ Repository (Data Access)
  └─→ Fiber App Configuration
```

## Capas y Responsabilidades

### 1. Entry Point (`cmd/server/main.go`)
**Responsabilidad**: Inicialización y orquestación
- Establece conexión a base de datos
- Configura servidor HTTP (Fiber)
- Registra middlewares globales (CORS, Logger)
- Inicializa módulos de negocio
- Define health checks

**Relaciones**: 
- Consume `internal/db` para conectar a PostgreSQL
- Inicializa handlers de cada módulo (`orders`, `incomes`)

---

### 2. Database Layer (`internal/db/`)
**Responsabilidad**: Gestión de conexiones y esquema
- Provee pool de conexiones PostgreSQL
- Contiene migraciones SQL versionadas

**Estructura de Migraciones**:
```
000001 → orders table
000002 → expense_categories table
000003 → income table
000004 → expenses table
000005 → user table
000006 → financial_percentages table
000007 → history_orders_status table
```

**Relaciones**:
- Inyectada en main → Pasada a repositories

---

### 3. Module Layer (Feature-based)
Cada módulo de negocio sigue el mismo patrón de 3 capas:

#### **Handler** (HTTP Transport)
- Recibe peticiones HTTP (Fiber Context)
- Parsea DTOs de request
- Valida formato básico
- Delega lógica al Service
- Formatea respuestas HTTP

**Responsabilidades**:
- Routing (`RegisterRoutes`)
- Serialización/deserialización
- Códigos de estado HTTP
- Manejo de errores HTTP

**Relaciones**: Handler → Service (unidireccional)

---

#### **Service** (Business Logic)
- Contiene reglas de negocio
- Validaciones complejas
- Orquestación de operaciones
- Transformación de datos

**Responsabilidades**:
- Validación de negocio (ej: `AmountCharged >= 0`)
- Lógica de workflows (ej: `FinishOrder`)
- Manejo de errores de negocio

**Relaciones**: Service → Repository (unidireccional)

---

#### **Repository** (Data Access)
- Abstracción de base de datos
- Queries SQL
- Mapeo de datos

**Responsabilidades**:
- CRUD operations
- Queries complejas
- Transacciones (si aplica)

**Relaciones**: Repository → DB Connection

---

#### **Models** 
- Entidades de dominio
- Estructuras de datos del negocio

---

#### **DTO** (Data Transfer Objects)
- Contratos de API
- Request/Response payloads
- Validación de estructura

---

## Módulos de Negocio

### Orders Module (`internal/orders/`)
**Domain**: Gestión de órdenes de trabajo

**Endpoints**:
- `POST /api/orders` - Crear orden
- `GET /api/orders` - Listar con filtros (status, from, to)
- `GET /api/orders/:id` - Obtener detalle
- `GET /api/orders/:id/payment-status` - **NUEVO**: Obtener estado de pago calculado
- `PUT /api/orders/:id` - Actualizar
- `POST /api/orders/:id/finish` - Finalizar orden (marca como `completed`)
- `DELETE /api/orders/:id` - Eliminar

**Dependencias:**
- `OrdersRepository` - Acceso a datos de orders
- `IncomesRepository` - Necesario para calcular PaymentStatus

**Relaciones con otras tablas**:
- `income` - Una orden puede tener múltiples income records (pagos parciales)

---

### Incomes Module (`internal/incomes/`)
**Domain**: Gestión de ingresos (pagos de órdenes)

**Endpoints**:
- `POST /api/incomes` - Registrar ingreso
- `GET /api/incomes` - Listar con filtros (from, to)
- `GET /api/incomes/:id` - Obtener detalle
- `PUT /api/incomes/:id` - Actualizar
- `DELETE /api/incomes/:id` - Eliminar

**Dependencias:**
- `IncomesRepository` - Acceso a datos de income

**Características importantes:**
- Una orden puede tener **múltiples** income records (pagos parciales)
- El repository incluye método `GetByOrderID()` para obtener todos los pagos de una orden

---

## Principios Aplicados

✅ **Separation of Concerns** - Cada capa tiene una única responsabilidad
✅ **Dependency Inversion** - Las capas internas no conocen las externas
✅ **Interface-based Design** - Services definen interfaces para testability
✅ **Feature-based Organization** - Código organizado por dominio, no por tipo

## Puntos de Mejora Identificados

### Pendientes
- [ ] Validación de transiciones de estado (ej: no permitir saltar de `new` a `completed` sin pasar por `active`)
- [ ] Sistema de migraciones automatizado (considerar `golang-migrate`)
- [ ] Gestión de variables de entorno (considerar `godotenv`)
- [ ] CORS más restrictivo para producción
- [ ] Tests unitarios y de integración
- [ ] Documentación de variables de entorno

### Completados
- [x] Separación de estado operacional y financiero
- [x] Soporte para múltiples pagos por orden
- [x] Cálculo de PaymentStatus en runtime
- [x] Eliminación de campos redundantes (`paid_50_percent`)
- [x] Estados operacionales claramente definidos

⚠️ **Falta capa de validación dedicada** - Los DTOs podrían usar tags de validación (ej: `validator` package)
⚠️ **Sin middleware de autenticación** - No hay control de acceso visible
⚠️ **Sin manejo centralizado de errores** - Cada handler maneja errores de forma similar (oportunidad de DRY)
⚠️ **Repository sin interfaz explícita** - Dificulta testing con mocks
⚠️ **Sin transacciones visibles** - Operaciones complejas podrían necesitar atomicidad

## Diagrama de Flujo de Request

```
HTTP Request
    ↓
[Fiber Middlewares: CORS, Logger]
    ↓
[Handler] - Parse DTO, valida formato
    ↓
[Service] - Valida reglas de negocio
    ↓
[Repository] - Ejecuta query SQL
    ↓
[PostgreSQL Database]
    ↓
[Repository] - Mapea resultado
    ↓
[Service] - Transforma datos
    ↓
[Handler] - Formatea JSON response
    ↓
HTTP Response
```
