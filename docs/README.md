# Documentación Ebzer API

Bienvenido a la documentación del backend de Ebzer, un sistema de gestión de órdenes, ingresos y gastos.

## ✅ Estado Actual

**Estado:** Ready for Testing & UI Integration

**Los problemas críticos del schema han sido resueltos.**

**Lee estos documentos primero:**
1. 🎯 [Estado Actual](./estado-actual.md) - Schema actualizado, diagramas y guía de integración **← EMPIEZA AQUÍ**
2. 📊 [Modelo de Datos](./modelo-de-datos.md) - Análisis completo del schema
3. 📝 [Decisiones de Schema](./decisiones-schema.md) - Historial de decisiones tomadas

## 📋 Contenido

### 🎯 Inicio Rápido
- **[Estado Actual](./estado-actual.md)** - Schema actualizado, diagramas, API endpoints y guía de integración
- **[Migración a SQLite](./migracion-sqlite.md)** - Guía completa de la migración PostgreSQL → SQLite

### Documentación Técnica
- [Tecnologías](./tecnologias.md) - Stack tecnológico y herramientas utilizadas
- [Arquitectura](./arquitectura.md) - Estructura del proyecto y relaciones entre componentes

### 🗄️ Base de Datos
- **[Modelo de Datos](./modelo-de-datos.md)** - Análisis detallado del schema y relaciones
- **[Decisiones de Schema](./decisiones-schema.md)** - Historial de decisiones y cambios aplicados

## ✅ Cambios Implementados

- ✅ **Fix #1:** Eliminado constraint UNIQUE en `income.order_id` - ahora soporta múltiples pagos por orden
- ✅ **Fix #2:** Estados operacionales simplificados (4 estados: new → active → ready → completed, cancelled) optimizado para emprendimiento pequeño
- ✅ **Fix #3:** Eliminado campo `paid_50_percent` - estado financiero calculado en runtime
- ✅ **Feature:** Nuevo endpoint `/api/orders/:id/payment-status` para obtener estado de pago calculado
- ✅ **Migration:** PostgreSQL → SQLite (simplifica despliegue)
- ✅ **Auto-migrations:** Sistema automático de migrations al iniciar
- ✅ **Fix #4:** Campo `client_name` ahora es requerido (NOT NULL) - una orden siempre está asociada a un cliente
- 🟡 **Tablas mantenidas sin uso:** `users`, `financial_percentages` (para futuro uso)

## Inicio Rápido

```bash
# Instalar dependencias
go mod download

# Iniciar servidor (migrations automáticas)
go run cmd/server/main.go

# La base de datos SQLite se creará en: ./data/ebzer.db
# Las migrations se ejecutan automáticamente

# Healthcheck
curl http://localhost:3000/ping
```

**Ver:** [Guía de Migración a SQLite](./migracion-sqlite.md) para detalles completos

## Estado de la Documentación

- [x] Tecnologías
- [x] Arquitectura
- [x] Modelo de Datos
- [x] Decisiones de Schema
- [x] Estado Actual (con diagramas y API endpoints)
- [ ] Variables de Entorno
- [ ] Guía de Desarrollo
- [ ] Testing
- [ ] Deployment

## 🧪 Próximos Pasos

1. **Testing** - Implementar tests unitarios, de integración y E2E
2. **UI Integration** - Actualizar frontend con nuevo schema y endpoints
3. **Validaciones** - Agregar validación de transiciones de estado
4. **Documentación** - Completar guía de variables de entorno y desarrollo
