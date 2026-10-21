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
