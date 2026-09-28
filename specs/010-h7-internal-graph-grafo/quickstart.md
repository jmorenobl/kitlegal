# Quickstart: validar H7 · grafo del mundo

Guía de validación de la entrega, escenario a escenario. Cada escenario remite a su contrato y a sus FR/SC; no repite
el detalle. Se ejecuta **después** de implementar el hito: antes, el applet `graph` no existe, `boe` y `territorio` no
dejan `world.db` y los escenarios 2 a 10 no dan lo esperado.

## Antes de empezar

- Desde la raíz del repositorio, en **una sola sesión** de `sh` (los escenarios comparten `REPO`, `T`, `W` y las
  variables de los binarios), con el toolchain de `go.mod`, `git`, `make`, `cksum` y `sed`, y como un usuario que no
  es `root` (a `root` no le afectan los permisos de escritura que ejercen los escenarios 7 y 8).
- **Sin red de ninguna fuente**: `boe` responde con el binario de e2e, que sirve las grabaciones de H4 desde
  `reproduccion/` y nunca abre una conexión; `territorio` y `graph` no usan la red. La única red es la de `make ci`
  (`make vuln` consulta su base de datos), como en todos los hitos.
- Efectos en el árbol: `bin/` (el `make build` de esta preparación) y los perfiles de cobertura de `make ci`,
  **ignorados por git**. Nada toca `~/.cache/kitlegal`: toda invocación del binario lleva `KITLEGAL_CACHE_DIR` bajo
  `$T`, y los objetivos de `make` se ejecutan **sin** esa variable (`env -u KITLEGAL_CACHE_DIR make …`), para que los
  tests vean el entorno de siempre. El último escenario borra `$T` y compara `git status --porcelain` con el del
  principio.

```sh
REPO=$(git rev-parse --show-toplevel)
cd "$REPO"
INICIAL=$(git status --porcelain)
T=$(mktemp -d)
make build
E2E=./internal/app/ejemplo/kitlegal-e2e
mkdir -p "$T/bin" "$T/t0" "$T/t1" "$T/t8"
go build -o "$T/bin/kitlegal" "$E2E"
go build -ldflags "-X main.reloj=2026-09-28T12:00:00Z" -o "$T/t0/kitlegal" "$E2E"
go build -ldflags "-X main.reloj=2026-09-29T12:00:00Z" -o "$T/t1/kitlegal" "$E2E"
go build -ldflags "-X main.reloj=2026-10-06T12:00:00Z" -o "$T/t8/kitlegal" "$E2E"
K="$T/bin/kitlegal"; K0="$T/t0/kitlegal"; K1="$T/t1/kitlegal"; K8="$T/t8/kitlegal"
W="$T/w"; mkdir -p "$W/reproduccion"
cp -R internal/source/boe/testdata/boe.legislacion-consolidada "$W/reproduccion/"
A21=GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json
export KITLEGAL_CACHE_DIR="$W/cache"
cd "$W"
```

Cada binario de e2e busca `reproduccion/` en el directorio de trabajo, por eso los escenarios se ejecutan desde `$W`.

## 1. `make ci` en verde (FR-094, SC-013)

```sh
env -u KITLEGAL_CACHE_DIR make -C "$REPO" ci
wc -l < "$REPO/skills/boe-legislacion/SKILL.md"
```

Esperado: `make ci` termina en verde, con `schema-check` y `skills-check` sin diferencias y los guiones `h7-*` dentro
de `TestEntregaDelHito`; `SKILL.md` por debajo de 300 líneas.

## 2. El binario recuerda lo que consulta (US1, SC-001, SC-002; contracts/emision.md)

```sh
"$K0" boe articulo BOE-A-2015-10565 a21 --json > a21.json
"$K" territorio resolver Leganés --json > leganes.json
"$K0" graph stats --json
"$K0" boe articulo BOE-A-2015-10565 a21 --json > /dev/null
"$K" territorio resolver Leganés --json > /dev/null
"$K0" graph stats --json
ls cache
```

