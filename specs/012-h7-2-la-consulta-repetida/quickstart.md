# Quickstart: validar H7.2

Escenarios para comprobar la entrega una vez implementada, desde la raíz del repositorio y en la rama del hito. Los §1
a §5 y el §7 no usan red ni modelo (el §7, `make ci`, sí red para `govulncheck`, como siempre); el §6 abre dos
conversaciones de Claude Code y lo ejecuta la persona al leer el informe final, fuera de `make ci` y del run (FR-062).

Efectos: ninguno en el índice ni en el historial de git. En el árbol de trabajo, solo lo que cada escenario declara: el
§2 reescribe la derivada con el mismo contenido (git no ve ningún cambio) y el §7 escribe `coverage.out` y
`coverage-integration.out`, que git ignora (`/*.out`). El §6 escribe solo en su directorio temporal, que borra al final.
Formatos: [contracts/lista-y-juicio.md](./contracts/lista-y-juicio.md),
[contracts/eval-y-derivada.md](./contracts/eval-y-derivada.md),
[contracts/skill-boe-legislacion.md](./contracts/skill-boe-legislacion.md) y
[contracts/comprobacion-del-quickstart.md](./contracts/comprobacion-del-quickstart.md).

## 1. La lista, la eval nueva y la skill, sin modelo (FR-002, FR-030, FR-043, FR-051, FR-084, FR-085)

```bash
go test -count=1 -v -run '^TestEvalsDelRepositorio$' ./internal/evals/ | grep -E -- '--- (PASS|FAIL)'
make skills-check
```

Esperado: `--- PASS` en `TestEvalsDelRepositorio` y en cada subtest, entre ellos `formato` (la lista bien formada),
`conjunto` (19 evals, 10 positivas que deciden), `grafo-previo` (con la eval 19, un `version-obsoleta` de 20180309 a
20200206), `expresiones-calibradas` (35 de 93 con el reparto de FR-084), `expresiones-en-los-bloques` y
`expresiones-de-la-skill`; ningún `--- FAIL`; y `make skills-check` en verde.

## 2. La derivada del grafo previo (FR-010 a FR-013, SC-005)

```bash
go test -count=1 -v -run '^TestGrabacionesDerivadas' ./internal/app/ | grep -E -- '--- (PASS|FAIL)'
go test -count=1 -run '^TestGrabacionesDerivadas$' ./internal/app/ -args -actualizar-derivadas
git status --porcelain -- testdata/evals/grafo-previo/
```

Esperado: `--- PASS` en `TestGrabacionesDerivadas` y en sus subtests `version-posterior`, `version-ulterior`,
`sin-eli`, `eli-sin-segmento` y `lcsp-a1-30-redaccion-original`, y en el test de las derivadas inventadas (las tres no
pasan, como deben); la segunda orden, `ok`, reescribe la derivada desde la grabación; la tercera no imprime nada: el
código del repositorio la reproduce byte a byte.

## 3. El juicio, el informe y la comprobación del quickstart, sin modelo (FR-050 a FR-055, FR-061, FR-082, FR-083)

```bash
go test -count=1 ./internal/evals/
```

Esperado: `ok`. Incluye los tests de formato de la lista, de `ExtraerExpresionesProhibidas`, de `Juzgar` (sin
expresiones, una de cada familia, las variantes toleradas y la línea `⚠ REDACCIÓN MODIFICADA:`), del informe (por
sesión y por modelo, con y sin prueba de red) y de las condiciones de la comprobación del §6.

## 4. Lo retirado no está (FR-020, SC-007)

```bash
test ! -e evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml && echo "retirada la eval"
test ! -e testdata/evals/grafo-previo/lpac-a21-version-anterior && echo "retirado su grafo previo"
git grep -n -e lpac-a21-version-anterior -e parrafoDeLaVersionAnterior -e 19-lpac-articulo-21-redaccion-cambiada \
  -- '*.go' '*.txtar' 'evals/' 'testdata/' 'scripts/'
echo "coincidencias: $?"
```

Esperado: las dos líneas `retirad…` y `coincidencias: 1` (ninguna).

## 5. La skill y el conjunto (FR-001, FR-030, FR-047)

```bash
wc -l < skills/boe-legislacion/SKILL.md
ls evals/boe-legislacion/[0-9][0-9]-*.yaml | wc -l
grep -n -e '^pregunta:' -e 'grabaciones:' -e '- version-obsoleta' evals/boe-legislacion/19-lcsp-contrato-menor-redaccion-cambiada.yaml
```

Esperado: menos de 300 líneas; `19`; la pregunta literal de FR-001, `grabaciones: lcsp-a1-30-redaccion-original` y
`- version-obsoleta`.

## 6. La consulta repetida en Claude Code (FR-060, FR-061, FR-062, SC-002)

Necesita Claude Code con la credencial de la cuenta de la persona y red hacia el modelo; abre dos conversaciones con
`claude-sonnet-5` (unos minutos, consumo de la suscripción). Todo va a un directorio temporal: el binario de la rama y el
guion que lo pone primero en el `PATH` de las conversaciones, la caché y el grafo preparados como la sesión de la eval
nueva en el job, la skill instalada en local y las dos conversaciones. Nada cambia fuera de él: ni el repositorio, ni la
instalación de kitlegal o las skills de la cuenta, ni queda ninguna conversación guardada (`--no-session-persistence`).

