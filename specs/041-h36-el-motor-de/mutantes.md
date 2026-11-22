# Mutantes y comprobación final del diff de H36

Registro de T013 (spec, «Controles»; plan.md «Mutantes» y «Controles de umbral»; ADR 0029). Todo lo que sigue se
ejecutó en la sesión de T013 y se vio terminar.

## La copia

- Árbol de prueba: `$TMPDIR/h36-mutantes`, un `git clone --no-hardlinks` del repositorio en la cabeza de la rama
  (`8bb44dbc`, T012), con `origin/main` en `82d4ccdd`, el `main` del repositorio. Nada se aplicó al árbol versionado.
- Los dos guiones de `specs/041-h36-el-motor-de/aceptacion/` copiados en la copia como
  `internal/app/testdata/script/h36-plazos-de-hoy.txtar` y `h36-describe.txtar`, donde los busca `TestEntregaDelHito`.
- Línea base, sin mutante: `TestEntregaDelHito/h36-describe` y `TestEntregaDelHito/h36-plazos-de-hoy` en `PASS`
  (`go test -v`), y todos los controles de la tabla en verde (`./internal/app/`: 75 pruebas; `./internal/core/plazos/`:
  211; `./internal/evals/`: 50; `./internal/skills/`: 24).
- Cada mutante se aplicó con `sed -i` (o `git checkout origin/main -- <fichero>`), su control se ejecutó con
  `go -C <copia> test -count=1 -run '<patrón>' <paquete>` y se deshizo con `git -C <copia> checkout HEAD -- .`; tras
  cada uno, `git -C <copia> status --short` solo daba los dos guiones activados.

## Mutantes de la tabla de plan.md

Las órdenes se ejecutan en la copia (`go -C <copia> test -count=1 -run …`); se escriben con su patrón y su paquete.