Esperado: las dos salidas de `graph stats` son idénticas y cuentan 5 nodos, 3 aristas y 1 texto: `Bloque`,
`BloqueVersion` y `Norma` con `boe.legislacion-consolidada`, `Municipio` y `Organo` con `kitlegal.territorio`, y
`eli:has_part`, `eli:has_version` y `lb:pertenece_a` con su fuente. En `cache/`, solo `cache.db` y `world.db`: ningún
temporal `world.db-nuevo-*` ni auxiliar (la primera entrega publicó `world.db` con `os.Link` y borró el temporal;
research D11).

## 3. `graph show`: procedencia del sobre y ningún texto legal (US1.4, US4, SC-006; contracts/applet-graph.md §3.1)

```sh
"$K0" graph show 'eli/es/l/2015/10/01/39#a21' --json
V=$(grep -o '"hash_texto":"[^"]*"' a21.json | sed 's/.*:"//; s/"$//')
"$K0" graph show "eli/es/l/2015/10/01/39#a21@20161002:$V" --json
"$K0" graph show ine:28074 --json
"$K0" graph show 'eli/es/l/2015/10/01/39#a21@20161002:'"$V" | grep -c 'Obligación de resolver' || true
sh -c '"$0" graph show no-existe --json; echo "código $?"' "$K0"
```

Esperado: el `Bloque` con su arista entrante `eli:has_part` y su saliente `eli:has_version`; la versión con
`fecha_vigencia` `20161002` y su `hash_texto`; en todos, `ultima_observacion` con la `fuente`, la `url` y la
`fecha_consulta` (`2026-09-28T12:00:00Z` para `boe`, `2026-02-04T00:00:00Z` para `territorio`) del sobre que los
creó; el `Municipio` con la arista entrante `lb:pertenece_a` desde `L01280745`; `0` coincidencias del texto;
`código 3`.

## 4. `version-obsoleta` con la versión posterior derivada (US2.1-US2.3, SC-005; contracts/applet-graph.md §3.3, §5)

```sh
"$K1" graph check --json
rm -f cache/cache.db
cp "$REPO/internal/app/testdata/derivadas/version-posterior/$A21" reproduccion/boe.legislacion-consolidada/
"$K1" boe articulo BOE-A-2015-10565 a21 --json > b21.json
"$K1" graph check --json
```

Esperado: el primer `check` sale con 0 y `data` es `[]` (una sola versión); el segundo sale con 0, `"ok":true`, y
`data` tiene exactamente un hallazgo `version-obsoleta` cuyo `id` es el de la versión de 2016 y cuya explicación
contiene `[BOE-A-2015-10565, bloque a21]`, `20161002`, `20250101`, la `url` del bloque y `2026-09-29T12:00:00Z`.

## 5. `fuente-caducada` (US2.4, US2.5; contracts/applet-graph.md §5)

```sh
"$K8" graph check --json
KITLEGAL_CACHE_DIR="$T/vacio" "$K8" graph check --json
test ! -e "$T/vacio" && echo "sin world.db"
```

Esperado: el primer `check` da, además del `version-obsoleta` del escenario 4, **un único** `fuente-caducada`: el de
la versión de 2016, observada por última vez en T0, con `"vigencia_segundos":604800` y una explicación con
`2026-09-28T12:00:00Z`, `604800 s` y `2026-10-05T12:00:00Z`. La `Norma`, el `Bloque` y la versión de 2025 se
observaron por última vez en T1 y caducan exactamente en `2026-10-06T12:00:00Z`, el instante de T8: el límite no
cuenta (FR-066) y no aparecen; `Municipio` y `Organo` no declaran vigencia. El segundo: `[]` y `sin world.db`.

## 6. El grafo no cambia la salida; `--no-graph` y `--dry-run` no lo tocan (US3.1-US3.4, US3.7, SC-003, SC-004)

```sh
"$K" boe articulo BOE-A-2015-10565 a21 --json > con.json
cksum cache/world.db > antes.txt
"$K" boe articulo BOE-A-2015-10565 a21 --json --no-graph > sin.json
"$K" boe articulo BOE-A-2015-10565 a21 --dry-run
"$K" territorio resolver Leganés --no-graph > /dev/null
cmp con.json sin.json && cksum cache/world.db | cmp - antes.txt && echo "intacto"
"$K0" graph stats --json > s1.json; "$K0" graph stats --no-graph --json > s2.json; cmp s1.json s2.json && echo iguales
ls cache
```

