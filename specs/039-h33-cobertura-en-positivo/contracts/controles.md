# Contrato: los controles que añade o cambia el hito

FR-002, FR-003, FR-006, FR-007, FR-010, FR-011, FR-013 a FR-021, FR-023; SC-001 a SC-003, SC-005 a SC-009. Todos en
`make ci` (`make test` y `make test-integration`), offline. Cada mutante sale en rojo si el código no da su defecto o su
clasificación exactos (ADR 0029).

## 1. `internal/skills`

| Test | Fichero | Qué fija |
|---|---|---|
| `TestLeerCobertura` | `cobertura_test.go` | subtest `repositorio`: `data/cobertura.yaml` se lee sin defecto y su motivo y sus cuatro filas son, campo a campo, las de FR-001; los seis mutantes de contracts/datos-y-regiones.md §2; los casos de hoy de `regulados_aparte`, con sus mensajes |
| `TestRegionDeLoCubierto` (nuevo) | `cubierto_test.go` | sobre una copia de `consulta-juridica`, `plazos-y-territorio` y `data/`: la región da las 4 líneas en su orden y la sincronía no cambia nada (dos filas de una capacidad las prueba `TestVerticalDeEnsayo`); mutantes de FR-003, cada uno con los defectos exactos de cada skill de la copia: `materia: Laboral` (el de la materia, que da la sincronía y no el esquema, research D5), sin `data/cobertura.yaml` (en `consulta-juridica`, el de la región y el de la plantilla; en `plazos-y-territorio`, el de la plantilla y el de hoy de su región `cobertura_por_materia`; nada regenerado) y el motivo de §4 cambiado en la prosa (el del motivo) |
| `TestRegiones` | `regiones_test.go` | sus árboles sintéticos con la plantilla y los dos motivos; un mutante por defecto del generador (contracts/datos-y-regiones.md §5): la plantilla quitada, una segunda plantilla, un motivo quitado, un marcador desconocido en la plantilla y otro en un motivo; la fila sin región, como hoy |
| `TestRegionDeLaCobertura` | `referencias_test.go` | `filaDeLosTributos`, byte a byte la del spec (historia 3, escenario 1); 9 filas del paquete; la referencia del repositorio es la generada; su mutante, la plantilla cambiada sin regenerar, da la deriva |
| `TestVerticalDeEnsayo` | `vertical_de_ensayo_test.go` | los tres cambios de contracts/datos-y-regiones.md §6, con la fila de la vertical como segunda línea de los plazos, detrás de las cuatro del repositorio (dos filas de una capacidad, FR-007); cada cuerpo, como mucho 11 000 caracteres con la vertical (10 756 el de `plazos-y-territorio`) |

## 2. `internal/app`

- `TestSkillsDelRepositorio` (`skills_test.go`): `partirElCuerpo` no parte las líneas entre las marcas de lo cubierto
  (research D16). Sin más cambios, sus subtests fijan lo nuevo: `deriva` (0 ficheros que cambie `make skills-sync`),
  `presupuestos` (11 000 caracteres), `formas-compartidas` (la plantilla, la misma línea en las dos skills),
  `forma-de-cita` y `ordenes-de-ejemplo`.
- `testdata/script/h32-paquetes.txtar`: en `-- tributos.txt --`, la fila de la tabla con la línea del spec (historia 3,
  escenario 1). Nada más.
- `testdata/script/h32-capacidades.txtar`: en `-- formas.txt --`, `plazos-y-territorio|⚠ NO CUBIERTO:|1`; el
  comentario de cabecera, que dice «dos en ⚠ NO CUBIERTO:» con sus dos formas, pasa a decir una, la plantilla.
  `formas: 16 de 16` no cambia.
- El guion nuevo `h33-cobertura-instalada.txtar`, que activa el workflow desde `aceptacion/` (contracts/aceptacion.md).

## 3. `TestSkillsDelProductoDesdeControl` (`internal/evals/laboratorio_familias_test.go`)

- `declaracionDelProducto` gana `enLasReferencias []cambioEnUnaReferencia`, con `cambioEnUnaReferencia{referencia,
  antes, despues string}`; `conLaDeclaracion` aplica cada uno a `references/<referencia>` de la base sin lo generado,
  con la regla de los del cuerpo (el `antes`, una sola vez; si no, el error con su número y su referencia).