| # | Control | Mutante | Orden | En rojo |
|---|---|---|---|---|
| M01 | El `data` de hoy (SC-001) | una palabra del texto del aviso del art. 183 en `data/plazos.yaml`: «actuaciones judiciales» → «actuaciones procesales» | `-run '^TestEntregaDelHito$/^h36-plazos-de-hoy$' ./internal/app/` | `TestEntregaDelHito/h36-plazos-de-hoy`: `h36-plazos-de-hoy.txtar:29: stdout and caso-15.json differ` |
| M02 | El `data` de hoy (SC-001) | el silencio con `da: vencimiento` en lugar de `dia-siguiente` | la de M01; `-run '^TestCasosDePlazos$' ./internal/app/` | `TestEntregaDelHito/h36-plazos-de-hoy`: `h36-plazos-de-hoy.txtar:26: stdout and caso-11.json differ`; `TestCasosDePlazos/11-silencio` |
| M03 | El `data` de hoy (SC-001) | el texto del municipio de `contencioso`: «el de la sede del juzgado» → «el de la sede del tribunal» | la de M01 | `TestEntregaDelHito/h36-plazos-de-hoy`: `h36-plazos-de-hoy.txtar:51: stdout and error-municipios-contencioso.json differ` |
| M04 | El esquema sin datos (SC-005) | la ayuda y los enumerados de hoy en las etiquetas: `internal/app/plazos.go` e `internal/core/plazos/salida.go` de `main`, que solo difieren de la rama en esas etiquetas | `-run '^TestDescribeDePlazosSinDatos$'`, `'^TestEsquemasPublicados$'`, `'^TestEntregaDelHito$/^h36-describe$'`, `'^TestNingunCodigoDePlazosEnGo$'` y `'^TestAppletPlazos$'`, en `./internal/app/` | `TestDescribeDePlazosSinDatos/los-del-binario`, `/sin-los-tipos` y `/con-un-tipo-y-un-aviso-mas`; `TestEsquemasPublicados/schemas`; `TestEntregaDelHito/h36-describe` (`h36-describe.txtar:23`, la ayuda del tipo de §C.7); `TestNingunCodigoDePlazosEnGo`; `TestAppletPlazos/el-golden-de-describe` y `/los-enumerados-publicados` |
| M05 | Ningún código en Go (SC-003) | `if tipo := c.tipo; tipo.Codigo == "silencio"` en `computar` de `internal/core/plazos/computo.go`, que decide el día siguiente | `-run '^TestNingunCodigoDePlazosEnGo$' ./internal/app/` | `TestNingunCodigoDePlazosEnGo`: `../core/plazos/computo.go:43: el literal es el código de plazos "silencio"` |
| M06 | Ningún código en Go (SC-003) | la ayuda de hoy: `internal/app/plazos.go` de `main` | la de M05 | `TestNingunCodigoDePlazosEnGo`: `plazos.go:133: la etiqueta lleva el código de plazos "acceso-reclamar"`, y uno por cada código que nombra |
| M07 | Las propiedades (SC-004) | el motor cuenta como hábiles los días de un periodo `inhabil`: en `enesimoHabil`, `c.calendario.Inhabil(dia, c.delCalendario)` en lugar de `c.inhabil(dia)` | `-run '^TestPropiedadesDelComputo$' ./internal/core/plazos/` | `TestPropiedadesDelComputo/civil_días_hábiles_81001`, `/civil_días_hábiles_81002` y `/civil_días_hábiles_82001` |
| M08 | El civil (SC-002) | agosto de `civil` como `aviso`, con un aviso completo (`agosto-civil`, `AGOSTO CIVIL`, su texto), que pide la regla 9 | `-run '^TestComputoCivil$' ./internal/app/` | `TestComputoCivil/C1_agosto_inhábil_en_días_hábiles`: da `2026-08-07`, espera `2026-09-07`; también C3, C5 y C9 |
| M09a | El civil (SC-002) | agosto de `civil` como `no-corre`, con `civil` en solo `meses`, que la regla 5 admite | la de M08 | `TestComputoCivil/C3_prórroga_de_los_meses_fuera_de_agosto`: da `2026-09-15`, espera `2026-09-01`; también C1, C2, C5, C7, C8, C9 y C10 |
| M09b | El civil (SC-002) | agosto de `civil` como `no-corre`, con sus dos unidades, que la regla 5 rechaza | la de M08 | `TestComputoCivil`, antes de sus filas: `data/plazos.yaml: el periodo 1 del cómputo "civil" no corre y el cómputo admite otras unidades que meses` |
| M10 | El civil (SC-002) | `civil` sin el periodo del `12-24` al `01-06` | la de M08 | `TestComputoCivil/C2_del_24_de_diciembre_al_6_de_enero_inhábiles`: da `2026-12-28`, espera `2027-01-08`; también C4 |
| M11 | El civil (SC-002) | agosto de `civil` desde el `08-02` | la de M08 | `TestComputoCivil/C9_primer_día_de_agosto`: da `2025-08-01`, espera `2025-09-01` |
| M12 | El civil (SC-002) | el calendario sin los locales de Madrid: fuera de `data/festivos/locales/13.yaml` las cuatro líneas de `municipio: "28079"` | la de M08 | `TestComputoCivil/C10_fiesta_local_de_la_sede`: da `2026-11-09`, espera `2026-11-10` |
| M13 | El ensayo (SC-003) | un motor que solo reconoce agosto como periodo: `contiene` y `pasaPor` de `computo.go` dan falso en un periodo que no es del `08-01` al `08-31` | `-run '^TestComputoDeEnsayo$' ./internal/core/plazos/` | `TestComputoDeEnsayo/E1_inhabil_en_días_hábiles`; también E2, E3, E4, E7, E8 y E11 |
| M14 | El ensayo (SC-003) | un motor que ignora `dia-siguiente` fuera del silencio: `\|\| c.tipo.Codigo != "silencio"` en la condición de `computar` | la de M13 | `TestComputoDeEnsayo/E5_dia-siguiente`: da `2026-06-08`, espera `2026-06-09`; también E6 |
| M15 | El ensayo (SC-003) | un motor que no comprueba las unidades del cómputo: fuera el `case` de `unidadQueSuComputoNoAdmite` de `internal/core/plazos/calcular.go` | la de M13 | `TestComputoDeEnsayo/E12_unidades_del_cómputo`: `An error is expected but got nil` |
| M16 | El ensayo (SC-003) | un motor que da los avisos de periodo con un texto fijo, el del art. 183 de H9: `avisoDeSuPeriodo` de `internal/core/plazos/avisos.go` | la de M13; la de M01 | `TestComputoDeEnsayo/E7_aviso_por_su_primer_día`, `/E8_aviso_por_su_último_día` y `/E11_orden_de_los_avisos`; con el mismo mutante, `h36-plazos-de-hoy` sigue en verde: solo el ensayo lo ve |
| R01 | `LeerTipos` (SC-005) | quitar la regla 1 | `-run '^TestLeerTipos$' ./internal/core/plazos/`; y `-run '^TestLeerTipos$/^regla_1_'` | `TestLeerTipos/regla_1_·_el_cómputo_del_tipo_existe`, con un `panic` de puntero nulo: sin la regla, `LeerTipos` copia en el tipo un cómputo que no está |
| R02 | `LeerTipos` (SC-005) | quitar la regla 2 | `-run '^TestLeerTipos$' ./internal/core/plazos/` | `TestLeerTipos/regla_2_·_la_unidad_de_la_duración_la_admite_su_cómputo` |
| R03 | `LeerTipos` (SC-005) | quitar la regla 3 | la de R02 | las cuatro de la regla: `regla_3_·_las_unidades_del_tipo_son_las_de_su_cómputo`, `…_una_unidad_que_su_cómputo_no_admite`, `…_las_unidades_vacías` y `…_unidades_y_duración` |
| R04 | `LeerTipos` (SC-005) | quitar la regla 4 | la de R02 | las tres: `regla_4_·_el_precepto_de_un_periodo_es_del_cómputo_o_de_cada_tipo_suyo`, `…_el_de_un_solo_tipo_de_un_cómputo_con_varios` y `…_en_un_cómputo_sin_tipos` |
| R05 | `LeerTipos` (SC-005) | quitar la regla 5 | la de R02 | las cuatro: `regla_5_·_un_no-corre_es_un_mes_entero`, `…_desde_el_día_1`, `…_de_30_o_31_días` y `…_en_un_cómputo_de_solo_meses` |
| R06 | `LeerTipos` (SC-005) | quitar la regla 6 | la de R02 | las dos: `regla_6_·_dos_periodos_no_comparten_ningún_día` y `…_al_cruzar_el_año` |
| R07 | `LeerTipos` (SC-005) | quitar la regla 7 | la de R02 | las cuatro: `regla_7_·_el_aviso_de_un_periodo_no_es_del_motor`, `…_la_etiqueta_de_un_aviso_del_motor`, `…_dos_avisos_con_el_mismo_código` y `…_dos_avisos_con_la_misma_etiqueta` |
| R08 | `LeerTipos` (SC-005) | quitar la regla 8 | la de R02 | las tres: `regla_8_·_la_gramática_de_un_código_de_tipo`, `…_la_de_un_código_de_cómputo` y `…_la_de_un_código_de_aviso` |
| R09 | `LeerTipos` (SC-005) | quitar la regla 9 | la de R02 | las veintiuna de la regla, de `regla_9_·_uno_o_dos_municipios` a `regla_9_·_un_texto_de_dos_líneas` |
| R10 | `LeerTipos` (SC-005) | quitar la regla 10 | la de R02 | `regla_10_·_da`, `regla_10_·_uno_de_sus_dos_valores` y `las_reglas_1_y_10_a_la_vez` |
| M27 | Los preceptos del civil (SC-006) | el `vigente_desde` del 133.1 con el 2010-05-04 | `-run '^TestPreceptosDeLosPlazos$' ./internal/evals/` | `TestPreceptosDeLosPlazos`: «el cómputo civil: BOE-A-2000-323 a133, apartado 133.1: vigente_desde es 2010-05-04, y las redacciones grabadas dicen el apartado así desde el 2001-01-08» |
| M28 | Los preceptos del civil (SC-006) | el `vigente_desde` del 130.2 con el 2015-10-07 | la de M27 | `TestPreceptosDeLosPlazos`: «el cómputo civil: BOE-A-2000-323 a130, apartado 130.2: vigente_desde es 2015-10-07, y las redacciones grabadas dicen el apartado así desde el 2022-12-23» |
| M29 | La búsqueda grabada de la unión (desviación de D10) | `rutaDeLaBusqueda` de `internal/skills/normas_test.go` con `/datosabiertos/api/otra`, que deja la unión sin búsquedas | `-run '^TestEsquemaDeNormas$' ./internal/skills/` | `TestEsquemaDeNormas/rangos-grabados`, por el `require` de la unión: «los conjuntos de grabaciones […] tienen entre todos alguna búsqueda» |
| M30 | El aviso de más de la copia (SC-005) | la sustitución de `TestDescribeDePlazosSinDatos` sin su periodo `aviso` (fuera `+periodoAnadido` en `internal/app/plazos_describe_test.go`) | `-run '^TestDescribeDePlazosSinDatos$' ./internal/app/` | `TestDescribeDePlazosSinDatos/con-un-tipo-y-un-aviso-mas`, su `require.Equal`: «la copia da un código de aviso más que data/plazos.yaml, el añadido, en el lugar de su periodo…» |

