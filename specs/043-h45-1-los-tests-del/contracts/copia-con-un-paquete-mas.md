# Contrato: la copia con un paquete más ejerce los cuatro casos (FR-010 a FR-013)

Los dos ficheros, `internal/app/testdata/script/h45-paquete-de-mas.txtar` y
`specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar`, son el mismo, byte a byte, y los cambia una sola tarea
`[datos]` (research D5, D6). Lo que no dice este contrato queda como hoy: la copia del módulo, `anadirEntradas`, la
sincronía, `ejecutar`, `principio`, `maxSalida`, `paquete/normas.yaml`, `paquete/plazos.yaml`, `paquete/cobertura.yaml`
y `paquete/prueba.md`.

## §1 · El comentario de cabecera

- El párrafo «El paquete solo añade: …» pasa a decir lo que añade (lo de hoy, la referencia en `normativa`, las cuatro
  evals de modo —la `99-` de hoy entre ellas— y la de lo no declarado) y lo que cambia en dos filas de la base de
  `data/cobertura.yaml`: estrecha la de extranjería y asilo y da su norma a la del orden social, que pasa a `prueba.md`.
- La lista de FR-016 de H45 pasa a ser «la de FR-016 de H45 con FR-011 de H45.1»: `TestControlesSobreLasReferenciasAMano`
  entre los tests enteros, y de `TestDefinicionDelJob` solo `cuatro-evals-mas`, entre los subtests.
- «Cubre:» suma «US6 · FR-010 a FR-013 · SC-003 de H45.1».

## §2 · El programa `paquete-de-mas.go`

**Filtros** (FR-011):

```go
filtroDeLosTestsEnteros = `^(TestLeerCobertura|TestCoberturaDelRepositorio|TestRegionDeLoCubierto|` +
    `TestRenderizarRecursosYPlazos|TestRegionDeLaCobertura|TestRegiones|TestVerticalDeEnsayo|` +
    `TestPlanEnDosModos|TestUmbralesDePlazosYTerritorio|TestUmbralesDeNormativa|` +
    `TestUmbralesDeLasSkills|TestInformeEnDosModos|TestMatrizDeLasPropuestas|` +
    `TestSkillsDelProductoDesdeControl|TestDescripcionesDelProducto|TestDescripcionesCortas|` +
    `TestControlesSobreLasReferenciasAMano)$`
filtroDeLosSubtests = `^(TestEvalsDelRepositorio|TestCasosDePlazos|TestEntregaDelHito|TestDefinicionDelJob)$/` +
    `^(precepto-de-la-cobertura|ordenes-para-powershell|conjunto-plazos-y-territorio|19-eval-c-tributos|` +
    `h32-paquetes|h33-cobertura-instalada|h36-plazos-de-hoy|h41-4-skills-instaladas|cuatro-evals-mas)$`
```

`go test -run` casa cada parte con un nivel (research V3): de `TestDefinicionDelJob` solo corre `cuatro-evals-mas`
(ningún otro de sus subtests se llama como uno de la lista), y nunca `del-repositorio`, `peor-caso` ni
`tope-con-las-evals-del-repositorio`.

**Los cambios de la base** (FR-010 a y b):

```go
var cambiosDeLaBase = []struct{ viejo, nuevo string }{
    {
        "  - procedimiento: actuaciones y procedimientos en materia de extranjería y asilo\n",
        "  - procedimiento: actuaciones y procedimientos en materia de extranjería y asilo, salvo los del paquete de prueba\n",
    },
    {
        "  - procedimiento: procedimientos sancionadores en el orden social\n" +
            "    precepto: {norma: BOE-A-2015-10565, bloque: daprimera, apartado: \"DA 1ª.2.c\"}\n",
        "  - procedimiento: procedimientos sancionadores en el orden social\n" +
            "    precepto: {norma: BOE-A-2015-10565, bloque: daprimera, apartado: \"DA 1ª.2.c\"}\n" +
            "    norma: {identificador: BOE-A-2015-11722, nombre: \"texto refundido de la Ley sobre Tráfico, Circulación de " +
            "Vehículos a Motor y Seguridad Vial\", aplicable: {bloque: a1, apartado: \"1\"}}\n",
    },
}
```

`cambiarUnaVez(fichero string, cambios []struct{ viejo, nuevo string }) error` sustituye cada `viejo`, que tiene que
estar una sola vez (`strings.Count`), por su `nuevo`; si no, devuelve un error que nombra el fichero y el texto, y la
instalación falla (`paquete-de-mas: instalación: …`, código 1, el guion cae en su `exec`).

**`instalar`**, tras `anadirEntradas`: `cambiarUnaVez` sobre `data/cobertura.yaml`; copia `paquete/prueba.md` en
`skills/plazos-y-territorio/references/prueba.md` (hoy) y `paquete/prueba-de-normativa.md` en
`skills/normativa/references/prueba.md`; copia la primera eval de `evals/plazos-y-territorio/` (`primerFichero`, hoy)
con los prefijos `96-`, `97-`, `98-` y `99-`; y `paquete/fuera-de-lo-declarado.yaml` en
`evals/plazos-y-territorio/95-fuera-de-lo-declarado.yaml`.

