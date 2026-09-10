# Validación del primer runtime local

10 de septiembre de 2026. Código validado: [`9815b65`](https://github.com/MartinRusso28/bandia/commit/9815b6501cf67cc416986396d7d3fb3d638c4c2c).

**Resultado: aprobado** en [GitHub Actions, ejecución 34429074792](https://github.com/MartinRusso28/bandia/actions/runs/34429074792). Los jobs `test` y `compose-smoke` finalizaron con `success`. Esto valida el primer incremento local, no agentes, música, redes ni operación autónoma.

## Evidencia

| Verificación | Entorno | Resultado |
| --- | --- | --- |
| Build Go y `go vet ./...` | Entorno de desarrollo y CI, Go 1.27.1 | Aprobado |
| Tests unitarios Go con `-race` | Desarrollo y CI | Aprobado |
| Formato Go | Desarrollo y CI | Sin diferencias de gofmt |
| Inicialización de credenciales | Desarrollo y CI | Secretos locales aleatorios, archivo protegido y sin sobreescritura |
| Tests PostgreSQL con `-race` | CI, PostgreSQL 17 real | Aprobado; no omitidos |
| Build de imagen | CI | Aprobado |
| Arranque Compose, migración y API | CI | Aprobado |
| Prueba HTTP y reinicio de API/PostgreSQL | CI | Mismos datos, IDs y job después del reinicio |
| Tests Python del sondeo de readiness | Desarrollo y CI | Dos tests aprobados |

Los tests PostgreSQL cubren migración repetida sin resetear identidad, snapshots, idempotencia de instrucciones y reuniones, 16 creaciones concurrentes con una misma clave, conflicto de payload, rollback sin registros parciales, workers concurrentes, recuperación con un pool nuevo, consultas inexistentes y rechazo de checksums alterados.

El smoke test recorre la API HTTP real: rechaza accesos sin token, guarda una instrucción, crea una reunión y su job, verifica replay y conflicto, espera el bloqueo por adaptador ausente y confirma una lista de mensajes vacía. Después reinicia los dos servicios y consulta nuevamente los mismos registros.

La primera ejecución detectó que el script no esperaba ante `ConnectionResetError` durante el reinicio. Se corrigió el sondeo para tolerar fallos transitorios con límite de 45 segundos y se agregaron pruebas de regresión. La ejecución enlazada arriba corresponde a la corrección.

## Límites de la validación

El entorno de desarrollo no tenía Docker ni PostgreSQL; los tests de base allí se informaron como omitidos. Su ejecución real y el reinicio de contenedores se comprobaron en CI, no en la computadora del manager.

No se validaron aún proveedores externos, rondas de agentes, scheduler diario, generación de audio, publicación, feedback, backups/restauración, carga productiva ni despliegue permanente. Tampoco se importó la conversación histórica. Una reunión bloqueada acredita persistencia y diagnóstico, no creación artística.
