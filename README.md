# BandIA

Repositorio inicial del proyecto BandIA.

Banda virtual de agentes de IA: autonomía creativa, memoria y aprendizaje del feedback de redes. Nombre artístico inicial: **Fuera de Hora**.

## Estado

Base inicial en **Go**, usando sólo la biblioteca estándar. Incluye servidor HTTP, apagado ordenado, logs JSON y pruebas del contrato HTTP.

**Todavía no ejecuta reuniones, genera música ni publica en X.** La persistencia, el scheduler, los agentes y las integraciones son próximos hitos. No necesita claves de API para arrancar esta base y no realiza llamadas pagas.

## Ejecutar

Instalar una versión soportada de [Go](https://go.dev/dl/) compatible con `go.mod` (Go 1.26 o posterior).

```sh
go run ./cmd/bandia
```

Por defecto escucha en `127.0.0.1:8080`. Para cambiarlo:

```sh
BANDIA_HTTP_ADDR=127.0.0.1:9090 go run ./cmd/bandia
```

`.env.example` documenta las variables; el servidor no carga archivos `.env` automáticamente. No hay autenticación en esta base local; la autenticación de administración debe implementarse antes de exponer operaciones de la banda.

| Endpoint | Resultado | Significado |
| --- | --- | --- |
| `GET /healthz` | 200 | El proceso HTTP responde. |
| `GET /readyz` | 503 | El runtime de la banda aún no está implementado. |

Ambos aceptan `HEAD` sin cuerpo. Las rutas desconocidas devuelven 404 y los métodos no admitidos, 405. No usar `/readyz` como condición para esperar una banda operativa en este hito: su 503 es intencional.

## Desarrollo

```sh
make fmt
make check
make build
```

`make check` requiere Go y un compilador C compatible para el detector de carreras. La validación en un entorno con Go queda pendiente de ejecución en esta entrega: el entorno de preparación no tenía el compilador disponible.

## Organización

- `cmd/bandia`: arranque, configuración HTTP y señales.
- `internal/httpapi`: endpoints operativos y pruebas.
- `docs/spec.md`: especificación del proyecto y decisiones acordadas.
- `docs/decisions/0001-go.md`: Go como lenguaje principal; Python auxiliar opcional.
- `docs/roadmap.md`: siguientes entregables y evidencia de aceptación.
- `docs/api-implementation-plan.md`: catálogo de endpoints, permisos, modelo de datos, estados y plan completo por fases.

Go controlará endpoints, orquestación, persistencia e integraciones. Python podrá usarse para procesamiento especializado cuando haya una necesidad concreta, con entradas/salidas explícitas; no se agregan scripts vacíos.

## Persistencia

GitHub guarda código y diseño. Una base persistente guardará memoria y decisiones; el almacenamiento de objetos guardará audio y medios. Esos componentes aún no existen en este repositorio. No subir credenciales, datos de producción, comentarios originales de terceros ni archivos de audio al historial Git.

## Próximo hito

Implementar repositorio durable de reuniones y su máquina de estados, con creación idempotente y recuperación después de reiniciar. Conectar agentes reales sólo después de acordar proveedor, credenciales y presupuesto. Ver [roadmap](docs/roadmap.md).

El [plan de implementación de la API](docs/api-implementation-plan.md) define el ciclo completo y el primer PR funcional. Los workers llevarán las decisiones a tareas y resultados sin requerir llamadas manuales por cada paso.