**La orden del paso 2** (US6.4): `const ordenDelPasoDos = "kitlegal boe articulo BOE-A-1978-31229 a140 --json && " +
"kitlegal graph check BOE-A-1978-31229 a140 --json"` y `lineaDeLaOrden(fichero string) (int, error)`, el número de la
única línea que es, sin los blancos de sus extremos, esa orden (error si son 0 o más de 1). En `main`, tras la
sincronía y antes de los tests: la del `skills/normativa/SKILL.md` del módulo (`os.Args[1]`) y la de la copia; un error
termina con 1; si no, imprime `orden del paso 2 de normativa, líneas más abajo: <copia − módulo>`.

**La salida** (lo lee el guion, en cada `make ci`): cinco líneas, una por paso, 109 bytes con todo en verde (research
P1), más, por cada orden de `go` que no sale con 0, su línea y el principio de su salida en la de errores, hasta 1 200
bytes, como hoy. No depende del uso del kit.

```text
copia: hecha
sincronía: 0
orden del paso 2 de normativa, líneas más abajo: 1
tests enteros: 0
subtests: 0
```

## §3 · Los ficheros nuevos del archivo

```text
-- paquete/prueba-de-normativa.md --
# Lo prueba

El paquete prueba en `normativa`: el de la copia con un paquete más, que comprueba que un paquete con referencia en
esta skill no pone en rojo ningún test ajeno. No es una materia, y ninguna pregunta lo pide.
-- paquete/fuera-de-lo-declarado.yaml --
# Eval de lo no declarado del paquete prueba: un plazo de una materia que kitlegal no declara cubierta. La respuesta lo
# declara con la plantilla y el motivo de lo no declarado, sin calcular ningún plazo.
pregunta: "Me perdieron la maleta en un vuelo del 3 de noviembre de 2026. ¿Hasta qué día puedo reclamar a la compañía aérea?"
activa: true
prohibidos:
  - applet: plazos
    verbo: calcular
no_cubierto:
  fuera_de_lo_declarado: true
```

La referencia de `normativa` no lleva ninguna expresión de `evals/normativa/expresiones-prohibidas.yaml`, ni órdenes, ni
líneas ⚠, ni fechas con cifras: sin mutar, ningún control de `TestControlesSobreLasReferenciasAMano` da un defecto con
ella (research D5).

## §4 · Las comprobaciones del guion

Tras las de hoy (`copia`, `sincronía`, `tests enteros`, `subtests` y las cinco de US2.1-US2.2 de H45):

```text
# --- la fila estrechada, en la referencia de la base; la que gana la norma, en la del paquete (US6.3 · FR-010) ---
grep '^\| actuaciones y procedimientos en materia de extranjería y asilo, salvo los del paquete de prueba \|' repo/skills/plazos-y-territorio/references/administrativo.md
grep '^\| procedimientos sancionadores en el orden social \|' repo/skills/plazos-y-territorio/references/prueba.md
! grep '^\| procedimientos sancionadores en el orden social \|' repo/skills/plazos-y-territorio/references/administrativo.md

# --- la referencia en normativa: su línea, en la región de los paquetes; la orden del paso 2, una más abajo (US6.4 · FR-010) ---
grep '^- Si la pregunta es de lo prueba, lee \[prueba\]\(references/prueba\.md\)\.$' repo/skills/normativa/SKILL.md
stdout '^orden del paso 2 de normativa, líneas más abajo: 1$'

# --- este guion y su copia congelada, iguales byte a byte (FR-013 de H45.1) ---
cmp $KITLEGAL_SKILLS/../internal/app/testdata/script/h45-paquete-de-mas.txtar $KITLEGAL_SKILLS/../specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar
```

Que la fila estrechada y la que gana norma sigan con sus datos de `c991808` en alguna referencia de un paquete lo
comprueba `h32-paquetes`, que corre en la copia entre los subtests (US6.3).

## §5 · Rojos

| Con | La salida | Y el guion |
|---|---|---|
| `TestCoberturaDelRepositorio` de `main` | `tests enteros: 1`; `--- FAIL: TestCoberturaDelRepositorio` | `h45-paquete-de-mas.txtar:<línea>: no match for ^tests enteros: 0$` |
| `mutantesDeLasOrdenes` con «línea 36» | `tests enteros: 1`; `--- FAIL: TestControlesSobreLasReferenciasAMano` | ídem |
| `cuatro-evals-mas` de `main` | `subtests: 1`; `--- FAIL: TestDefinicionDelJob/cuatro-evals-mas` | `… no match for ^subtests: 0$` |
| `llevaNoCubierto` de `main` | `subtests: 1`; `--- FAIL: TestEvalsDelRepositorio/precepto-de-la-cobertura` | ídem |
| la copia congelada distinta | — | `… and …/paquete-de-mas.txtar differ` |
| una fila de `cambiosDeLaBase` que ya no está una vez en `data/cobertura.yaml` (la cambió otro hito) | `paquete-de-mas: instalación: … no lleva una sola vez …` | `exec go run` sale con 1: la copia no deja de ejercer un caso de FR-010 sin decirlo |
