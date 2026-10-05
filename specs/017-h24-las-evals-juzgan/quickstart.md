# Quickstart: comprobar H24

Cómo ver que el hito funciona, con el árbol de la rama `017-h24-las-evals-juzgan` ya implementado. Los escenarios 1 a
6 no usan la red ni abren ninguna sesión con modelo, y no cambian el árbol de trabajo, el índice ni la historia: los
tests escriben solo en directorios temporales y en la caché de construcción de Go. El 7 y el 9 los lanza una persona;
el 8 lo hace el workflow al cerrar. Las órdenes se ejecutan desde la raíz del repositorio.

Los contratos: [contracts/](./contracts/). Los tipos: [data-model.md](./data-model.md).

## 1. `make ci` en verde (SC-013)

```bash
make ci
```

Esperado: termina con 0. Dentro van todos los tests de abajo, `schema-check` y `skills-check`.

## 2. El voto, la frase y la regla (SC-002, SC-003, SC-007)

```bash
go test -count=1 -run '^(TestTextosDeLaSesion|TestMensajeDelVoto|TestOrdenDelVoto|TestFraseEnLaRespuesta|TestVotoDelJuez|TestReglaDeLosVotos)$' ./internal/evals/
```

Esperado: `ok  	github.com/jmorenobl/kitlegal/internal/evals`. Fijan lo de
[contracts/juez-y-voto.md](./contracts/juez-y-voto.md) §9: el mensaje y la orden de la validación, la frase que cruza
un salto de línea o pierde un acento grave, el voto nulo que se repite una vez, el que no llega a darse y las cinco
filas de la regla. `TestOrdenDelVoto` ejecuta `scripts/evals-voto.sh` con un `claude` sustituto: no abre ninguna
sesión.

## 3. El informe con el juez y sin la lista (SC-004, SC-011)

```bash
go test -count=1 -run '^(TestInformeConElJuez|TestUmbralesDelInforme|TestEjecucionSinMedir|TestJuzgarSinLaLista|TestJuicioDelSondeo|TestSalidaDelSondeo)$' ./internal/evals/
```

Esperado: `ok`. Con una respuesta marcada, veredicto `fallo` con la sesión y sus tres frases; con una medida que no
corresponde, `fallo` sin ninguna sesión abierta; y ninguna sesión deja de pasar por una expresión de la lista.

## 4. La medida, sus copias y los casos (SC-005, SC-006, SC-008, SC-009)

```bash
go test -count=1 -run '^(TestMedidaVersionada|TestCopiasDelJuez|TestGrabacionesDerivadas|TestEjecucionDeLaMedida)$' ./internal/evals/
go test -count=1 -run '^TestGrabacionesDerivadas$' ./internal/app/
```

Esperado: `ok` en las dos. La primera reconstruye en proceso los textos de los 259 casos y vota con un votante que
responde según la etiqueta: imprime en el test la medida con 0 de 212 y 0 de 47.

Las cuatro copias, a mano:

```bash
cmp evidencias/adr-0037/rubrica.md evals/boe-legislacion/juez/rubrica.md
cmp evidencias/adr-0037/esquema.json evals/boe-legislacion/juez/esquema.json
cmp evidencias/adr-0037/casos.yaml evals/boe-legislacion/juez/casos.yaml
cmp evidencias/adr-0037/medida.json evals/boe-legislacion/juez/medida.json
```

Esperado: ninguna escribe nada y las cuatro terminan con 0.

Las huellas de la medida, a mano:

```bash
shasum -a 256 evals/boe-legislacion/juez/rubrica.md evals/boe-legislacion/juez/casos.yaml
jq -r '.rubrica.sha256, .casos.sha256, .modelo_del_juez, .version_de_claude_code, .defectos.sin_marcar, .correctos.marcados' evals/boe-legislacion/juez/medida.json
```

Esperado: las dos huellas de la primera orden son las dos primeras líneas de la segunda
(`5f1e2115…06ee` y `4827894a…401a`), seguidas de `claude-opus-5-5`, `2.1.289`, `0` y `0`.

## 5. La skill (SC-010)

