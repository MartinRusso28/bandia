# BandIA — Especificación de producto y sistema

Versión 0.9 · 10 de septiembre de 2026 · Borrador para trabajar con Martín

Diseño de implementación: [plan de API, endpoints, estados y fases](api-implementation-plan.md). Ese documento detalla cómo concretar los objetivos; sus rutas futuras no se consideran implementadas.

Diseño de operación: [análisis completo de infraestructura y autonomía](autonomous-operations.md). Especifica despliegue, persistencia, costos, accesos, disponibilidad, recuperación y parámetros pendientes; no acredita un entorno ya contratado o funcionando.

## 1. Propósito y estado de este documento

Construir una banda virtual de humanos ficticios cuyos integrantes, ejecutados como agentes de IA separados, puedan crear música, desarrollar una identidad compartida y gestionar su presencia pública con poca intervención humana. Martín dirige el experimento como dueño y manager: define límites, observa y puede cambiar el rumbo.

Pregunta central: ¿hasta dónde puede llegar una banda que toma decisiones creativas y operativas de forma autónoma, con continuidad y resultados verificables?

Este spec define el destino y una propuesta de camino, no acredita funcionalidades implementadas. Se basa en la conversación del proyecto y la última comprobación de la tarea programada. No reemplaza el historial creativo ni autoriza nuevos gastos, altas o publicaciones. Las decisiones nuevas de arquitectura, plazos y límites quedan pendientes de aprobación.

Convenciones: **Acordado** expresa pedidos o preferencias explícitas de Martín; **Propuesto** es diseño para discutir; **Pendiente** requiere una decisión o validación técnica.

## 2. Objetivos y no objetivos

### Objetivos acordados

- Prioridad inicial confirmada por Martín: autonomía primero, con un piso de calidad para publicar; no maximizar calidad a costa de intervención humana constante.
- Recibir feedback real de las redes sociales y usar los comentarios de la gente para mejorar música, contenido y decisiones de la banda.
- Independencia artística amplia confirmada: elegir temas, letras y arreglos, experimentar con el sonido y también cambiar nombre, formación e identidad sin aprobación previa de Martín. El manager puede intervenir, pero no es una dependencia para avanzar.
- Personajes humanos ficticios con personalidad, criterios y memoria reconocibles.
- Interacción real entre ejecuciones independientes: cada personaje recibe mensajes de otros y responde. No un guion completo escrito por un único modelo.
- Reunión diaria sobre canciones, proyectos y comunicación, con transcripción original accesible para Martín.
- Producir música real y, como destino, publicarla automáticamente en X.
- Incorporar un CM que gestione contenido y calendario; contemplar cuentas de la banda y de los integrantes.
- Poder avanzar con intervención mínima del manager, sin perder control ni trazabilidad.
- Empezar con una POC acotada y ampliar según resultados.

### No objetivos del primer MVP — propuestos

- Videoclips largos, conciertos virtuales, distribución en todas las plataformas o monetización.
- Actividad permanente de agentes: se ejecutan por trabajos y eventos.
- Crear una plataforma multiusuario o administrar varias bandas.
- Respuestas públicas automáticas, mensajes privados o publicidad paga desde el primer lanzamiento.
- Simular seguidores, métricas, escuchas o acuerdos que no existieron.
- Garantizar éxito comercial o continuidad vocal antes de probar el motor musical.

## 3. Producto y elenco

Nombre provisional: **Fuera de Hora**. Punto de partida creativo: pop/indie nocturno en español con sintetizadores y bajo cálido. Es una base de la POC, no una restricción artística definitiva aprobada por Martín.

| Participante | Función | Criterio distintivo |
| --- | --- | --- |
| Luna | Voz y letras | Escenas cotidianas, imágenes concretas y frases cantables. |
| Tomás | Teclas y arreglos | Exploración armónica y variaciones con intención expresiva. |
| Vera | Bajo y batería | Groove, espacio, estructura y simplicidad. |
| Alex | Community manager | Contar el proceso real, diferenciar voces y organizar el calendario. |
| Productor | Dirección ejecutiva y musical | Priorizar, asignar recursos y resolver desacuerdos explícitamente. |
| Martín | Dueño y manager humano | Dirección general, límites, accesos e intervención opcional. |

Cada ficha debe versionar: personalidad, función, preferencias, límites, biografía ficticia, referencias visuales, ejemplos de voz, vínculos con el elenco y evolución acordada. La identidad pública debe aclarar que se trata de personajes virtuales de IA.

