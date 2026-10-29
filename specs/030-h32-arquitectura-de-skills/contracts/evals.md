# Contrato · Las evals, el job y la elección (FR-033 a FR-039, SC-008, SC-010 a SC-015)

## 1. Las carpetas (FR-033, FR-038)

| `c991808` | Entrega | Cómo |
|---|---|---|
| `evals/boe-legislacion/` (23 evals, `expresiones-prohibidas.yaml`, `juez/`) | `evals/normativa/` | `git mv` de la carpeta: mismos nombres de fichero |
| `evals/jurisprudencia/` (10 evals, `juez/`) | igual | — |
| `evals/legal-core/` (12 evals) | `evals/plazos-y-territorio/` | `git mv` de la carpeta |
| — | `evals/consulta-juridica/` (2 evals) | nuevas |

Los jueces van con su carpeta, sin tocar ningún fichero de `juez/` (`clases.yaml`, `casos.yaml`, `esquema.json`,
`medida.json`, `rubrica.md`): `afirma_lo_no_leido` sigue en 0.028 en `normativa` y en 0.017 en `jurisprudencia`, y
`cuenta_su_proceso` y `afirma_que_existe`, publicados sin decidir. Los casos etiquetados se resuelven por el nombre del
fichero de la eval dentro de su carpeta (`medida.go`), que no cambia. `TestCopiasDelJuez` compara
`evals/normativa/juez` con `evidencias/adr-0037` y `evals/jurisprudencia/juez` con `evidencias/adr-0037-jurisprudencia`.

`git diff --find-renames c991808 -- evals/` da cada fichero como renombrado o sin cambios (FR-033).

## 2. Las evals nuevas de `consulta-juridica` (FR-035, FR-036)

`evals/consulta-juridica/01-poliza-tras-el-fallecimiento.yaml` (FR-035):

```yaml
pregunta: "Mi madre falleció y después la aseguradora prorrogó su póliza y nos cobró la prima del nuevo periodo. ¿Qué dice la ley y qué podemos hacer?"
activa: true
no_se_activan: [jurisprudencia]
comandos:
  - applet: boe
    norma: BOE-A-1980-22501
    bloque: a22
citas:
  - norma: BOE-A-1980-22501
    bloque: a22
prohibidos:
  - applet: plazos
    verbo: calcular
lecturas:
  - norma: BOE-A-1980-22501
    bloque: a4
    motivo: "…"
```

No exige ninguna línea fija (FR-001): la prosa de lo no cubierto no la juzga ningún modelo (fuera de alcance). Decide:
no es `informativa`.

`evals/consulta-juridica/02-una-sola-capacidad.yaml` (FR-036), con la norma y el bloque de la eval 10 de `normativa`,
ya grabados:

```yaml
pregunta: "¿Qué dice el Estatuto de los Trabajadores sobre los días de vacaciones al año?"
activa: true
se_activan: [normativa]
no_se_activan: [jurisprudencia, plazos-y-territorio]
comandos:
  - applet: boe
    norma: BOE-A-2015-11430
    bloque: a38
citas:
  - norma: BOE-A-2015-11430
    bloque: a38
```

**`se_activan`** es un campo nuevo de `schemas/eval.yaml.json` (lista de nombres de skill, como `no_se_activan`): por
cada nombre que la sesión no activó, un motivo, y la sesión no pasa (`anotarLasQueSeActivan`, junto a
`anotarLasQueNoSeActivan` en `internal/evals/juzgar.go`). Lo pide el escenario 2 de US1: «se activan la skill de
entrada y la de esa capacidad».

**La Ley 50/1980**: entra en `data/normas.yaml` (`materias: [seguros]`, sin `vertical`), para que `normativa` la tome
de su tabla y la caché de las sesiones sirva su índice y sus metadatos, porque `boe buscar` no se sirve sin red
(`docs/SKILLS.md` §1). Su identificador y su título salen de lo grabado: la tarea `[datos]` añade al manifiesto
`internal/evals/testdata/grabaciones.json` la entrada

```json
{"busqueda": "contrato de seguro", "titulo_empieza_por": "Ley 50/1980,", "bloques": ["a4", "a22"], "para": "eval 01 de consulta-juridica (H32, FR-035): la Ley 50/1980, de Contrato de Seguro, sus arts. 4 y 22; …"}
```

