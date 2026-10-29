# Contrato · La suite de aceptación (ADR 0018; SC-006, SC-017, SC-018)

La primera tarea escribe seis guiones `testscript` en `specs/030-h32-arquitectura-de-skills/aceptacion/`; el workflow
comprueba que cada uno falla hoy por una aserción, los congela y, al terminar el bucle, los activa como
`internal/app/testdata/script/h32-<nombre>.txtar`, donde los corre `TestEntregaDelHito` en `make ci`. Leen el
repositorio por `$KITLEGAL_SKILLS` (el `skills/` del repositorio) y `$KITLEGAL_SKILLS/..` (su raíz), y el binario por
`$KITLEGAL_BIN`. Cada uno nombra en un comentario los FR y SC que cubre.

Reglas de todos:

- **El veredicto, por una aserción**: `exists`, `grep`, `! grep`, `stdout` o `cmp`. Una orden que puede salir con
  distinto de 0 por un resultado válido —`git grep` sin coincidencias sale con 1— va dentro de un `sh` que sale con 0
  y el veredicto lo da `stdout`.
- **Un paso del cuerpo** es el tramo entre su encabezado `### <n>.` y el siguiente `###` o `##`; antes de buscar una
  frase en él, el guion junta sus líneas con un espacio y colapsa los blancos (las líneas del cuerpo van cortadas a 120
  columnas).
- **Lo de `c991808`** (los pasos 3 y 6, la primera viñeta del paso 4, las referencias generadas de `legal-core`) va
  copiado en el guion como sección `-- c991808/<fichero> --`: el `make ci` de la propuesta corre sobre un clon
  superficial, sin ese commit.

| Guion | Historia | Cubre | Qué comprueba |
|---|---|---|---|
| `entrada.txtar` | US1 | FR-001 (empotrada e instalada), FR-004, FR-005, FR-007 · SC-017 | `kitlegal skills install --dir <tmp>` deja `consulta-juridica/SKILL.md`; ningún frontmatter de `skills/*/SKILL.md` contiene «derecho público»; cada applet de `boe`, `graph`, `cita`, `plazos` y `territorio` está en el `kitlegal-applets` de exactamente una skill, y `consulta-juridica` no declara ninguno; ningún nombre de `skills/` contiene `boe`, `cendoj` ni `eur` |
| `capacidades.txtar` | US2 | FR-006 · SC-017 | las formas de [reparto.md](reparto.md) §5, cada una como línea de un bloque `text` del `SKILL.md` de su skill; ningún cuerpo ni referencia de `normativa`, `jurisprudencia` ni `plazos-y-territorio` lleva el nombre de otra skill de la entrega como código en línea solo (`` `<nombre>` ``) o detrás de la palabra «skill» |
| `normativa.txtar` | US3 | FR-015, FR-016, FR-017, FR-018, FR-021, FR-040 · SC-017 | el paso 2 contiene el texto N4 de [textos.md](textos.md); el paso 5 contiene las dos frases de N8 y no contiene «o el número, o la materia» ni «Puedes avisar de que una materia se regula en otra parte»; los pasos 3 y 6 y la primera viñeta del paso 4, iguales a los de `c991808`, línea a línea; las formas de FR-018 en un bloque `text`; la tabla de comandos entre sus marcas, con una fila por verbo de `boe` y de `graph`; ningún cuerpo de capacidad contiene una norma de las evals por materia (la lista mínima de FR-021) |
| `paquetes.txtar` | US4 | FR-019, FR-021, FR-023 · SC-017 | cada fila de `cobertura_por_materia.md`, `recursos_y_plazos.md` y `leyes_vertebrales.md` de `c991808` tiene sus datos en `plazos-y-territorio/references/administrativo.md` —el procedimiento, el tipo o la norma, y cada pareja identificador y bloque como `[<identificador>, bloque <id>]`—; la línea de la fila de los tributos es la del ejemplo de [paquetes.md](paquetes.md) §3; el cuerpo de cada skill de capacidad no contiene ninguna norma con `vertical` de `data/normas.yaml` de la rama (identificador, título hasta la primera coma o abreviatura como palabra) ni ninguna de la lista mínima; el `SKILL.md` de `plazos-y-territorio` lleva, entre sus marcas, la línea del enlace a `administrativo.md` |
| `retiradas.txtar` | US5 | FR-026, FR-027, FR-028, FR-029, FR-031 · SC-006 | [instalacion-retiradas.md](instalacion-retiradas.md) §5: huellas que coinciden, en el proyecto y en `-g` (sale con 0; no están el directorio, el enlace ni la entrada del manifiesto; `data` lo dice); un fichero editado (sale con 0; nada cambia; el aviso con su ruta en `data` y en la salida legible); `--dry-run` (lo dice en la salida de error y el árbol no cambia); `install boe-legislacion` (sale con 2) |
| `nombres-retirados.txtar` | US7 | FR-032 · SC-018 | la búsqueda de abajo no da ninguna línea |

US6 no tiene guion: su aceptación son `TestDefinicionDelJob`, `TestEvalsDelRepositorio` y los tests de los umbrales en
`make ci`, y el job de cierre. US8 la mide una persona.

## La búsqueda de FR-032

```sh
git -C "$RAIZ" grep -n -E -e "$P1" -e "$P2" -e "$P3" -e "$P4" -- \
  skills evals internal cmd .github/workflows Makefile README.md web/src/pages \
  ':(exclude)evals/*/juez/*' ':(exclude)internal/evals/testdata' \
  ':(exclude)internal/app/retiradas.go' ':(exclude)internal/app/retiradas_test.go' \
  ':(exclude)internal/app/testdata/script/h32-retiradas.txtar' \
  ':(exclude)internal/app/testdata/script/h32-nombres-retirados.txtar' | grep -v 'alt="'
```

con los nombres retirados `boe-legislacion|legal-core` en `N`:

| Patrón | Lo que cuenta | Expresión |
|---|---|---|
| P1 | `skills/<nombre>` o `evals/<nombre>` del repositorio, sin otro segmento delante | `(^\|[^A-Za-z0-9_./-])(skills\|evals)/(N)([^A-Za-z0-9-]\|$)` |
| P2 | `.agents/skills/<nombre>` de un ámbito, sin otro segmento delante | `(^\|[^A-Za-z0-9_./-])\.agents/skills/(N)([^A-Za-z0-9-]\|$)` |
| P3 | el nombre como código en línea solo | `` `(N)` `` |
| P4 | el nombre detrás de la palabra «skill» | ``(^\|[^A-Za-z0-9_])[Ss]kills?[[:space:]]+`?(N)([^A-Za-z0-9-]\|$)`` |

Las exclusiones son las de FR-032: la lista de FR-025 y sus tests, el e2e de FR-031, los ficheros de los jueces, lo
grabado bajo `internal/evals/testdata/`, el texto alternativo de una captura y —porque el patrón no cuenta una ruta con
un segmento delante— las rutas a `specs/`. El guion se excluye a sí mismo porque escribe los patrones. Un test que lee
una sesión grabada sigue comparando con el nombre que la grabación lleva (`"boe-legislacion"` en una cadena de Go), que
no casa con ningún patrón; su comentario no lo dice detrás de «skill».