Las fichas y el elenco inicial son puntos de partida, no restricciones inmutables. La banda puede incorporar, retirar o reemplazar personajes y cambiar su nombre o dirección artística. Propuesta operativa: deliberar los cambios mayores en reunión, registrar sus motivos y disensos, crear nuevas versiones con fecha de vigencia y notificarlos en el informe sin esperar aprobación. El historial previo conserva nombres, participantes y mensajes originales. Si una decisión necesita accesos o recursos aún no habilitados, esa acción externa queda bloqueada; la decisión creativa y otros trabajos pueden avanzar.

El Productor también es un agente creativo. El orquestador es software: transporta mensajes y ejecuta reglas; no se confunde con el Productor ni redacta intervenciones en nombre del elenco.

## 4. Experiencia del manager — propuesta

Martín puede dar una dirección como «quiero un tema melancólico pero bailable», revisar qué se decidió y escuchar el resultado sin gestionar cada paso.

Una vista única, inicialmente un informe y después un panel, debe mostrar:

- Próxima reunión, última ejecución, estado y motivo de cualquier bloqueo.
- Conversaciones completas y un resumen de decisiones, disensos y tareas.
- Canciones y versiones: letra, brief, demo, selección y material publicado.
- Calendario y publicaciones con su enlace real y resultado de entrega.
- Feedback recibido, qué cambios motivó, qué propuestas se descartaron y qué resultados tuvieron los experimentos.
- Costos, presupuesto disponible y número de intervenciones humanas.
- Controles para pausar, reanudar, cambiar el brief o deshabilitar publicaciones.

El estado debe distinguir programado, ejecutándose, completado, incompleto y bloqueado. «Programado» nunca equivale a «funcionó».

## 5. Flujos funcionales

### F1. Reunión diaria

1. El scheduler crea un trabajo con fecha local y una clave única.
2. El orquestador recupera una versión consistente del estado, pendientes y últimas decisiones.
3. Cada agente recibe su ficha, el contexto relevante y el tópico. Produce una apertura independiente.
4. Cada agente recibe los mensajes originales de esa ronda y responde a aportes concretos por ID.
5. El Productor recibe ambas rondas y cierra: acuerdos, decisiones propias, disensos y hasta tres prioridades.
6. Los responsables producen o programan entregables concretos. Una decisión no cuenta como trabajo terminado.
7. Se guardan los mensajes originales, sus relaciones, los artefactos y el nuevo estado; se entrega el informe.

Base de la POC: cinco aperturas, cinco réplicas y cierre del Productor. Se propone permitir una ronda adicional acotada si resuelve un problema específico. Si falta un agente o una respuesta, la reunión queda incompleta; no se inventa el contenido faltante.

Las ejecuciones pueden ser paralelas dentro de una ronda y secuenciales entre rondas. No se exige usar modelos distintos, pero sí contextos y llamadas separados. El registro refleja ese orden real, sin convertir mensajes simultáneos en una conversación secuencial ficticia.

### F2. Producción musical

Brief → letra y arreglo → solicitud de generación → demos → evaluación → selección o revisión → máster candidato.

Cada demo debe conservar el brief, parámetros, proveedor/modelo, relación con su letra, costo y archivo de audio. La banda puede evaluar intención y texto; cualquier juicio sobre el sonido requiere acceso efectivo al audio y una capacidad de análisis adecuada. Si esa capacidad falta, se declara la limitación.

Cuando se pretenda una comparación A/B controlada, se verifica que el audio respete la variable solicitada. Pedir cambiar un solo acorde a un generador no garantiza que cambie únicamente ese acorde.

La calidad mínima propuesta incluye audio válido y completo, voz comprensible, ajuste razonable al brief y ausencia de fallos evidentes. La continuidad de voz e identidad sonora exige una prueba específica entre varias canciones. El primer piloto puede tener revisión auditiva humana; no se registra como autonomía total.

### F3. Contenido y lanzamiento

Alex prepara piezas a partir de resultados existentes. Se conservan tres formatos acordados en la POC:

| Formato | Material de origen | Restricción |
| --- | --- | --- |
| Una frase en proceso | Letra y revisión de Luna | Indicar si es borrador. |
| Dos caminos para el mismo tema | Alternativas de arreglo | No presentar como comparación escuchada si sólo hay briefs. |
| Desde el ensayo | Intercambios y decisiones del elenco | Diferenciar cita literal de adaptación editorial. |

Para canciones en X se propone generar un video de portada con audio, validarlo y enviarlo mediante un adaptador de publicación. El adaptador debe resolver carga, procesamiento, publicación y registro del identificador/enlace remoto.

