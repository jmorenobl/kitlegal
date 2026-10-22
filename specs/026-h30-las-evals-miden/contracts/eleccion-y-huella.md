# Contrato: conjunto, elección y huella de cada eval

Lo calcula el arnés (`internal/evals`) sobre el árbol de la cabeza, sin red salvo la lista de ficheros de la propuesta.
Entidades en [../data-model.md](../data-model.md) §1 a §6; decisiones en [../research.md](../research.md) D2 a D7 y D19.

## 1. Entradas

| Entrada | De dónde | Quién la da |
|---|---|---|
| Árbol de la cabeza | `git ls-files -z --cached --others --exclude-standard` en la raíz, sha256 de cada fichero que mide; `go mod edit -json go.mod`; frontmatter de cada `skills/*/SKILL.md` | el checkout del trabajo; en `make ci`, el árbol de trabajo |
| Grafo de paquetes | `go list -tags evals -deps -json -e ./...` con `GOOS=linux GOARCH=amd64 CGO_ENABLED=0`, una vez | el módulo, sin red (caché de módulos) |
| Ficheros de `internal/app` | `go/parser`, solo importaciones y métodos `Nombre()` | el árbol |
| Árbol de la base | base de fusión `git merge-base <base> HEAD`; `git show <base>:<ruta>`, `git ls-tree --name-only <base> skills/ evals/`; `go.mod` de la base con `go mod edit -json` sobre un temporal | el checkout con `fetch-depth: 0` (S4) |
| Rutas cambiadas | `gh api --paginate "repos/{owner}/{repo}/pulls/$PROPUESTA/files" --jq '.[] \| [.filename, (.previous_filename // "")] \| @tsv'` | la plataforma (FR-007; S3) |
| Fijados de la huella | banderas `-modelo-que-decide`, `-version-de-claude-code`, `-modelo-del-juez`, `-version-del-juez`, desde el `env` del trabajo | `.github/workflows/evals.yml` |

Cada orden con un valor variable va por `sh -c` con su texto constante y el valor en una variable del entorno, como las
de la tanda (gosec G204).

## 2. Conjuntos

Uno por eval de `evals/*/[0-9][0-9]-*.yaml`, con sus seis partes (data-model §3). Ejemplo, la eval
`legal-core/05-plazo-reposicion-leganes` (extracto ilustrativo de la forma; las entradas exactas, con todos los paquetes
del cierre de `boe`, `graph`, `plazos` y `territorio`, las da el cálculo):

```text
fichero    evals/legal-core/05-plazo-reposicion-leganes.yaml
fichero    skills/legal-core/SKILL.md
directorio skills/legal-core/references/
fichero    skills/boe-legislacion/SKILL.md
directorio skills/boe-legislacion/references/
fichero    data/plazos.yaml
paquete    internal/core/plazos
paquete    internal/source/boe
directorio internal/source/boe/testdata/boe.legislacion-consolidada/
directorio testdata/evals/boe.legislacion-consolidada/
directorio internal/evals/testdata/boe.legislacion-consolidada/
modulo     go.yaml.in/yaml/v3 v3.0.5
```

No lleva `evals/legal-core/` salvo su propio fichero (legal-core no tiene juez ni otros ficheros), ni
`skills/jurisprudencia/`. La de `jurisprudencia/01-existe-con-numero-y-fecha` lleva `skills/boe-legislacion/SKILL.md`,
el paquete `internal/core/cita` y el directorio `evals/jurisprudencia/juez/`, y no `skills/legal-core/SKILL.md`.

Error (FR-005): `make ci` falla con `<skill>/<eval>: <causa>` por cada eval sin conjunto; en el job, la elección es
todas con `sin_conjunto` y la causa va al informe (`origen.sin_conjunto`).

## 3. Elección

Reglas en data-model §5. Salida en el registro del paso, una línea por eval elegida, en orden de skill y de fichero:

```text
elegida legal-core/05-plazo-reposicion-leganes: conjunto: skills/legal-core/references/recursos_y_plazos.md
```

