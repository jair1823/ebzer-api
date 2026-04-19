#!/bin/bash

# Script de gestión del servidor ebzer-api

case "$1" in
  start)
    echo "🚀 Iniciando servidor ebzer-api..."
    if lsof -Pi :3000 -sTCP:LISTEN -t >/dev/null 2>&1 ; then
      echo "❌ El servidor ya está corriendo en puerto 3000"
      exit 1
    fi
    
    nohup go run cmd/server/main.go > server.log 2>&1 &
    SERVER_PID=$!
    echo $SERVER_PID > .server.pid
    
    # Esperar a que el servidor inicie
    sleep 2
    
    if curl -s http://localhost:3000/ping > /dev/null 2>&1; then
      echo "✅ Servidor iniciado exitosamente (PID: $SERVER_PID)"
      echo "📊 Servidor corriendo en http://localhost:3000"
      echo "📝 Logs en: server.log"
    else
      echo "❌ Error al iniciar el servidor. Ver server.log para detalles."
      exit 1
    fi
    ;;
    
  stop)
    echo "🛑 Deteniendo servidor ebzer-api..."
    if [ -f .server.pid ]; then
      PID=$(cat .server.pid)
      if kill -0 $PID 2>/dev/null; then
        kill $PID
        rm .server.pid
        echo "✅ Servidor detenido (PID: $PID)"
      else
        echo "⚠️  El proceso $PID no existe, limpiando archivos..."
        rm .server.pid
      fi
    else
      # Intentar detener por puerto
      if lsof -Pi :3000 -sTCP:LISTEN -t >/dev/null 2>&1 ; then
        lsof -ti :3000 | xargs kill -9 2>/dev/null
        echo "✅ Servidor detenido"
      else
        echo "⚠️  No hay servidor corriendo"
      fi
    fi
    
    # Limpiar procesos huérfanos
    pkill -9 -f "cmd/server/main.go" 2>/dev/null
    ;;
    
  restart)
    echo "🔄 Reiniciando servidor..."
    $0 stop
    sleep 1
    $0 start
    ;;
    
  status)
    if lsof -Pi :3000 -sTCP:LISTEN -t >/dev/null 2>&1 ; then
      PID=$(lsof -ti :3000)
      echo "✅ Servidor corriendo (PID: $PID)"
      echo "📊 URL: http://localhost:3000"
      
      # Probar el servidor
      if curl -s http://localhost:3000/ping > /dev/null 2>&1; then
        echo "✅ Health check OK"
      else
        echo "❌ El servidor no responde correctamente"
      fi
    else
      echo "❌ Servidor no está corriendo"
      exit 1
    fi
    ;;
    
  logs)
    if [ -f server.log ]; then
      tail -f server.log
    else
      echo "❌ No se encontró server.log"
      exit 1
    fi
    ;;
    
  test)
    echo "🧪 Ejecutando pruebas de API..."
    if ! lsof -Pi :3000 -sTCP:LISTEN -t >/dev/null 2>&1 ; then
      echo "❌ El servidor no está corriendo. Inicia el servidor primero:"
      echo "   ./server.sh start"
      exit 1
    fi
    
    if [ -x ./test-api.sh ]; then
      ./test-api.sh
    else
      echo "❌ test-api.sh no encontrado o no es ejecutable"
      exit 1
    fi
    ;;
    
  reset-db)
    echo "🗑️  Resetear base de datos..."
    read -p "¿Estás seguro? Esto eliminará todos los datos. (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
      $0 stop
      rm -f ./data/ebzer.db
      echo "✅ Base de datos eliminada"
      echo "💡 Inicia el servidor para recrear la base de datos:"
      echo "   ./server.sh start"
    else
      echo "❌ Cancelado"
    fi
    ;;
    
  db)
    if [ -f ./data/ebzer.db ]; then
      sqlite3 ./data/ebzer.db
    else
      echo "❌ Base de datos no encontrada en ./data/ebzer.db"
      echo "💡 Inicia el servidor primero para crear la base de datos:"
      echo "   ./server.sh start"
      exit 1
    fi
    ;;
    
  *)
    echo "Script de gestión del servidor ebzer-api"
    echo ""
    echo "Uso: $0 {start|stop|restart|status|logs|test|reset-db|db}"
    echo ""
    echo "Comandos:"
    echo "  start     - Iniciar el servidor en background"
    echo "  stop      - Detener el servidor"
    echo "  restart   - Reiniciar el servidor"
    echo "  status    - Ver el estado del servidor"
    echo "  logs      - Ver los logs en tiempo real (Ctrl+C para salir)"
    echo "  test      - Ejecutar pruebas de API"
    echo "  reset-db  - Eliminar la base de datos (requiere confirmación)"
    echo "  db        - Abrir consola interactiva de SQLite"
    echo ""
    exit 1
    ;;
esac
