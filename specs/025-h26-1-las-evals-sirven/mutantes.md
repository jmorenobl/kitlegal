# Mutantes de H26.1

Cada mutante se aplica en una copia fuera del árbol, uno cada vez, y su fichero se restaura antes del siguiente
(tasks.md, «Batería de verificación por tarea»; quickstart.md §7; research D16). Ninguno queda en el diff: esta tarea
no cambia nada bajo `internal/` ni bajo `data/`.

```bash
copia="$TMPDIR/h261-copia"
git worktree add --detach "$copia" HEAD
# … cada mutante, con su comprobación de que se aplicó y su fichero restaurado tras él …
git worktree remove --force "$copia"
git worktree list
git status --short
```

Lo que dice cada sección «Visto en T015» se ejecutó en la sesión de T015 (2026-10-21) sobre una copia de la cabeza de
la rama, `e3f8f58` (T014), con las órdenes de cada sección. El sandbox de la sesión no deja comprobar una orden con una
variable antes de ejecutarla, así que `$copia` y `$f` se escribieron con la ruta entera de `$TMPDIR/h261-copia`; y la
salida de `git diff` llega resumida, así que la línea cambiada se miró con `rtk proxy git diff --no-index -U0` (o
`--stat`). Una sustitución sencilla se hizo con `sed -i.orig`; la de un rango de líneas (`sed '/^func …/,/^}/ s/…/'`)
pide aprobación en el sandbox, y esa y las de más de una línea se hicieron con el editor sobre una copia `.orig` hecha
con `cp`.

## Sin mutantes

```bash
go -C "$copia" test -count=1 -v \
  -run '^(TestEvalsDelRepositorio|TestCacheDeLasSesiones|TestUmbralesDeLasLecturas|TestUmbralesDeLasTresSkills|TestLineaDeLaSalidaNoEscrita)$' \
  ./internal/evals/ 2>&1 | grep -E '^--- |^(ok|FAIL)'
go -C "$copia" test -count=1 -v -run '^TestTuberiaCerrada$' ./internal/app/ 2>&1 | grep -E '^--- |^(ok|FAIL)'
```

Visto en T015, antes del primer mutante: los cinco de `internal/evals` en `PASS` y `ok`, con los tres subtests de
`TestEvalsDelRepositorio/lo-que-lleva-a-leer` (`boe-legislacion`, `legal-core`, `jurisprudencia`) y los cuatro de
`TestCacheDeLasSesiones/sin-servir` (05, 07, 11 y 12) en `PASS`; `TestTuberiaCerrada` en `PASS` y `ok`. Tras el último,
con todos los ficheros restaurados, `ok` en los dos paquetes con
`-run '^(TestEvalsDelRepositorio|TestCacheDeLasSesiones|TestLoQueLlevaALeerSinGrabar|TestBaseDeLaReconstruccion|TestLecturasDeLosDatos|TestUmbralesDeLasLecturas|TestUmbralesDeLasTresSkills|TestLineaDeLaSalidaNoEscrita|TestJuzgarElMotivoDeLasFallidas)$'`
en `./internal/evals/` y `-run '^TestTuberiaCerrada$'` en `./internal/app/`.

## 1 · FR-003 · Un precepto de `data/` sin grabar (T004; quickstart.md §7.a)

Cambio: en `data/plazos.yaml:29`, `bloque: a128,` → `bloque: a129,` (la única aparición).

```bash
f="$copia/data/plazos.yaml"
sed -i.orig 's/bloque: a128,/bloque: a129,/' "$f"
git diff --no-index --quiet "$f.orig" "$f" && echo "LA SUSTITUCIÓN NO SE HA APLICADO"
go -C "$copia" test -count=1 -v -run '^TestEvalsDelRepositorio$' ./internal/evals/ 2>&1 |
  grep -E -- '--- FAIL|^(ok|FAIL)|a129'
mv "$f.orig" "$f"
git diff --no-index --exit-code data/plazos.yaml "$f"
```

Visto en T015: el diff contra el `.orig` salió con 1 (una línea más y una menos), sin «NO SE HA APLICADO». En `FAIL`:

