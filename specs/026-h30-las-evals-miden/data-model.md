# Data model: H30 · Las evals miden solo lo que cambia

Entidades del spec («Key Entities») con sus campos, reglas y relaciones. Todo vive en `internal/evals` (research D1),
sin exportar salvo las dos lecturas de `internal/skills` (D11). Las formas JSON que ve una persona o el cierre están en
[contracts/informe-del-job.md](contracts/informe-del-job.md); el cálculo, en
[contracts/eleccion-y-huella.md](contracts/eleccion-y-huella.md).

## 1. Rutas que miden las evals (FR-004)

Predicado sobre una ruta relativa a la raíz y los directorios de las entradas paquete de los conjuntos (§3 (d)), sin
leer el disco:

- miden: `skills/**`, `evals/**`, `data/**`, `cmd/**`, `internal/**`, `skills.go`, `Makefile`, `go.mod`, `go.sum`,
  `.github/workflows/evals.yml`, `scripts/evals.sh`, `scripts/evals-sesion.sh`, `scripts/evals-voto.sh`,
  `scripts/evals-medir-juez.sh`, `schemas/eval.yaml.json`, `schemas/expresiones-prohibidas.yaml.json`,
  `schemas/juez-clases.yaml.json` y las raíces de grabaciones de §3(e) fuera de `internal/`
  (`testdata/evals/boe.legislacion-consolidada/**`, `testdata/evals/grafo-previo/**`);
- no miden: `internal/workflow/**`, y, del directorio de un paquete que es entrada de algún conjunto por §3 (d), sus
  `*_test.go` y su `testdata/**`, salvo las raíces de grabaciones de §3(e) (FR-001(d); research D4). Los `_test.go` y
  `testdata/` de un paquete de lo común son rutas de lo común (§2), y los de un paquete del módulo que no es de ningún
  conjunto ni de lo común, rutas sin asignar (§6);
- cualquier otra ruta no mide.

## 2. Lo común (FR-002)

| Campo | Contenido |
|---|---|
| `rutas` | Prefijos y ficheros fijos: `cmd/`, `internal/app/`, `internal/cli/`, `internal/httpx/`, `internal/cache/`, `internal/mcp/`, `internal/evals/`, `skills.go`, `Makefile`, `.github/workflows/evals.yml`, los cuatro `scripts/evals*.sh` de §1 y los tres esquemas de §1. Solo cuentan las rutas que miden (§1) |
| `paquetes` | Los del módulo cuyo directorio cae en esos prefijos, y el paquete raíz para recorrer sus importaciones |
| `modulos` | Ruta y versión de cada módulo externo alcanzado desde `paquetes` (más `TestImports` y `XTestImports` de `internal/evals`) sin entrar en un paquete del módulo que no es de lo común (research D4, V8) |
| `go`, `toolchain` | Los de `go.mod` |

## 3. Conjunto de dependencias de una eval (FR-001)

Se calcula sobre la cabeza; nunca se escribe. `ConjuntoDeEval{Eval, Skill, Entradas []Entrada, Modulos map[string]string,
Error string}`; `Entrada{Clase: fichero|directorio|paquete, Ruta}` (research D2).

| Parte | Entradas |
|---|---|
| (a) | fichero `evals/<skill>/<NN-nombre>.yaml` |
| (b) | Skills de la eval: la que mide y, por cierre, cada skill cuyo nombre (el de su directorio de `skills/`) aparece en el `SKILL.md` de otra del conjunto, entre límites que no son `[a-z0-9-]`. De cada una: fichero `skills/<s>/SKILL.md`, directorio `skills/<s>/references/` |
| (c) | fichero `data/<x>` de la cabecera `<!-- generado desde data/<x>, no editar -->` de cada referencia (`skills.DatosDeLaReferencia`) |
| (d) | De cada applet de la tabla de comandos de cada skill de (b) (`skills.AppletsDeLaTabla`): el fichero de `internal/app` con `Nombre()` que devuelve ese literal; sus importaciones del módulo fuera de lo común; su cierre por `Imports` en el `go list` de research D3. De cada paquete del módulo del cierre que no es de lo común ni el raíz: entrada paquete (su directorio) y un fichero por cada `EmbedFiles`. De cada paquete externo: `Modulos[path] = version` |
| (e) | directorios de las raíces de grabaciones (`GrabacionesDeH4`, `GrabacionesDeH5`, `GrabacionesDeH9` de `grabaciones.go`, sin `../../`) y, si la eval tiene `grafo_previo`, `testdata/evals/grafo-previo/<nombre>/` |
| (f) | si `evals/<skill>/juez/` existe, ese directorio; y cada fichero de `evals/<skill>/` que no es una eval (no casa con `[0-9][0-9]-*.yaml`) |