Quitar una regla de `LeerTipos` (R01 a R10) es, en la copia, un fichero `internal/core/plazos/zz_mutante_regla.go` y
una guarda al principio de `defecto` de `internal/core/plazos/tipos.go` que no anota ningún defecto cuyo formato empieza
por los de esa regla en la tabla de contracts/datos-de-plazos.md §2 (en la 9, también «%s no tiene %s» cuando es de un
periodo o de un aviso; el mismo formato de un precepto es de H9 y se queda). Antes de aplicarlas se comprobó que cada
uno de los 33 formatos está en `tipos.go`. En cada regla cayeron todas sus subpruebas: 1, 1, 4, 3, 4, 2, 4, 3, 21 y 2,
las que tiene `TestLeerTipos`.

## Los controles de umbral que no están en la tabla

| # | Control | Mutante | Orden | En rojo |
|---|---|---|---|---|
| M31 | FR-009: el esquema de eval enumera los códigos de `CodigosDeAviso` | un periodo `aviso` más en `lpac` de `data/plazos.yaml`: del `03-01` al `03-02`, con el 30.2 y el aviso `aviso-de-prueba` | `-run '^TestFormatoDePlazos$' ./internal/evals/` | `TestFormatoDePlazos/enumerado-de-los-avisos`: «el esquema de eval no enumera el código de aviso aviso-de-prueba en plazo.avisos» |
| M32 | SC-007: las regiones de `skills/` y lo que se genera desde `data/` | el nombre de un tipo en `data/plazos.yaml`: «recurso potestativo de reposición» → «recurso de reposición potestativo» | `-run '^TestRegiones$' ./internal/skills/` | `TestRegiones/sincronia` |
| M33 | SC-009: `make ci` con 0 fallos | M32 | `make -C <copia> ci` | `make: *** [Makefile:97: test] Error 1`, tras `fmt-check` y `lint` (`0 issues.`) en verde; en `make test`, `TestRegiones/sincronia`, `TestRegionDeLoCubierto`, `TestRegionDeLaCobertura` y `TestRenderizarRecursosYPlazos` (`internal/skills`), `TestSkillsDelRepositorio` (`internal/app`) y `TestPreceptosDeLosPlazos` (`internal/evals`) |

