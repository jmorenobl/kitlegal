# Mutantes de H26.1

Cada mutante se aplica en una copia fuera del árbol y se retira al terminar (tasks.md, «Batería de verificación por
tarea»; quickstart.md §7). Ninguno queda en el diff.

```bash
copia="$TMPDIR/h261-copia"
git worktree add --detach "$copia" HEAD
# … cada mutante, restaurando su fichero tras ejecutarlo …
git worktree remove --force "$copia"
git worktree list
git status --short
```

## FR-002 · Sin la derivación de `data/` (T004)

`consultasDeLaCache` devuelve solo las consultas necesarias de las evals, sin lo que la skill lleva a leer de `data/`.

```bash
f="$copia/internal/evals/consultas.go"
sed -i.orig '/^func consultasDeLaCache/,/^}/ s/return necesarias.consultas, nil/return ConsultasNecesarias(conjunto), nil/' "$f"
git diff --no-index --quiet "$f.orig" "$f" && echo "EL MUTANTE NO SE HA APLICADO"
go -C "$copia" test -count=1 \
  -run '^(TestEvalsDelRepositorio|TestCacheDeLasSesiones|TestLoQueLlevaALeerSinGrabar|TestBaseDeLaReconstruccion|TestLecturasDeLosDatos)$' \
  ./internal/evals/ 2>&1 | grep -E '^\s*--- FAIL|^(ok|FAIL)'
mv "$f.orig" "$f"
```

Esperado: ninguna línea «NO SE HA APLICADO» (el diff es una sola línea, la 161: `return necesarias.consultas, nil` →
`return ConsultasNecesarias(conjunto), nil`), y en `FAIL`:

- `TestEvalsDelRepositorio/lo-que-lleva-a-leer`, con sus tres subtests, `boe-legislacion`, `legal-core` y
  `jurisprudencia`: la caché de su sesión no sirve los bloques ni las normas de `data/` con `--offline` (código 4);
- `TestCacheDeLasSesiones/sin-servir/07-cobertura-liquidacion-tributaria-daprimera`: `daprimera` no la nombra ninguna
  eval de `legal-core` (la de `a124` de la 05 sigue en verde, porque la 05 la nombra);
- `TestBaseDeLaReconstruccion/sirve-lo-que-la-skill-lleva-a-leer`;
- `TestLoQueLlevaALeerSinGrabar`: sin la lectura de `a129`, no hay falta que la nombre;
- `TestLecturasDeLosDatos/detras-de-las-evals`.

Ejecutado en la sesión de T004 sobre la copia con el diff de la tarea aplicado, con la misma sustitución hecha con el
editor (`sed -i` pide aprobación en el sandbox de la sesión): las líneas `--- FAIL` fueron exactamente esas. Retirado:
`git diff --no-index --exit-code` de la copia contra `internal/evals/consultas.go` del árbol de trabajo salió con 0.

## FR-003 · Un precepto de `data/` sin grabar (T004; quickstart.md §7.a)

```bash
f="$copia/data/plazos.yaml"
sed -i.orig 's/bloque: a128,/bloque: a129,/' "$f"
git diff --no-index --quiet "$f.orig" "$f" && echo "LA SUSTITUCIÓN NO SE HA APLICADO"
go -C "$copia" test -count=1 -run '^TestEvalsDelRepositorio$' ./internal/evals/ 2>&1 | grep -E 'a129|FAIL' | head -5
mv "$f.orig" "$f"
```

Esperado: ninguna línea «NO SE HA APLICADO»; `TestEvalsDelRepositorio/lo-que-lleva-a-leer` en `FAIL` en sus tres
subtests, porque la sesión no se prepara, con la falta

```text
data/plazos.yaml: lo que la skill lleva a leer boe articulo BOE-A-1998-16718 a129: «boe articulo BOE-A-1998-16718 a129» terminó con código 1: …
```

y la misma con `--fecha 9999-12-31`. Ejecutado en la sesión de T004 igual que el anterior: salieron las dos faltas, con
«ha fallado la petición del bloque a129 de la norma BOE-A-1998-16718» detrás, y `FAIL` en `jurisprudencia` y en
`legal-core`, los dos que enseñó la salida, cortada con `head` antes del tercero. Retirado con el `mv`.

## FR-033 · `lectura_sin_intentar` con `decide: false`, y retirado (T007)

Dos mutantes de `internal/evals/umbrales.go`, uno cada vez:

- (a) en `lecturasMedidas.umbrales`, `Decide: grupo.modelo == l.modeloQueDecide,` → `Decide: false,`;
- (b) en `umbralesDelInforme`, `return append(umbrales, lecturas.umbrales()...)` → `return umbrales`.

```bash
f="$copia/internal/evals/umbrales.go"
cp "$f" "$f.orig"
# (a) o (b), con el editor
git diff --no-index --quiet "$f.orig" "$f" && echo "EL MUTANTE NO SE HA APLICADO"
go -C "$copia" test -count=1 -run '^(TestUmbralesDeLasLecturas|TestUmbralesDeLasTresSkills)$' ./internal/evals/ 2>&1 |
  grep -E '^\s*--- FAIL|^(ok|FAIL)'
mv "$f.orig" "$f"
```

Esperado: ninguna línea «NO SE HA APLICADO», y con cada uno, en `FAIL`, `TestUmbralesDeLasLecturas` en cinco de sus seis
subtests: `intentan-la-lectura`, `una-no-la-intenta`, `el-informativo-no-la-intenta`, `ninguna-la-intenta` y
`en-los-dos-modos`; `sin-una-eval-asi`, cuyo informe no lleva el umbral, sigue en verde. En `una-no-la-intenta`, con
(a), el umbral del modelo que decide sale con `decide: false` (`exigirUmbrales`, en la lista y en informe.json), sin el
motivo `umbral lectura_sin_intentar:claude-sonnet-5-5:orden: 1 de 3 (33,3 %), y tiene que ser ≤ 0,0 %:
01-sintetica-claude-sonnet-5-5-02` (`exigirMotivosDeLaRaiz`: `actual: []string{}`) y con el veredicto `aprobado` en
lugar de `fallo`; con (b), la lista de umbrales sale vacía (`actual: []evals.Umbral{}` y `"[]"` en informe.json), y sin
ese motivo. `TestUmbralesDeLasTresSkills` sigue en verde con los dos, porque ninguna eval del repositorio lleva aún
`sin_servir`: lo pondrá en rojo T008, con la 11 (T015 lo comprueba).

Ejecutado en la sesión de T007 sobre una copia (`git worktree add --detach` en `$TMPDIR`) con el diff de la tarea
aplicado, con los dos cambios hechos con el editor: las líneas `--- FAIL` fueron exactamente esas cinco y `FAIL` del
paquete, con cada uno. Retirados con el `mv`: `git diff --no-index --exit-code` de la copia contra
`internal/evals/umbrales.go` del árbol de trabajo salió con 0, y `git worktree remove --force` retiró la copia.
