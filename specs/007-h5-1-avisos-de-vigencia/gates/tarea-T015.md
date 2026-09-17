# T015 · La traza con la llamada desconocida cerrada

## Qué pasó

La revisión final corrigió ficheros fuera del directorio del hito (`CHANGELOG.md`, `CONTRIBUTING.md`, `README.md`,
`internal/evals/conjunto.go`, `internal/evals/formato.go`), así que la aceptación de la ejecución `35148840549` sobre
`8272bc8` dejó de cubrir la cabeza y se repitió por etiqueta (quickstart §11.6).

- Ejecución: https://github.com/jmorenobl/kitlegal/actions/runs/35156339496 (`evals`, evento `pull_request` por la
  etiqueta `evals`, `headSha` `064308fb3c64d4db69451eead2da660add14b512`).
- Conclusión: `failure`. Veredicto del informe: `fallo`.
- Único motivo de la raíz del informe:

  ```text
  04-lgt-prescripcion-claude-haiku-4-5-20251001-03: sesión ilegible: traza: traza ilegible: /home/runner/work/_temp/kitlegal-evals-boe-legislacion/sesiones/04-lgt-prescripcion-claude-haiku-4-5-20251001-03/traza/t.19547, línea 1: no es ninguna de las formas de línea de la traza: execve, clone, clone3, fork, vfork o connect con su resultado o sin terminar, ???( <unfinished ...>, una señal o la línea final: «???()                                   = ?»
  ```

- Ninguna serie por debajo de su umbral (`jq '.tasas[] | select(.pasa | not)'` vacío) y `red` vacío (`[]`).

## La línea

`???()`, 35 espacios, `= ?`: 43 caracteres, con el igual en la columna 41, el relleno de alineación de `-a 40` (research
V63 de H5). Es la llamada desconocida de la forma D —un hilo que murió en la parada de entrada de una llamada que strace
no llegó a identificar— cerrada con el resultado de la llamada sin terminar, `?`, como strace cierra la forma A de una
llamada conocida. T045 de H5 la dejó ilegible (`desconocida-con-resultado`) porque la sonda de V65 no la produjo en mil
repeticiones; el runner la ha producido. No se tiene el resto del fichero: el informe solo publica la línea del defecto,
y la ejecución no sube las trazas.

## Decisión

- **Leerla como la forma D**: sin argumentos y solo con el resultado `?`. Es lo que la línea dice y lo único
  observado; `? <unavailable>`, un resultado numérico, un error, argumentos o la marca dentro siguen ilegibles, porque
  no hay evidencia de ellos.
- **Rechazado relanzar sin arreglar**: repetir para buscar otro resultado lo prohíbe el procedimiento de la aceptación
  (T014, D17); la sesión volvería a ser ilegible en cuanto strace escribiera otra vez la línea.
- **Rechazado quitar la sesión ilegible del veredicto** en los modelos informativos: una traza ilegible no deja
  comprobar que la sesión no llegó a la red (FR-076 de H5), y eso vale para cualquier modelo.
- **Rechazado arreglarlo en una propuesta de cambio aparte**: la aceptación de H5.1 exige veredicto `aprobado` sobre la
  cabeza y el defecto la bloquea; es el mismo tipo de arreglo que T044 y T045 hicieron dentro de H5 cuando la prueba de
  red descubrió otras formas de línea.

## Después

`make ci` en verde, commit, push de la rama y repetición por etiqueta sobre el commit nuevo; su evidencia va como
sección vigente de `gates/evals-aceptacion.md`.