```bash
d=$(cd "$(mktemp -d)" && pwd -P)
mkdir -p "$d/bin" "$d/sesion/cache" "$d/trabajo" "$d/primera" "$d/segunda"

# El binario de la rama, y el guion que Claude Code ejecuta antes de cada orden de Bash de las conversaciones
# (CLAUDE_ENV_FILE) y lo pone primero en su PATH.
CGO_ENABLED=0 go build -trimpath -o "$d/bin/kitlegal" ./cmd/kitlegal
printf 'export PATH="%s/bin:$PATH"\n' "$d" > "$d/entorno.sh"

# La preparación del job: el grafo con la redacción original del art. 118 LCSP y la caché con la vigente.
go test -tags evals -count=1 -run '^TestPrepararSesion$' ./internal/evals/ -args \
  -skill boe-legislacion -eval 19-lcsp-contrato-menor-redaccion-cambiada.yaml -modelo claude-sonnet-5 \
  -sesion "$d/sesion"

# La skill de la rama, con ese binario, en ámbito local en el directorio de las conversaciones.
(cd "$d/trabajo" && export PATH="$d/bin:$PATH" && kitlegal skills install boe-legislacion --host claude)
test -L "$d/trabajo/.claude/skills/boe-legislacion" && test -f "$d/trabajo/.claude/skills/boe-legislacion/SKILL.md" \
  && echo "skill instalada en $d/trabajo"

# Dos conversaciones nuevas, una tras otra, con la pregunta literal de la eval. Antes, sin modelo, `kitlegal` resuelto
# como en el Bash de las conversaciones —los ficheros de arranque del shell de la persona y después el guion—: si no es
# el de la rama, no se abre ninguna.
(cd "$d/trabajo" && export CLAUDE_ENV_FILE="$d/entorno.sh" KITLEGAL_CACHE_DIR="$d/sesion/cache" &&
  k=$("${SHELL:-/bin/sh}" -l -i -c '. "$CLAUDE_ENV_FILE" && command -v kitlegal' < /dev/null 2> /dev/null |
    tail -n 1) &&
  if [ "$k" != "$d/bin/kitlegal" ]; then
    echo "las conversaciones ejecutarían ${k:-ningún kitlegal} y no $d/bin/kitlegal: no se abren" >&2
    exit 1
  fi &&
  for c in primera segunda; do
    claude -p "$(cat "$d/sesion/pregunta.txt")" --model claude-sonnet-5 \
      --output-format stream-json --verbose --no-session-persistence --setting-sources project \
      --permission-mode dontAsk --allowedTools 'Bash(kitlegal *)' Read Skill \
      < /dev/null > "$d/$c/sesion.jsonl" 2> "$d/$c/sesion.err"
    echo $? > "$d/$c/codigo-de-la-sesion"
  done)

# La comprobación mecánica de las dos respuestas.
go test -tags evals -count=1 -v -run '^TestComprobarConsultaRepetida$' ./internal/evals/ -args \
  -skill boe-legislacion -primera "$d/primera" -segunda "$d/segunda" \
  -fecha-superada 20180309 -fecha-leida 20200206
```

Esperado: `skill instalada en …`; y de la comprobación, `--- PASS: TestComprobarConsultaRepetida` con la línea
`se cumplen las tres condiciones: la forma con 20180309 y 20200206 en la primera respuesta, sin ella en la segunda, y
ninguna expresión prohibida en las dos`, y `ok`. Si algo no se cumple, `--- FAIL` con una línea por condición que
falla (contracts/comprobacion-del-quickstart.md §2); las respuestas se leen en el campo `result` del último mensaje de
`"$d/primera/sesion.jsonl"` y `"$d/segunda/sesion.jsonl"`. Si `kitlegal` no resuelve al de la rama, la orden de las
conversaciones termina sin abrir ninguna con `las conversaciones ejecutarían … y no …/bin/kitlegal: no se abren`, y la
comprobación no tiene nada que leer.

Por qué estas banderas (research D17): `CLAUDE_ENV_FILE` es lo que garantiza que las conversaciones ejecutan el binario
de `"$d/bin"`, y no el `PATH` heredado: el Bash de Claude Code carga antes de cada orden una instantánea del shell de la
persona con el `PATH` que dejan sus ficheros de arranque —un `brew shellenv` puede poner delante `/opt/homebrew/bin` y,
con él, el `kitlegal` que la persona tenga instalado, otra versión: la v0.3.1 no tiene el applet `graph`— y después el
guion de `CLAUDE_ENV_FILE`, que deja `"$d/bin"` como primera entrada; la comprobación previa resuelve `kitlegal` del
mismo modo, con el shell de la persona como shell de inicio de sesión interactivo y el guion. `--setting-sources project`
hace que la conversación cargue la `boe-legislacion` de `"$d/trabajo"` y no la de la cuenta —el cargador de Claude Code
solo lee `~/.claude/skills` con la fuente `user`—; `--permission-mode dontAsk` con `--allowedTools` le deja activar la
skill, leer sus `references/` y ejecutar `kitlegal`, y deniega sin preguntar todo lo demás; `stream-json` con
`--verbose` es el formato del job, y la respuesta que se comprueba es la misma que juzgaría. La caché sirve el bloque
siete días; si una conversación pide `metadatos` o `buscar` pasados cinco minutos desde la preparación, `kitlegal` los
pide al BOE como en cualquier uso, y nada de lo que se comprueba depende de ellos.

Cuando ya no hagan falta las respuestas:

```bash
rm -rf "$d"
```

## 7. El veredicto del repositorio (FR-080, SC-008)

```bash
make ci
```

Esperado: `ci: todos los controles en verde`. El job de evals de la propuesta de cambio (FR-087) lo lanza el workflow
tras la revisión final; su informe da lo que pide SC-001.
