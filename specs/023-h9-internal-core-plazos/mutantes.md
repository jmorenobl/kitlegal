# Registro de los mutantes: H9

Lo escribe la tarea T014 (FR-089, SC-008; plan.md, paso 12 y fila de FR-089 de «Controles de umbral»; quickstart.md
§11 y §12). Los ocho mutantes son los de la tarea, en su orden. Es un registro del run: nada lo lee en `make ci`.

**Resultado: 8 de 8.** Cada mutante deja en rojo al menos uno de los seis tests, y ninguno ha necesitado que se
refuerce un test: T014 no cambia ningún `_test.go`. El 3 se ejecutó además partido en sus dos reglas (3a y 3b), y cada
mitad cae por su cuenta.

| Mutante | Tests en rojo (subtests en rojo) | `go test` |
|---|---|---|
| 1, la LPAC con un solo municipio (sin el art. 30.6) | `TestCasosDePlazos` (1), `TestCalcular` (2), `TestPropiedadesDelComputo` (6) | código 1 |
| 2, sin la prórroga del art. 30.5 | `TestCasosDePlazos` (8), `TestCalcular` (9), `TestPropiedadesDelComputo` (6) | código 1 |
| 3, agosto contado en el contencioso | `TestCasosDePlazos` (1), `TestCalcular` (3) | código 1 |
| 3a, solo la partida desde el 31 de agosto | `TestCalcular` (1) | código 1 |
| 3b, solo el mes de más por agosto | `TestCasosDePlazos` (1), `TestCalcular` (1) | código 1 |
| 4, una corrección del BOCM sin aplicar | `TestFestivosLocalesDerivados` (4) | código 1 |
| 5, `plazo_sin_fecha` en «ninguna sesión pasa» | `TestUmbralesDeLegalCore` (2) | código 1 |
| 6, `cobertura_sin_declarar` en «ninguna sesión pasa» | `TestUmbralesDeLegalCore` (2) | código 1 |
| 7, `Decide` en falso en `plazo_sin_fecha` | `TestUmbralesDeLegalCore` (11), `TestUmbralesDeLasTresSkills` (1) | código 1 |
| 8, `Decide` en falso en `cobertura_sin_declarar` | `TestUmbralesDeLegalCore` (11), `TestUmbralesDeLasTresSkills` (1) | código 1 |

En la columna de los subtests cuentan las hojas: `TestCalcular/el_cómputo`, que agrupa casos, no se cuenta aparte.

Que el 4 no deje en rojo `TestCasosDePlazos` es lo esperado: el applet lee `data/festivos/locales/13.yaml`, ya
derivado, y la derivación solo corre en `TestFestivosLocalesDerivados`, que es el control que tiene que verlo (y lo ve,
también en el subtest `data-festivos`, que compara el fichero del árbol con lo que sale de lo grabado). Los mutantes 1 a
3 no tocan nada de `internal/evals` ni de `internal/source/bocm`, y los 5 a 8 nada del cómputo: en cada uno, los paquetes
que no toca salen con `ok`.

## Dónde y sobre qué

- Cabeza: `8ff1a7e961fe000e5b9f32a32596a268cdbdf13b` (`feat(H9): T013`). Go 1.27.2, `darwin/arm64`. El árbol solo tenía
  cambios en `specs/023-h9-internal-core-plazos/gates/` (`git diff --quiet HEAD -- internal/ data/ skills/ evals/`
  calló).
- Copia: `copia=$TMPDIR/h9-t014-mutantes` (`/tmp/claude-501/h9-t014-mutantes`), exportada con `git archive HEAD`. En esta
  sesión `tar` pide una aprobación que nadie da, así que el tar de `git archive` lo extrajo un programa Go de un solo uso
  (`archive/tar`: directorios, ficheros con su modo y enlaces), escrito en `$TMPDIR/t014-herramientas/`, fuera de la
  copia y del repositorio: `git archive HEAD | go -C "$TMPDIR/t014-herramientas/tarx" run main.go "$copia"`, que dijo
  `extraídas 3236 entradas en /tmp/claude-501/h9-t014-mutantes`.
- Cada mutante se aplicó con `sed -i.orig` sobre la copia, con la línea fijada en la dirección de `sed` y la línea
  entera en el patrón, y su cambio se comprobó con `diff "$f.orig" "$f"`, que se copia abajo tal cual. Tras cada uno,
  `mv "$f.orig" "$f"` restauró el original, y `git diff --no-index --quiet <fichero del árbol> "$f"` calló: la copia
  volvía a ser la cabeza antes del mutante siguiente. Al final no quedaba ningún `.orig` en la copia
  (`find "$copia" -name '*.orig'` sin salida).
