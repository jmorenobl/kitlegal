# Contrato · La instalación con un nombre retirado (FR-025 a FR-031)

Cambia el verbo `install` del applet `skills` (H19): `internal/app/instalacion.go`, `internal/app/instalacion_legible.go`
y el dominio `internal/core/instalacion/`. `list` y `doctor` no cambian.

## 1. La lista de los nombres retirados (FR-025)

En `internal/app/retiradas.go`, la composición:

```go
// skillsRetiradas son los nombres de skill que este binario ya no empotra y,
// por cada uno, las skills de la entrega que heredan sus preguntas (H32, ADR 0048).
var skillsRetiradas = instalacion.Retiradas{
	"boe-legislacion": {"normativa"},
	"legal-core":      {"plazos-y-territorio"},
}
```

`instalacion.Retiradas` es `map[string][]string` y llega a `instalacion.Planificar` como un argumento más: el dominio no
lleva ningún nombre de skill. `TestSkillsRetiradas` (`internal/app/retiradas_test.go`): con los nombres de `skills/` en
`c991808` escritos en el test (`boe-legislacion`, `jurisprudencia`, `legal-core`), sale en rojo si uno que el binario
no empotra falta en la lista, si la lista nombra una skill que el binario empotra, o si una heredera no es una skill que
el binario empotra.

## 2. Qué hace `install` (FR-026 a FR-029)

En cada ámbito —proyecto, `-g` o `--dir`—, con su manifiesto, sin mirar qué skills pide la invocación ni cuáles quedan:

| Estado de una skill que el manifiesto declara con un nombre retirado | Sin `--dry-run` | `data` | Salida legible |
|---|---|---|---|
| Huellas que coinciden: su directorio tiene exactamente los ficheros que el manifiesto le declara, cada uno con su huella, y cada entrada de host es la declarada (el enlace con su destino; la copia con sus ficheros y sus huellas) | quita sus ficheros y sus directorios, de abajo arriba; sus entradas de host; y su entrada del manifiesto | `{nombre, ruta, estado: "retirada", enlaces: [las quitadas]}` | `  <nombre>: retirada; la sustituye <herederas>` |
| Huellas que no coinciden (un fichero editado, añadido o que falta; una entrada de host distinta) | no toca nada de ella; conserva su entrada del manifiesto | `{nombre, ruta, estado: "sin retirar", enlaces: [], aviso}` | `  <nombre>: sin retirar. <aviso>` |
| Ni su directorio ni ninguna de sus entradas de host están | quita su entrada del manifiesto | nada | nada |

El `aviso`, con su ruta y la heredera:

```text
<ruta> no es lo que instaló kitlegal: su protocolo puede activarse junto al de <herederas>, que la sustituye; retirarla es cosa tuya.
```

Sale con 0 (ADR 0023). Un nombre retirado pedido en la invocación (`kitlegal skills install boe-legislacion`) sale con
2 (`argumentos`), como hoy toda skill que el binario no empotra: `ValidarInvocacion` no cambia (FR-029). Una skill
declarada que no se empotra y no es un nombre retirado sigue como hoy: ni se retira, ni se cambia, ni se nombra
(FR-030, `h19-skills-no-empotrada.txtar`).

Con `--dry-run` (FR-028), nada se escribe ni se borra, y el kernel lleva a la salida de error, con las líneas de hoy,
una por cada skill de nombre retirado que retiraría o de la que avisaría:

```text
--dry-run: se habría pedido retirar boe-legislacion de .agents/skills/boe-legislacion; Claude Code .claude/skills/boe-legislacion (enlace)
--dry-run: se habría pedido conservar boe-legislacion en .agents/skills/boe-legislacion: <aviso>
```

La comprobación de las huellas usa el puerto `Disco` del dominio (`internal/core/instalacion/puertos.go`, que
implementa `internal/disco`: la huella y el listado del directorio, sin seguir enlaces) contra el `SkillDeclarada` del
manifiesto. La retirada reutiliza la fase «Retirar» de `Aplicar` y el orden de `retiradaDeCopia`
(ficheros, después directorios de abajo arriba). Una escritura o un borrado que fallan son el error de hoy de `Aplicar`
(conflicto, 7, o defecto inesperado, 1): ningún estado nuevo.

## 3. La forma de `data` (FR-026, FR-027)

`instalacion.SkillInstalada` gana dos valores de `estado` —`retirada` y `sin retirar`— y un campo opcional:

```go
// Aviso dice, en una skill de nombre retirado que no se retira, por qué y qué
// hacer; vacío en las demás (H32).
Aviso string `json:"aviso,omitempty"`
```

`schemas/instalacion.json` se regenera desde `--describe` (`make schema-check`): `estado` con sus cinco valores y
`aviso` opcional. Ejemplo de `data` en un ámbito con `boe-legislacion` intacta y `legal-core` editada, con sus bytes
(las de la entrega, `estado` «instalada», se omiten):

```json
[
  {"nombre":"boe-legislacion","ruta":".agents/skills/boe-legislacion","estado":"retirada",
   "enlaces":[{"host":"claude","ruta":".claude/skills/boe-legislacion","modo":"enlace"}]},
  {"nombre":"legal-core","ruta":".agents/skills/legal-core","estado":"sin retirar","enlaces":[],
   "aviso":".agents/skills/legal-core no es lo que instaló kitlegal: su protocolo puede activarse junto al de plazos-y-territorio, que la sustituye; retirarla es cosa tuya."}
]
```

## 4. Tamaño y caducidad (criterio de uso)

Una entrada en `data` y una línea en la salida legible por cada skill de nombre retirado que el manifiesto del ámbito
declara: como mucho dos en la entrega, con cualquier volumen de lo consultado. La retirada se dice una vez: su entrada
sale del manifiesto. El aviso se repite en cada instalación mientras sus ficheros no tengan la huella del manifiesto, y
deja de darse cuando la persona quita su directorio y sus entradas de host: la instalación siguiente quita la entrada
sin decir nada.

## 5. El e2e (FR-031)

El guion de aceptación `retiradas.txtar` ([aceptacion.md](aceptacion.md)), activado como
`internal/app/testdata/script/h32-retiradas.txtar`. Siembra cada ámbito con un manifiesto escrito en el guion —una
sección `-- proyecto/.agents/skills/kitlegal.json --`, como `h19-skills-no-empotrada.txtar`— y los ficheros de la skill
con sus huellas calculadas en el propio guion con un `sh` incluido (`declarar.sh`, el patrón de
`h19-skills-idempotencia.txtar`), y el enlace de host con `symlink`. Casos: huellas que coinciden en el proyecto y en
`-g`; un fichero editado; `--dry-run`, que no toca nada (`arbol` antes y después); y `install boe-legislacion`, con 2.
