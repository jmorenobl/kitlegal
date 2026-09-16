# Pendientes de estructura

Cosas de la estructura del repositorio que hoy no son un defecto pero hay que decidir o vigilar en un
hito concreto. Cada entrada nombra el hito en el que se resuelve y se borra cuando se resuelve. Lo que
es una decisión cerrada no está aquí: está en `docs/ADR/`.

Origen: revisión de la estructura frente a las convenciones de Go y del estándar Agent Skills
(2026-09-12), tras cerrar H1.

## Cuando existan `docs/ARCHITECTURE.md` y `docs/SOURCES.md` · Absorber `refs/`

`refs/` son los documentos semilla del proyecto. Cuando la arquitectura y la tabla de fuentes tengan su
documento propio bajo `docs/`, lo que quede vigente de `refs/` debe pasar ahí y el directorio
desaparecer.

## Cuando molesten · Artefactos del workflow en `specs/*/gates/`

Los veredictos, rondas e intentos del workflow `hito` se versionan a propósito (workflow 1.5.0 y
posteriores) y crecen hito a hito. Si ensucian los diffs de las propuestas de cambio, un
`.gitattributes` con `linguist-generated` sobre `specs/*/gates/` los pliega sin dejar de versionarlos.