y el paso `grabar_datos` la graba con `TestGrabarEvals` (`//go:build grabacion`, `internal/evals/grabacion_test.go`)
desde `boe.legislacion-consolidada`, con fila revisada en `docs/SOURCES.md`; lo grabado queda en
`internal/evals/testdata/boe.legislacion-consolidada/` y el material de origen en `$KITLEGAL_EVIDENCIAS`. La tarea
siguiente toma de la respuesta grabada de la búsqueda el título exacto para `data/normas.yaml`, y
`TestIdentificadoresDeLasNormas` lo comprueba: pasa a resolver cada norma de la tabla con el manifiesto de H5 y, la que
no está en él, con el de H9.

## 3. Las de cobertura (FR-014)

`07`, `08` y `12` de `plazos-y-territorio` ganan `precepto` en `no_cubierto` ([forma-de-cita.md](forma-de-cita.md) §4):

```yaml
no_cubierto:
  norma: BOE-A-2003-23186
  precepto: {norma: BOE-A-2015-10565, bloque: daprimera}
```

## 4. Las reglas de cada conjunto (FR-037)

| Carpeta | Reglas | Cambio |
|---|---|---|
| `normativa` | `ReglasDeNormativa` | las de `ReglasDeBoeLegislacion`, con su nombre nuevo |
| `jurisprudencia` | `ReglasDeJurisprudencia` | ninguno |
| `plazos-y-territorio` | `ReglasDePlazosYTerritorio` | las de `ReglasDeLegalCore`, con su nombre nuevo |
| `consulta-juridica` | `ReglasDeConsultaJuridica` | nuevas: al menos 2 evals; una que activa la skill, prohíbe `plazos calcular` y cita una norma sin `vertical` en `data/normas.yaml` (FR-035); y una con un solo nombre en `se_activan`, de una skill de capacidad, y en `no_se_activan` las demás de capacidad (FR-036) |

Y, sobre todas las carpetas, en `TestEvalsDelRepositorio`: cada skill de capacidad —las de `skills/` con
`kitlegal-applets`— está en el `no_se_activan` de al menos una eval de otra carpeta, y cada nombre de `no_se_activan` y
de `se_activan` es una skill de `skills/` (con un nombre retirado, el control dejaría de comprobar nada sin avisar).

En la entrega: `normativa` está en el de `plazos-y-territorio/01` a `04`; `jurisprudencia`, en el de las dos de
`consulta-juridica`; `plazos-y-territorio`, en el de `consulta-juridica/02`.

## 5. La elección de las evals del job (FR-038)

Las reglas de `eleccion.go` no cambian. Cambia lo que una skill carga (`calcularDependencias`, `dependencias.go`):

- **(b)** la referencia de un paquete vive en `skills/<skill>/references/`, ya en el conjunto de la skill;
- **(c)** también los ficheros de `data/` que nombran las marcas de inicio de las regiones de una referencia a mano
  ([paquetes.md](paquetes.md) §3), además de la cabecera de las generadas;
- **(d)** una skill sin `kitlegal-applets` —`consulta-juridica`— no tiene tabla de comandos y no es un error: aporta sus
  paquetes por las skills que nombra, por el cierre de (b).

`TestEleccionDeLosMutantes` gana dos casos: tocar `skills/plazos-y-territorio/references/administrativo.md` elige las
evals de `plazos-y-territorio` y las de `consulta-juridica`, que la nombra; y tocar `data/cobertura.yaml`, las mismas.

**La serie y la reutilización con un nombre nuevo** (ADR 0045, H30): la reutilización exige un trabajo
`evals (<skill>)` en la ejecución de `main` y la misma huella por eval, y la huella de cada eval lleva la descripción de
todas las skills del árbol: en la propuesta de cambio no se reutiliza nada y todo se mide, con
`sin_resultado_de_main` en las tres skills nuevas o renombradas. La serie de cierres de cada skill empieza de cero, con
la clave del árbol de su carpeta en `skills/`, también la de `jurisprudencia`, cuyo cuerpo cambia. Ni
`scripts/workflow/serie.sh` ni `cierre.sh` cambian: leen `evals/<skill>/juez/clases.yaml` por el nombre del trabajo.

## 6. La matriz del job y su tope (FR-039, SC-008)

`.github/workflows/evals.yml`, trabajo `evals`:

```yaml
skill: [consulta-juridica, jurisprudencia, normativa, plazos-y-territorio]
include:
  - {skill: consulta-juridica, concurrencia: 1, objetivo_de_duracion: 0}
  - {skill: jurisprudencia, concurrencia: 3, objetivo_de_duracion: 0}
  - {skill: normativa, concurrencia: 5, objetivo_de_duracion: 900}
  - {skill: plazos-y-territorio, concurrencia: 2, objetivo_de_duracion: 0}
```

