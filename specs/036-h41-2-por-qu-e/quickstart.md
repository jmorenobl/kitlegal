# Quickstart: validar H41.2

Desde la raíz del repositorio, con Go 1.27 (`go.mod`). Los escenarios 1 a 6 no usan modelo ni credencial, no escriben en
el índice ni en el historial de git y salen con 0 cuando el hito está hecho. Los escenarios 1 a 5 no usan red ni
escriben en el árbol de trabajo: solo en la caché de compilación de Go y, el 3 y los tests, en el directorio temporal. El
6, `make ci`, además: su objetivo `vuln` consulta la base de vulnerabilidades de Go, que requiere red (`Makefile:128`), y
`test` y `test-integration` dejan `coverage.out` y `coverage-integration.out` en la raíz (`Makefile:97` y `:101`),
ignorados por git (`.gitignore`: `/*.out`, `/coverage.*`). Lo que comprueba cada uno, en contracts/; los nombres de los
tests, en plan.md, «Gates». Antes del hito ya salen en verde los tests de H39 a H41 de los escenarios 1, 2 y 3 y las
suites del 4; lo demás nombra tests, ficheros o argumentos que entrega el hito.

## 1. Las skills activadas en el informe del job

```sh
go test -count=1 -run '^(TestActivacionesDeLosUsos|TestInformeConLasActivaciones|TestInforme|TestInformeDeLaSuiteCompletaNoCambia)$' ./internal/evals/
```

Verde: cada sesión fija de los dos modos da en `informe.json` y en `informe.md` su lista exacta —nombres, orden y
«antes»/«después», también con `Skill` en el mismo turno que la primera herramienta—; la sesión ilegible, sin lista y
«sin leer»; `activada`, `sin_activar` (2 de 5 en cada modo), las tasas, los motivos y el veredicto, los de hoy; y el
informe de la suite completa, sin la lista, igual byte a byte que antes (contracts/informe.md §5).

## 2. El arnés, con el sustituto de `claude`

```sh
go test -count=1 -run '^(TestArnesConLaInstalacionDelProducto|TestArnesConUnaEval|TestArnesDeLasFamilias|TestArnesDeNormativa|TestArnesDelLaboratorio|TestRecetasDelSondeo|TestGuionDelLaboratorio)$' ./internal/evals/
```

Verde: con `INSTALACION=producto`, cada sesión de una variante versionada y de la `-producto` de cada familia tiene las
cuatro skills, la de su variante en su lugar, byte a byte y sin repetir, y `sondeo.json` las nombra; sin la opción, lo de
hoy; con `EVAL`, solo sesiones de esa eval, tantas como `SESIONES`; cada petición inválida, ninguna sesión y su mensaje;
y la receta pasa los nueve argumentos y el guion los da a `go test`, con `-instalacion` y `-eval` delante de `-temporal`
(contracts/arnes.md §5). `TestRecetasDelSondeo` ejecuta `make -n`, que solo imprime; `TestGuionDelLaboratorio`
ejecuta el guion con un sustituto de go delante en el `PATH`, que solo anota sus argumentos.

## 3. La tabla

```sh
go test -count=1 -run '^(TestTablaConLasActivaciones|TestTablaDeLasFamilias|TestTablaDeNormativa|TestTablaDelLaboratorio|TestLecturaConCd)$' ./internal/evals/
go run ./cmd/laboratorio tabla internal/evals/testdata/laboratorio/familias/activaciones/sondeo > "${TMPDIR:-/tmp}/h41-2-tabla.md" \
  && cmp "${TMPDIR:-/tmp}/h41-2-tabla.md" internal/evals/testdata/laboratorio/familias/activaciones/tabla.md \
  && echo igual
```

Verde e `igual`: una fila por instalación, con las skills instaladas, «Activan otra skill», «Otras skills» y «No activan
ninguna» con sus intervalos, la fila de H41 sin constancia con `normativa` sola y sin sesiones terminadas, y el detalle
con la lista de cada sesión (contracts/tabla.md §4). Escribe solo `h41-2-tabla.md` en el directorio temporal.

## 4. Las suites congeladas y la del hito

```sh
go test -count=1 -run '^TestEntregaDelHito$/^h(39|40|41)-' ./internal/app/
```

Verde: las de H39, H40, H41 y H41.1 sin cambiar un fichero (FR-032) y `h41-2-laboratorio-tabla` (contracts/aceptacion.md).
Construye los binarios del arnés e2e en el temporal del test.

## 5. Que el arnés no se lanza desde `make ci`

```sh
grep -n 'evals-laboratorio\|TestSondeoDelLaboratorio' Makefile .github/workflows/ci.yml
grep -n '^ci:' -A2 Makefile
```

Solo lee. Da tres líneas del `Makefile` —la de `.PHONY`, la de la ayuda y la del objetivo— y ninguna de `ci.yml`; `ci`
no tiene `evals-laboratorio` entre sus objetivos, y `TestSondeoDelLaboratorio` solo compila con `-tags evals`
(`internal/evals/job_test.go:1`) (FR-014).

## 6. Todo

```sh
make ci
```

## 7. Lo que deja el cierre (fuera de los escenarios anteriores)

En el job de evals del cierre, `informe.json` e `informe.md` de cada skill llevan la lista de cada sesión
(contracts/informe.md §2 y §3); si `sin_activar` de `normativa` en modo herramienta sale en rojo, esas sesiones dicen si
activaron otra skill o ninguna (FR-033). No se reproduce en local: lo mide el workflow.

## 8. La aceptación de la persona (fuera del run; FR-030, FR-031)

Lo hace una persona con la propuesta de cambio abierta, con la credencial del sondeo y, en cada invocación, la versión de
Claude Code que pide `CLAUDE_CODE` en el `PATH` (el arnés lo comprueba antes de abrir nada); abre 160 sesiones con modelo,
que consumen su suscripción. Las órdenes, las de contracts/arnes.md §1: cuatro invocaciones y una tabla, con las sesiones
fuera del repositorio, bajo `../kitlegal-sondeos/h41-2/`; el `DIRECTORIO` de cada invocación no puede existir antes. La
del día se lee con `claude --version | awk '{print $1}'`.

La persona deja la tabla en `evidencias/`, con el modelo, la versión de Claude Code y las skills instaladas de cada fila,
y su lectura en la propuesta de cambio, que no se fusiona sin ellas. La tabla va a un ADR: si una fila reproduce la caída
—sesiones en «No activan ninguna» o en «Activan otra skill» sin «Activan antes»— y otra no, con intervalos que no se
tocan, qué diferencia la causa y qué hito o propuesta la arregla; si ninguna la reproduce, la tasa que queda acotada con
las 160 sesiones (0 de 160 da [0,0–2,3 %], research V3) y si la serie se sigue llevando a una persona o se decide sobre el
umbral.
