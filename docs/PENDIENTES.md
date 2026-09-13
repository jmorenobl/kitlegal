# Pendientes de estructura

Cosas de la estructura del repositorio que hoy no son un defecto pero hay que decidir o vigilar en un
hito concreto. Cada entrada nombra el hito en el que se resuelve y se borra cuando se resuelve. Lo que
es una decisión cerrada no está aquí: está en `docs/ADR/`.

Origen: revisión de la estructura frente a las convenciones de Go y del estándar Agent Skills
(2026-09-12), tras cerrar H1.

## En H5 · Tres directorios llamados `skills`

Cuando exista `skills/` (el producto) coexistirán tres árboles con tres significados: `skills/` (skills
que se distribuyen), `.agents/skills/` (skills de agente vendorizadas para trabajar en este repo,
registro en `skills-lock.json`) y `.claude/skills/` (symlinks a las anteriores más las de spec-kit).
Documentar la diferencia en el `README.md` cuando aparezca el primero.

## En H5 · Peso de las skills vendorizadas

`.agents/skills/` es casi la mitad de los ficheros versionados del repositorio y no es producto. Con
`skills-lock.json` como manifiesto, valorar instalarlas en local en vez de versionarlas. Si se quedan,
un `.gitattributes` con `linguist-vendored` las saca de las estadísticas y de los diffs de las
propuestas de cambio.

## En H5 · Cómo llama cada skill al binario

El diseño prevé `skills/<skill>/scripts/<applet>` como symlink a `bin/kitlegal`. Un symlink a un binario
que no está versionado no sobrevive a un zip, a Windows ni a la instalación desde un marketplace. Un
wrapper de una línea en shell que localice `kitlegal` en el `PATH` es más robusto; decidirlo con la
primera skill.

## Cuando existan `docs/ARCHITECTURE.md` y `docs/SOURCES.md` · Absorber `refs/`

`refs/` son los documentos semilla del proyecto. Cuando la arquitectura y la tabla de fuentes tengan su
documento propio bajo `docs/`, lo que quede vigente de `refs/` debe pasar ahí y el directorio
desaparecer.

## Cuando molesten · Artefactos del workflow en `specs/*/gates/`

Los veredictos, rondas e intentos del workflow `hito` se versionan a propósito (workflow 1.5.0 y
posteriores) y crecen hito a hito. Si ensucian los diffs de las propuestas de cambio, un
`.gitattributes` con `linguist-generated` sobre `specs/*/gates/` los pliega sin dejar de versionarlos.