```bash
make skills-check
wc -l skills/boe-legislacion/SKILL.md
go test -count=1 -run '^(TestProsaDeLaSkill|TestEvalsDelRepositorio)$' ./internal/evals/
grep -n -o -e 'el sobre' -e 'fecha_vigencia' -e 'norma_modificadora' skills/boe-legislacion/SKILL.md | sort -u
```

Esperado: `skills-check` termina con 0; `wc` da 298 o menos; los tests, `ok`; y `grep` solo encuentra los dos nombres
de campo en las filas de la tabla generada (con el prototipo, `fecha_vigencia` en las líneas 240, 241 y 242 y
`norma_modificadora` en la 240 y la 241), y ninguna vez «el sobre».

## 6. La definición del job (SC-012)

```bash
go test -count=1 -run '^TestDefinicionDelJob$' ./internal/evals/
grep -n -e 'MODELO_DEL_JUEZ' -e 'VERSION_DE_CLAUDE_CODE_DEL_JUEZ' -e 'evals-medir-juez' -e 'medir_al_juez' -e 'timeout-minutes' .github/workflows/evals.yml
```

Esperado: `ok`; y en `grep`, `MODELO_DEL_JUEZ: claude-opus-5-5` y `VERSION_DE_CLAUDE_CODE_DEL_JUEZ: 2.1.289` en los
trabajos `evals` y `medida`, la etiqueta y la entrada en la condición de `medida`, y los topes `352` (`evals`) y `269`
(`medida`), con los de `cambios` y `tanda` como están.

## 7. El sondeo, con modelo (lo lanza una persona; FR-075, FR-076)

Abre sesiones con modelo y consume la suscripción: no lo ejecuta ningún paso del workflow.

```bash
make evals-sondeo SKILL=boe-legislacion EVALS=03,15 MODELO=claude-sonnet-5-5 REPETICIONES=1
```

Esperado: sale con 0; tras las tasas, las líneas del juez de
[contracts/informe-del-job.md](./contracts/informe-del-job.md) §8, con dos respuestas juzgadas, y la que dice si la
medida versionada corresponde al Claude Code del equipo. No escribe nada en el repositorio.

## 8. El job de cierre (lo lanza el workflow; SC-001)

Tras la revisión final, el workflow pone la etiqueta `evals` y recoge los informes en `gates/evals/`. Con ellos:

```bash
jq -r '.veredicto, (.umbrales[] | [.nombre, .medida, (.total // "-"), .cumple, .decide] | @tsv)' specs/017-h24-las-evals-juzgan/gates/evals/boe-legislacion.json
jq -r '(.juez.sin_juzgar | length), (.juez.respuestas | length), (.red | length)' specs/017-h24-las-evals-juzgan/gates/evals/boe-legislacion.json
jq -r '.veredicto, (.umbrales | length), .juez' specs/017-h24-las-evals-juzgan/gates/evals/legal-core.json
```

Esperado, en la primera: `aprobado` y doce filas, en el orden de
[contracts/informe-del-job.md](./contracts/informe-del-job.md) §2; las diez que deciden con `true` en `cumple`;
`afirma_lo_no_leido:claude-sonnet-5-5:orden` y `…:herramienta` con medida 0; las dos de `medida_del_juez:…` con 0 de
212 y 0 de 47; ninguna `expresiones_prohibidas:…` ni `redaccion_no_leida:…`. En la segunda: `0`, el número de
respuestas con algún voto afirmativo y `0`. En la tercera: `aprobado`, `0` y `null`.

## 9. La medida con el código del job (la lanza Jorge antes de fusionar; SC-014)

Fuera del run. Con la etiqueta `evals-medir-juez` en la propuesta de cambio, o con el flujo `evals` lanzado a mano con
su entrada `medir_al_juez`. Esperado, en el registro del trabajo `medida del juez (boe-legislacion)`: la medida entre
`--- inicio de medida.json ---` y `--- fin de medida.json ---`, con `"sin_marcar": 0` de 212 y `"marcados": 0` de 47, y
el trabajo en verde; ningún trabajo `evals (<skill>)` en esa ejecución. Si no se cumple, no se fusiona. Después, Jorge
lee en el informe del cierre las respuestas de `juez.respuestas` y las tres de las evals 19 y 20 en cada modo, y anota
en `docs/USO.md` si el juez acertó.
