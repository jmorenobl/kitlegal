# Pendiente de la revisión final de H32

## [h] El verbo del aviso de la retirada, partido en dos literales (juez B)

**Motivo.** `internal/core/instalacion/retiradas.go:35` (`avisoDeRetirada`) escribe «instaló» como `"insta" + "ló"`
para que misspell no lea «instal» como errata de «install», y lo explica en el comentario de las líneas 31-33: es una
supresión del linter por otra vía. La misma partición está en `internal/core/instalacion/retiradas_test.go:111`
(`avisoDeVieja`). El juez pide unir el literal y neutralizar el falso positivo donde lo hace el repositorio.

**Por qué no se arregla aquí.** Las dos salidas posibles quedan fuera de lo que puede tocar el corrector:

- El arreglo que fija el repositorio es la regla de ignorar de misspell en `.golangci.yml`
  (`linters.settings.misspell.ignore-rules`, «el español se neutraliza falso positivo a falso positivo»), que es
  configuración de lint: el corrector no puede modificarla, y el workflow desharía el cambio.
- Cambiar el texto del aviso para no usar el verbo no es posible: lo fija byte a byte la suite de aceptación congelada
  (`internal/app/testdata/script/h32-retiradas.txtar`, líneas 47, 60 y 80, con su huella en
  `gates/aceptacion-congelada.json`) y `contracts/instalacion-retiradas.md` §2.

Comprobado en esta sesión: con el literal unido, `go tool -modfile=tools/golangci-lint/go.mod golangci-lint run
--enable-only misspell ./internal/core/instalacion/...` da `retiradas.go:35:31: `instal` is a misspelling of `install`
(misspell)`; el fichero se restauró después y el código queda como estaba.

**Lo que hace falta, en una propuesta aparte o en la rama del hito, por una persona.**

1. En `.golangci.yml`, en `misspell.ignore-rules`, por orden alfabético, entre las entradas que ya hay:

   ```yaml
   # «Instaló», del aviso de la skill de nombre retirado que install no retira
   # (H32, FR-027; contracts/instalacion-retiradas.md §2), que fija la suite de
   # aceptación; misspell lee su raíz como «install».
   - instal
   ```

2. En `internal/core/instalacion/retiradas.go`, `avisoDeRetirada` con el literal unido
   (`" no es lo que instaló kitlegal: su protocolo puede activarse junto al de "`) y sin las líneas del comentario que
   explican la partición; en `internal/core/instalacion/retiradas_test.go`, `avisoDeVieja` con el literal unido.

3. `make ci` en verde: el texto emitido no cambia, así que `h32-retiradas.txtar` y la salida legible siguen iguales.
