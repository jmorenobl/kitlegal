# Quickstart: comprobar H36

Escenarios que se ejecutan tal cual desde la raíz del repositorio, sin red y sin modelo. Ninguno escribe en el árbol, el
índice ni el historial de git salvo donde se dice; los que construyen algo lo hacen en `$TMPDIR` o en la caché de Go.
Los que nombran un test que entrega el hito (marcados **H36**) solo existen en la rama ya implementada; los demás se
han ejecutado en la sesión del plan con el código de hoy (research V18).

## 0. Precondición de la captura de la tarea `[aceptacion]`

```bash
git diff --name-only cd3ad148 HEAD | grep -v '^docs/' | grep -v '^specs/' ; echo "fin"
```

Antes de la primera tarea de producto imprime solo `fin`: el árbol es el código de `main` (`cd3ad148`) y su binario es
el que captura los goldens (contracts/aceptacion.md §1). Solo lee.

## 1. Con los datos de hoy, lo mismo que `main` (US1; SC-001)

```bash
go test -count=1 -run '^TestEntregaDelHito$/^h36-plazos-de-hoy$' ./internal/app/      # H36, tras la activación
go test -count=1 -run '^TestEntregaDelHito$/^h9-plazos' ./internal/app/
go test -count=1 -run '^(TestCasosDePlazos|TestAppletPlazos)$' ./internal/app/
go test -count=1 -run '^(TestCalcular|TestComprobarPeticion)$' ./internal/core/plazos/
```

En verde: cada sobre capturado con `main` sale igual salvo `fecha_consulta`, los guiones de H9 y sus copias pasan sin
cambiar, `caso-a.json`, `caso-b.json` y `caso-e.json` son los mismos (subtest `los-golden-de-los-casos`) y la tabla de
casos da sus fechas y avisos. Antes de activarse, el guion vive en `specs/041-h36-el-motor-de/aceptacion/`.

## 2. El cómputo civil (US2; SC-002)

```bash
go test -count=1 -run '^TestComputoCivil$' ./internal/app/                                # H36
```

En verde, C1 a C10 de spec §C.9 y ningún tipo de `data/plazos.yaml` con el civil (US2.5).

## 3. El cómputo de ensayo y ningún código en Go (US3; SC-003)

```bash
go test -count=1 -run '^TestComputoDeEnsayo$' ./internal/core/plazos/                     # H36
go test -count=1 -run '^TestNingunCodigoDePlazosEnGo$' ./internal/app/                    # H36
```

En verde, E1 a E13 de spec §C.10, y ningún literal ni etiqueta de Go con un código de `data/plazos.yaml` o del fichero de
ensayo.

## 4. Las propiedades (SC-004)

```bash
go test -count=1 -run '^TestPropiedadesDelComputo$' ./internal/core/plazos/
```

Hoy, las de H9; en la rama, también las del civil, con un municipio cada vez.

## 5. El esquema de `calcular` sin datos y `LeerTipos` (US4, US5; SC-005)

```bash
go test -count=1 -run '^TestDescribeDePlazosSinDatos$' ./internal/app/                    # H36
go test -count=1 -run '^TestEntregaDelHito$/^h36-describe$' ./internal/app/               # H36, tras la activación
go test -count=1 -run '^TestEsquemasPublicados$' ./internal/app/
go test -count=1 -run '^TestLeerTipos$' ./internal/core/plazos/
```

En verde: `--describe` igual con los tres juegos de datos y a `describe.json`; `schemas/plazo.json` igual a lo que
emite; un subtest por regla de §C.2, el de dos defectos y el de la copia del contrato.

## 6. Los preceptos del civil contra lo grabado (US6; SC-006)

```bash
go test -count=1 -run '^(TestPreceptosDeLosPlazos|TestGrabacionesSinSolape|TestManifiestoDeGrabaciones)$' ./internal/evals/
go test -count=1 -run '^TestEsquemaDeNormas$' ./internal/skills/
```

En la rama, con `evidencias/h36/boe.legislacion-consolidada/` grabado por `grabar_datos`, pasa con los cinco preceptos
del civil y la unión de cuatro conjuntos, y `rangos-grabados` con la búsqueda exigida a la unión, no al conjunto de
H36, que no tiene ninguna (research D10). Solo lee.

## 7. Las referencias no cambian (SC-007)

```bash
go test -count=1 -run '^TestRegiones$' ./internal/skills/
```

Es la comprobación de `make ci` (`make skills-check` no escribe nada). Quien quiera regenerar y mirar: `make skills-sync
&& git diff --exit-code -- skills/`, que reescribe `skills/*/references/` con los mismos bytes y sale con 0 si no hay
diferencias; escribe en el árbol, aunque no lo cambie.

## 8. Todo

```bash
make ci
```

El veredicto del repositorio (`Makefile`, objetivo `ci`): no modifica ningún fichero versionado y pasa con los tests de
H9 que nombra FR-017 cambiados, más las tres desviaciones de research D10 (la raíz de `TestGrabarEvals`, la lista de
`TestGrabacionesSinSolape` y la búsqueda de la unión en `rangosDeLasBusquedasGrabadas`) y la entrada de
`probarLosDatosQueNoSeLeen` (`gates/supuestos.md`), y los demás intactos (SC-009). Deja `coverage.out`, que no está versionado.
