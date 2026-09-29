# T001 · intento 1: `make lint` en rojo por `misspell` (`conversacion`); tarea redelimitada con `.golangci.yml`

## Qué quedó hecho (en el árbol, sin commitear)

- `schemas/expresiones-prohibidas.yaml.json`, tal cual contracts/lista-y-juicio.md §1.
- `internal/evals/formato_test.go`: `TestEsquemaDeExpresionesProhibidas` (con `patronDeExpresion`,
  `maquinariaBienFormada` y `otraConversacionBienFormada`, e importa `jsonschema/v6` y `jsonschema/v6/kind`). Lee el
  esquema de `../../schemas/expresiones-prohibidas.yaml.json`, lo compila con `skills.CompilarEsquema`, exige que la
  lista con las dos familias valide y se lea entera (premisa) y, en subtests paralelos, que diez listas con un solo
  defecto cada una no validen, con el incumplimiento exacto y su ruta: `sin-maquinaria` y `sin-otra-conversacion`
  (`kind.Required`, en la raíz), `maquinaria-vacia` y `otra-conversacion-vacia` (`kind.MinItems{0, 1}`, en la
  familia), `clave-de-mas` (`kind.AdditionalProperties{avisos}`), y `blanco-al-principio`, `blanco-al-final`,
  `dos-blancos-entre-palabras`, `asterisco` y `guion-bajo` (`kind.Pattern` con el patrón del contrato, en
  `<familia>/0`; repartidos entre las dos familias para que las dos `$ref` queden fijadas).
- Rojo → verde: sin el esquema, falla al leerlo (`no such file or directory`); con él, los once casos pasan
  (`go test ./internal/evals/ -run TestEsquemaDeExpresionesProhibidas`).
- Mutante momentáneo (restaurado y comprobado con `cmp`): `otra_conversacion` con `{"type": "array"}` en lugar de la
  `$ref` hace fallar `otra-conversacion-vacia`, `blanco-al-final` y `asterisco`.

## Por qué no quedó en verde

`make ci` sale con 2 en `lint`: `misspell` da 11 hallazgos, todos `conversacion` → «conversation», en las líneas del
test que escriben la clave `otra_conversacion` y los nombres de subtest `sin-otra-conversacion` y
`otra-conversacion-vacia`:

```text
internal/evals/formato_test.go:817:38: `conversacion` is a misspelling of `conversation` (misspell)
			otraConversacionBienFormada = "otra_conversacion:\n" +
internal/evals/formato_test.go:869:25: `conversacion` is a misspelling of `conversation` (misspell)
			nombre:    "sin-otra-conversacion",
…
11 issues:
* misspell: 11
make: *** [lint] Error 1
```

La clave la fija el contrato (lista-y-juicio §1, research D2) y T002 la escribe en la etiqueta
`yaml:"otra_conversacion"` del campo `OtraConversacion`, que no se puede partir en dos literales ni esquivar sin
deformar el código de producto. La solución de raíz es la de `observacion` en H7: la palabra en
`misspell.ignore-rules` de `.golangci.yml`, en la primera tarea que la escribe en Go, que es esta. `.golangci.yml` no
estaba entre las rutas de T001 (ni el plan lo preveía: `conversacion` está en el diccionario de misspell v0.8.0,
`words.go` línea 3192; del vocabulario del hito no salta ninguna otra de las comprobadas: `expresiones`,
`prohibidas`, `maquinaria`, `calibradas`, `familia(s)`, `encontradas`, `respuestas`, `repetida`, `derivada`,
`condiciones`).

## Redelimitación

La línea de T001 en `tasks.md` declara ahora `.golangci.yml` (comprobado con el extractor de `tarea.sh`: añade esa
ruta y ninguna otra). El intento siguiente solo tiene que añadir esta entrada, entre `controles` y `defectos`:

```yaml
        # «Conversación», el tramo de la clave otra_conversacion de la lista de
        # expresiones prohibidas de una skill, que fija el contrato lista-y-juicio
        # §1 de H7.2 (FR-050, research.md D2): en los documentos que la escriben
        # en los tests y en la etiqueta YAML del campo que la lee; misspell lo lee
        # como «conversation», también entre guiones bajos y guiones.
        - conversacion
```

Comprobado sobre una copia desechable (`rsync` del repositorio con estos cambios a `/tmp/kitlegal-t001-copia` y la
entrada añadida con Edit): `make -C /tmp/kitlegal-t001-copia ci` sale con 0 y termina en
«ci: todos los controles en verde» (lint, tests con `-race`, `schema-check`, `skills-check`, secretos y módulos).

Si el árbol llega al intento 2 sin los dos ficheros de arriba, se rehacen como se describe; si llega con ellos, basta
la entrada, `make ci` en primer plano y marcar la tarea.
