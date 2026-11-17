# Quickstart: validar H41.3

Desde la raíz del repositorio, con Go 1.27.4 (`go.mod`). Los escenarios 1 a 6 no usan modelo ni credencial, no escriben
en el índice ni en el historial de git y salen con 0 cuando el hito está hecho. Los escenarios 1 a 5 no usan red ni
escriben en el árbol de trabajo: solo en la caché de compilación de Go y, el 3 y los tests, en el directorio temporal. El
6, `make ci`, además: su objetivo `vuln` consulta la base de vulnerabilidades de Go, que requiere red (`Makefile:128`), y
`test` y `test-integration` dejan `coverage.out` y `coverage-integration.out` en la raíz (`Makefile:97` y `:101`),
ignorados por git (`.gitignore`: `/*.out`, `/coverage.*`). Lo que comprueba cada uno, en contracts/; los nombres de los
tests, en plan.md, «Gates». Antes del hito ya salen en verde los tests de H39 a H41.2 que nombran los escenarios 2 y 3 y
las suites del 4; lo demás nombra tests, ficheros o argumentos que entrega el hito.

## 1. La variante

```sh
LC_ALL=C.UTF-8 awk '{print FILENAME, length($0); total += length($0)} END {print "total", total}' \
  laboratorio/producto-descripciones-cortas/*.txt
go test -count=1 -run '^(TestDescripcionesCortas|TestConLaDescripcion)$' ./internal/evals/
```

`awk`, en Linux con GNU awk, que en `C.UTF-8` cuenta caracteres (en macOS, el `awk` del sistema puede contar bytes, y
la cifra que vale es la del test), solo lee, y da cada fichero con su ruta `laboratorio/producto-descripciones-cortas/`
delante: `consulta-juridica.txt 417`, `jurisprudencia.txt 420`,
`normativa.txt 426`, `plazos-y-territorio.txt 427`, `total 1690` (contracts/variante.md §2). Verde: la variante del
repositorio no tiene ningún defecto; cada caso roto de contracts/variante.md §3, su línea exacta; y cada `SKILL.md` de
`skills/` compuesto con textos con `: `, comillas, `#` y `«»` da su texto exacto, pasa `skills.ValidarFrontmatter` y deja
lo demás igual.

## 2. El arnés, con el sustituto de `claude`

```sh
go test -count=1 -run '^(TestArnesConLasDescripcionesCortas|TestArnesConElPresupuesto|TestArnesConLaInstalacionDelProducto|TestArnesConUnaEval|TestArnesDeLasFamilias|TestArnesDeNormativa|TestArnesDelLaboratorio|TestRecetasDelSondeo|TestGuionDelLaboratorio)$' ./internal/evals/
```

Verde: con `INSTALACION=producto-descripciones-cortas`, cada sesión tiene las cuatro skills de `skills/`, cada una con la
`description` de la variante y lo demás byte a byte, y `sondeo.json` lleva `instalacion`; con `PRESUPUESTO=1000000`, cada
sesión tiene `SLASH_COMMAND_TOOL_CHAR_BUDGET=1000000` y `sondeo.json` lleva `presupuesto`; sin `PRESUPUESTO`, ninguna
sesión tiene la variable aunque la tenga quien lanza, y el entorno de cada una es, valor a valor, el que componen las
funciones de hoy, salvo `PWD`, `SHLVL` y `_`; sin las dos opciones, lo de hoy; cada petición inválida, ninguna
sesión y su mensaje; y la receta pasa diez argumentos y el guion los da a `go test`, con `-presupuesto` delante de
`-temporal` (contracts/arnes.md §6). `TestRecetasDelSondeo` ejecuta `make -n`, que solo imprime;
`TestGuionDelLaboratorio` ejecuta el guion con un sustituto de go delante en el `PATH`, que solo anota sus argumentos.

## 3. La tabla

