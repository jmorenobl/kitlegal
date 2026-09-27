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

## Cuando salga GoReleaser v2.19 · El `postflight` del cask de Homebrew

Homebrew 7.0 (2026-09-13) deprecó los bloques `preflight`/`postflight` de los casks en favor de
`postflight_steps`, declarativo, y cada `brew update` o `brew upgrade` avisa con «Calling `postflight` is
deprecated!». El cask de `jmorenobl/homebrew-tap` lo genera GoReleaser desde `homebrew_casks.hooks.post.install`
de `.goreleaser.yaml`, y la v2.18.1 fijada en `tools/goreleaser/` solo sabe escribir la forma antigua
([goreleaser#6870](https://github.com/goreleaser/goreleaser/issues/6870), previsto para la v2.19.0). Ese bloque
retira la cuarentena del binario, que no está notarizado: sin él, el binario no arranca en un Mac
(research.md D25 de H19). Cuando Dependabot proponga la v2.19, pasar el hook a la forma que genere
`postflight_steps` y comprobar con `brew install` en un Mac que el binario arranca y el aviso desaparece. La
alternativa, notarizar el binario, lo haría innecesario, y es una decisión aparte.