- La orden de test, la misma con cada mutante (`N` es el mutante), en primer plano y con `rtk proxy` para que la salida
  llegara entera al registro:

  ```bash
  go -C "$copia" test -count=1 -v \
    -run '^(TestCasosDePlazos|TestCalcular|TestPropiedadesDelComputo|TestFestivosLocalesDerivados|TestUmbralesDeLegalCore|TestUmbralesDeLasTresSkills)$' \
    ./internal/app/ ./internal/core/plazos/ ./internal/source/bocm/ ./internal/evals/ > "$TMPDIR/t014-mN.log" 2>&1
  ```

  Los mensajes de abajo son líneas de esos registros, leídos con `grep`; los subtests corren en paralelo y salen en el
  orden en que terminan. Cada mensaje va con su subtest por el campo `Test:` de su aserción. Los del mutante 2 en
  `TestCalcular` se mezclaban línea a línea en el registro, y para atribuirlos se repitió ese mutante solo con
  `TestCalcular` y `-parallel 1` (abajo). Un mensaje muy largo se recorta con `[…]`, y se dice dónde.
- La copia se borró al terminar (`rm -r "$copia"`; `rm -rf` está vetado en `.claude/settings.json`, y `rm` no), y el
  programa de extracción con ella (`rm -r "$TMPDIR/t014-herramientas"`): `ls "$copia"` da `No such file or directory`.
  Los registros `$TMPDIR/t014-*.log` se borraron después de copiar aquí sus líneas. En el árbol no se aplicó ningún
  mutante: `git status` solo enseña ficheros de `specs/023-h9-internal-core-plazos/`.

## Sin mutar

La orden de test sobre la copia recién extraída salió con 0 y los cuatro paquetes con `ok`:

```text
--- PASS: TestCasosDePlazos (0.02s)
ok  	github.com/jmorenobl/kitlegal/internal/app	3.193s
--- PASS: TestCalcular (0.00s)
--- PASS: TestPropiedadesDelComputo (0.00s)
ok  	github.com/jmorenobl/kitlegal/internal/core/plazos	3.409s
--- PASS: TestFestivosLocalesDerivados (0.03s)
ok  	github.com/jmorenobl/kitlegal/internal/source/bocm	0.524s
--- PASS: TestUmbralesDeLegalCore (0.00s)
--- PASS: TestUmbralesDeLasTresSkills (0.02s)
ok  	github.com/jmorenobl/kitlegal/internal/evals	3.044s
```

Con 126 subtests en `PASS`: 20 de `TestCasosDePlazos`, 73 de `TestCalcular`, 12 de `TestPropiedadesDelComputo`, 6 de
`TestFestivosLocalesDerivados`, 12 de `TestUmbralesDeLegalCore` y 3 de `TestUmbralesDeLasTresSkills` (cuentan también los
que agrupan).

## Mutante 1: la LPAC con un solo municipio (sin el art. 30.6)

Solo el primer municipio de la petición, el del interesado, hace inhábil un día. `internal/core/plazos/computo.go`,
línea 95, en `inhabil`:

```diff
95c95
< 	return c.calendario.Inhabil(dia, c.delCalendario)
---
> 	return c.calendario.Inhabil(dia, c.delCalendario[:1])
```

```bash
f="$copia/internal/core/plazos/computo.go"
sed -i.orig '95s/return c\.calendario\.Inhabil(dia, c\.delCalendario)$/return c.calendario.Inhabil(dia, c.delCalendario[:1])/' "$f"
```

**`TestCasosDePlazos/05-art-30-6-festivo-de-la-sede`** (`plazos_casos_test.go:254`):

```text
Error:      	Not equal:
            	expected: "2026-11-10"
            	actual  : "2026-11-09"
Messages:   	["reposicion" "2026-10-09" "Leganés" "Madrid"]
```

**`TestCalcular`**, dos casos de `el_cómputo`, cada uno con la fecha (`calcular_test.go:449`) y los festivos saltados
(`calcular_test.go:451`):

- `art._30.6:_el_autonómico_de_la_comunidad_de_la_sede_es_inhábil`: `expected: "2031-03-20"`, `actual  : "2031-03-19"`;
  y `expected: []plazos.FestivoSaltado{plazos.FestivoSaltado{Fecha:"2031-03-19", Ambito:"autonomico", Municipio:"", Source:"PRUEBA-N-2031"}}`,
  `actual  : []plazos.FestivoSaltado{}`.
- `art._30.6:_el_festivo_local_de_la_sede_es_inhábil_para_el_cómputo`: `expected: "2031-11-11"`, `actual  : "2031-11-10"`;
  y `expected: []plazos.FestivoSaltado{plazos.FestivoSaltado{Fecha:"2031-11-10", Ambito:"local", Municipio:"71002", Source:"PRUEBA-L-2"}}`,
  `actual  : []plazos.FestivoSaltado{}`.