107 bytes más el salto; con varias rutas, separadas por `, `; con una ruta común, `elegida <skill>/<eval>: comun: internal/cli/salida.go`. Si `gh` no da los
ficheros de la propuesta, una línea `error de la plataforma: <orden>: <salida de error>` y, después, `elegida
<skill>/<eval>: plataforma` por cada eval (data-model §5; research D17). Con *e* evals
elegidas y *r* rutas que eligen cada una, unos *e* × (60 + 50·*r*) bytes: hoy, como mucho 45 líneas; con cientos de evals,
crece con las elegidas y con las rutas de la propuesta, no con la historia. En el trabajo `cambios`, después,
`elige=si` o `elige=no` en `$GITHUB_OUTPUT`. Deja de darse cuando la propuesta se fusiona: sobre `main` no hay propuesta
y la suite completa no elige.

## 4. Huella

sha256 en hexadecimal del texto siguiente, en UTF-8, una línea por elemento terminada en `\n`; dentro de cada bloque,
las líneas ordenadas por bytes; los bloques, en este orden:

```text
huella-de-eval v1
modelo_que_decide claude-sonnet-5-5
claude_code 2.1.284
juez claude-opus-5-5 2.1.289
go 1.27.0
toolchain go1.27.3
modulo github.com/alecthomas/kong v1.16.1
modulo modernc.org/sqlite v1.60.1
go_mod_resto 8b0e…(64)
go_sum_resto d41c…(64)
descripcion boe-legislacion 4f1c…(64)
descripcion jurisprudencia 0a9e…(64)
descripcion legal-core 77d2…(64)
ruta .github/workflows/evals.yml 1b3a…(64)
ruta evals/jurisprudencia/01-existe-con-numero-y-fecha.yaml 9c0d…(64)
ruta evals/jurisprudencia/juez/rubrica.md 51aa…(64)
```

- `juez`: solo si la skill de la eval tiene `evals/<skill>/juez/`; la rúbrica, los casos y la medida entran como rutas
  de su conjunto (parte (f)).
- `modulo`: los del conjunto y los de lo común, cada uno una vez.
- `go_mod_resto`: sha256 del JSON de `go mod edit -json go.mod` sin `Go`, `Toolchain` ni las entradas de `Require`
  de los módulos de algún conjunto o de lo común; `go_sum_resto`: sha256 de las líneas de `go.sum`, ordenadas, cuyo
  módulo no es de ningún conjunto ni de lo común (research D7). Las dos, iguales en todas las evals.
- `descripcion`: de cada skill de la cabeza, el sha256 de su `description`.
- `ruta`: cada fichero de la cabeza que pertenece al conjunto, a lo común o a ninguno de los dos (sin asignar), con el
  sha256 de su contenido; nunca un fichero que no mide, ni `go.mod` ni `go.sum`, que entran por sus dos restos.

Ejemplo de valor: `3f6d2b1c…` (64 caracteres). Lo que no entra: la fecha, el runner, la ruta absoluta, el orden del
recorrido, el commit, los modelos informativos (el job los fija en `evals.yml`, que es una ruta de lo común).

## 5. Garantías que comprueban los tests de `make ci`

| Garantía | Test |
|---|---|
| Cada eval del repositorio tiene conjunto no vacío y las seis partes; una de cada skill lleva lo que pide FR-001 y no lo que excluye; una skill con un applet inexistente da error que nombra la eval | `TestConjuntoDeCadaEval` |
| La elección de cada mutante de SC-003 y de FR-002 a FR-007, eval por eval | `TestEleccionDeLosMutantes` |
| La matriz de FR-019 de un conjunto fijo de propuestas, byte a byte | `TestMatrizDeLasPropuestas` |
| Cada clase de ruta cambia la huella de las evals que elige y no la de las demás; `go.mod` de lo común cambia todas; `docs/` ninguna; dos clones en directorios distintos dan las mismas | `TestHuellaDeLasEvals` |