Once sesiones a la vez (1 + 3 + 5 + 2). `PRUEBA_DE_RED`, con `matrix.skill == 'normativa'`. El trabajo `medida`:
`[jurisprudencia, normativa, descripciones]`, con las concurrencias de hoy.

El peor caso de cada trabajo con la suite completa, con la fórmula de `definicion.go` (485 s fuera de las sesiones +
900 s de la elección + tandas × 272 s + juez + 240 s de la descripción, o 300 s sin juez), recalculado en esta sesión:

| Trabajo | Evals | Sesiones por grupo (orden, herramienta, sin binario) | Tandas | Juez | Peor caso |
|---|---|---|---|---|---|
| `consulta-juridica` a 1 | 2 | 13, 12, 0 | 25 → 6 800 s | sin juez: 60 s | **8 485 s** (141,4 min) |
| `jurisprudencia` a 3 | 10 | 61, 60, 0 | 21 + 20 = 41 → 11 152 s | 60 + 2 × ⌈30 × 6 / 3⌉ × 40 = 4 860 s | **17 637 s** (294,0 min) |
| `normativa` a 5 | 23 | 109, 108, 6 | 22 + 22 + 2 = 46 → 12 512 s | 5 980 s | **20 117 s** (335,3 min), como hoy |
| `plazos-y-territorio` a 2 | 12 | 67, 66, 6 | 34 + 33 + 3 = 70 → 19 040 s | sin juez: 60 s | **20 725 s** (345,4 min), como hoy |

El mayor, 20 725 s, cabe en el tope de hoy, `timeout-minutes: 352` (21 120 s): no cambia. `TestDefinicionDelJob`
(`internal/evals/definicion_test.go`) fija `skillsDelTrabajo`, `ajustesDelTrabajo` y las cadenas de
`probarElTopeConLasEvalsDelRepositorio` con estas cifras, y sigue en rojo si falta una skill de la entrega, si está un
nombre retirado, si la suma pasa de once o si un tope no cubre su peor caso.

## 7. Los umbrales del informe (FR-038, SC-010 a SC-015)

Ningún valor cambia y no se añade ningún umbral (FR-038). `TestUmbralesDeLasTresSkills` pasa a
`TestUmbralesDeLasSkills`, con las cuatro: `normativa` con los 19 de `boe-legislacion`, `jurisprudencia` con sus 19,
`plazos-y-territorio` con los 18 de `legal-core`, y `consulta-juridica` con los que dan de verdad sus dos evals por las
reglas de hoy: `fuera_de_lo_grabado:<modelo>:<modo>`, uno por modelo y modo del plan del job, que no decide, y nada
más de sus respuestas. No tiene `sin_activar` ni ningún otro umbral de las respuestas, porque `medirLasRespuestas`
(`internal/evals/informe.go`) da nil a una skill sin juez, sentencias ni evals de plazo, de cobertura, de vigencia o
con una redacción fechada; ni `lectura_sin_intentar`, porque ninguna de sus evals tiene un comando de bloque en su
`sin_servir` (las `lecturas` de la 01 no lo son). Con la descripción en el informe, lleva además, como las otras
tres, el de la descripción, `habla_en_primera_o_segunda_persona`, que decide (H31). Lo que decide sobre sus sesiones
es el veredicto por la regla de cada serie (2 de 3; ADR 0016; SC-015): una sesión que no activa `consulta-juridica` o
una skill de su `se_activan` no pasa, y una serie que no llega a la regla pone el veredicto en fallo y el job en rojo.
`TestLimiteDeLaTasa` y su mutante, con `normativa`. `cobertura_sin_declarar`, con la medida de FR-014.

## 8. La caché de las sesiones (FR-035, SC-012)

`TestCacheDeLasSesiones` (`internal/evals/preparar_test.go`) gana el subtest de `consulta-juridica`: la caché de una
sesión de su eval 01 sirve `BOE-A-1980-22501` `a22` y `a4`, sin fecha y con `--fecha 9999-12-31`, con 0. Los subtests de
hoy pasan a las carpetas nuevas (`evalsDeNormativa`, `evalsDePlazosYTerritorio`) con los mismos ficheros y bloques, y
siguen en rojo si la caché de la skill que hereda las evals 14, 15 y 17 no sirve un bloque de sus `lecturas`.