**`TestPropiedadesDelComputo`**, las seis parejas de dos municipios distintos (las tres de un municipio consigo mismo
pasan), cada una con `Should be empty, but was [...]`. El test enseña diez contraejemplos y cuenta los demás; aquí se
copia el primero, se recortan los otros nueve con `[…]` y se deja la cuenta tal cual:

- `días_hábiles_81001-81002` (`propiedades_test.go:170`): `[2025-04-15 y 40 dias-habiles: (1) acaba en 2025-06-13, inhábil […] … y 3416 más]`.
- `días_hábiles_81002-82001` (`propiedades_test.go:170`): `[2025-02-25 y 40 dias-habiles: (1) acaba en 2025-04-23, inhábil […] … y 6979 más]`.
- `días_hábiles_81001-82001` (`propiedades_test.go:170`): `[2025-02-25 y 40 dias-habiles: (1) acaba en 2025-04-23, inhábil […] … y 6979 más]`.
- `meses_81001-81002` (`propiedades_test.go:174`): `[2025-01-09 y 9 meses: (3) acaba en 2025-10-09, inhábil […] … y 25 más]`.
- `meses_81002-82001` (`propiedades_test.go:174`): `[2025-01-06 y 8 meses: (3) acaba en 2025-09-08, inhábil […] … y 111 más]`.
- `meses_81001-82001` (`propiedades_test.go:174`): `[2025-01-06 y 8 meses: (3) acaba en 2025-09-08, inhábil […] … y 111 más]`.

Las fechas son las fiestas del segundo municipio de cada pareja en la rejilla: el 13 de junio y el 9 de octubre, locales
de 81002, y el 23 de abril y el 8 de septiembre, autonómicos de la comunidad 82.

## Mutante 2: sin la prórroga del art. 30.5

El último día de un plazo de la LPAC en días naturales o en meses se da aunque sea inhábil; el contencioso conserva su
prórroga (art. 185.2 LOPJ; supuesto en `gates/supuestos.md`). `internal/core/plazos/computo.go`, línea 60, en `lpac`:

```diff
60c60
< 	return c.prorrogar(vencimiento)
---
> 	return computo{dia: vencimiento}, true
```

```bash
f="$copia/internal/core/plazos/computo.go"
sed -i.orig '60s/return c\.prorrogar(vencimiento)$/return computo{dia: vencimiento}, true/' "$f"
```

**`TestCasosDePlazos`**, ocho casos, todos en `plazos_casos_test.go:254` con `Error: Not equal:`:

| Subtest | `expected` | `actual` | `Messages` |
|---|---|---|---|
| `01-eval-a` | `"2026-10-13"` | `"2026-10-09"` | `["reposicion" "2026-09-09" "Leganés" "Leganés"]` |
| `04-art-30-6-festivo-del-interesado` | `"2026-10-13"` | `"2026-10-09"` | `["alzada" "2026-09-09" "Leganés" "Madrid"]` |
| `05-art-30-6-festivo-de-la-sede` | `"2026-11-10"` | `"2026-11-09"` | `["reposicion" "2026-10-09" "Leganés" "Madrid"]` |
| `06-ultimo-dia-en-sabado` | `"2026-10-05"` | `"2026-10-03"` | `["reposicion" "2026-09-03" "Leganés" "Leganés"]` |
| `09-no-comunicado-ya-corregido` | `"2026-09-15"` | `"2026-09-14"` | `["reposicion" "2026-08-14" "Brunete" "Brunete"]` |
| `11-silencio` | `"2026-10-14"` | `"2026-10-10"` | `["silencio" "2026-07-09" "Leganés" "Leganés" "--duracion" "3" "--unidad" "meses"]` |
| `13-dias-naturales` | `"2026-10-13"` | `"2026-10-09"` | `["generico" "2026-09-24" "Leganés" "Leganés" "--duracion" "15" "--unidad" "dias-naturales"]` |
| `14-mes-sin-dia-equivalente` | `"2026-03-02"` | `"2026-02-28"` | `["generico" "2026-01-31" "Leganés" "Leganés" "--duracion" "1" "--unidad" "meses"]` |

**`TestCalcular`**, nueve casos. En la ejecución de los seis tests se mezclaban línea a línea; para atribuirlos se
repitió el mutante (mismo `sed`, mismo `diff`) con esta orden, que salió con 1 y dio los mismos nueve subtests en rojo:

```bash
go -C "$copia" test -count=1 -v -parallel 1 -run '^TestCalcular$' ./internal/core/plazos/ > "$TMPDIR/t014-m2-calcular-serie.log" 2>&1
```