```sh
go test -count=1 -run '^(TestTablaConLaInstalacion|TestTablaConLasActivaciones|TestTablaDeLasFamilias|TestTablaDeNormativa|TestTablaDelLaboratorio)$' ./internal/evals/
go run ./cmd/laboratorio tabla internal/evals/testdata/laboratorio/familias/instalacion/sondeo > "${TMPDIR:-/tmp}/h41-3-tabla.md" \
  && cmp "${TMPDIR:-/tmp}/h41-3-tabla.md" internal/evals/testdata/laboratorio/familias/instalacion/tabla.md \
  && echo igual
```

Verde e `igual`: cinco filas, con «Instalación» `normativa`, `producto`, `producto`, `producto-descripciones-cortas` y
`producto` y «Presupuesto» vacío, vacío, `1000000`, vacío y vacío; los recuentos y los intervalos de contracts/tabla.md
§4; y las tablas de antes, sin las columnas, byte a byte. Escribe solo `h41-3-tabla.md` en el directorio temporal.

## 4. Las suites congeladas y la del hito

```sh
go test -count=1 -run '^TestEntregaDelHito$/^h(39|40|41)-' ./internal/app/
```

Verde: las de H39 a H41.2 sin cambiar un fichero (FR-032) y `h41-3-laboratorio-tabla` (contracts/aceptacion.md), activada
al final del run. Construye los binarios del arnés e2e en el temporal del test.

## 5. Que el arnés no se lanza desde `make ci`

```sh
grep -n 'evals-laboratorio\|TestSondeoDelLaboratorio' Makefile .github/workflows/ci.yml
grep -n '^ci:' -A2 Makefile
```

Solo lee. Da tres líneas del `Makefile` —la de `.PHONY`, la de la ayuda y la del objetivo— y ninguna de `ci.yml`; `ci` no
tiene `evals-laboratorio` entre sus objetivos, y `TestSondeoDelLaboratorio` solo compila con `-tags evals`
(`internal/evals/job_test.go:1`) (FR-017).

## 6. Todo

```sh
make ci
```

## 7. La aceptación de la persona (fuera del run; FR-030, FR-031)

Lo hace una persona con la propuesta de cambio abierta, con la credencial del sondeo y Claude Code 2.1.284 en el `PATH`
(el arnés lo comprueba antes de abrir nada); abre 640 sesiones con modelo, que consumen su suscripción, y las deja fuera
del repositorio, bajo `../kitlegal-sondeos/h41-3/`, cuyos cuatro `DIRECTORIO` no pueden existir antes. Las órdenes, las de
contracts/arnes.md §1: cuatro invocaciones de `make evals-laboratorio` y `make laboratorio-tabla
DIRECTORIO=../kitlegal-sondeos/h41-3`.

La persona deja la tabla en `evidencias/`, con el modelo, la versión de Claude Code, la instalación y el presupuesto de
cada fila, y su lectura en la propuesta de cambio, que no se fusiona sin ellas. La tabla va a un ADR con las tres
comparaciones de FR-031 en «No activan ninguna» —dos intervalos no se tocan si el extremo superior del de abajo es menor
que el inferior del de arriba, con los extremos que imprime la tabla—: control sobre `sola` (la caída es de la
instalación completa; si se tocan, no se demuestra, y las otras dos no deciden nada); diagnóstico bajo control (la causa
es el presupuesto del listado); y remedio bajo control, con su «Activan otra skill» no por encima de la del control (el
hito siguiente lo aplica y lo valida con las evals de las cuatro skills en el job). Con 160 por fila, 0 de 160 da
[0,0–2,3 %] y 8 de 160, [2,6–9,6 %], que ya no lo toca (research V6).

## 8. Lo que deja el cierre (fuera de los escenarios anteriores)

En el job de evals del cierre, las evals de las cuatro skills, en los dos modos, con sus umbrales de hoy. Si
`sin_activar` de la eval 04 de `normativa` en modo herramienta sale en rojo, el informe final da el umbral sin cumplir con
sus cifras y la deja dicha para la persona (FR-034). No se reproduce en local: lo mide el workflow.