- `TestEvalsDelRepositorio/lo-que-lleva-a-leer/jurisprudencia`, `…/legal-core` y `…/boe-legislacion` (y con ellos
  `TestEvalsDelRepositorio/lo-que-lleva-a-leer`, `TestEvalsDelRepositorio` y el paquete), cada uno con
  `Should be empty, but was [` y las dos faltas de la sesión que no se prepara:

  ```text
  data/plazos.yaml: lo que la skill lleva a leer boe articulo BOE-A-1998-16718 a129: «boe articulo BOE-A-1998-16718 a129» terminó con código 1: … clase=inesperado
  ha fallado la petición del bloque a129 de la norma BOE-A-1998-16718 (https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-1998-16718/texto/bloque/a129): … no hay grabación de GET https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-1998-16718/texto/bloque/a129: falta el fichero …
  data/plazos.yaml: lo que la skill lleva a leer boe articulo BOE-A-1998-16718 a129 --fecha 9999-12-31: «boe articulo BOE-A-1998-16718 a129 --fecha 9999-12-31» terminó con código 1: … clase=inesperado
  ```

Restaurado con el `mv`: `git diff --no-index --exit-code` contra `data/plazos.yaml` del árbol de trabajo salió con 0.

Antes, en la sesión de T004, con el diff de la tarea sobre la copia: las dos faltas y `FAIL` en `jurisprudencia` y
`legal-core`, los dos que enseñó la salida, cortada con `head` antes del tercero.

## 2 · FR-002 · Sin la derivación de `data/` (T004)

Cambio: en `consultasDeLaCache` (`internal/evals/consultas.go:169`), `return necesarias.consultas, nil` →
`return ConsultasNecesarias(conjunto), nil`: la caché recibe solo las consultas necesarias de las evals, sin lo que la
skill lleva a leer de `data/` (`anotarLosDatos` sigue ejecutándose, pero su resultado se descarta).

```bash
f="$copia/internal/evals/consultas.go"
cp "$f" "$f.orig"
# el cambio, con el editor, en la línea 169 (la misma línea se repite fuera de la función)
git diff --no-index --quiet "$f.orig" "$f" && echo "EL MUTANTE NO SE HA APLICADO"
go -C "$copia" test -count=1 -v \
  -run '^(TestEvalsDelRepositorio|TestCacheDeLasSesiones|TestLoQueLlevaALeerSinGrabar|TestBaseDeLaReconstruccion|TestLecturasDeLosDatos)$' \
  ./internal/evals/ 2>&1 | grep -E -- '--- FAIL|^(ok|FAIL)'
go -C "$copia" test -count=1 -v -run '^TestCacheDeLasSesiones$' ./internal/evals/ 2>&1 | grep -E 'expected:|actual  :'
mv "$f.orig" "$f"
git diff --no-index --exit-code internal/evals/consultas.go "$f"
```

Visto en T015: el diff, `@@ -169 +169 @@ func consultasDeLaCache(…)`, solo esa línea. En `FAIL`:

- `TestEvalsDelRepositorio/lo-que-lleva-a-leer/jurisprudencia`, `…/legal-core` y `…/boe-legislacion`
  (`conjunto_test.go:2366`): `Should be empty, but was [«boe articulo BOE-A-2015-10565 a122» terminó con código 4: …
  clase=fuente-no-disponible …`, con el mensaje `lo que la skill lleva a leer de data/ que la caché de una sesión de
  01-existe-con-numero-y-fecha.yaml de ../../evals/jurisprudencia no sirve sin red:` (en `legal-core`,
  `01-territorio-municipio-cubierto.yaml`; en `boe-legislacion`, `01-lpac-articulo-21.yaml`);
- `TestCacheDeLasSesiones/sin-servir/07-cobertura-liquidacion-tributaria-daprimera` (`preparar_test.go:1131`), tres
  veces `expected: 0`, `actual  : 4`, con el mensaje `boe articulo BOE-A-2015-10565 daprimera --offline desde la caché
  de la sesión de 07-cobertura-liquidacion-tributaria.yaml: … clase=fuente-no-disponible` y el mismo con
  `--fecha 2026-09-09` y con `--fecha 9999-12-31`: `daprimera` no la nombra ninguna eval de `legal-core`; la de `a124`
  de la 05, que la 05 nombra, siguió en verde, igual que la 11 y la 12;
- `TestBaseDeLaReconstruccion/sirve-lo-que-la-skill-lleva-a-leer` (`medida_test.go:3153`): `Should be empty, but was
  [«boe articulo BOE-A-2015-10565 daprimera» terminó con código 4: …`, con el mensaje `la base de la medida de
  ../../evals/boe-legislacion sirve sin red daprimera de BOE-A-2015-10565:`;