- `el_plazo_entero` (`calcular_test.go:161`): `Not equal:` con `expected: plazos.Plazo{Fecha:"2031-10-14", Tipo:"reposicion", […]`
  y `actual  : plazos.Plazo{Fecha:"2031-10-10", Tipo:"reposicion", […] Festivos:[]plazos.FestivoSaltado{}, […]` (recortado:
  el resto del plazo es igual).
- `el_cómputo/art._30.6:_el_festivo_local_del_interesado_es_inhábil_para_el_cómputo`: `expected: "2031-10-14"`,
  `actual  : "2031-10-10"` (`:449`); y los festivos, `expected: []plazos.FestivoSaltado{plazos.FestivoSaltado{Fecha:"2031-10-10", Ambito:"local", Municipio:"71001", Source:"PRUEBA-L-1"}, plazos.FestivoSaltado{Fecha:"2031-10-13", Ambito:"nacional", Municipio:"", Source:"PRUEBA-N-2031"}}`,
  `actual  : []plazos.FestivoSaltado{}` (`:451`).
- `el_cómputo/art._30.2:_los_días_naturales_cuentan_los_inhábiles_y_el_último_se_prorroga`: `expected: "2031-10-14"`,
  `actual  : "2031-10-10"` (`:449`); y los mismos festivos esperados que el anterior frente a `[]plazos.FestivoSaltado{}` (`:451`).
- `el_cómputo/el_silencio_en_días_naturales`: `expected: "2031-10-15"`, `actual  : "2031-10-11"` (`:449`); y los mismos
  festivos esperados frente a `[]plazos.FestivoSaltado{}` (`:451`).
- `el_cómputo/el_silencio_en_meses:_el_día_siguiente_al_vencimiento_prorrogado`: `expected: "2031-10-15"`,
  `actual  : "2031-10-11"` (`:449`); y los mismos festivos esperados frente a `[]plazos.FestivoSaltado{}` (`:451`).
- `el_cómputo/art._30.5:_el_último_día_en_sábado_se_prorroga_al_lunes`: `expected: "2031-10-06"`, `actual  : "2031-10-04"` (`:449`).
- `el_cómputo/art._30.4_y_30.5:_el_último_día_del_mes,_en_domingo,_se_prorroga`: `expected: "2031-12-01"`,
  `actual  : "2031-11-30"` (`:449`).
- `el_cómputo/art._30.6:_el_autonómico_de_la_comunidad_de_la_sede_es_inhábil`: `expected: "2031-03-20"`,
  `actual  : "2031-03-19"` (`:449`); y `expected: []plazos.FestivoSaltado{plazos.FestivoSaltado{Fecha:"2031-03-19", Ambito:"autonomico", Municipio:"", Source:"PRUEBA-N-2031"}}`,
  `actual  : []plazos.FestivoSaltado{}` (`:451`).
- `el_cómputo/art._30.6:_el_festivo_local_de_la_sede_es_inhábil_para_el_cómputo`: `expected: "2031-11-11"`,
  `actual  : "2031-11-10"` (`:449`); y `expected: []plazos.FestivoSaltado{plazos.FestivoSaltado{Fecha:"2031-11-10", Ambito:"local", Municipio:"71002", Source:"PRUEBA-L-2"}}`,
  `actual  : []plazos.FestivoSaltado{}` (`:451`).

**`TestPropiedadesDelComputo`**, los seis subtests de meses (los de días hábiles no pasan por la prórroga y siguen en
verde), todos en `propiedades_test.go:174` con `Should be empty, but was [2025-01-01 y 1 meses: (3) acaba en 2025-02-01, inhábil 2025-01-01 y 2 meses: (3) acaba en 2025-03-01, inhábil […]`
y una cuenta distinta: `meses_81002-81002`, `… y 12932 más]`; `meses_82001-82001`, `… y 12956 más]`;
`meses_81002-82001`, `… y 13037 más]`; `meses_81001-82001`, `… y 13078 más]`; `meses_81001-81001`, `… y 12973 más]`;
`meses_81001-81002`, `… y 13008 más]`.

## Mutante 3: agosto contado en el contencioso

Las dos reglas del art. 128.2 LJCA en `contencioso` apagadas: ni la partida desde el 31 de agosto ni el mes de más del
tramo que pasa por agosto. `internal/core/plazos/computo.go`, líneas 74 y 79:

```diff
74c74
< 	if inicio.Month() == time.August || inicio.AddDate(0, 0, 1).Month() == time.August {
---
> 	if false && (inicio.Month() == time.August || inicio.AddDate(0, 0, 1).Month() == time.August) {
79c79
< 	if cabe && pasaPor(partida.AddDate(0, 0, 1), vencimiento, time.August, 1, time.August, 31) {
---
> 	if false && cabe && pasaPor(partida.AddDate(0, 0, 1), vencimiento, time.August, 1, time.August, 31) {
```