Esperado: `intacto` e `iguales`; ni `world.db-wal` ni `world.db-shm` en `cache/` (sí pueden estar `cache.db-wal` y
`cache.db-shm`, que deja `boe articulo --dry-run` como en `main`: es la caché de H3 y H4, no el grafo).

## 7. Una entrega fallida no cambia el código ni la salida (US3.5, US3.6, SC-011; contracts/resultado-y-entrega.md §4)

```sh
mkdir "$T/roto" && printf 'no es una base de datos\n' > "$T/roto/world.db"
cp cache/cache.db "$T/roto/"
KITLEGAL_CACHE_DIR="$T/roto" "$K" boe articulo BOE-A-2015-10565 a21 --json > roto.json; echo "código $?"
cmp roto.json con.json && echo "misma salida"
KITLEGAL_CACHE_DIR= "$K" territorio resolver Leganés --json > /dev/null; echo "código $?"
env -u KITLEGAL_CACHE_DIR -u HOME "$K" territorio resolver Leganés --json > /dev/null; echo "código $?"
KITLEGAL_CACHE_DIR= "$K" territorio resolver Leganés --json --no-graph > /dev/null 2> "$T/sin-grafo.err"
wc -c < "$T/sin-grafo.err"
cat "$T/roto/world.db"
mkdir "$T/ro" && cp cache/world.db "$T/ro/" && chmod 0400 "$T/ro/world.db"
cksum "$T/ro/world.db" > "$T/ro.antes"
KITLEGAL_CACHE_DIR="$T/ro" "$K" territorio resolver Leganés --json > /dev/null 2> "$T/ro.err"; echo "código $?"
wc -l < "$T/ro.err"
ls "$T/ro"; cksum "$T/ro/world.db" | cmp - "$T/ro.antes" && echo "sin cambios"
```

Esperado: `código 0` y `misma salida`, con una sola línea en la salida de error que empieza por
`kitlegal: lo observado no ha llegado al grafo del mundo:` y nombra `world.db`; los dos `territorio resolver` salen
con 0 y escriben esa única línea; con `--no-graph`, `0` bytes en la salida de error; y `$T/roto/world.db` sigue
diciendo `no es una base de datos`. Con `world.db` sin permiso de escritura (copia del de `cache/`, en WAL y sin
auxiliares): `código 0`, `1` línea en la salida de error, que dice que no se puede escribir `world.db`; en `$T/ro`
solo `world.db`, y `sin cambios` (contracts/almacen-world-db.md §4.1, research D11, V46).

## 8. Códigos de `graph` (US4.3, US4.4, US6.4, FR-004, FR-010, FR-011; contracts/applet-graph.md §4)

```sh
"$K" graph show; echo "código $?"
"$K" graph show a b; echo "código $?"
"$K" graph stats x; echo "código $?"
"$K" graph check x; echo "código $?"
"$K" graph show ' '; echo "código $?"
"$K" graph show "$(printf '\302\240')"; echo "código $?"
"$K" graph show "$(printf 'a\177')"; echo "código $?"
"$K" graph show 'a b'; echo "código $?"
KITLEGAL_CACHE_DIR="$T/roto" "$K" graph stats; echo "código $?"
mkdir -p "$T/esdir/world.db"; KITLEGAL_CACHE_DIR="$T/esdir" "$K" graph check; echo "código $?"
KITLEGAL_CACHE_DIR= "$K" graph stats --no-graph; echo "código $?"
mkdir "$T/cero" && : > "$T/cero/world.db"
KITLEGAL_CACHE_DIR="$T/cero" "$K" graph stats --json; wc -c < "$T/cero/world.db"
"$K0" graph stats --json > "$T/s-cache.json"
KITLEGAL_CACHE_DIR="$T/ro" "$K0" graph stats --json > "$T/s-ro.json"; echo "código $?"
cmp "$T/s-cache.json" "$T/s-ro.json" && echo "misma lectura"
ls "$T/ro"; cksum "$T/ro/world.db" | cmp - "$T/ro.antes" && echo "sin cambios"
```