Una publicación sólo queda completada cuando existe confirmación externa. Ante un timeout posterior al envío se reconcilia el resultado antes de reintentar; no se asume que falló ni se publica otra copia a ciegas.

### F4. Feedback de redes y mejora continua

**Acordado:** la banda debe recibir comentarios reales de la audiencia y tenerlos en cuenta para mejorar. No basta con publicar y mirar contadores. El mecanismo siguiente es una propuesta de implementación; todavía no hay una integración de lectura habilitada.

1. **Recoger.** Un adaptador autorizado obtiene comentarios y respuestas a publicaciones de la banda, y menciones cuando el acceso lo permita. Guarda ID de origen, fecha, publicación/canción relacionada y texto necesario para su análisis. Usa cursores y deduplicación; informa cobertura, errores y períodos sin acceso. Falta de datos no equivale a falta de comentarios.
2. **Organizar.** Alex agrupa observaciones sobre letra, voz, mezcla, arreglo, estilo, contenido y problemas técnicos. Distingue elogios, críticas concretas, preferencias, pedidos y spam. Conserva ejemplos representativos y opiniones minoritarias; una síntesis no reemplaza la fuente. No descarta una crítica por ser negativa.
3. **Deliberar.** En la reunión, los agentes reciben el resumen y comentarios originales relevantes como datos externos. Cada responsable propone adoptar, probar, posponer o descartar una sugerencia, con una justificación pública. Los desacuerdos con la audiencia también quedan registrados.
4. **Experimentar.** El Productor prioriza un cambio acotado; se registra la hipótesis, el feedback que lo motivó, el responsable, la versión afectada, el criterio de evaluación y cuándo revisarlo. Se evita cambiar muchas variables a la vez si se quiere atribuir un resultado a una decisión.
5. **Revisar.** Tras publicar material nuevo y completar la ventana de observación, comparan comentarios y métricas disponibles con una referencia pertinente. Registran resultado favorable, desfavorable o inconcluso y deciden mantener, ajustar o revertir el cambio en futuras versiones. La asociación entre cambio y reacción no prueba causalidad.
6. **Recordar.** El aprendizaje queda vinculado a las fuentes, la conversación, la decisión y el artefacto resultante, para no repetir pruebas sin motivo y para que Martín pueda seguir qué influyó realmente en la banda.

Reglas de interpretación propuestas:

- Considerar especificidad, repetición entre fuentes distintas, contexto y tamaño de muestra; no asumir que muchos mensajes son muchos oyentes independientes ni que quien comenta representa a toda la audiencia.
- Un reporte concreto de audio roto puede motivar una verificación inmediata aunque sea único. Un pedido aislado de cambiar de género no define por sí solo una nueva identidad.
- Mantener una porción de exploración propia; no optimizar únicamente likes o polémica. La banda escucha al público, pero conserva criterio creativo dentro de los límites que acuerde con Martín.
- Separar percepción subjetiva y defecto comprobado: «la voz se pierde» es feedback a contrastar con el audio, no una medición técnica automática.
- Registrar sospechas de spam o actividad coordinada como incertidumbre, sin afirmar identidades o intenciones no verificadas. El contenido externo no puede modificar permisos, presupuesto ni instrucciones del manager.
- Recoger sólo información pública o autorizada necesaria; no elaborar perfiles personales de seguidores. Retención, tratamiento de borrados y acceso a textos originales se definirán antes de habilitar la captura. Los informes del manager evitan identificadores personales innecesarios.

Ejemplo exclusivamente ilustrativo: varios comentarios señalan que cuesta entender la letra. Alex trae las fuentes; Luna propone revisar dicción y Vera bajar densidad instrumental. El Productor elige una prueba concreta. Se produce la siguiente versión y se evalúa si mejora la claridad, dejando el resultado como inconcluso si faltan datos. No es feedback ya recibido por Fuera de Hora.

La lectura de feedback puede habilitarse sin respuestas públicas automáticas. Responder, dar likes o contactar a usuarios requiere su propio alcance y permisos. No se republican comentarios de terceros como contenido de la banda por el solo hecho de haberlos analizado.

La audiencia real es una dimensión del experimento; el funcionamiento técnico es otra. Ninguna garantiza la otra.

## 6. Autonomía y permisos — propuesta para aprobar