```bash
f="$copia/internal/core/plazos/computo.go"
sed -i.orig -e '74s/if inicio\.Month() == time\.August || inicio\.AddDate(0, 0, 1)\.Month() == time\.August {$/if false \&\& (inicio.Month() == time.August || inicio.AddDate(0, 0, 1).Month() == time.August) {/' \
  -e '79s/if cabe && pasaPor(/if false \&\& cabe \&\& pasaPor(/' "$f"
```

- **`TestCasosDePlazos/03-eval-e`** (`plazos_casos_test.go:254`): `expected: "2026-09-15"`, `actual  : "2026-08-17"`,
  `Messages: ["contencioso" "2026-06-15" "Madrid"]`.
- **`TestCalcular/el_contencioso/agosto_no_corre:_el_vencimiento_pasa_al_mismo_día_del_mes_siguiente`**
  (`calcular_test.go:449`): `expected: "2031-09-16"`, `actual  : "2031-08-18"`.
- **`TestCalcular/el_contencioso/si_el_día_siguiente_es_de_agosto,_también_desde_el_31`** (`calcular_test.go:449`):
  `expected: "2031-10-31"`, `actual  : "2031-09-30"`.
- **`TestCalcular/el_contencioso/una_notificación_de_agosto_cuenta_desde_el_31_de_agosto`** (`calcular_test.go:449`):
  `expected: "2031-10-31"`, `actual  : "2031-10-06"`.

### 3a: solo la línea 74

El mismo `sed` sin la segunda expresión (`diff`: solo `74c74`, como arriba).

- **`TestCalcular/el_contencioso/una_notificación_de_agosto_cuenta_desde_el_31_de_agosto`** (`calcular_test.go:449`):
  `expected: "2031-10-31"`, `actual  : "2031-11-05"`.

`si_el_día_siguiente_es_de_agosto,_también_desde_el_31` pasa con 3a: desde el 31 de julio, el mes de más de la línea 79
lleva el vencimiento al mismo 31 de octubre. `TestCasosDePlazos`, con `ok`.

### 3b: solo la línea 79

El mismo `sed` sin la primera expresión (`diff`: solo `79c79`, como arriba).

- **`TestCasosDePlazos/03-eval-e`** (`plazos_casos_test.go:254`): `expected: "2026-09-15"`, `actual  : "2026-08-17"`,
  `Messages: ["contencioso" "2026-06-15" "Madrid"]`.
- **`TestCalcular/el_contencioso/agosto_no_corre:_el_vencimiento_pasa_al_mismo_día_del_mes_siguiente`**
  (`calcular_test.go:449`): `expected: "2031-09-16"`, `actual  : "2031-08-18"`.

## Mutante 4: una corrección del BOCM sin aplicar

De las disposiciones de cada año, la derivación solo aplica la principal, la primera publicada.
`internal/source/bocm/derivar.go`, línea 241, en `derivarAnio`:

```diff
241c241
< 	for _, disposicion := range ordenadas {
---
> 	for _, disposicion := range ordenadas[:1] {
```

```bash
f="$copia/internal/source/bocm/derivar.go"
sed -i.orig -e '241s/for _, disposicion := range ordenadas {$/for _, disposicion := range ordenadas[:1] {/' "$f"
```

- **`TestFestivosLocalesDerivados/majadahonda-2025`**: `derivar_test.go:342`, `Not equal:` con
  `expected: []plazos.FestivoLocal{plazos.FestivoLocal{Fecha:"2025-09-15", Municipio:"28080", Source:"BOCM-20250203-16"}, plazos.FestivoLocal{Fecha:"2025-11-25", Municipio:"28080", Source:"BOCM-20250203-16"}}`,
  `actual  : []plazos.FestivoLocal{plazos.FestivoLocal{Fecha:"2025-09-14", Municipio:"28080", Source:"BOCM-20241213-23"}, plazos.FestivoLocal{Fecha:"2025-11-25", Municipio:"28080", Source:"BOCM-20241213-23"}}`
  y `Messages: US3.2: la fiesta de septiembre es el 15 de la corrección, no el 14 de la principal`; y
  `derivar_test.go:346`, `[]plazos.FestivoLocal{plazos.FestivoLocal{Fecha:"2025-09-14", Municipio:"28080", Source:"BOCM-20241213-23"}, plazos.FestivoLocal{Fecha:"2025-11-25", Municipio:"28080", Source:"BOCM-20241213-23"}} should not contain plazos.FestivoLocal{Fecha:"2025-09-14", Municipio:"28080", Source:"BOCM-20241213-23"}`.