- `TestLoQueLlevaALeerSinGrabar` (`preparar_test.go:1049`): `Should be true`, con el mensaje `alguna falta empieza por
  "data/plazos.yaml: lo que la skill lleva a leer boe articulo BOE-A-1998-16718 a129: «boe articulo BOE-A-1998-16718
  a129» terminó con código "; faltas:`; y tres veces en `preparar_test.go:1053`, `file "…/002/eval.txt" exists` (y
  `modelo.txt`, `pregunta.txt`), con el mensaje `la sesión no se abre`;
- `TestLecturasDeLosDatos/detras-de-las-evals` (`consultas_test.go:644`): `Not equal`, con el mensaje `boe articulo
  BOE-A-2015-10565 a122 suma el origen de data/ a los de la eval`; y en `consultas_test.go:651`, sobre la lista de
  consultas, que empieza `[{boe articulo [BOE-A-2015-10565 a122] [{01-un-bloque-de-los-datos.yaml comando esperado}]}
  …`.

Restaurado con el `mv`: `git diff --no-index --exit-code` contra `internal/evals/consultas.go` del árbol de trabajo salió
con 0.

Antes, en la sesión de T004, con el diff de la tarea sobre la copia, cuando la línea era la 161: las mismas líneas
`--- FAIL`.

## 3 · FR-013 · El `sin_servir` de cualquier eval de la skill aplicado a todas (T008)

Cambio: en `prepararLaCache` (`internal/evals/preparar.go:359`), la caché de la sesión de cada eval se prepara sin lo
que deja sin servir cualquier eval de la skill, no solo la suya:
`return Preparar(dirCache, s.Grabaciones, sinLoQueNoSeSirve(consultas, eval))` →

```go
todas := eval
for _, otra := range evals {
	todas.SinServir = append(slices.Clone(todas.SinServir), otra.SinServir...)
}

return Preparar(dirCache, s.Grabaciones, sinLoQueNoSeSirve(consultas, todas))
```

```bash
f="$copia/internal/evals/preparar.go"
cp "$f" "$f.orig"
# el cambio, con el editor
git diff --no-index --quiet "$f.orig" "$f" && echo "EL MUTANTE NO SE HA APLICADO"
go -C "$copia" test -count=1 -v -run '^TestCacheDeLasSesiones$' ./internal/evals/ 2>&1 |
  grep -E -- '--- (FAIL|PASS)|^(ok|FAIL)|Messages:|expected:|actual  :'
mv "$f.orig" "$f"
git diff --no-index --exit-code internal/evals/preparar.go "$f"
```

Visto en T015: el diff, 6 líneas más y 1 menos. En `FAIL`, `TestCacheDeLasSesiones/sin-servir` con:

- `05-plazo-reposicion-leganes-a124`: la 05 deja de servir `a124`, que la 11 lleva en `sin_servir`; tres veces
  `expected: 0`, `actual  : 4`, con el mensaje `boe articulo BOE-A-2015-10565 a124 --offline desde la caché de la sesión
  de 05-plazo-reposicion-leganes.yaml: … clase=fuente-no-disponible`, y el mismo con `--fecha 2026-09-09` y con
  `--fecha 9999-12-31`;
- `07-cobertura-liquidacion-tributaria-daprimera`: la 07 deja de servir `daprimera`, que lleva la 12; las mismas tres
  lecturas, con `daprimera` y `07-cobertura-liquidacion-tributaria.yaml`.

En `PASS`, `11-plazo-reposicion-sin-leer-el-precepto-a124` y `12-cobertura-seguridad-social-sin-leer-la-da-primera-daprimera`,
que esperan 4, y `TestCacheDeLasSesiones/lecturas`.

Restaurado con el `mv`: `git diff --no-index --exit-code` contra `internal/evals/preparar.go` del árbol de trabajo salió
con 0.

Antes, en la sesión de T008: las mismas líneas `--- FAIL`, con las tres lecturas de cada uno en 4 donde se esperaba 0.

## 4 · FR-033 · `lectura_sin_intentar` con `decide: false`, y retirado (T007)

Tres mutantes de `internal/evals/umbrales.go`, uno cada vez, restaurando el fichero desde el `.orig` entre uno y otro:

- (a) en `lecturasMedidas.umbrales` (línea 984), `Decide:      grupo.modelo == l.modeloQueDecide,` →
  `Decide:      false,`;