Deshecho M32, los seis tests que cayeron en M33 salen en verde en la copia
(`-run '^(TestSkillsDelRepositorio|TestPreceptosDeLosPlazos|TestRenderizarRecursosYPlazos|TestRegionDeLoCubierto|TestRegionDeLaCobertura|TestRegiones)$'`
sobre `./internal/app/ ./internal/evals/ ./internal/skills/`: `ok` en los tres): el rojo es del mutante.

Cada fila `ci:` de «Controles de umbral» tiene así su mutante: SC-001 (M01 a M03), SC-002 (M08 a M12), SC-003 y FR-022
(M13 a M16), SC-003 y FR-010 (M05, M06), SC-004 (M07), SC-005 y FR-012 (M04, M30), SC-005 y FR-013 (R01 a R10), SC-006
(M27, M28), FR-009 (M31), SC-007 (M32) y SC-009 (M33). Ningún mutante dejó su control en verde: ningún control tiene un
defecto que arreglar.

## La comprobación del diff

En el árbol versionado, con la rama en `8bb44dbc` y `main` en `82d4ccdd`, que es también la base de la rama.

1. `git diff --name-only main -- specs/023-h9-internal-core-plazos/ skills/ evals/ data/festivos/ data/cobertura.yaml
   data/normas.yaml docs/ ':!docs/roadmap/in-progress/H36-el-motor-de-computo-descrito.md'
   ':!docs/roadmap/done/H36-el-motor-de-computo-descrito.md'`: vacío (FR-018, FR-020, FR-025).
2. `git diff -M --name-status main -- docs/roadmap/`: una línea, `R100
   docs/roadmap/in-progress/H36-el-motor-de-computo-descrito.md docs/roadmap/done/H36-el-motor-de-computo-descrito.md`.
3. `git log --format=%h --grep='^chore(H36): grabaciones de' main..HEAD` da `304df573`, y
   `git show --name-only --format= 304df573` sus seis ficheros: los tres grabados bajo
   `evidencias/h36/boe.legislacion-consolidada/` (`robots.txt` y los bloques `a130` y `a133` de `BOE-A-2000-323`),
   `evidencias/h36/manifiesto.json`, y `gates/grabaciones.log` y `gates/grabaciones.md` de este directorio. Nada fuera de
   eso. `git log main..HEAD -- evidencias/` solo da ese commit.