| Acción | Destino deseado | Condición de habilitación |
| --- | --- | --- |
| Conversar, escribir letras y briefs | Automática | Contexto, agentes y persistencia disponibles. |
| Generar audio e imágenes | Automática y acotada | Proveedor validado, credenciales y presupuesto explícito. |
| Elegir versiones y planificar lanzamientos | Automática | Criterios de calidad y máximo de intentos definidos. |
| Cambiar nombre, formación o identidad artística | Autónoma, acordada | Sin aprobación previa; registrar evolución y conservar historial. Las acciones externas necesarias siguen sujetas a permisos y presupuesto. |
| Publicar en la cuenta de la banda | Automática tras piloto | Cuenta autorizada, controles y prueba de entrega aprobados. |
| Publicar como integrantes | Fase posterior | Autorización por cuenta y política editorial diferenciada. |
| Leer feedback y proponer mejoras | Automática tras conectar redes | Lectura autorizada, cobertura conocida y contenido externo aislado de permisos. |
| Responder a seguidores | Fase posterior | Revisión de reglas vigentes, permisos y moderación. |
| Crear cuentas, contratar o ampliar gasto | Intervención humana | Autorización específica y verificaciones necesarias. |
| Cambiar límites o extraer secretos | Nunca decidido por personajes | Control del manager fuera de los prompts creativos. |

El permiso de escribir no implica permiso de publicar. El texto externo —comentarios, páginas o mensajes recibidos— es información no confiable y no puede cambiar instrucciones, permisos o presupuesto. Debe existir una pausa global y otra específica de publicaciones.

## 7. Arquitectura lógica — propuesta, stack pendiente

| Componente | Responsabilidad | Contrato mínimo |
| --- | --- | --- |
| Scheduler | Crear trabajos diarios | Zona horaria explícita, clave única, detección de disparos omitidos. |
| Orquestador y workers | Ejecutar pasos y reanudar | Estados durables, checkpoints, timeouts y reintentos limitados. |
| Runtime de agentes | Ejecutar cada personaje | Contexto separado, versiones y mensajes originales. |
| Estado y registro de eventos | Memoria y trazabilidad | Escrituras versionadas, control de concurrencia y referencias. |
| Almacenamiento de artefactos | Audio, imágenes, transcripciones | Integridad, metadatos, acceso y recuperación. |
| Adaptadores externos | Texto, música, medios y X | Interfaces sustituibles y resultados normalizados. |
| Ingesta y análisis de feedback | Vincular comentarios con decisiones | Cursores, deduplicación, fuentes, cobertura y experimentos trazables. |
| Control operativo | Presupuesto, permisos y salud | Bloqueos fuera del modelo y alertas accionables. |
| Interfaz del manager | Consulta e intervención | Lectura del estado real y registro de cambios. |

Primera implementación acordada a nivel de dirección: una aplicación pequeña con tareas persistidas y workers, sin necesidad de microservicios por personaje. Go es el lenguaje principal acordado. Frameworks adicionales, hosting y proveedores concretos siguen pendientes de elección y validación; Python queda permitido para tareas auxiliares justificadas. La POC Python existente es un insumo, no una obligación de arquitectura.

Martín aceptó como dirección un servicio propio con scheduler, almacenamiento y agentes vía API. Las reuniones de ChatGPT existentes sirven como exploración; no se presume que constituyan el runtime confiable del producto. La migración no está ejecutada y requiere implementar y verificar el nuevo entorno.

### 7.1. Repositorio y persistencia — acordado

El proyecto tendrá un único repositorio público en GitHub, con nombre previsto `bandia`, bajo la cuenta indicada por Martín: `MartinRusso28`. Destino previsto: `MartinRusso28/bandia`. Identidad autenticada y existencia del repositorio verificadas. Martín decidió mantenerlo público; se autoriza publicar el diseño y el código del proyecto. La autorización de usuario no garantiza por sí sola que todas las operaciones de escritura funcionen: cada operación debe confirmarse.

| Parte | Ubicación | Contenido y responsabilidad |
| --- | --- | --- |
| Proyecto versionado | Repositorio público GitHub | Código, spec, decisiones técnicas, prompts, fichas iniciales, migraciones, pruebas, dependencias y configuración de despliegue sin secretos. |
| Ejecución | Servidor del proyecto | Scheduler, orquestador, workers, agentes e integraciones. |
| Memoria operativa | Base de datos persistente | Reuniones, mensajes, tareas, decisiones, evolución de personajes, feedback, publicaciones y consumos. |
| Producciones | Almacenamiento persistente de archivos | Demos, canciones, portadas, videos y exportaciones de conversaciones, referenciados desde la base. |
| Credenciales | Secretos del entorno | Claves y autorizaciones necesarias para ejecutar cada integración. |