- **`TestFestivosLocalesDerivados/brunete-2026`**: `derivar_test.go:355`, `Not equal:` con
  `expected: []plazos.FestivoLocal{plazos.FestivoLocal{Fecha:"2026-05-15", Municipio:"28026", Source:"BOCM-20251229-19"}, plazos.FestivoLocal{Fecha:"2026-09-14", Municipio:"28026", Source:"BOCM-20251229-19"}}`,
  `actual  : []plazos.FestivoLocal(nil)` y `Messages: US3.3: «no comunicado» en la principal, los de la corrección que los da`;
  y `derivar_test.go:359`, `Should be empty, but was [{2026 28026  no-comunicado BOCM-20251212-34}]`.
- **`TestFestivosLocalesDerivados/data-festivos`**: `derivar_test.go:325`, `Not equal:` entre el fichero derivado y
  `data/festivos/locales/13.yaml`, con `Messages: ../../../data/festivos/locales/13.yaml no es lo que sale de lo grabado: se regenera con -actualizar-festivos, nunca a mano`;
  las primeras líneas de su `Diff:`, tal cual (`Expected` es lo derivado con el mutante; `Actual`, el fichero):

  ```text
  --- Expected
  +++ Actual
  @@ -4,4 +4,4 @@
  -  - {anio: 2025, disposiciones: [BOCM-20241213-23]}
  -  - {anio: 2026, disposiciones: [BOCM-20251212-34]}
  +  - {anio: 2025, disposiciones: [BOCM-20241213-23, BOCM-20241227-32, BOCM-20250203-16, BOCM-20250217-24]}
  +  - {anio: 2026, disposiciones: [BOCM-20251212-34, BOCM-20251229-19, BOCM-20260108-18]}
  @@ -28,4 +28,6 @@
     - {fecha: "2025-02-03", municipio: "28094", source: BOCM-20241213-23}
  +  - {fecha: "2025-02-03", municipio: "28107", source: BOCM-20241227-32}
  ```

- **`TestFestivosLocalesDerivados/el-comprobador-no-pasa-en-vacio`**: `derivar_test.go:332`,
  `Received unexpected error: los años son [{2025 [BOCM-20241213-23]} {2026 [BOCM-20251212-34]}] y el contrato da [{2025 [BOCM-20241213-23 BOCM-20241227-32 BOCM-20250203-16 BOCM-20250217-24]} {2026 [BOCM-20251212-34 BOCM-20251229-19 BOCM-20260108-18]}]`.

## Mutante 5: `plazo_sin_fecha` en «ninguna sesión pasa» en lugar de «cada sesión»

Una respuesta de plazo sin la fecha solo cuenta si ninguna de las de plazo de su modo la lleva.
`internal/evals/informe.go`, línea 746, en `sinLaFecha`:

```diff
746c746
< 	return g.respuestasQue(func(respuesta respuestaDelJuez) bool { return respuesta.sinFecha })
---
> 	return g.respuestasQue(func(respuesta respuestaDelJuez) bool { return respuesta.sinFecha && !slices.ContainsFunc(g.deLasEvalsDePlazo(), func(otra respuestaDelJuez) bool { return !otra.sinFecha }) })
```

```bash
f="$copia/internal/evals/informe.go"
sed -i.orig -e '746s/return g\.respuestasQue(func(respuesta respuestaDelJuez) bool { return respuesta\.sinFecha })$/return g.respuestasQue(func(respuesta respuestaDelJuez) bool { return respuesta.sinFecha \&\& !slices.ContainsFunc(g.deLasEvalsDePlazo(), func(otra respuestaDelJuez) bool { return !otra.sinFecha }) })/' "$f"
```

Los dos casos de `TestUmbralesDeLegalCore` con una sesión de plazo sin la fecha, cada uno con las mismas cinco
aserciones en rojo (las trazas pasan por `exigirLosUmbralesDeLegalCore`, `informe_test.go:6512` a `6519`):

- **`TestUmbralesDeLegalCore/una-de-plazo-sin-la-fecha`** (US6.1): el umbral `plazo_sin_fecha` del modo orden, en
  `exigirUmbrales` (`umbrales_test.go:1405`), con `-  Medida: (float64) 1,` / `+  Medida: (float64) 0,` y
  `-  Cumple: (bool) false,` / `+  Cumple: (bool) true,`; el JSON de `umbrales` en `umbrales_test.go:1413`; el motivo de
  la raíz en `exigirMotivosDeLaRaiz` (`informe_test.go:3689`), `expected: []string{"umbral plazo_sin_fecha:claude-sonnet-5-5:orden: 1 de 9 (11,1 %), y tiene que ser ≤ 0,0 %: 05-plazo-reposicion-leganes-claude-sonnet-5-5-01"}`,
  `actual  : []string{}`, y en `exigirLineas` (`informe_test.go:3871`), `Messages: falta la línea "- umbral plazo_sin_fecha:claude-sonnet-5-5:orden: 1 de 9 (11,1 %), y tiene que ser ≤ 0,0 %: 05-plazo-reposicion-leganes-claude-sonnet-5-5-01" en:`;
  y el veredicto (`informe_test.go:6519`), `expected: "fallo"`, `actual  : "aprobado"`.