- (b) en `lecturasMedidas.umbrales`, retirado el `append` del umbral `lectura_sin_intentar` (líneas 976 a 985: la
  variable `conSinServir` y el `compararUmbral` con `nombreDelUmbralDeLecturaSinIntentar`), con el de
  `fuera_de_lo_grabado` en su sitio;
- (c) el que apuntó T007 como (b): en `umbralesDelInforme` (línea 423), `return append(umbrales,
  lecturas.umbrales()...)` → `return umbrales`. Desde T010, `lecturas.umbrales()` lleva también `fuera_de_lo_grabado`,
  así que este retira los dos; (b) es el que retira solo `lectura_sin_intentar`.

```bash
f="$copia/internal/evals/umbrales.go"
cp "$f" "$f.orig"
# (a), (b) o (c), con el editor
git diff --no-index --quiet "$f.orig" "$f" && echo "EL MUTANTE NO SE HA APLICADO"
go -C "$copia" test -count=1 -v -run '^(TestUmbralesDeLasLecturas|TestUmbralesDeLasTresSkills)$' ./internal/evals/ 2>&1 |
  grep -E -- '--- (FAIL|PASS)|^(ok|FAIL)'
go -C "$copia" test -count=1 -v -run '^TestUmbralesDeLasLecturas$/^una-no-la-intenta$' ./internal/evals/ 2>&1 |
  grep -E '_test.go:[0-9]+:|Messages:|expected|actual|ecide'
go -C "$copia" test -count=1 -v -run '^TestUmbralesDeLasTresSkills$/^legal-core$' ./internal/evals/ 2>&1 |
  grep -E '_test.go:[0-9]+:|Messages:|expected|actual|ecide|Nombre'
cp "$f.orig" "$f"   # entre uno y otro; tras el último, mv "$f.orig" "$f"
git diff --no-index --exit-code internal/evals/umbrales.go "$f"
```

Visto en T015, ninguno con «NO SE HA APLICADO» (diff de (a): `@@ -984 +984 @@`, esa línea; de (b): 11 líneas menos;
de (c): 1 más y 1 menos):

- Con (a), `TestUmbralesDeLasLecturas` en `FAIL` en nueve de sus diez subtests —`fuera-de-lo-grabado`,
  `intentan-la-lectura`, `sin-binario-ni-servidor`, `ninguna-la-intenta`, `una-no-escribe-su-salida`,
  `una-la-llama-sin-sobre`, `el-informativo-no-la-intenta`, `una-no-la-intenta` y `en-los-dos-modos`—, y
  `sin-una-eval-asi`, cuyo informe no lleva el umbral, en `PASS`. En `una-no-la-intenta`: `umbrales_test.go:2255`, dos
  veces, `-  Decide: (bool) true` / `+  Decide: (bool) false`; `umbrales_test.go:2262`, `expected:
  []string{"umbral lectura_sin_intentar:claude-sonnet-5-5:orden: 1 de 3 (33,3 %), y tiene que ser ≤ 0,0 %:
  01-sintetica-claude-sonnet-5-5-02"}`, `actual  : []string{}`, con el mensaje `falta la línea "- umbral
  lectura_sin_intentar:claude-sonnet-5-5:orden: …" en:`; y `umbrales_test.go:2263`, `expected: "fallo"`, `actual  :
  "aprobado"`. `TestUmbralesDeLasTresSkills/legal-core` en `FAIL`, con `jurisprudencia` y `boe-legislacion` en `PASS`:
  `informe_test.go:6874`, en la lista de umbrales y en su JSON, dos veces `-  Decide: (bool) true` /
  `+  Decide: (bool) false`; `informe_test.go:6905`, sobre los nombres de los que deciden, con el mensaje `deciden los
  doce del modelo que decide que no son fuera_de_lo_grabado, y los demás solo se publican`; e `informe_test.go:6909`,
  con el mensaje `doce deciden`.
- Con (b), `TestUmbralesDeLasLecturas` en `FAIL` en los mismos nueve y `sin-una-eval-asi` en `PASS`; y
  `TestUmbralesDeLasTresSkills/legal-core` en `FAIL`, con `jurisprudencia` y `boe-legislacion` en `PASS`:
  `informe_test.go:6874` echa en falta los umbrales `lectura_sin_intentar:claude-sonnet-5-5:orden`,
  `lectura_sin_intentar:claude-haiku-4-5-20251001:orden`, `lectura_sin_intentar:claude-sonnet-5-5:herramienta` y
  `lectura_sin_intentar:claude-haiku-4-5-20251001:herramienta`; `informe_test.go:6891`, los mismos nombres en la lista;
  e `informe_test.go:6904`, `6908` (`dieciocho umbrales`) y `6909` (`doce deciden`).