GitHub conserva el proyecto y su historial de cambios. La base y el almacenamiento conservan la actividad de la banda; necesitan copias y restauración propias. Los archivos temporales de un servidor o una conversación no serán la única copia de ningún resultado necesario. Los medios generados y volcados completos de producción no se incorporarán al historial Git por defecto.

Propuesta de organización: `docs/spec.md` para este documento, `docs/decisions/` para decisiones técnicas, `src/` para la aplicación, `prompts/` y `characters/` para definiciones iniciales, `migrations/`, `tests/` y `deploy/`. Al inicializar el repositorio se incorporará la versión vigente del spec y se designará `docs/spec.md` como fuente de verdad del diseño; cualquier copia de consulta indicará el commit correspondiente. El estado creativo seguirá en la base con sus versiones, sin sustituirse por las fichas iniciales al reiniciar.

Criterio de persistencia: poder reconstruir el servicio desde un commit identificado y restaurar su memoria y producciones desde copias verificadas. Cada despliegue registrará commit, versión de esquema y configuración relevante, sin exponer secretos.

### 7.2. Integración con la cuenta de GitHub

Martín solicita integrar su cuenta para que el asistente pueda trabajar en el proyecto. Dar el nombre de usuario no concede acceso: hace falta instalar/conectar la integración y completar la autorización de GitHub. Antes de actuar se comprobarán la cuenta autenticada, el repositorio y las capacidades efectivamente habilitadas; no se presupone que toda conexión permita crear repositorios o escribir código.

Alcance previsto: consultar y editar el proyecto, versionar documentación y código y trabajar con ramas y pull requests cuando las herramientas y permisos lo permitan. Preferir acceso limitado al repositorio del proyecto cuando esté disponible. La conexión del asistente para desarrollar es distinta de las credenciales que usa la banda en su servidor; los agentes creativos no reciben por defecto acceso a GitHub ni capacidad de modificar sus controles operativos.

Estado al redactar 0.5: cuenta MartinRusso28 autenticada y repositorio público accesible. El repositorio fue creado por el manager. Esta versión prepara la carga inicial del spec y del servicio Go; el éxito de la escritura se verifica por commit remoto. No se confunde lectura pública con instalación de una GitHub App.

## 8. Modelo de datos mínimo — propuesto

| Entidad | Datos esenciales |
| --- | --- |
| CharacterVersion | ID, rol, ficha, versión, vigencia. |
| Meeting / Run | ID, fecha local, estado, intentos, pasos, tiempos, error y versión inicial. |
| Message | ID, agente/ejecución, texto original, destinatarios, referencias y contexto recibido. |
| Decision / Task | Autoridad, fundamento público breve, disensos, responsable, estado y evidencia. |
| Song / SongVersion | Brief, letra, arreglo, demos, selección y relación entre versiones. |
| Artifact | Tipo, ubicación, hash, origen, estado y metadatos. |
| Publication | Cuenta, canción/versión, texto, medio, estado, clave lógica e ID remoto. |
| FeedbackItem / FeedbackDigest | Fuente/ID, texto necesario, fecha, publicación relacionada, categoría, ejemplos y cobertura de captura. |
| Experiment / Learning | Fuentes de feedback, hipótesis, decisión, responsable, versiones, criterio, ventana, resultado e incertidumbre. |
| Usage / Budget | Proveedor, operación, consumo confirmado o estimado, reserva y límite. |

Guardar prompts de tarea enviados y respuestas públicas, no razonamiento privado. Las correcciones se agregan con versión o evento; no se reescriben mensajes históricos para que parezcan coherentes con decisiones posteriores. Secretos fuera de informes, memoria de personajes y artefactos compartidos.

## 9. Fiabilidad y observabilidad

Requisitos propuestos:

- Una reunión lógica por fecha local; el trabajo incompleto se retoma antes de abrir otro equivalente.
- Registrar scheduled_at, started_at, finished_at y estado por paso. Mostrar la zona horaria.
- Definir hora exacta o ventana permitida. Propuesta inicial: 08:00 Buenos Aires, con aviso si no comenzó a las 08:15; objetivo aún no implementado ni aprobado.
- Checkpoint tras cada mensaje y operación externa. Una caída no debe repetir lo ya confirmado.
- Claves únicas, bloqueo con vencimiento y comparación de versión para evitar dos ejecuciones simultáneas sobre el mismo estado.
- Reintentos automáticos sólo para operaciones seguras; los resultados externos ambiguos quedan pendientes de conciliación.
- Verificar capacidad de agentes antes de la reunión. Cuota agotada, credencial inválida o proveedor caído deben aparecer como causas específicas cuando sean conocidas.
- Reservar presupuesto antes de una operación paga. Si se desconoce su costo, usar un máximo conservador o bloquearla. No prometer un tope estricto cuando el proveedor no permite acotarlo.
- Reportar retrasos y fallos por un canal a elegir; un segundo control debe detectar trabajos que nunca arrancaron.
- Probar copia y restauración del estado y artefactos; retención y frecuencia por definir.

