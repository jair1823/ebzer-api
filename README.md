# ebzer-api

API backend para gestión de pedidos y finanzas de creaciones artesanales.

## 🚀 Inicio Rápido

```bash
# Iniciar servidor
./server.sh start

# Ver estado
./server.sh status

# Ejecutar pruebas
./server.sh test

# Ver logs
./server.sh logs

# Detener servidor
./server.sh stop
```

El servidor iniciará en `http://localhost:3000`

## 📚 Documentación Completa

Ver [SETUP.md](SETUP.md) para documentación detallada sobre:
- Configuración del entorno
- Endpoints disponibles
- Ejemplos de uso
- Solución de problemas

## 🛠️ Stack Tecnológico

- **Lenguaje**: Go 1.25+
- **Framework Web**: Fiber v2
- **Base de Datos**: SQLite 3
- **Driver**: mattn/go-sqlite3
- **Arquitectura**: Clean Architecture

## 📊 Estructura del Proyecto

```
cmd/server/          - Punto de entrada de la aplicación
internal/
  db/                - Conexión y migraciones de base de datos
    migrations/      - Scripts SQL de migraciones  
  orders/            - Dominio de órdenes/pedidos
  incomes/           - Dominio de ingresos/pagos
docs/                - Documentación del proyecto
```

## 🗄️ Base de Datos

Base de datos SQLite en `./data/ebzer.db`

Las migraciones se ejecutan automáticamente al iniciar el servidor.

```bash
# Acceder a la base de datos
./server.sh db

# Resetear la base de datos
./server.sh reset-db
```

## 🧪 Pruebas

```bash
# Pruebas automatizadas
./server.sh test

# O directamente
./test-api.sh
```

## 📝 API Endpoints

### Health Checks
- `GET /ping` - Verificar servidor
- `GET /dbping` - Verificar base de datos

### Orders (Pedidos)
- `POST /api/orders` - Crear pedido
- `GET /api/orders` - Listar pedidos
- `GET /api/orders/:id` - Obtener pedido
- `PUT /api/orders/:id` - Actualizar pedido
- `DELETE /api/orders/:id` - Eliminar pedido
- `GET /api/orders/:id/payment-status` - Ver estado de pago
- `POST /api/orders/:id/finish` - Finalizar pedido

### Incomes (Pagos)
- `POST /api/incomes` - Registrar pago
- `GET /api/incomes` - Listar pagos
- `GET /api/incomes/:id` - Obtener pago
- `PUT /api/incomes/:id` - Actualizar pago
- `DELETE /api/incomes/:id` - Eliminar pago

## ⚙️ Configuración

### Variables de Entorno

```bash
# Ruta de la base de datos (opcional)
export SQLITE_DB_PATH=./data/ebzer.db

# Iniciar con configuración personalizada
SQLITE_DB_PATH=/custom/path.db ./server.sh start
```

## 🐛 Solución de Problemas

### Puerto 3000 en uso
```bash
./server.sh stop
```

### Base de datos corrupta
```bash
./server.sh reset-db
./server.sh start
```

### Ver logs
```bash
./server.sh logs
# O
cat server.log
```

## 📖 Más Información

- [SETUP.md](SETUP.md) - Guía completa de configuración y uso
- [docs/](docs/) - Documentación técnica del proyecto

## ✅ Estado del Proyecto

- ✅ Configuración completada
- ✅ Migraciones SQLite implementadas
- ✅ API de Orders funcional
- ✅ API de Incomes funcional
- ✅ Cálculo de estado de pago
- ✅ Filtros por status y fechas
- ✅ Pruebas automatizadas
- ✅ Scripts de gestión

---

**Servidor**: http://localhost:3000

**Base de Datos**: `./data/ebzer.db`
