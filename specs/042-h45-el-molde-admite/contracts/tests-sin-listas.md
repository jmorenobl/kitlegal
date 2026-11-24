# Contrato: los tests de la lista dejan de fijar lo que hay en `data/`, `skills/` y `evals/` (FR-016, FR-017)

Regla: un test que comprueba el repositorio deriva lo esperado de `data/` y `evals/`; uno que comprueba una regla usa
sus datos de `testdata/`. Cada uno sigue en rojo con el mutante que lo ponía en rojo (columna de la derecha), y la
comprobación de SC-004 (contracts/aceptacion.md §1) lo corre con un paquete más. «Del paquete» quiere decir de las
normas con `vertical: administrativo` (data-model §2 y §3), con un filtro escrito en el test, no con la función de
producción, para que un mutante de la regla de producción no lo arrastre.

## `internal/skills`

| Test | Literal de hoy | Pasa a | Rojo con |
|---|---|---|---|
| `TestLeerCobertura/repositorio` (`cobertura_test.go:92`) | las 4 filas de `cubierto` | las 4 de la base, comparadas con las filas sin `materia` o de `administrativo`, en su orden; las demás, con `materia` que es un `vertical` de `data/normas.yaml` | un texto, el orden o la `materia` de una fila de la base cambiados en `data/cobertura.yaml` |
| `TestCoberturaDelRepositorio` (`:264`) | las 9 tuplas de `regulados_aparte` | las mismas 9 tuplas, sobre las filas del paquete (por su norma, si la lleva con `vertical`, o por su precepto) | una fila del paquete quitada, cambiada o sin su `norma`; `LeerCobertura` sin `Norma` |
| `TestRegionDeLoCubierto/repositorio` (`cubierto_test.go:52`) | 4 líneas de lo cubierto; la copia solo con `plazos-y-territorio` y `consulta-juridica` | una línea por fila de `cubierto` de `data/cobertura.yaml`, con la forma de `lineasDeLoCubierto`; la copia, con cada skill que tiene una referencia de un paquete | `lineasDeLoCubierto` sin una fila, en otro orden o con la materia dentro |
| `TestRegionDeLoCubierto/sin-data-cobertura` (`:89`) | el defecto de `administrativo.md` solo | uno por `references/*.md` de la copia con la marca de `cobertura_por_materia`, en orden de nombre, tras el de la plantilla | el defecto de los datos que faltan callado |
| `TestRenderizarRecursosYPlazos/desde-data` (`referencias_test.go:447`) | todos los tipos; `2+9` | los tipos del paquete; `2 +` una fila por tipo y una por unidad de cada tipo con unidades; la región de `references/administrativo.md` | `filasDeLosTipos` que da una fila menos; una región sin regenerar |
| `TestRegionDeLaCobertura` (`referencias_test.go:528`) | `filaDeLosTributos` en `administrativo.md`; `2+len(ReguladosAparte)` y `9`; una deriva | la fila de cada fila del paquete en `administrativo.md`, con la línea que da la plantilla; `2 + n` y `n` con `n` las del paquete; una deriva por referencia con la marca | el filtro de `filasDeLaCobertura` quitado (con `n` derivado sí lo ve); el motivo con norma sin sustituir; una región sin regenerar |
| `TestRegiones` (`regiones_test.go:193`; P2) | el `data/plazos.yaml` real (`arbolDeRegiones`, `:160-164`) | `testdata/regiones/plazos.yaml`, copia de los tipos y cómputos de hoy, nuevo en una tarea `[datos]` | sus mutantes de hoy, y los de contracts/sincronia-de-los-paquetes.md §6 |
| `TestVerticalDeEnsayo` (P3) | la fila de la vertical «detrás de la del territorio» (`vertical_de_ensayo_test.go:151-154`) | detrás de la que era la última de la región en la copia antes de instalar | la fila de la vertical en otro sitio |

## `internal/app`

| Test | Literal de hoy | Pasa a | Rojo con |
|---|---|---|---|
| `TestCasosDePlazos/19-eval-c-tributos` (`plazos_casos_test.go:219`) | la fila por el texto de su procedimiento | la única fila con precepto DA 1ª.2.a de la LPAC y `norma` la LGT | la fila sin `norma`, con otro `aplicable` u otro precepto, o repetida |
| `h32-paquetes.txtar` (`[datos]`, y `specs/030-h32-arquitectura-de-skills/aceptacion/paquetes.txtar`) | las filas de c991808 y la de tributos, dentro de `administrativo.md` (`:28-37`) | dentro de alguna referencia de un paquete de `plazos-y-territorio` (`references/*.md` con una región) | una fila del generador cambiada o perdida |
| `h33-cobertura-instalada.txtar` (`[datos]`, y `specs/039-h33-cobertura-en-positivo/aceptacion/cobertura-instalada.txtar`) | la región de lo cubierto con 4 líneas (`:35`); «filas: 9 de 9, 2 con norma» en `administrativo.md` (`:38-39`) | las 4 líneas de la base, en su orden, al principio de la región; las filas de todas las referencias de paquetes instaladas, tantas como `- procedimiento:` de `$KITLEGAL_SKILLS/../data/cobertura.yaml` y con norma tantas como sus `norma:` | lo cubierto o una fila perdidos en la instalación |
| `h36-plazos-de-hoy.txtar` (`[datos]`, y `specs/041-h36-el-motor-de/aceptacion/plazos-de-hoy.txtar`) | `error-tipo.json` con el mensaje que nombra los 7 tipos y su `hash` (`:35-36`, `:69`) | código 2, `clase` `argumentos` y el mensaje con su forma: `el tipo \"inventado\" no es ` seguido de códigos separados por «, » y « ni » el último, con `stdout` y expresiones | otra clase o código; la enumeración sin « ni »; el mensaje sin el tipo pedido |
| `h41-4-skills-instaladas.txtar` (`[datos]`, y su copia de `specs/038-h41-4-las-skills-del/aceptacion/`) | la descripción de hoy (`:34`) | la de FR-023 | la de hoy (la que nombra la LPAC) |