**Pertenencia.** Una ruta pertenece al conjunto si es igual a una entrada fichero, empieza por una entrada directorio, o
está en el directorio de una entrada paquete (sin subdirectorios), termina en `.go` y no en `_test.go`. Vale para rutas
que ya no existen en la cabeza.

**Error (FR-005).** El conjunto no se puede calcular si falta `skills/<skill>/` de su eval, si un applet de una tabla no
tiene fichero en `internal/app`, si una referencia no tiene cabecera, si la tabla no tiene sus marcas, o si un paquete
del cierre trae `Error` o `DepsErrors`. `Error` nombra la causa; un conjunto sin entradas también es error.

## 4. Árbol

`Arbol{Ficheros map[ruta]sha256, Descripciones map[skill]string, Skills []string, DirectoriosDeEvals []string,
Modulos map[path]version, Go, Toolchain string}`.

- **Cabeza** (job y `make ci`): `Ficheros` son las rutas que miden (§1) de `git ls-files -z --cached --others
  --exclude-standard`, con el sha256 de su contenido; `Descripciones`, el `description` de `skills.LeerFrontmatter` de
  cada `SKILL.md`; `Modulos`, `Go` y `Toolchain`, de `go mod edit -json go.mod`.
- **Base** (solo en el job): lo mismo de la base de fusión, sin `Ficheros`, con `git show` y `git ls-tree` (research D5).
- En los tests, la base y la cabeza son el árbol del repositorio con mutaciones en memoria.

## 5. Cambio y elección (FR-006 a FR-008)

`Cambio{Ruta, Anterior string}` (FR-007: `Anterior` es `previous_filename`; las dos rutas cuentan).

`EvalElegida{Eval, Skill, Motivo, Rutas []string}`; `Eleccion{Elegidas []EvalElegida, SinConjunto string,
ErrorDeLaPlataforma *ErrorDeLaPlataforma}`; `ErrorDeLaPlataforma{Orden, Error string}` (la orden de `gh` y su salida de
error, o «plazo de 10 minutos agotado» con la orden en curso).

Regla, para cada ruta *r* de cada cambio (las dos de un renombrado), la primera que se aplica:

1. *r* no mide (§1) → nada.
2. *r* es `go.mod` o `go.sum` → se comparan `Modulos`, `Go` y `Toolchain` de base y cabeza: si cambian `go`,
   `toolchain` o un módulo de lo común → todas, `comun`; si cambia algún módulo de algún conjunto → esas evals,
   `conjunto`; si no cambia ningún módulo de un conjunto ni de lo común → todas, `sin_asignar`.
3. *r* es de lo común (§2) → todas, `comun`.
4. *r* es `skills/<s>/SKILL.md` y la descripción de *s* difiere entre base y cabeza → todas, `descripcion`.
5. El primer componente bajo `skills/` o `evals/` de *r* está en uno solo de los dos árboles (directorio añadido o
   retirado) → todas, `sin_asignar`.
6. *r* pertenece al conjunto de alguna eval → esas evals, `conjunto`.
7. Si no → todas, `sin_asignar`.