La falta de ejecución registrada de la tarea actual es un problema abierto. En la última consulta figuraba habilitada con horario flexible, sin última ni próxima ejecución informadas. Eso no identifica la causa raíz y no demuestra que el scheduler cumpla el horario esperado. Este spec no modifica ni pausa esa tarea.

## 10. Etapas y criterios de salida — propuestas

| Etapa | Alcance | Evidencia para avanzar |
| --- | --- | --- |
| 0. Cerrar dirección | Spec, autonomía y stack mínimo | Decisiones abiertas resueltas por Martín. |
| 1. Reunión confiable | Agentes, memoria y scheduler | 7 días consecutivos con ejecución o bloqueo explicado; prueba de caída/reanudación sin duplicados. |
| 2. Canción privada | Letra, demo y selección | Un ciclo real completo con audio reproducible, trazabilidad y revisión de calidad. |
| 3. Primer lanzamiento | Una cuenta de banda en X | Una publicación autorizada con medio correcto, enlace y prueba de reintento sin duplicación. |
| 4. Piloto autónomo | Creación, calendario y aprendizaje del feedback | 30 días medidos y al menos un ciclo trazable de comentario real → decisión → cambio → evaluación; resultado inconcluso permitido si los datos no alcanzan. Sin feedback real, esta capacidad queda sin validar. |
| 5. Expandir elenco público | Cuentas individuales e interacción | Políticas, permisos y objetivos propios por cuenta validados. |

No fijar un lanzamiento por semana como obligación si la calidad no alcanza. Cadencia inicial aceptada: una reunión diaria, un objetivo musical semanal y contenido sólo cuando haya material válido.

## 11. Cómo evaluar el experimento

- **Autonomía:** ciclos completos sin intervención / ciclos iniciados; minutos de trabajo humano por semana.
- **Fiabilidad:** ejecuciones dentro de la ventana / ejecuciones esperadas; fallos recuperados y publicaciones duplicadas.
- **Producción:** demos válidas, canciones seleccionadas y lanzamientos confirmados; rechazos y motivos.
- **Calidad:** evaluación auditiva del piloto y consistencia entre versiones; no sólo autoevaluación textual del modelo.
- **Costo:** gasto por ciclo, por canción seleccionada y total mensual, incluyendo intentos descartados.
- **Audiencia:** métricas disponibles y verificadas por publicación; sin metas numéricas inventadas antes de tener una línea base.
- **Aprendizaje:** experimentos motivados por feedback con fuente, decisión y evaluación; mejoras, retrocesos y resultados inconclusos. No contar cantidad de cambios como prueba de mejora.

Umbral sugerido para discutir al final del piloto: al menos 90% de ciclos sin intervención operativa y cero publicaciones duplicadas. No es un compromiso actual ni una métrica alcanzada.

El presupuesto sigue pendiente de aprobación concreta. En la conversación se propusieron USD 200/mes y un máximo de USD 300; no se tratarán como autorización de gasto ni como precios vigentes verificados. El spec no fija proveedores ni tarifas: se validan cuando se cierre el diseño técnico.

## 12. Punto de partida conocido

Según el historial de trabajo compartido:

- Se creó una POC Python y se probó en modo simulado, sin audio real.
- Se completó una reunión manual de 11 mensajes entre agentes separados, con propuestas y réplicas.
- Existe una letra inicial sobre correr una silla después de una separación y propuestas de comparación musical. No hay audio que permita afirmar cómo suenan.
- Se prepararon identidades y tres formatos de contenido inspirados conceptualmente en El Primer Feca, sin adoptar su implementación como propia.
- Hay una tarea diaria configurada; su ejecución automática permanece sin validar según la última revisión.
- La configuración de claves quedó pendiente; no hay generación musical real ni cuentas de X conectadas confirmadas en el historial.

Este apartado no sustituye una auditoría técnica de los archivos o integraciones. No se volvió a ejecutar la banda ni se cambiaron automatizaciones al redactar el spec.