Esperado: `código 2` en las siete invocaciones mal formadas —entre ellas un id de U+00A0 solo (espacio en blanco) y
uno con U+007F (control), contracts/applet-graph.md §4, research D35—; `código 3` con `a b` (un id válido que no
está); `código 1` con `world.db` que no es base y con un
directorio, y el fichero intacto; `código 2` con la variable vacía también con `--no-graph`; con 0 bytes, tres ceros y
listas vacías, y `0` bytes después. Sobre el `world.db` sin permiso de escritura del escenario 7: `código 0`,
`misma lectura` que sobre `cache/` (de donde se copió), solo `world.db` en `$T/ro` —ni `-wal` ni `-shm`— y
`sin cambios` (contracts/almacen-world-db.md §3, research D10, V46).

## 9. Ocho a la vez (US6.3, SC-009)

```sh
for c in 28074 28079 28065 28007 28092 28058 28106 47165; do
  KITLEGAL_CACHE_DIR="$T/ocho" "$K" territorio resolver "$c" --json > /dev/null 2>> "$T/ocho.err" &
done; wait
wc -c < "$T/ocho.err"
KITLEGAL_CACHE_DIR="$T/ocho" "$K" graph stats --json
ls "$T/ocho"
```

Esperado: `0` bytes de salida de error y ocho `Municipio` y ocho `Organo` con `kitlegal.territorio`; en `$T/ocho`,
que no existía, `world.db` y ningún temporal `world.db-nuevo-*`: una invocación lo publicó y las otras siete aplicaron
sobre él (research D11). Si las últimas conexiones cierran a la vez, SQLite puede dejar además `world.db-wal` y
`world.db-shm`, que ningún requisito prohíbe y que `grafo-concurrencia` no afirma (supuesto de T001).

## 10. El binario distribuido y la skill (FR-050, FR-083, FR-085)

```sh
KITLEGAL_CACHE_DIR="$T/prod" "$REPO/bin/kitlegal" territorio resolver Tordesillas --json > /dev/null
KITLEGAL_CACHE_DIR="$T/prod" "$REPO/bin/kitlegal" graph stats --json
"$REPO/bin/kitlegal" boe articulo --help | tr -s ' \n' '  ' | grep -c 'No entrega al grafo del mundo nada de lo que observa la invocación'
grep -c 'kitlegal graph check' "$REPO/skills/boe-legislacion/SKILL.md"
env -u KITLEGAL_CACHE_DIR make -C "$REPO" skills-check
```

Esperado: `graph stats` cuenta el `Municipio` y el `Organo` de Tordesillas; `1` (la ayuda parte la frase en dos
líneas cuando no cabe en una, como tiene en cuenta `grafo-applet`; `tr` las une antes de buscarla); al menos una línea
con `kitlegal graph check`; `skills-check` en verde (incluye la eval informativa nueva y su grafo previo).

## 11. Costes (SC-007, SC-008)

```sh
env -u KITLEGAL_CACHE_DIR make -C "$REPO" test-tiempos
(cd "$REPO" && env -u KITLEGAL_CACHE_DIR go test -race -count=1 -v -run '^TestCosteDelGrafo$' ./internal/app/) | grep 'SC-00[78]: mediana'
```

Esperado: `test-tiempos` en verde. `TestCosteDelGrafo` publica las medianas medidas con `t.Logf`, que `go test` solo
muestra con `-v`: la segunda orden lo ejecuta otra vez, solo, y da dos líneas, `SC-007: mediana de 20 boe articulo
desde la caché: …` con la diferencia y su máximo de 150 ms, y `SC-008: mediana de 5 sobre 10000 nodos y 10000
aristas: …` con `graph check` (máximo 3 s) y `graph stats` (máximo 1 s).

## 12. Limpieza

```sh
cd "$REPO"
rm -rf "$T"
test "$(git status --porcelain)" = "$INICIAL" && echo "árbol intacto"
```

Esperado: `árbol intacto`.