4. `git diff --name-only main`, fichero a fichero:

   | Fichero | Lo admite |
   |---|---|
   | `CHANGELOG.md` | el guardián («Format»); T012 |
   | `data/plazos.yaml` | T004, T009 |
   | `docs/roadmap/done/H36-el-motor-de-computo-descrito.md` | el renombrado del workflow (commit `3c80c3f6`) |
   | `evidencias/h36/**` (cuatro ficheros) | el paso `grabar_datos` (`304df573`) |
   | `internal/app/plazos.go` | T003 |
   | `internal/app/plazos_test.go` | T003, T008 (`x_test.go` de `plazos.go`) |
   | `internal/app/plazos_civil_test.go` | T010 |
   | `internal/app/plazos_codigos_test.go` | T008 |
   | `internal/app/plazos_describe_test.go` | T005, T008 |
   | `internal/app/testdata/plazos/describe.json`, `internal/app/testdata/script/h39-describe.txtar` | T003 |
   | `internal/core/plazos/avisos.go`, `avisos_test.go` | T007, T008 |
   | `internal/core/plazos/calcular.go`, `calcular_test.go` | T004, T007, T008 |
   | `internal/core/plazos/computo.go`, `errores.go`, `ensayo_test.go` | T007 |
   | `internal/core/plazos/doc.go` | T004 |
   | `internal/core/plazos/propiedades_test.go` | T008, T011 |
   | `internal/core/plazos/salida.go`, `salida_test.go` | T003, T008 |
   | `internal/core/plazos/testdata/computo-de-ensayo.yaml`, `tipos-de-prueba-civil.yaml` | T006 |
   | `internal/core/plazos/tipos.go`, `tipos_test.go` | T004, T007, T008 |
   | `internal/evals/formato_test.go`, `plazos.go`, `plazos_test.go` | T008 |
   | `internal/evals/grabacion_test.go`, `testdata/grabaciones.json` | T002 |
   | `internal/evals/grabaciones.go`, `grabaciones_test.go` | T009 |
   | `internal/skills/normas_test.go` | T009 |
   | `internal/skills/referencias_test.go` | T008 |
   | `internal/skills/regiones_test.go` | T004 |
   | `schemas/plazo.json`, `specs/031-h39-c-omo-se/aceptacion/describe.txtar` | T003 |
   | `specs/041-h36-el-motor-de/**` | el directorio del feature |
   | `.golangci.yml` | T004, que lo nombra en su línea: ver abajo |

   **`.golangci.yml`** lo cambió T004 (`902d8282`): una sola entrada, `periodos`, en `ignore-rules` de `misspell`, con su
   comentario. La línea de T004 nombra el fichero (por la regla `unused` que retirar `fijaSuDuracion` exige respetar) y
   el guardián la admitió, pero las convenciones de tasks.md dicen «No se tocan `.golangci.yml`» y plan.md («Controles
   mecánicos») dice que `make lint` pasa «con la configuración de hoy, que no cambia». Medido en la copia: con el
   `.golangci.yml` de `main`, `make lint` da `41 issues`, todos de `misspell` sobre `periodos` y `Periodos`
   (`make: *** [Makefile:113: lint] Error 1`). La clave `periodos` la fija spec §C.1 y la lleva el Go que la lee y lo
   escribe; el diccionario de `misspell` es solo inglés y el proyecto neutraliza el español con `ignore-rules`, una
   entrada por falso positivo (como `patrones` de H5 o `producto` de H41). Sin la entrada, `make ci` no puede estar en
   verde, y no cambia nada observable. Se toma como ruta declarada por T004 (`gates/supuestos.md`); queda aquí para la
   revisión final.

