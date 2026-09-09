# Roadmap de BandIA

Contrato detallado: [plan de API e implementación](api-implementation-plan.md). Objetivos y acuerdos: [spec](spec.md).

La base actual sólo ofrece salud HTTP. Las fases siguientes incorporan comportamiento real; una respuesta simulada no satisface sus criterios de salida.

| Fase | Entrega | Resultado verificable |
| --- | --- | --- |
| P0 | Base ejecutable, CI, entorno local, autenticación, PostgreSQL y migraciones | Checkout compilable y pruebas ejecutadas; acceso y readiness correctos. |
| P1 | Reunión durable y agentes reales | Crear reunión, leer respuestas originales y decisiones; recuperar tras reinicio sin duplicación. |
| P2 | Scheduler, tareas escritas, informes, eventos y control | Rutina diaria sin disparo humano, retrasos detectados y resultados persistidos. |
| P3 | Versiones musicales, generación, evaluación y selección | Demo real reproducible, con gasto trazable y calidad evaluada. |
| P4 | Cuenta X, borradores, imagen/video y publicación | Lanzamiento confirmado, snapshot e ID remoto; reintentos sin duplicados. |
| P5 | Feedback, síntesis, experimentos y aprendizaje | Comentario real → decisión → cambio → evaluación con fuentes. |
| P6 | Piloto de 30 días y expansión | Medir autonomía, costo, calidad y fallos; sumar cuentas/respuestas sólo cuando estén habilitadas. |

## Primer PR funcional

Preparar P0 y el núcleo de P1: migraciones, configuración, autenticación, fichas, instrucciones, creación de reuniones, consulta de mensajes/jobs y worker durable. Si falta proveedor debe guardar e informar el bloqueo. Publicar OpenAPI de las rutas entregadas y probar idempotencia y reinicios contra almacenamiento real.

El PR siguiente conecta el adaptador de texto, rondas de agentes y cierre con decisiones/tareas. Sólo entonces se cumple P1; una cola que guarda jobs sin ejecutar agentes no acredita autonomía creativa.

## Reglas de avance

La banda inicia tareas por decisiones y agendas; el manager puede observar e intervenir sin ser requisito de cada paso. El cierre de una fase exige resultados de su prueba real, no sólo mocks o cantidad de endpoints. Los adaptadores fake se usan en tests aislados.

PostgreSQL, cola inicial en base y almacenamiento compatible con S3 son propuestas técnicas del plan; todavía no se provisionaron. Proveedores, presupuesto numérico, umbrales de calidad y canal de alertas siguen pendientes. La prueba original de Go tampoco se declaró exitosa: P0 debe compilar, formatear y ejecutar los tests antes de continuar.