- `sinLoGenerado` reduce a sus marcas, además de las del `SKILL.md`, las regiones de cada referencia escrita a mano:
  desde una línea que empieza por `<!-- inicio de ` y termina en ` con make skills-sync, no editar -->` hasta la
  `<!-- fin de <generador> -->` de su generador; una región sin cerrar es el error de hoy, con la ruta.
- La declaración de `plazos-y-territorio` gana los tres cambios de contracts/skills.md §2 en `cambios` y el de §3 en
  `enLasReferencias` (`administrativo.md`: el `antes`, la línea en blanco y la marca de inicio de la región; el
  `despues`, la línea en blanco, la frase, su línea en blanco, el bloque, su línea en blanco y la marca).
- Los subtests de `normativa` y `jurisprudencia` no cambian: sus declaraciones no tienen `enLasReferencias`, y sus
  referencias con regiones, si las hay, son las de su base.

## 4. `TestConsultaJuridicaDesdeLaBase` (`internal/evals/descripciones_del_producto_test.go`)

- `huellasDeConsultaJuridicaEnLaBase[SKILL.md]`: la SHA-256 del `SKILL.md` de la entrega sin las líneas de la clave
  `description` (`descripcionYCuerpo`) y con sus dos regiones reducidas a sus marcas (`conLasRegionesGeneradasReducidas`),
  calculada por la tarea dos veces (research D14).
- `regionesGeneradas` (`laboratorio_familias_test.go`) gana `{inicio: "<!-- lo que cubre la base: ", fin: "<!-- fin de lo
  que cubre la base -->"}` y `{inicio: "<!-- paquetes de la base: ", fin: "<!-- fin de los paquetes de la base -->"}`.
- Los mensajes: `… falta: está en la entrega de H33`, `… sobra: no está en la entrega de H33` y `… difiere de la
  entrega de H33`. Los cinco casos de hoy dan su defecto exacto con esos textos; «solo la clave», ninguno.

## 5. `internal/evals`

| Test | Fichero | Qué fija |
|---|---|---|
| `TestFormatoDeLaCobertura` (nuevo) | `formato_test.go` | contracts/evals.md §1 |
| `TestFormatoDePlazos` | `formato_test.go` | sin cambios |
| `TestJuzgarLaDeclaracionDeNoCubierto` (nuevo) | `plazos_test.go` | contracts/evals.md §2 |
| `TestJuzgarPlazos`, `TestJuzgarLasLecturasSinServir`, `TestUmbralesDePlazosYTerritorio`, los de `informe_test.go` | sus ficheros | sus constantes con la plantilla; lo que comprueban, igual |
| `TestActivadaPor`, `TestSkillsNombradasEnTambienLaResuelven` (nuevos, los del prototipo) | `tambien_la_resuelven_test.go` | `activadaPor` con y sin la otra skill declarada; un nombre que no es una skill de `skills/` es un defecto del conjunto |
| `TestUmbralesDeLasSkills` | `informe_test.go` | subtest `consulta-juridica`: los cuatro de contracts/evals.md §3, `Decide: true`, delante de los cuatro `fuera_de_lo_grabado`; su comentario deja de decir «sin `sin_activar`» |
| `TestEvalsDelRepositorio` | `conjunto_test.go` | «formato» con las cuatro evals; el mutante de «no activación de cada capacidad» sobre la 02 y la 03; la regla nueva y su mutante sin la 03 (contracts/evals.md §4) |
| `TestCacheDeLasSesiones` | `preparar_test.go` | subtest `consulta-juridica` con la 04 (contracts/evals.md §6) |
| `TestDefinicionDelJob`, `TestMatrizDeLasPropuestas` | `definicion_test.go`, `eleccion_test.go` | contracts/evals.md §7 |

## 6. Lo que no cambia

`.golangci.yml`, `Makefile` y `.github/workflows/ci.yml`; ningún `//nolint`, `t.Skip` ni test desactivado. Las copias de
`specs/030-h32-*/aceptacion/` y las demás suites congeladas no se tocan: las cuatro de FR-020 son las únicas que cambian.