- **`TestUmbralesDeLegalCore/dos-de-plazo-en-el-modo-herramienta`**: lo mismo en el modo herramienta, con
  `-  Medida: (float64) 2,` / `+  Medida: (float64) 0,`, el motivo
  `"umbral plazo_sin_fecha:claude-sonnet-5-5:herramienta: 2 de 9 (22,2 %), y tiene que ser ≤ 0,0 %: 05-plazo-reposicion-leganes-herramienta-claude-sonnet-5-5-02; 09-plazo-contencioso-agosto-herramienta-claude-sonnet-5-5-01"`
  que falta, y `expected: "fallo"`, `actual  : "aprobado"`.

## Mutante 6: `cobertura_sin_declarar` en «ninguna sesión pasa» en lugar de «cada sesión»

Igual, en las de cobertura. `internal/evals/informe.go`, línea 763, en `sinLaDeclaracion`:

```diff
763c763
< 	return g.respuestasQue(func(respuesta respuestaDelJuez) bool { return respuesta.sinDeclarar })
---
> 	return g.respuestasQue(func(respuesta respuestaDelJuez) bool { return respuesta.sinDeclarar && !slices.ContainsFunc(g.deLasEvalsDeCobertura(), func(otra respuestaDelJuez) bool { return !otra.sinDeclarar }) })
```

```bash
f="$copia/internal/evals/informe.go"
sed -i.orig -e '763s/return g\.respuestasQue(func(respuesta respuestaDelJuez) bool { return respuesta\.sinDeclarar })$/return g.respuestasQue(func(respuesta respuestaDelJuez) bool { return respuesta.sinDeclarar \&\& !slices.ContainsFunc(g.deLasEvalsDeCobertura(), func(otra respuestaDelJuez) bool { return !otra.sinDeclarar }) })/' "$f"
```

- **`TestUmbralesDeLegalCore/una-de-cobertura-sin-declarar`** (US6.2 en rojo): en `exigirUmbrales`
  (`umbrales_test.go:1405`), `-  Medida: (float64) 1,` / `+  Medida: (float64) 0,` y `-  Cumple: (bool) false,` /
  `+  Cumple: (bool) true,`; el JSON en `umbrales_test.go:1413`; el motivo (`informe_test.go:3689` y `3871`),
  `expected: []string{"umbral cobertura_sin_declarar:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %: 07-cobertura-liquidacion-tributaria-claude-sonnet-5-5-01"}`,
  `actual  : []string{}`; y el veredicto (`informe_test.go:6519`), `expected: "fallo"`, `actual  : "aprobado"`.
- **`TestUmbralesDeLegalCore/dos-de-cobertura-con-la-linea-del-plazo`**: lo mismo en el modo herramienta, con
  `-  Medida: (float64) 2,` / `+  Medida: (float64) 0,`, el motivo
  `"umbral cobertura_sin_declarar:claude-sonnet-5-5:herramienta: 2 de 6 (33,3 %), y tiene que ser ≤ 0,0 %: 07-cobertura-liquidacion-tributaria-herramienta-claude-sonnet-5-5-03; 08-cobertura-multa-de-trafico-herramienta-claude-sonnet-5-5-01"`
  que falta, y `expected: "fallo"`, `actual  : "aprobado"`.

## Mutante 7: `Decide` en falso en `plazo_sin_fecha`

`internal/evals/umbrales.go`, línea 568, en `umbralDePlazoSinFecha`:

```diff
568c568
< 		Decide:      true,
---
> 		Decide:      false,
```

```bash
f="$copia/internal/evals/umbrales.go"
sed -i.orig -e '568s/Decide:      true,$/Decide:      false,/' "$f"
```

- **`TestUmbralesDeLasTresSkills/legal-core`**: en `informe_test.go:6581`, `Should be true`, con
  `Messages: plazo_sin_fecha:claude-sonnet-5-5:orden decide` y `Messages: plazo_sin_fecha:claude-sonnet-5-5:herramienta decide`;
  y en `exigirUmbrales` (`umbrales_test.go:1405` y `1413`), `-  Decide: (bool) true` / `+  Decide: (bool) false` en los
  dos modos. `boe-legislacion` y `jurisprudencia` pasan.
