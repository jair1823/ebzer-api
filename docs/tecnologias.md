# Stack Tecnológico - Ebzer API

## Lenguaje Principal
- **Go 1.25.5** - Lenguaje de programación compilado, tipado estático y alto rendimiento

## Framework Web
- **Fiber v2.52.10** - Framework web HTTP de alto rendimiento inspirado en Express
  - Construido sobre Fasthttp, uno de los motores HTTP más rápidos de Go
  - Manejo eficiente de rutas y middlewares
  - Timeouts configurados (15s lectura/escritura)

## Base de Datos
- **SQLite** - Base de datos relacional embebida
- **go-sqlite3 v1.14.22** - Driver CGO de SQLite para Go
  - Base de datos en archivo único
  - Sin servidor separado necesario
  - ACID compliant
  - Soporte completo de foreign keys
  - WAL mode para mejor concurrencia

**Configuración SQLite:**
- Foreign keys habilitadas
- WAL (Write-Ahead Logging) mode para mejor performance
- Conexión única (óptimo para SQLite)
- Archivo de base de datos: `./data/ebzer.db` (configurable vía env)

## Arquitectura y Patrones
- **Clean Architecture** - Separación por capas (Handler → Service → Repository)
- **Dependency Injection** - Inyección manual de dependencias
- **Migration-based Schema** - Control de versiones de base de datos con archivos `.sql`
- **Auto-migrations** - Las migrations se ejecutan automáticamente al iniciar la aplicación

## Middlewares
- **CORS** - Configurado para aceptar cualquier origen (⚠️ considerar restringir en producción)
- **Logger** - Registro de todas las peticiones HTTP

## Estructura del Módulo
```
Module: creaciones-api
```

## Variables de Entorno

### Base de Datos
- `SQLITE_DB_PATH` - Ruta al archivo de base de datos SQLite (default: `./data/ebzer.db`)

## Notas de Mejora
⚠️ **CORS configurado con `AllowOrigins: "*"`** - Potencial riesgo de seguridad para producción  
✅ **Sistema de migraciones automático** - Migrations se ejecutan al iniciar  
⚠️ **Sin gestión de secrets/env** - Considera usar `godotenv` o similar para otras configuraciones