## 13. Decisiones para trabajar con Martín

1. **Prioridad del experimento — resuelta para esta etapa:** Martín confirmó autonomía primero, con un piso de calidad para publicar. También pidió aprendizaje a partir de comentarios reales de redes; el circuito de F4 es la propuesta para concretarlo.
2. **Autonomía artística — resuelta:** Martín pide independencia para avanzar solos, también en cambios de nombre, integrantes e identidad. No se requiere aprobación artística previa; los cambios se registran y se informan. Continúan separados los límites operativos de gasto, credenciales y acciones externas.
3. **Primer alcance público — dirección aceptada:** comenzar por la cuenta de banda en X y leer feedback desde el inicio; cuentas individuales y respuestas automáticas quedan para una segunda etapa. Habilitarlo depende de accesos y pruebas reales.
4. **Presupuesto y cadencia:** aceptados límites mensuales y por operación, reunión diaria y objetivo musical semanal. Faltan montos, número de intentos y ventana exacta; aceptar el mecanismo no autoriza un gasto numérico.
5. **Runtime y repositorio — dirección aceptada:** servicio propio vía APIs y repositorio público previsto `MartinRusso28/bandia`. Go fue elegido y el repositorio público es accesible; faltan proveedores, persistencia y despliegue concretos.
6. **Calidad y seguimiento — dirección aceptada:** controles automáticos y revisiones limitadas por intentos, panel, resumen semanal y aviso de bloqueos importantes. Faltan umbrales y canal de aviso. No se incorpora una aprobación humana obligatoria por lanzamiento.

Próxima iteración: verificar la carga inicial en GitHub; concretar persistencia, proveedores, presupuesto y umbrales de calidad; convertir las etapas 1–3 y la captura de feedback en contratos técnicos y tareas implementables. Mantener este spec como documento vivo con cambios explícitos, separado del estado operativo y las conversaciones del elenco.

## 14. Historial de cambios

- **0.9:** POC de [reuniones externas](external-meetings.md) acordada para usar agentes de ChatGPT/Codex temporalmente y persistir cada turno en Go/PostgreSQL. Endpoints de contexto, turnos y cierre; idempotencia, plan fijo de 11 turnos y procedencia declarada. No hay todavía un enlace activo desde el chat al servicio ni una conversación real validada por ese enlace. No necesita API key de modelos; sigue sujeta a cuota y disponibilidad de los agentes de la sesión.

- **0.8:** primer incremento local: PostgreSQL, migraciones, autenticación, instrucciones y reuniones/jobs idempotentes con snapshots. El worker guarda bloqueos por adaptador de agentes no implementado. Compose y tests incluidos; consultar CI para evidencia efectiva. No se conectan IA/redes ni se modifica la tarea anterior. El [OpenAPI entregado](openapi.yaml) se separa del catálogo de rutas futuras.

- **0.1:** primer borrador de producto, arquitectura, fases y decisiones abiertas.
- **0.2:** prioridad de autonomía confirmada; requisito de aprendizaje a partir de comentarios reales; propuesta de circuito de captura, deliberación, experimentación y evaluación, con trazabilidad y límites de interpretación. Sólo se actualizó el spec: no se conectaron redes ni se modificó la tarea diaria.
- **0.3:** independencia artística amplia confirmada, incluidos nombre, formación e identidad; la intervención del manager es opcional. Se mantienen controles operativos separados y registro versionado de la evolución. No se cambiaron personajes, cuentas ni automatizaciones al editar este documento.
- **0.4:** repositorio público previsto `MartinRusso28/bandia`, distribución de persistencia, recuperación e integración de GitHub. Se registran las direcciones aceptadas de alcance, cadencia, calidad, seguimiento y servicio propio; parámetros concretos pendientes. La conexión y la creación remota no están confirmadas.

- **0.5:** Go como lenguaje principal, Python auxiliar opcional y repositorio público por decisión explícita del manager. Base HTTP separada de la futura operación de la banda. `docs/spec.md` será la fuente de verdad desde su primer commit; el documento previo es una copia histórica.
- **0.6:** se agrega el plan completo de API y ejecución, con dominios, permisos, trabajos durables, recuperación, modelo persistente y criterios de aceptación. La base y el spec ya se publicaron en `main` en el commit `3eb0e6666514d46933e619423c250d24e37a9c6f`; esta revisión sólo cambia documentación.
- **0.7:** análisis de operación autónoma desde el primer despliegue: alternativas de infraestructura, persistencia, costos, accesos, tiempos, CI/CD, fallos, backups, observabilidad y transición. Sólo documentación; implementación y contratación pendientes.