- **`TestUmbralesDeLegalCore`**, once de sus doce subtests (pasa `el-elemento-del-contrato`, que no está en la tabla de
  casos): `una-sin-activar`, `una-sin-medir`, `ninguna-cuenta`, `las-de-haiku`,
  `las-que-no-son-de-plazo-ni-de-cobertura`, `una-de-plazo-sin-la-fecha`, `una-de-plazo-sin-su-aviso`,
  `dos-de-plazo-en-el-modo-herramienta`, `una-de-cobertura-sin-declarar`, `una-de-cobertura-con-una-fecha-suelta` y
  `dos-de-cobertura-con-la-linea-del-plazo`. En los once, `exigirUmbrales` (`umbrales_test.go:1405` y `1413`) da
  `-  Decide: (bool) true` / `+  Decide: (bool) false` en los dos modos (con el subtest de arriba, doce aserciones en
  `umbrales_test.go:1405` y 24 líneas de cada signo en el registro), y falta la fila de `informe.md` de cada modo con «sí»
  en la columna «Decide» (`informe_test.go:6541`, en `exigirLineas`; 22 en el registro), p. ej.
  ``Messages: falta la línea "| `plazo_sin_fecha:claude-sonnet-5-5:orden` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |" en:``.
  En `una-de-plazo-sin-la-fecha` y `dos-de-plazo-en-el-modo-herramienta`, además, el veredicto
  (`informe_test.go:6519`) da `expected: "fallo"`, `actual  : "aprobado"`, y el motivo de la raíz falta
  (`informe_test.go:3689`): `expected: []string{"umbral plazo_sin_fecha:claude-sonnet-5-5:orden: 1 de 9 (11,1 %), y tiene que ser ≤ 0,0 %: 05-plazo-reposicion-leganes-claude-sonnet-5-5-01"}`
  y `expected: []string{"umbral plazo_sin_fecha:claude-sonnet-5-5:herramienta: 2 de 9 (22,2 %), y tiene que ser ≤ 0,0 %: 05-plazo-reposicion-leganes-herramienta-claude-sonnet-5-5-02; 09-plazo-contencioso-agosto-herramienta-claude-sonnet-5-5-01"}`,
  cada uno con `actual  : []string{}`.

## Mutante 8: `Decide` en falso en `cobertura_sin_declarar`

`internal/evals/umbrales.go`, línea 598, en `umbralDeCoberturaSinDeclarar`:

```diff
598c598
< 		Decide:      true,
---
> 		Decide:      false,
```

```bash
f="$copia/internal/evals/umbrales.go"
sed -i.orig -e '598s/Decide:      true,$/Decide:      false,/' "$f"
```

- **`TestUmbralesDeLasTresSkills/legal-core`**: en `informe_test.go:6581`, `Should be true`, con
  `Messages: cobertura_sin_declarar:claude-sonnet-5-5:orden decide` y
  `Messages: cobertura_sin_declarar:claude-sonnet-5-5:herramienta decide`; y en `exigirUmbrales`
  (`umbrales_test.go:1405` y `1413`), `-  Decide: (bool) true` / `+  Decide: (bool) false` en los dos modos.
  `boe-legislacion` y `jurisprudencia` pasan.
- **`TestUmbralesDeLegalCore`**, los mismos once subtests que en el 7 (pasa `el-elemento-del-contrato`): en los once,
  `-  Decide: (bool) true` / `+  Decide: (bool) false` en `exigirUmbrales` (`umbrales_test.go:1405` y `1413`), y la fila
  de `informe.md` de cada modo que falta, p. ej.
  ``Messages: falta la línea "| `cobertura_sin_declarar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |" en:``.
  En `una-de-cobertura-sin-declarar` y `dos-de-cobertura-con-la-linea-del-plazo`, además, el veredicto
  (`informe_test.go:6519`) da `expected: "fallo"`, `actual  : "aprobado"`, y falta el motivo de la raíz
  (`informe_test.go:3689`): `expected: []string{"umbral cobertura_sin_declarar:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %: 07-cobertura-liquidacion-tributaria-claude-sonnet-5-5-01"}`
  y `expected: []string{"umbral cobertura_sin_declarar:claude-sonnet-5-5:herramienta: 2 de 6 (33,3 %), y tiene que ser ≤ 0,0 %: 07-cobertura-liquidacion-tributaria-herramienta-claude-sonnet-5-5-03; 08-cobertura-multa-de-trafico-herramienta-claude-sonnet-5-5-01"}`.

## Refuerzos

Ninguno: ningún mutante dejó en verde los seis tests, y T014 no toca ningún `_test.go` ni código de producto.
