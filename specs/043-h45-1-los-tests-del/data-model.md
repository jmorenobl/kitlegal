# Data model: H45.1 · Los tests del molde, sin el estado de hoy del repositorio

El hito no cambia ningún dato del repositorio (`data/`, `evals/`, `skills/`), ningún esquema ni ningún tipo de
producción. Las entidades son las de los cuatro tests y las de la copia con un paquete más; viven en código de test y en
el archivo del guion.

## 1. Tupla de lo regulado aparte (`TestCoberturaDelRepositorio`)

Una fila de `regulados_aparte` de `data/cobertura.yaml` sin el texto de su procedimiento (FR-001).

| Campo | De dónde sale en lo leído | Regla |
|---|---|---|
| `delPrecepto` | `ReguladoAparte.Precepto.Norma` | identificador BOE |
| `apartado` | `ReguladoAparte.Precepto.Apartado` | «DA 1ª.2.a», «112.4»… |
| `queLaRige` | `ReguladoAparte.Norma.Identificador`, o vacío si la fila no lleva `norma` | |

- **Las nueve** (escritas a mano, en su orden; contracts/tests-genericos.md §1): `(LPAC, DA 1ª.2.a, LGT)`,
  `(LPAC, DA 1ª.2.b, —)`, `(LPAC, DA 1ª.2.c, —)` dos veces, `(LPAC, DA 1ª.2.c, BOE-A-2015-11722)`, `(LPAC, DA 1ª.2.c, —)`,
  `(LPAC, DA 1ª.2.d, —)`, `(LPAC, 112.2, —)`, `(LPAC, 112.4, —)`, con LPAC = `BOE-A-2015-10565` y LGT =
  `BOE-A-2003-23186`.
- **De la base** (filtro escrito en el test, FR-002): la fila leída cuyo `queLaRige` tiene en `data/normas.yaml` el
  `vertical` `administrativo`, o, si no lo lleva o no tiene `vertical`, cuyo `delPrecepto` lo tiene.
- **Relación**: las tuplas leídas de la base son una subsecuencia de las nueve (research D1). Una fila que pasa a otro
  paquete sale de lo leído; una de las nueve puede no estar.

## 2. Orden de Bash del paso 2 (`mutantesDeLasOrdenes`)

La línea de `skills/normativa/SKILL.md` que es, sin los blancos de sus extremos,
`kitlegal boe articulo BOE-A-1978-31229 a140 --json && kitlegal graph check BOE-A-1978-31229 a140 --json`
(`bashDelPasoDos`). Está una sola vez; su número, contado desde 1 sobre el fichero entero, es 36 con `main` y uno más
por cada línea de la región de los paquetes (`<!-- paquetes: … -->`, que va antes, en el paso 1).

## 3. La carpeta de 16 evals (`cuatro-evals-mas`)

Una copia de `evals/` en el temporal en la que `plazos-y-territorio/` tiene exactamente 16 evals, `01-copia-<d>` a
`16-copia-<d>`, todas con el contenido de la primera eval de modo que decide del módulo (`!SinBinarioNiServidor &&
!Informativa`; con `main`, `01-territorio-municipio-cubierto.yaml`, y `<d>` es `territorio-municipio-cubierto.yaml`). Lo
que no es una eval de la carpeta se queda.

## 4. Eval de lo no declarado (`llevaNoCubierto`)

Una eval con `no_cubierto: {fuera_de_lo_declarado: true}` (`NoCubiertoEsperado.FueraDeLoDeclarado`,
`internal/evals/formato.go:244`): no lleva precepto. `llevaNoCubierto` la deja fuera; la de lo regulado aparte (con
`norma` y `precepto`) sigue dentro.

## 5. El paquete de prueba de la copia (`h45-paquete-de-mas`)

`vertical: prueba`, con la norma de tráfico `BOE-A-2015-11722`. Lo de hoy (data-model §6 de H45): la norma en
`data/normas.yaml`, el tipo `prueba-de-un-paquete-mas`, su fila de lo cubierto, una fila de lo regulado aparte con su
norma, `prueba.md` en `plazos-y-territorio` y `99-<primera eval>`. Lo que gana (FR-010; contracts/copia-con-un-paquete-mas.md):

| Cambio | Dónde, en la copia | Lo que queda tras la sincronía |
|---|---|---|
| (a) estrecha una fila sin norma, sin partirla | `data/cobertura.yaml`: «actuaciones y procedimientos en materia de extranjería y asilo» → «…, salvo los del paquete de prueba» | la fila, en `administrativo.md`, con el texto de `c991808` dentro |
| (b) da la norma del paquete a otra fila sin norma | `data/cobertura.yaml`: «procedimientos sancionadores en el orden social» gana `norma: {identificador: BOE-A-2015-11722, …, aplicable: {bloque: a1, apartado: "1"}}` | la fila, en `prueba.md` y no en `administrativo.md` |
| (c) una referencia en `normativa` | `skills/normativa/references/prueba.md` | la línea del paquete en la región de los paquetes del `SKILL.md`; la orden del paso 2, una línea más abajo |
| (d) evals en `plazos-y-territorio` | `96-` a `99-<primera eval>` y `95-fuera-de-lo-declarado.yaml` | 12 + 5 = 17 evals con `main` |

Ninguno toca la fila de los tributos, la del sancionador tributario y aduanero ni la del art. 112.4.