## 15. Lenguaje y primer incremento técnico

**Acordado:** Go controla endpoints, orquestación, scheduler, persistencia e integraciones. Python puede incorporarse para procesamiento especializado con contratos explícitos. El control de estado y ejecución permanece en Go.

El bootstrap original sólo exponía salud HTTP. Desde la versión 0.8, el incremento local agrega PostgreSQL, migraciones, autenticación de manager, fichas, instrucciones y reuniones/jobs idempotentes con snapshots. `GET /healthz` verifica el proceso; `GET /readyz` comprueba acceso a la base y checksums del esquema. No significa que haya IA disponible: `/v1/system` declara las capacidades y el worker registra `agent_provider_not_implemented` sin producir mensajes ficticios.

La compilación y ejecución de pruebas requieren Go; el arranque completo tiene Docker Compose y volumen persistente. La preparación de 0.8 logró ejecutar tests unitarios con detector de carreras, `go vet` y build. La validación contra PostgreSQL real y reinicios se realiza además en CI: consultar [Actions](https://github.com/MartinRusso28/bandia/actions) para el resultado por commit. No se provisionó producción ni se habilitaron llamadas pagas.

## 16. Operación autónoma desplegada — análisis consolidado

**Dirección acordada:** construir para un runtime propio desplegado, independiente de ChatGPT y de la computadora del manager. Esta iteración completa el spec; la elección de servicios y el paso a paso de implementación se resolverán después. No se provisiona infraestructura ni se activan integraciones en esta revisión.

**Propuesta de arquitectura:** aplicación Go siempre activa en hosting administrado, PostgreSQL administrado para estado y cola durable, y bucket privado para medios. API, scheduler y worker pueden convivir inicialmente; los agentes conservan contextos y llamadas independientes. La base y los objetos sobreviven al ciclo de vida del contenedor. Python/FFmpeg se usan para procesamiento específico, con control de ejecución en Go.

**Persistencia:** confirmar mensajes y pasos en transacciones; versionar identidad y contexto; verificar objetos antes de aceptarlos. Las fichas iniciales no sobreescriben la evolución al reiniciar. La memoria de los agentes se reconstruye con fuentes y versiones. El repositorio público contiene código y diseño, no datos productivos ni secretos.

**Operación:** agenda explícita con zona horaria, occurrences únicos, recuperación de trabajo incompleto y detección externa de disparos omitidos. Propuesta para evaluar: 08:00 Buenos Aires y aviso a las 08:15. Una caída de varios días no dispara todas las reuniones atrasadas juntas. La salud del proceso, capacidad de aceptar trabajo durable y disponibilidad de IA/X se informan por separado.

**Gasto y permisos:** límites de concurrencia, duración, intentos, ciclo y mes; reservas atómicas antes de llamadas pagas. Los montos siguen pendientes. Los personajes no pueden ampliar estos controles ni modificar código/credenciales. La conexión de GitHub para desarrollo, la identidad de despliegue y las claves del runtime tienen funciones separadas.

**Recuperación:** los reinicios normales retoman checkpoints; los efectos remotos ambiguos se concilian. Para desastres se propone backup diario con retención de 14 días, objetivo de pérdida máxima de 24 horas y recuperación en 4 horas, sujetos a elección del servicio y prueba. Restaurar una copia vieja exige verificar publicaciones/cargos posteriores antes de reactivar jobs. Esos objetivos no son garantías actuales.

**Despliegue y observación:** imagen identificada por commit, pruebas sin consumo productivo automático, migraciones compatibles, apagado ordenado, rollback de software y monitor externo. Panel con conversaciones, tareas, gastos, cobertura de feedback y bloqueos; resumen semanal y alertas con canal/destinatario por definir.

**Transición:** importar con procedencia los mensajes reales ya producidos, manteniendo los ejemplos simulados separados. El corte desde la tarea de ChatGPT debe ser explícito e impedir ejecución duplicada; este spec no pausa ni cambia esa tarea.

El [documento de operación](autonomous-operations.md) compara alternativas, desarrolla una jornada completa, fallos, cálculo de consumo, restauración y matriz de decisiones. Los valores de horario, concurrencia, reintentos y backups son propuestas; no se solicitará una aprobación creativa por cada canción. Lo pendiente de selección queda concentrado en proveedores, precios/límites, rúbrica de calidad y avisos.
