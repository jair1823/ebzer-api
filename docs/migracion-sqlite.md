# Migración a SQLite - Guía Completa

**Fecha:** 18 de abril, 2026  
**Estado:** ✅ **COMPLETADO**

---

## 📋 Resumen de Cambios

La migración de PostgreSQL a SQLite se ha completado exitosamente. Este cambio simplifica el despliegue y reduce la complejidad de infraestructura para ebzer-api.

### ✅ Cambios Implementados

| Componente | Antes (PostgreSQL) | Ahora (SQLite) |
|------------|-------------------|----------------|
| **Driver** | `pgx/v5` | `mattn/go-sqlite3` |
| **Conexión** | TCP (host:port) | Archivo local |
| **Config** | 5 variables env | 1 variable env |
| **Pooling** | 25 conexiones | 1 conexión |
| **Migrations** | Manuales | Automáticas |

---

## 🔄 Adaptaciones del Schema

### Mapeo de Tipos de Datos

| PostgreSQL | SQLite | Notas |
|------------|--------|-------|
| `SERIAL` | `INTEGER PRIMARY KEY AUTOINCREMENT` | Auto-incremento |
| `NUMERIC(10,2)` | `REAL` | Números decimales |
| `TIMESTAMP` | `TEXT` | Formato ISO8601 |
| `VARCHAR(n)` | `TEXT` | Sin límite de tamaño |
| `ENUM` | `TEXT + CHECK` | Validación con constraints |
| `NOW()` | `datetime('now')` | Función de fecha actual |

### Ejemplo de Conversión

**PostgreSQL:**
```sql
CREATE TYPE order_status AS ENUM ('confirmed', 'in_progress', 'ready');

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    status order_status NOT NULL DEFAULT 'confirmed',
    amount NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**SQLite:**
```sql
CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    status TEXT NOT NULL DEFAULT 'confirmed' 
        CHECK(status IN ('confirmed', 'in_progress', 'ready')),
    amount REAL NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
```

---

## 🗄️ Sistema de Migrations

### Automático al Iniciar

Las migrations ahora se ejecutan **automáticamente** cuando inicias la aplicación:

```go
// En cmd/server/main.go
conn, err := db.Connect()
if err != nil {
    log.Fatalf("Failed to connect to the database: %v", err)
}

// Automáticamente ejecuta todas las migrations pendientes
if err := db.RunMigrations(conn, "internal/db/migrations"); err != nil {
    log.Fatalf("Failed to run migrations: %v", err)
}
```

### Tabla de Control

SQLite mantiene un registro de migrations aplicadas:

```sql
SELECT * FROM schema_migrations;

-- Resultado:
-- version                          | applied_at
-- ---------------------------------|-------------------------
-- 000001_create_orders_table       | 2026-04-18 10:30:00
-- 000002_create_expense_categories | 2026-04-18 10:30:00
-- 000003_create_income_table       | 2026-04-18 10:30:01
-- ...
```

### Gestión de Migrations

```bash
# Las migrations están en:
internal/db/migrations/
├── 000001_create_orders_table.up.sql
├── 000001_create_orders_table.down.sql
├── 000002_create_expense_categories_table.up.sql
├── 000002_create_expense_categories_table.down.sql
└── ...

# Se ejecutan automáticamente en orden numérico
# Solo las pendientes se aplican (idempotente)
```

---

## ⚙️ Configuración

### Variable de Entorno

```bash
# Ruta al archivo de base de datos
export SQLITE_DB_PATH="./data/ebzer.db"

# O usa el default (./data/ebzer.db)
```

### Características Habilitadas

El sistema configura SQLite con:

- ✅ **Foreign Keys ON** - Integridad referencial
- ✅ **WAL Mode** - Write-Ahead Logging para mejor concurrencia
- ✅ **Single Connection** - Óptimo para SQLite
- ✅ **Auto-create directory** - Crea `./data/` si no existe

---

## 🚀 Cómo Usar

### Iniciar desde Cero

```bash
# 1. Elimina base de datos anterior (si existe)
rm -rf ./data/ebzer.db

# 2. Inicia la aplicación
go run cmd/server/main.go

# Output esperado:
# Connecting to SQLite: ./data/ebzer.db
# 🔄 Running database migrations...
# 🔄 Applying migration: 000001_create_orders_table
# ✅ Successfully applied: 000001_create_orders_table
# 🔄 Applying migration: 000002_create_expense_categories_table
# ✅ Successfully applied: 000002_create_expense_categories_table
# ...
# ✅ Migrations completed successfully
```

### Reiniciar con Datos Limpios

```bash
# Simplemente elimina el archivo
rm ./data/ebzer.db

