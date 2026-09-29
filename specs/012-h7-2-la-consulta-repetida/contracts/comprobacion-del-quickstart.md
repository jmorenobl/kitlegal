# Contrato: la comprobación de la consulta repetida (quickstart)

FR-060 a FR-062; SC-002; US5. Decisiones en research D17 y D18. Las órdenes, en [quickstart.md](../quickstart.md) §6.

## 1. Punto de entrada

`TestComprobarConsultaRepetida`, en `internal/evals/job_test.go` (`//go:build evals`): fuera de `make ci` y del run,
como `TestPlanDeSesiones`, `TestPrepararSesion` y `TestInformeDelJob` (FR-062). Se invoca así:

```bash
go test -tags evals -count=1 -v -run '^TestComprobarConsultaRepetida$' ./internal/evals/ -args \
  -skill boe-legislacion -primera "$d/primera" -segunda "$d/segunda" \
  -fecha-superada 20180309 -fecha-leida 20200206
```

| Bandera | Qué es |
|---|---|
| `-skill` | la de siempre del job: de ella sale `evals/<skill>/`, donde está la lista |
| `-primera`, `-segunda` | los directorios de las dos conversaciones, cada uno con `sesion.jsonl`, `sesion.err` y `codigo-de-la-sesion` |
| `-fecha-superada`, `-fecha-leida` | las dos fechas de vigencia que la primera respuesta lleva en la línea de la forma |

Todas obligatorias (`exigirBanderas`). Lee la lista con `LeerConjunto(../../evals/<skill>)` —si hay ficheros mal
formados, falla nombrándolos, como `TestPlanDeSesiones`— y cada conversación con `LeerSesion`, la misma lectura del job:
la respuesta es el `result` del último mensaje `result` del transcript. Un error de `LeerSesion` hace fallar el test con
`la <primera|segunda> conversación no se ha podido leer: <error>`.

## 2. Las condiciones

`comprobarConsultaRepetida(primera, segunda Sesion, prohibidas ExpresionesProhibidas, superada, leida string) []string`,
en `internal/evals/consulta_repetida.go`, devuelve una línea por lo que falla, en este orden, y ninguna si se cumple
todo:

| Qué se mira | Línea si falla |
|---|---|
| La primera conversación terminó (`Terminada`) | `la primera conversación no terminó: <MotivoSinTerminar>` |
| La segunda conversación terminó | `la segunda conversación no terminó: <MotivoSinTerminar>` |
| 1 · Alguna línea de la primera respuesta casa con la forma de `version-obsoleta` (`formasDeHallazgo`, la del juicio) y contiene `superada` y `leida` como palabras (`contieneComoPalabra`) | `la primera respuesta no lleva ⚠ REDACCIÓN MODIFICADA: en una línea con <superada> y <leida>` |
| 2 · La segunda respuesta no lleva la forma (`ExtraerHallazgos`) | `la segunda respuesta lleva ⚠ REDACCIÓN MODIFICADA:` |
| 3 · La primera respuesta no lleva expresiones de la lista (`ExtraerExpresionesProhibidas`, la del juicio) | `la primera respuesta lleva expresiones prohibidas: <expresiones, separadas por «, »>` |
| 3 · La segunda, tampoco | `la segunda respuesta lleva expresiones prohibidas: <…>` |

Una conversación que no terminó no tiene respuesta que mirar: sus condiciones no se dan por cumplidas, y la línea de
«no terminó» basta. El test falla con todas las líneas, una por renglón; si no hay ninguna, registra con `-v`
`se cumplen las tres condiciones: la forma con <superada> y <leida> en la primera respuesta, sin ella en la segunda, y
ninguna expresión prohibida en las dos` y termina con `ok`.

No añade criterios a `Juzgar`, ni campos al formato de eval o al informe (FR-056): aplica a dos respuestas la lista y su
comparación, y la forma de H7.1.

## 3. Tests en `make ci`

`TestCondicionesDeLaConsultaRepetida`, en `internal/evals/consulta_repetida_test.go`, con sesiones construidas en el
test y la lista del repositorio: las dos
respuestas buenas → ninguna línea; la primera sin la forma; la forma en una línea y las fechas en otra; la segunda con la
forma; una expresión en la primera; una en la segunda; la primera sin terminar; y tres fallos a la vez → sus tres líneas
en orden (FR-061).