- Con (c), `TestUmbralesDeLasLecturas` en `FAIL` en sus diez subtests, `sin-una-eval-asi` incluido, porque se va también
  `fuera_de_lo_grabado`; y `TestUmbralesDeLasTresSkills` en `FAIL` en sus tres subtests.

`TestUmbralesDeLasTresSkills` sale en rojo, con `legal-core`, desde que T008 dio a la 11 y a la 12 su `sin_servir`: lo
que T007 dejó para que lo comprobara T015, comprobado. Restaurado: tras (a) y (b) con `cp` desde el `.orig` y tras (c)
con el `mv`, `git diff --no-index --exit-code` contra `internal/evals/umbrales.go` del árbol de trabajo salió con 0 las
tres veces.

Antes, en la sesión de T007, con (a) y con (c) (su (b)): cinco de los seis subtests que había entonces en `FAIL` y
`sin-una-eval-asi` en verde; `TestUmbralesDeLasTresSkills` en verde, porque ninguna eval del repositorio llevaba aún
`sin_servir`.

## 5 · FR-025 · El mensaje del presentador cambiado (T013; quickstart.md §7.b)

Cambio: en `enDescriptor` (`internal/render/render.go:99`), `render: no se pudo escribir en %s` →
`render: no se ha podido escribir en %s`. El arnés no copia el mensaje: pide el prefijo a `render.Nuevo` con una salida
que falla (`prefijoDeLaSalidaNoEscrita`), así que lo que hace fallar el cambio es el test del binario y no el del arnés.

```bash
f="$copia/internal/render/render.go"
sed -i.orig 's/render: no se pudo escribir en %s/render: no se ha podido escribir en %s/' "$f"
git diff --no-index --quiet "$f.orig" "$f" && echo "EL MUTANTE NO SE HA APLICADO"
go -C "$copia" test -count=1 -v -run '^TestTuberiaCerrada$' ./internal/app/ 2>&1 |
  grep -E -- '--- (FAIL|PASS)|^(ok|FAIL)|does not contain'
go -C "$copia" test -count=1 -v -run '^TestLineaDeLaSalidaNoEscrita$' ./internal/evals/ 2>&1 | grep -E -- '--- (FAIL|PASS)|^(ok|FAIL)'
go -C "$copia" test -count=1 -run '^TestJuzgarElMotivoDeLasFallidas$' ./internal/evals/ 2>&1 | grep -E -- '--- FAIL|^(ok|FAIL)'
mv "$f.orig" "$f"
git diff --no-index --exit-code internal/render/render.go "$f"
```

Visto en T015: el diff, `@@ -99 +99 @@ func enDescriptor(…)`, esa línea. `TestTuberiaCerrada` en `FAIL` en seis de sus
siete subtests —`--help`, `echo_--help`, `version`, `echo_hola_--json`, `echo_hola` y `echo_hola_--describe`—, cada
uno con

```text
"render: no se ha podido escribir en la salida estándar: write /dev/stdout: broken pipe\n" does not contain "no se pudo escribir en la salida estándar"
```

(uno de ellos con la línea del registro `… applet=echo verbo=repetir … clase=inesperado` delante), y
`echo_repetir_--help` en `PASS`, porque esa ayuda la escribe el analizador y no el presentador (`mensajeDeSalidaRota`,
`internal/app/tuberia_unix_test.go`). `TestLineaDeLaSalidaNoEscrita` en `PASS` (`ok`) y
`TestJuzgarElMotivoDeLasFallidas` en `ok`, porque el prefijo cambia con el mensaje. Restaurado con el `mv`:
`git diff --no-index --exit-code` contra `internal/render/render.go` del árbol de trabajo salió con 0.

Antes, en la sesión de T013, con el diff de la tarea sobre la copia: los mismos seis subtests en `FAIL`, el séptimo en
`PASS` y el arnés en verde.

## La copia, retirada

Visto en T015: `git -C "$copia" status --short --ignored` no dio nada tras restaurar el último fichero;
`git worktree remove --force "$copia"` salió con 0; `git worktree list` nombra solo el árbol de trabajo
(`~/Projects/kitlegal e3f8f58 [025-h26-1-las-evals-sirven]`), y `$TMPDIR/h261-copia` ya no existe.
