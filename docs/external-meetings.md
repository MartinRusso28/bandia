# POC: generación externa, memoria en BandIA

El motor temporal puede ser una sesión de ChatGPT/Codex con agentes separados. Go no llama a la API de OpenAI: recibe y persiste los turnos. **El puente HTTP está implementado; no hay todavía una conexión activa desde este chat a tu localhost ni una nueva reunión real ejecutada.** Las pruebas automáticas usan fixtures explícitos.

## Contrato

Todas las operaciones requieren el token de manager. Los POST requieren una clave idempotente, conservada en reintentos. El cuerpo máximo sigue siendo 64 KiB.

| Operación | Uso |
| --- | --- |
| `POST /v1/meetings` con `mode: external` | Crea reunión y job en `waiting_external`, sin intervención del worker interno. |
| `GET /v1/meetings/{id}/context` | Snapshot consistente de fichas, instrucciones, turnos originales, próximo personaje/secuencia y hash del contexto. |
| `POST /v1/meetings/{id}/turns` | Guarda un turno original y su procedencia declarada. Devuelve 201. |
| `POST /v1/meetings/{id}/close` | Guarda resumen/decisiones y completa reunión y job en una transacción. Devuelve 200. |
| `GET /v1/meetings/{id}/messages` | Transcripción original ordenada, incluidos prompts y metadatos. |

El plan de esta POC es fijo: dos rondas del elenco de cinco personas y una intervención final del productor (11 turnos). Queda congelado al crear la reunión. No se convierte una reunión interna bloqueada: se crea otra externa explícitamente.

Un turno requiere `sequence`, `history_through`, `character_id`, `context_hash`, `agent_run_id`, `prompt` y `content`. Debe coincidir con el siguiente turno y hash entregados por `/context`. Prompt y respuesta se conservan sin recortar espacios ni reescribirlos. Límites: prompt 32000 caracteres, respuesta 8000, ID de ejecución 200; todos no vacíos. El mismo ID de ejecución no se puede usar para dos turnos de una reunión.

La API serializa escrituras por reunión y rechaza con 409 un personaje fuera de orden, contexto viejo, ejecución reutilizada, turno después del final o cierre anticipado. La misma clave y entrada reproducen la respuesta original incluso después del cierre; cambiar la entrada devuelve `idempotency_conflict`. Otros conflictos devuelven `turn_conflict`: consultar el contexto antes de decidir qué hacer.

El cierre requiere `last_sequence: 11`, `summary` (hasta 8000 caracteres) y `decisions` (1–20 textos de hasta 2000 caracteres). Esas decisiones quedan registradas, **no se ejecutan como tareas ni autorizan gasto/publicaciones**. La última respuesta del productor permanece como evidencia separada del resumen suministrado por el orquestador.

## Protocolo del orquestador

1. Crear la reunión externa y guardar su ID y clave de creación.
2. Obtener `/context` y conservar el archivo original de ese contexto.
3. Ejecutar sólo el personaje indicado como próximo. Pasarle su ficha, tema, instrucciones y los contenidos originales de los mensajes anteriores; no insertar todos los prompts anteriores recursivamente.
4. Guardar el prompt exacto enviado, la respuesta exacta recibida y el identificador real de esa ejecución. No presentar texto simulado como respuesta de un agente.
5. Persistir ese turno y confirmar la respuesta HTTP antes de ejecutar el siguiente agente.
6. Al reanudar, consultar primero el contexto. Si una escritura tuvo resultado ambiguo, reintentar con la misma clave y entrada, sin volver a generar la respuesta.
7. Después de los 11 mensajes, guardar el cierre. No sustituir ni editar mensajes originales.

El hash prueba a qué contexto persistido se vinculó un envío; **no prueba que el modelo realmente lo haya leído**. `origin: external_agent_reported` y `agent_run_id` son procedencia declarada por el cliente autenticado, no verificación criptográfica de Codex. Para auditar la ejecución deben conservarse los registros originales del orquestador.

Si la cuota de agentes se agota, la reunión queda esperando sin perder turnos confirmados. No hay reintentos automáticos de generación, llamadas pagas ni garantía de continuidad después de cerrar el chat. Un resultado generado pero perdido antes de persistir no puede recuperarse desde la base: hay que conservar su archivo local y reutilizarlo, o documentar una nueva ejecución.

## Uso del puente local

Actualizar el código y la migración sin borrar el volumen:

```sh
git pull
docker compose stop api
docker compose build
docker compose run --rm migrate
docker compose up -d api
set -a
. ./.env
set +a
python3 scripts/external_client.py create --topic 'Definir el rumbo de la banda' --key rumbo-externo-001
```

Crear una carpeta privada de trabajo (por ejemplo `artifacts/`, ignorada por Git). Sustituir `ID` por el ID devuelto:

```sh
mkdir -p artifacts
python3 scripts/external_client.py context ID --out artifacts/contexto-001.json
python3 scripts/external_client.py turn ID --context artifacts/contexto-001.json --input artifacts/respuesta-001.json --key turno-001
```

El archivo de respuesta debe contener exactamente `prompt`, `content` y `agent_run_id`, copiados de la ejecución real. El cliente toma la secuencia, personaje y hash del contexto guardado. Después se obtiene un **nuevo** contexto para el siguiente turno. Los archivos de contexto no se sobrescriben.

Después del turno 11, descargar el contexto final y enviar un archivo con `summary` y `decisions`:

```sh
python3 scripts/external_client.py close ID --context artifacts/contexto-final.json --input artifacts/cierre.json --key cierre-001
```

`BANDIA_API_URL` permite seleccionar otro origen; por defecto usa localhost. El cliente exige HTTPS fuera de localhost y rechaza redirecciones para no reenviar credenciales. El token se lee del entorno, nunca de argumentos. No subir transcripciones o contextos privados al repositorio público.

## Para llamarlo desde ChatGPT

Hace falta publicar el servicio en un origen HTTPS accesible y configurar una herramienta autenticada que invoque estas rutas. El cliente incluido puede cumplir esa función desde el entorno de ejecución del asistente si tiene conectividad y credenciales configuradas; no exige desarrollar otro conector. **Una URL pública por sí sola no conecta esta conversación**, ni la conexión de GitHub da acceso al servicio. No abrir 8080/5432 a Internet, pegar tokens en el chat ni desactivar autenticación. El despliegue y la configuración de acceso quedan pendientes; esta entrega no contrata ni publica servicios.

Prueba de infraestructura: `python3 scripts/smoke_external.py` con Compose levantado. Crea exclusivamente fixtures de test y reinicia API/PostgreSQL a mitad de la reunión; no genera una conversación creativa real.