Si los conjuntos no se pueden calcular en el job → todas, `sin_conjunto`, y `SinConjunto` lleva la causa. Si `gh` no da
los ficheros de la propuesta (la orden falla o, en la decisión de un trabajo, se agota su plazo) → todas,
`plataforma`, sin rutas, y
`ErrorDeLaPlataforma` lleva la orden y su error (research D17). El motivo de una eval es el de más prioridad entre los
que la eligen (`comun` > `sin_conjunto` > `plataforma` > `sin_asignar` > `descripcion` > `conjunto`, research D19); `Rutas`, todas las que la eligen, ordenadas. La elección se ordena por skill y por fichero de
eval (FR-008).

## 6. Huella (FR-009)

sha256 en hexadecimal (64 caracteres) del texto de [contracts/eleccion-y-huella.md](contracts/eleccion-y-huella.md)
§4: modelos y versiones fijados, juez de la skill, `go`/`toolchain`, módulos del conjunto y de lo común, el resto de
`go.mod` y de `go.sum` (research D7), descripción de cada skill, y rutas del conjunto, de lo común y sin asignar con el
sha256 de su contenido. **Rutas sin asignar** (en el árbol): las de `Arbol.Ficheros` que no pertenecen a ningún
conjunto ni a lo común, salvo `go.mod` y `go.sum`, que entran por su resto; entran en la huella de todas.
`Arbol` guarda para eso `GoModResto` y `GoSumLineas` (las líneas de `go.sum` por módulo).

## 7. Medida de `main` (FR-010)

`MedidaDeMain{Ejecucion int64, Commit string, Propuesta int, Fecha time.Time, Trabajos map[skill]int64}`.

Válida si: la ejecución (`gh run list`) terminó (`status: completed`) y no se canceló; su `headSha` es el `head.sha`
de una propuesta fusionada hacia `main` (`merged_at` no nulo); y cada trabajo `evals (<skill>)` de la ejecución terminó
con conclusión `success` o `failure` y su paso «El informe del trabajo está escrito» en `success`; y su `Fecha` es
de menos de 90 días antes de la decisión (de una medida más antigua ningún resultado es reutilizable, §8, y su registro
ya no se lee). Se ordenan por `Fecha` (la `createdAt` de la ejecución) y por `Ejecucion`, de la más reciente a la más
antigua; se recorren como mucho 10 (FR-012).

## 8. Origen de las evals (FR-011, FR-022)

`OrigenDeLasEvals{Modo, SinConjunto, ErrorDeLaPlataforma *ErrorDeLaPlataforma, EvalsMedidas, EvalsReutilizadas,
SesionesAbiertas, Evals []OrigenDeEval}`;
`OrigenDeEval{Eval, Huella, Medida bool, Motivo, Rutas, Procedencia *Procedencia, Aportes []Aporte,
Tasas []TasaDelInforme}`; `Procedencia{Ejecucion, Trabajo int64, Commit string, Propuesta int, Fecha string}`;
`Aporte{Umbral string, Medida float64, Total int}`.

- `Modo`: `completa` o `reutilizacion`.
- `ErrorDeLaPlataforma` (§5): el de la lectura de los ficheros de la propuesta, o el primero de la búsqueda de medidas
  de `main` —una orden de `gh` que falla o el plazo de 10 minutos agotado—, que la para; ausente si no hubo. Con él,
  las evals con `sin_resultado_de_main` son las que quedaban sin resultado al pararse.
- Una eval medida lleva `Motivo` (§5, `sin_resultado_de_main` o `suite_completa`) y no lleva `Procedencia` ni `Tasas`
  (sus series están en `tasas` del informe).
- Una eval reutilizada lleva `Procedencia` y `Tasas`, copiadas con sus `Aportes` de la medida en que se abrieron sus
  sesiones (FR-011): si la medida de `main` la había reutilizado, se copian su procedencia y sus datos.
- `Aportes`: por cada umbral de las respuestas del informe (todos menos los de la duración y los de la medida del juez),
  la medida y el total que dan las respuestas de esa eval; solo los de total mayor que 0.
