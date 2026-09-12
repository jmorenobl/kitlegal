# Pendientes de estructura

Cosas de la estructura del repositorio que hoy no son un defecto pero hay que decidir o vigilar en un
hito concreto. Cada entrada nombra el hito en el que se resuelve y se borra cuando se resuelve. Lo que
es una decisión cerrada no está aquí: está en `docs/ADR/`.

Origen: revisión de la estructura frente a las convenciones de Go y del estándar Agent Skills
(2026-09-12), tras cerrar H1.

## Antes de la primera tarea `[datos]` de H4 · Dónde viven los fixtures grabados

`CLAUDE.md` y `docs/ROADMAP.md` §1.2 fijan `testdata/<fuente>/` en la raíz del repositorio. La convención
de Go es `internal/source/<fuente>/testdata/`, al lado de los tests que los usan y sin rutas con
`../../..`. Hay que elegir antes de grabar el primer fixture, porque moverlos después toca la
grabación (`KITLEGAL_RECORD=1`), el replay y el guardián de `[datos]` del workflow, que hoy distingue
«`testdata/` de raíz» de «`internal/<pkg>/testdata/`» (`docs/WORKFLOW.md`). Recomendación: junto al
paquete.

H2 cerró sin grabar nada contra una fuente real (FR-044) y sin crear `testdata/` en la raíz: lo único que
dejó es `internal/httpx/testdata/reproduccion/`, material de reproducción escrito a mano, que no
compromete la decisión. Por eso el plazo pasa de H2 a la primera tarea `[datos]` de H4, la que grabe el
primer fixture de una fuente de verdad.

## En H4 · Lo que el primer adaptador de fuente retira y amplía

Anotado al cerrar H2 (plan, obligación 8) y ampliado en H3 (plan, obligación 10), para que no se pierda
entre hitos:

- **`TestElBinarioNoEnlazaHTTPX` se retira.** Hoy (`internal/arch_test.go`) comprueba que el binario
  distribuido no enlaza `internal/httpx`, cosa cierta solo mientras ningún applet lo use. El primer
  adaptador lo enlazará a propósito, y ese hito lo sustituye por lo que sí seguirá siendo cierto: la
  ampliación de `modulosDelBinario` con `golang.org/x/time` y `github.com/temoto/robotstxt`, que entran
  con él, **con la justificación por escrito** que exige la constitución §V. Esa lista es la que queda
  vigilando la superficie del binario (research.md D18).
- **`TestElBinarioNoEnlazaCache` se retira.** Hoy (`internal/arch_test.go`) comprueba que el binario
  distribuido no enlaza `internal/cache` ni ningún paquete bajo `modernc.org/`, cosa cierta solo mientras
  ningún applet use la caché. Cuando el primer adaptador la enlace, `modulosDelBinario` se amplía con **los
  módulos que muestre entonces `go list -deps -f '{{if .Module}}{{.Module.Path}}{{end}}' ./cmd/kitlegal`**
  —la misma orden de `TestDependenciasDelBinario`—, **justificados uno a uno** por escrito (constitución
  §V). Esta nota no trae la lista a propósito: el `go.mod` del driver declara módulos
  (`modernc.org/fileutil`, `github.com/google/pprof`) que `go mod tidy` no incorpora al grafo (H3,
  research.md, sonda 6), y una lista copiada de ahí o escrita a mano fijaría dependencias que el binario
  nunca enlaza. La misma orden sobre `./internal/cache` orienta antes de enlazar, pero la que vale es la
  del binario.
- **La clave de caché y su vigencia salen de la fuente.** `internal/cache` guarda con la clave que recibe,
  sin interpretarla, derivarla ni normalizarla, y no tiene vigencia por omisión: `Put` la exige. El esquema
  de claves lo fija cada fuente y la vigencia sale de la propia fuente (`Source.TTL()`), no de un valor
  global (H3 FR-010, FR-011).
- **El ritmo por fuente sale de la tabla de fuentes.** `httpx.ConIntervalo` deja de darse por omisión y
  toma su valor de `docs/SOURCES.md`, que crea H4: un intervalo por fuente, no uno global. H2 no toca esa
  tabla.

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
