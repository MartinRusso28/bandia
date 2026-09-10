# BandIA

Banda virtual de agentes de IA, con autonomía creativa, memoria y aprendizaje del feedback de redes. Nombre artístico inicial: **Fuera de Hora**.

## Estado de esta entrega

Primera implementación local en **Go + PostgreSQL**. Guarda instrucciones, reuniones, snapshots del elenco y trabajos en transacciones. Incluye autenticación de manager, idempotencia, migraciones versionadas y un worker que registra bloqueos reales.

**Todavía no hay agentes conectados, conversaciones generadas, scheduler, música ni publicaciones en X.** Una reunión pasa de `queued` a `blocked`, con motivo `agent_provider_not_implemented`; no se inventan mensajes. No hace llamadas pagas ni modifica la automatización anterior de ChatGPT.

`/readyz` comprueba base y versión del esquema, no disponibilidad artística. `/v1/system` declara por separado las capacidades implementadas. Esto completa la infraestructura inicial, no P1 ni la operación autónoma del spec.

## Arranque local

Requisitos: Docker con Compose v2 y Python 3. No necesitás instalar Go si ejecutás todo en contenedores. Los puertos locales 8080 y 5432 deben estar libres.

```sh
git clone https://github.com/MartinRusso28/bandia.git
cd bandia
python3 scripts/init-local.py
docker compose up --build -d
docker compose ps -a
curl --fail http://127.0.0.1:8080/readyz
```

El inicializador genera credenciales aleatorias sólo para desarrollo, con permisos de archivo `0600`, y nunca reemplaza una configuración existente. No copies `.env.example` sobre la configuración generada. Compose espera a PostgreSQL, aplica las migraciones y después arranca la API con el worker integrado. `migrate` debe terminar con código 0; esperar unos segundos si la API todavía está iniciando.

Para ver fallos de arranque: `docker compose logs migrate api`. No compartas la salida de `docker compose config` ni dumps de variables: contienen credenciales. Este Compose es local, con servicios vinculados a `127.0.0.1`, y no debe exponerse a Internet.

## Primera prueba

```sh
python3 scripts/smoke.py
```

Crea una instrucción y una reunión; verifica autenticación, replay idempotente, conflicto con otro payload, snapshots y bloqueo sin mensajes ficticios. Imprime los IDs para inspección. Es una prueba sobre datos nuevos, no una reunión de agentes.

Para comprobar también reinicios, con el stack de este repositorio ya levantado:

```sh
python3 scripts/smoke.py --restart
```

Ese flag reinicia **la API y PostgreSQL de este Compose**, provocando una breve interrupción local, y vuelve a leer la misma reunión y job. No elimina datos.

También podés llamar la API desde bash/zsh:

```sh
set -a
. ./.env
set +a

curl --fail http://127.0.0.1:8080/v1/characters \
  -H "Authorization: Bearer $BANDIA_MANAGER_TOKEN"

curl --fail http://127.0.0.1:8080/v1/meetings \
  -H "Authorization: Bearer $BANDIA_MANAGER_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: primera-reunion-001' \
  -d '{"topic":"Definir el rumbo inicial de la banda","instruction_ids":[]}'
```

La aceptación devuelve `meeting_id`, `job_id` y `status: queued`. Consultá los IDs devueltos en las rutas de detalle. Repetir la misma clave y contenido devuelve la aceptación original, aunque el estado actual ya sea `blocked`; no crea ni reintenta otro trabajo. Otra entrada con esa clave devuelve 409. Las claves no expiran en este hito.

## Endpoints entregados

| Método y ruta | Resultado |
| --- | --- |
| `GET/HEAD /healthz` | Salud del proceso, sin autenticación. |
| `GET/HEAD /readyz` | 200 con esquema compatible; 503 si la base/esquema no está disponible. |
| `GET /v1/system` | Capacidades disponibles y pendientes. |
| `GET /v1/band` | Identidad persistida. |
| `GET /v1/characters` | Fichas persistidas del elenco inicial. |
| `POST /v1/manager-instructions` | Guarda `{text}`; devuelve 201. |
| `POST /v1/meetings` | Guarda reunión y job atómicamente; devuelve 202. |
| `GET /v1/meetings/{id}` | Estado, participantes e instrucciones congelados al crear. |
| `GET /v1/meetings/{id}/messages` | Mensajes originales; por ahora `data: []`. |
| `GET /v1/jobs/{id}` | Estado, intentos y motivo de bloqueo. |