- Invariantes: para cada umbral de las respuestas, la suma de los aportes de las evals medidas es su `medida` y su
  `total`; `EvalsMedidas + EvalsReutilizadas` es el número de evals de la skill en la cabeza.
- Es a la vez el fichero de la decisión (`-decision`, sin `SesionesAbiertas`) y el bloque `origen` del informe.

**Resultado reutilizable** de la eval *X* en un informe de `main`: la entrada de *X* en `origen.evals` con la misma huella;
si se midió allí, sus series (`tasas` con `eval` *X*) y sus aportes, y está completo si ninguna entrada de
`sesiones_sin_medir` es de *X* y, con juez, ninguna de `juez.sin_juzgar` es de una sesión de *X* del modelo que decide;
si se reutilizó allí, su entrada tal cual. En los dos casos solo es reutilizable si su procedencia —la propia medida si
se midió allí, o la `procedencia` copiada si se reutilizó— tiene una `Fecha` de menos de 90 días antes de la decisión,
la vida del registro de la plataforma (FR-012): un resultado no pasa de una medida a la siguiente más allá de la vida
del registro en que se abrieron sus sesiones. Si no, la búsqueda sigue con la medida anterior, y sin resultado
reutilizable la eval se mide con `sin_resultado_de_main`. La decisión recibe el instante con el que compara (en el job,
el de su arranque; en los tests, uno fijo).

## 9. Umbrales con lo reutilizado (FR-015, FR-024)

- Los umbrales de hoy cuentan solo lo medido.
- De cada umbral de las respuestas al que aporta alguna eval reutilizada, un umbral `<nombre>:reutilizadas` con la suma
  de esos aportes, el mismo `umbral` y `decide` igual que el original, salvo `<clase>:<modelo>:reutilizadas`, que no
  decide.
- `<clase>:<modelo>` (la clase que decide): `medida` y `total` de las respuestas medidas del modelo que decide en los dos
  modos (*n*); se cumple si `medida ≤ k(n)`; publica `umbral` = el de `clases.yaml` si el mayor *m* con
  `m/n ≤ umbral` es k(*n*), y si no `k(n)/n` (research D12).
- `limiteDeLaTasa`: *N* = respuestas del modelo que decide del plan completo de la skill, las que juzga el juez; *K* = el
  mayor *m* ≤ *N* con `m/N ≤ umbral`; *p*ref = el mayor `q/10000` (q entero ≥ 0) con P(Bin(*N*, *p*ref) > *K*) ≤ 0,02;
  k(*n*) = el menor *k* con P(Bin(*n*, *p*ref) > *k*) ≤ 0,02. Hoy: `boe-legislacion` *N* = 120, *K* = 3,
  *p*ref = 0,0085; `jurisprudencia` *N* = 60, *K* = 1, *p*ref = 0,0036 (research V5).

## 10. Tanda (FR-021)

`estadoDeLaTanda` gana el modo: `tandaSinDecidir`, `tandaQueNoMide`, `tandaQueMideConReutilizacion`,
`tandaQueMideLaSuiteCompleta`. Se lee de los pasos del trabajo `tanda` de cada ejecución: «Esta ejecución mide el
commit» en `success` (mide) y «Esta ejecución mide la suite completa» en `success` (completa). Una ejecución con modo
*m* no mide si una anterior sin terminar mide con modo ≥ *m* (completa > reutilización).

## 11. Comprobación de la suite completa (FR-025)

Para el commit *c*: las ejecuciones del flujo `evals` sobre *c*, terminadas, de la más reciente a la más antigua. Vale la
primera cuyo `headSha` es *c*, cuyos trabajos `evals (<s>)` de cada skill con evals en *c* terminaron en `success`, y en
cuyo informe de cada una `origen.modo` es `completa`, toda eval de `evals/<s>/` en *c* está medida y ninguna
reutilizada. Si ninguna vale, el motivo de cada una (la primera condición que no cumple) va al mensaje de error.