# Las migrations se re-ejecutarán automáticamente
go run cmd/server/main.go
```

---

## 📊 Ventajas de SQLite

### ✅ Pros

| Ventaja | Detalle |
|---------|---------|
| **Sin servidor** | No necesitas PostgreSQL corriendo |
| **Portabilidad** | Un solo archivo = toda tu base de datos |
| **Simplicidad** | Cero configuración de infraestructura |
| **Performance** | Muy rápido para workloads pequeños/medianos |
| **Backups fáciles** | `cp ebzer.db ebzer.backup.db` |
| **Desarrollo local** | Ideal para desarrollo y testing |

### ⚠️ Limitaciones (para este proyecto NO son problema)

| Limitación | ¿Aplica a Ebzer? |
|------------|------------------|
| Sin concurrencia de escritura masiva | ❌ No aplica (negocio pequeño) |
| Sin replicación nativa | ❌ No aplica (single instance) |
| Sin usuarios/permisos | ❌ No aplica (auth en app layer) |
| Límite de tamaño ~140TB | ❌ No aplica (nunca alcanzaremos esto) |

---

## 🔍 Verificación Post-Migración

### 1. Verifica que la Base de Datos Existe

```bash
ls -lh ./data/ebzer.db
# -rw-r--r-- 1 user user 20K Apr 18 10:30 ./data/ebzer.db
```

### 2. Inspecciona el Schema

```bash
# Instala sqlite3 CLI si no lo tienes
sudo apt-get install sqlite3  # Ubuntu/Debian
brew install sqlite3          # macOS

# Conecta a la base de datos
sqlite3 ./data/ebzer.db

# Comandos útiles:
.tables                    # Lista todas las tablas
.schema orders            # Ver schema de una tabla
SELECT * FROM schema_migrations;  # Ver migrations aplicadas
```

### 3. Verifica Foreign Keys

```sql
PRAGMA foreign_keys;
-- Debería retornar: 1 (habilitado)

PRAGMA foreign_key_list(income);
-- Debería mostrar FK a orders(id)
```

### 4. Test Endpoints

```bash
# Prueba que todo funciona
curl http://localhost:3000/ping
curl http://localhost:3000/dbping

# Crea una orden
curl -X POST http://localhost:3000/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Test order",
    "amount_charged": 100.00
  }'
```

---

## 🧪 Testing con SQLite

### In-Memory Database para Tests

```go
// Para tests unitarios, puedes usar una DB en memoria
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatal(err)
    }
    
    // Ejecuta migrations
    if err := RunMigrations(db, "../../internal/db/migrations"); err != nil {
        t.Fatal(err)
    }
    
    return db
}
```

---

## 📦 Backup y Restore

### Backup

```bash
# Simple copy (asegúrate de que la app no esté escribiendo)
cp ./data/ebzer.db ./backups/ebzer-$(date +%Y%m%d).db

# O usa el comando SQLite
sqlite3 ./data/ebzer.db ".backup ./backups/ebzer-backup.db"
```

### Restore

```bash
# Detén la aplicación
# Reemplaza el archivo
cp ./backups/ebzer-20260418.db ./data/ebzer.db
# Reinicia la aplicación
```

---

## 🔧 Troubleshooting

### Error: "database is locked"

**Causa:** Múltiples procesos intentando escribir simultáneamente.

**Solución:**
```bash
# 1. Asegúrate de que solo hay una instancia corriendo
ps aux | grep "cmd/server/main"

# 2. Verifica WAL mode está habilitado
sqlite3 ./data/ebzer.db "PRAGMA journal_mode;"
# Debería retornar: wal
```

### Error: "foreign key constraint failed"

**Causa:** Intentando insertar/actualizar con FK inválida.

**Debug:**
```sql
PRAGMA foreign_keys = ON;
PRAGMA foreign_key_check;
```

### Error: "unable to open database file"

**Causa:** Directorio no existe o permisos incorrectos.

**Solución:**
```bash
# Crea el directorio
mkdir -p ./data

# Verifica permisos
chmod 755 ./data
```

---

## 🎯 Próximos Pasos

- [x] Migrar schema a SQLite
- [x] Implementar sistema de migrations automáticas
- [x] Actualizar documentación
- [ ] Implementar tests con in-memory DB
- [ ] Crear script de backup automático
- [ ] Documentar estrategia de deploy

---

## 📚 Referencias

- [SQLite Docs](https://www.sqlite.org/docs.html)
- [go-sqlite3 GitHub](https://github.com/mattn/go-sqlite3)
- [SQLite Datatypes](https://www.sqlite.org/datatype3.html)
- [SQLite Foreign Keys](https://www.sqlite.org/foreignkeys.html)

---

**Última actualización:** 18 de abril, 2026  
**Estado:** ✅ Migration Completed Successfully