## `internal/evals`

| Test | Literal de hoy | Pasa a | Rojo con |
|---|---|---|---|
| `TestDefinicionDelJob/peor-caso`, `/del-repositorio`, `/tope-con-las-evals-del-repositorio` | planes, segundos y minutos escritos | contracts/evals.md §2 | `peorCaso` con otra cuenta; la concurrencia de hoy con cuatro evals más; 13 sesiones a la vez |
| `TestPlanEnDosModos/del-repositorio` (`plan_test.go:295`) | series y sesiones por modo | desde `LeerConjunto`, como en contracts/evals.md §2 | el plan sin la sesión de red o sin un modo |
| `TestEvalsDelRepositorio/precepto-de-la-cobertura` (`probarLosPreceptosDeLaCobertura`, `conjunto_test.go:2912`) | el mapa de las tres evals de cobertura (07, 08, 12) con la DA 1ª (`:2930-2934`) | las evals con `no_cubierto` de `LeerConjunto`, al menos una, y el precepto de cada una el de una fila de `data/cobertura.yaml` (`preceptosFueraDeLaCobertura`, de hoy) | su mutante de hoy: la 12 con el bloque `a999`, que ninguna fila tiene (`:2939-2944`) |
| `TestEvalsDelRepositorio/conjunto-plazos-y-territorio` (`:2670`; P1, P2) | «las doce» (`evalsDePlazosYTerritorioConLaDeVigencia`) | tantas evals como `*.yaml` del directorio, y ninguna mal formada | `LeerConjunto` que pierde un fichero |
| `TestEvalsDelRepositorio/ordenes-para-powershell` | — (no es una lista) | contracts/skills-y-laboratorio.md §2 | §C.3 |
| `TestUmbralesDePlazosYTerritorio` (`informe_test.go:6790`) | totales (30, 12, 9, 3, 9, 33), nombres de eval y `modelosDePlazosYTerritorio` (`:7579`) | desde `LeerConjunto(evals/plazos-y-territorio)` por 3 repeticiones, con los predicados de producción de cada clase (`esDePlazo`, `esDeCobertura`, `esDeVigencia`, `esperaUnaRedaccionFechada`, `comandosSinServir`); los modelos, por eval de modo | quitar `if r.conVigencia` (`umbrales.go:625`); `lectura_sin_intentar` que decide para todo modelo (`:1164`) |
| `TestUmbralesDeNormativa` (`:7907`) | `60, 6, 66, 42`, las dos evals fechadas, `900/522/511`, `Len 19`/`11` | desde `LeerConjunto(evals/normativa)` y el `ObjetivoDeDuracion` de la definición leída | `esperaUnaRedaccionFechada` cambiada; el umbral de duración fuera |
| `TestUmbralesDeLasSkills` (`:7150`) | 18 nombres y sus cuentas por skill | desde el conjunto de cada skill, como arriba | `umbralesDelInforme` en otro orden; `fuera_de_lo_grabado` que decide |
| `TestInformeEnDosModos/plazos-y-territorio` (`:2233`) | `deModo`, las 11 evals de modo | las evals de `LeerConjunto` sin `SinBinarioNiServidor`, en su orden | una serie en otro orden o que no decide |
| `TestMatrizDeLasPropuestas` (`eleccion_test.go:406`) | `goldenDeLaMatriz` con los ficheros de las cuatro skills | la parte de los ficheros, desde `repositorio.evals[skill]`; la forma de cada sección, literal | el formato de `matrizDeLaEjecucion` (`eleccion.go:374-395`) o la regla de elección cambiados |
| `TestSkillsDelProductoDesdeControl` (`laboratorio_familias_test.go:649`) | cada fichero del producto, declarado | contracts/skills-y-laboratorio.md §3 | §3 |
| `TestDescripcionesDelProducto`, `TestDescripcionesCortas` | sumas escritas | contracts/skills-y-laboratorio.md §4 | una descripción por encima de su tope o la suma por encima de 1 700 |

Fuera de la lista: `TestNormasDelRepositorio/paquete-de-lo-administrativo` (el spec lo deja como está) y los tests que
comprueban un dato contra lo grabado (`TestPreceptosDeLosPlazos`, `TestIdentificadoresDeLasNormas`,
`TestEvalsDelRepositorio/lo-que-lleva-a-leer`, `/ordenes-de-ejemplo`), que con un paquete sin grabaciones caen por lo que
comprueban (P1).

Las reglas del conjunto de `normativa` (`internal/evals/conjunto.go`: como mucho 23 evals, 10 positivas de normas
distintas) no cambian (FR-018).