Todas las rutas `/v1` requieren `Authorization: Bearer …`. Ambos POST requieren `Idempotency-Key`. El contrato de **lo implementado** está en [OpenAPI](docs/openapi.yaml); el catálogo completo futuro sigue en [plan de API](docs/api-implementation-plan.md). Listados paginados, edición del elenco, retry y control del worker no están entregados aún.

## Persistencia y recuperación

PostgreSQL utiliza el volumen `bandia_pgdata`. `docker compose down` detiene los servicios pero conserva ese volumen; al volver a levantar, las reuniones y claves siguen allí. **No uses `down -v` ni borres el volumen si querés conservar la memoria.** Esto no sustituye un backup: todavía falta automatizar y probar backup/restauración para producción.

Las migraciones se ejecutan explícitamente, en transacción y con lock, y se comprueba su checksum. Un cambio a una migración aplicada falla: agregá una nueva migración. El seed inicial no se reaplica sobre cambios de identidad al reiniciar. El esquema sólo contempla los estados entregados; las próximas fases ampliarán la máquina de estados mediante migraciones.

La reserva de idempotencia, reunión y job se confirman juntos. El worker procesa la transición a bloqueo dentro de una transacción con `FOR UPDATE SKIP LOCKED`; si el proceso cae antes de commit, el trabajo sigue pendiente. Los bloqueados no se reintentan automáticamente. Antes de conectar IA hay que implementar leases/checkpoints y el adaptador; no se deben mantener locks mientras se espera un proveedor.

Las fichas de esta migración son un bootstrap del elenco. Todavía no se importaron la conversación histórica de 11 mensajes ni los perfiles ampliados de los archivos anteriores; ese traspaso necesita conservar su procedencia y no tratar ejemplos simulados como memoria real.

GitHub guarda código y diseño, **no la base de datos**. No subir claves, audio, transcripciones privadas o comentarios originales de terceros al repositorio público. El almacenamiento de objetos se incorpora con la producción musical, no en este hito.

## Desarrollo nativo y tests

Go 1.26+ compatible con `go.mod`, PostgreSQL 17 y compilador C para `-race`:

```sh
docker compose up -d postgres
set -a
. ./.env
set +a
go run ./cmd/bandia migrate
go run ./cmd/bandia
```

Si la API de Compose está corriendo, detenela primero con `docker compose stop api` para liberar 8080. El binario nativo no carga `.env` por sí solo. Para aplicar futuras migraciones con Compose: `docker compose run --rm migrate`.

```sh
make fmt
make check
make build
```

Los tests HTTP usan fakes **sólo en tests**. Los de almacenamiento requieren PostgreSQL real y una variable explícita; sin ella se informan como omitidos. Para crear una base descartable, una sola vez:

```sh
docker compose exec postgres createdb -U bandia bandia_test
export BANDIA_TEST_DATABASE_URL="postgres://bandia:$POSTGRES_PASSWORD@127.0.0.1:5432/bandia_test?sslmode=disable"
make integration
```

Cada test crea y borra únicamente su esquema aleatorio `bandia_test_*`; no trunca tablas de la banda. No apuntar tests a producción. La CI incluye PostgreSQL real, detector de carreras, build de imagen y prueba de Compose con reinicio. Ver su resultado en [Actions](https://github.com/MartinRusso28/bandia/actions); que exista el workflow no significa que haya pasado.

## Siguiente incremento

Conectar llamadas independientes de texto y sus checkpoints, conservar prompts/respuestas originales y cerrar decisiones/tareas. Configurar proveedor y límites de gasto antes de cualquier llamada paga. Después: scheduler propio, música, publicación y feedback. [Roadmap](docs/roadmap.md) · [Spec](docs/spec.md) · [Operación autónoma](docs/autonomous-operations.md).