5. Cada test y cada fichero de datos de un test, cambio a cambio (`git diff main -- <fichero>`):
   - `internal/core/plazos/calcular_test.go`: solo constantes por su literal, palabra a palabra (`TipoGenerico` →
     `"generico"` 27 veces, `TipoContencioso` 20, `TipoReposicion` 15, `TipoSilencio` y `TipoAlzada` 6, `TipoAccesoResolver`
     2, `TipoAccesoReclamar` 1, `AvisoDiasDelArt183` 6); ninguna entrada ni fecha esperada (FR-017).
   - `internal/core/plazos/propiedades_test.go`: quita solo `TipoGenerico` (por `"generico"`) y reescribe un comentario;
     lo demás es la propiedad del civil de T011 (FR-024), sin tocar la rejilla ni las de H9.
   - `internal/core/plazos/salida_test.go`: fuera los enumerados de `tipo` y `codigo`, que ya no tienen, y
     `TipoReposicion` por `"reposicion"` (FR-017).
   - `internal/core/plazos/avisos_test.go`: las firmas de `CodigosDeAviso(tipos)` y `EtiquetaDeAviso(tipos, …)`; los
     cuatro códigos de hoy en su orden y sus etiquetas se exigen igual, y lo añadido es lo que promete la firma nueva
     (`gates/supuestos.md`, T008).
   - `internal/core/plazos/tipos_test.go`: `plazosDelContrato` solo gana líneas (la forma de §C.1 y el civil; ninguna de
     hoy se quita); las variantes que desaparecen son las de las reglas que dependen de un código («falta un código»,
     «un código de más», «falta un cómputo», «cómputo desconocido», «un cómputo desconocido», «un tipo que fija su
     duración sin ella», «duración en un tipo que no la fija», «unidades fuera de generico» y las cuatro de «generico»,
     que pasan a la regla 3); «clave que el fichero no declara» sigue como «clave que un tipo no declara», con la misma
     entrada y el mismo motivo; las de H9 que no dependen de un código siguen con su entrada y su motivo, ahora en
     `motivos`; «los vocabularios» se queda con `Unidades()` (T008); lo nuevo son los subtests de las diez reglas, la
     copia del contrato y el de las reglas 1 y 10 (FR-013).
   - `internal/app/plazos_test.go`: la ayuda del tipo deja de exigir los códigos y los enumerados de `tipo` y
     `avisos[].codigo` pasan al `pattern`, en `schemas/plazo.json` y en lo emitido (T003, FR-017).
   - `internal/skills/referencias_test.go`: `plazos.ComputoLPAC`, `plazos.ComputoContencioso` y `plazos.TipoGenerico` por
     su literal (FR-017).
   - `internal/evals/formato_test.go` y `internal/evals/plazos_test.go`: `CodigosDeAviso(tipos)` con los tipos de
     `data/plazos.yaml`, y comentarios (FR-017).
   - `internal/evals/grabaciones_test.go`: la lista esperada de `TestGrabacionesSinSolape` gana `GrabacionesDeH36` (D10);
     `verificarPreceptos` añade `lecturasDeLosComputosSinTipos` y `palabrasClaveDeLosPreceptos` gana los arts. 130 y 133
     LEC con sus palabras (FR-015, que FR-017 admite).
   - `internal/evals/grabacion_test.go`: solo la raíz del cliente, `KITLEGAL_EVIDENCIAS` cuando no está vacía, y sus
     comentarios (D10, T002).
   - `internal/skills/normas_test.go`: el `require.NotEmpty` pasa de cada conjunto a la unión, y su comentario (D10, T009).
   - `internal/skills/regiones_test.go`: solo `tipos: []` → `tipos: [{codigo: inventado}]` en `probarLosDatosQueNoSeLeen`
     (T004, `gates/supuestos.md`).
   - `internal/evals/testdata/grabaciones.json`: carácter a carácter, una coma y la entrada de la Ley 1/2000 (T002).
   - `internal/app/testdata/plazos/describe.json` y `schemas/plazo.json`: los dos `enum` por el `pattern` de §C.8 y las
     cinco ayudas de §C.7; nada más.
   - `internal/app/testdata/script/h39-describe.txtar` y `specs/031-h39-c-omo-se/aceptacion/describe.txtar`: las líneas 39
     a 43, las cinco ayudas de §C.7; los cambios de los dos son idénticos (mismo `md5sum` de sus líneas cambiadas).
   - Nuevos de «Trazabilidad»: `internal/app/plazos_describe_test.go` (`TestDescribeDePlazosSinDatos`),
     `internal/core/plazos/ensayo_test.go` (`TestComputoDeEnsayo`), `internal/app/plazos_codigos_test.go`
     (`TestNingunCodigoDePlazosEnGo`) e `internal/app/plazos_civil_test.go` (`TestComputoCivil`), con los ficheros de
     prueba de T006.

   Ningún otro cambio en un test, y en ninguno cambian las entradas ni lo esperado de un caso de cálculo de H9.
6. `make skills-check` en verde (`ok` en `internal/app`, `internal/skills` e `internal/evals`) y `make ci` en verde en el
   árbol (`ci: todos los controles en verde`), los dos en la sesión de T013.
