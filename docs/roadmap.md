# Plan de implementación

## 0. Base HTTP y diseño

Incluido: servidor Go, salud, preparación, cierre ordenado, pruebas y spec. Pendiente: ejecutar compilación y pruebas en un entorno con Go. No se declara validación exitosa sin resultado real.

## 1. Estado durable de reuniones

Elegir almacenamiento y documentar el contrato. Agregar autenticación de administración antes de exponer endpoints operativos de escritura. Crear reuniones con ID y clave idempotente; estados queued/running/incomplete/completed/blocked y errores explícitos. Guardar cada mensaje y el contexto que recibió.

Criterio: repetir una solicitud no duplica la reunión; reiniciar recupera el avance confirmado; las escrituras concurrentes no pierden mensajes. No usar memoria de proceso como persistencia final.

## 2. Orquestación real y scheduler

Adaptador de agentes con credenciales configuradas de forma segura, límites de uso y tiempo. Aperturas, réplicas y cierre con mensajes originales y fichas versionadas. Scheduler diario con zona horaria, detección de retrasos y recuperación de trabajos incompletos.

Criterio: los personajes responden a mensajes efectivamente recibidos; los errores quedan visibles y no se reemplazan por diálogos inventados. Probar recuperación y siete días de operación antes de declarar automatización confiable.

## 3. Audio privado

Generar, almacenar y evaluar demos reales; registrar costos, parámetros y selección. Validar qué controles musicales y continuidad vocal ofrece el proveedor antes de prometerlos.

Criterio: audio reproducible vinculado a sus decisiones, dentro del presupuesto y con calidad mínima verificable.

## 4. X y aprendizaje

Autorizar cuenta, cargar medio, confirmar procesamiento y publicar con reconciliación de resultados ambiguos. Capturar feedback con fuentes y cobertura; discutirlo y experimentar con cambios.

Criterio: un lanzamiento confirmado y un ciclo comentario real → decisión → cambio → evaluación. Las respuestas públicas automáticas y cuentas individuales quedan para después.

## Parámetros pendientes

Presupuesto numérico, proveedores, almacenamiento, tolerancia de retraso, umbrales de calidad y canal de alertas. El spec distingue acuerdos de propuestas y sigue siendo la referencia funcional.
