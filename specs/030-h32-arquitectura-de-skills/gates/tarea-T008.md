# T008: por qué no quedó en verde

## Intento 1: la tarea necesita una ruta que no declaraba (redelimitada)

**Lo que queda hecho en el árbol, que hereda el intento siguiente** (sin commitear):

- `internal/skills/paquetes.go`, nuevo: `leerPaquetes` (los valores de `vertical` de `data/normas.yaml`, en orden,
  y por skill los que tienen `references/<paquete>.md` como fichero regular), `regionDeMarcas.buscar` (las marcas de
  una región y el defecto de las que no van en pareja), `paquetes.regenerar` (los enlaces entre las marcas de los
  paquetes, la lista entre las de la base, el defecto `referencia de paquete sin enlazar: <nombre>.md` y el de la
  tabla de normas que no se lee) y `paquetes.lineas`.
- `internal/skills/sincronia.go`: `Regenerar` lee los paquetes una vez (`sync.OnceValues`), con la primera skill
  cuyo frontmatter no tiene defectos, para que un `data/` que no es un directorio siga dando el error de hoy
  (`TestRegenerarSinPoderListar`); `regenerarSkill` sustituye las marcas tras la tabla de comandos y antes de medir
  el cuerpo y los enlaces; `sinRepetir` deja una vez el defecto de `data/normas.yaml` si también lo da una referencia
  generada desde ella.
- `internal/skills/paquetes_test.go`, nuevo: `TestPaquetes` con los casos de la tarea (un paquete con su línea de 93
  caracteres y su salto, dos en orden, la lista de la entrada, el texto editado a mano, la referencia sin marcas, las
  marcas sin pareja) y los de la tabla de normas con defectos, sin tabla y sin poder leer.
- Los supuestos de la tarea, en `gates/supuestos.md` (tres líneas T008).

Medido en esta sesión: `go test -count=1 ./internal/skills/` → `ok`; `golangci-lint run ./internal/skills/` → `0
issues`; dos mutantes (sin `slices.Sort(nombres)`; sin `sinRepetir`) ponen en rojo `dos-paquetes-en-orden`,
`normas-con-defectos` y `referencia-sin-marcas`, y restaurados vuelve el verde.

**Lo que sigue en rojo**: `make skills-check`, por `TestSkillsDelRepositorio` de `internal/app`. `go test -count=1
-skip TestMedidasDeTiempo ./...` (sin `-race`, en esta sesión) da en rojo solo ese test:

```
--- FAIL: TestSkillsDelRepositorio/sincronia-con-referencias-a-mano/normativa/desde-el-byte-cambiado
    actual  : []string{"normativa: SKILL.md: las marcas de los paquetes no van en pareja: falta la de inicio"}
--- FAIL: TestSkillsDelRepositorio/sincronia-con-referencias-a-mano/normativa/desde-la-cabecera-retirada
    actual  : []string{"normativa: SKILL.md: las marcas de los paquetes no van en pareja: falta la de inicio"}
--- FAIL: TestSkillsDelRepositorio/referencias/normativa/cien-lineas-sin-apartados
--- FAIL: TestSkillsDelRepositorio/referencias/normativa/ciento-una-lineas-sin-apartados
--- FAIL: TestSkillsDelRepositorio/referencias/normativa/ciento-una-lineas-con-indice
--- FAIL: TestSkillsDelRepositorio/referencias/normativa/cien-lineas-sin-indice
--- FAIL: TestSkillsDelRepositorio/referencias/normativa/enlazada-desde-otra
--- FAIL: TestSkillsDelRepositorio/referencias/normativa/ciento-una-lineas-sin-indice
    (cada uno, con el defecto de arriba delante de los que espera)
FAIL	github.com/jmorenobl/kitlegal/internal/app
```

**Causa**: el código hace lo que pide la tarea; lo que falla es un ayudante de test. `enlazarDesdeElSkillMd`
(`internal/app/skills_test.go`, hacia la línea 1234) sustituye por `[](references/<fichero>)` la primera línea de
«prosa» del cuerpo —fuera de la tabla y de los bloques, sin `[`, `` ` ``, `~` ni `|`, y con al menos tantos
caracteres como el enlace—. En `normativa` todas las líneas anteriores a la marca llevan corchete o comilla
invertida, así que la primera de prosa es `<!-- paquetes: lo genera make skills-sync -->` (T006): la copia pierde
la marca de inicio y la sincronía da, con razón, el defecto de las marcas sin pareja. Antes de T008 nadie leía la
marca y el test pasaba. `internal/app/skills_test.go` no estaba en las rutas de la tarea (`tarea-actual.json`), y
cambiar el cuerpo de `normativa` para esquivarlo no cabe (su texto lo fija T006, contracts/textos.md N1).

**Redelimitación**: la línea de T008 en `tasks.md` declara ahora `internal/app/skills_test.go` y dice qué cambiar.

**Arreglo probado** (en un worktree desechable bajo `$TMPDIR` con estos tres ficheros copiados, ya retirado): un
caso más en el `switch` de `enlazarDesdeElSkillMd`, antes del de la prosa, que no toma por prosa un comentario
HTML (las marcas):

```go
		case strings.HasPrefix(texto, "<!--"):
```

con su comentario y la frase del de la función («Una línea de prosa … no es un comentario HTML, como las marcas de
los paquetes …»). Con él, `go test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/` → `ok` y
`golangci-lint run ./internal/app/ ./internal/skills/` → `0 issues`. Falta, en el intento siguiente: aplicarlo,
`make skills-check` y `make ci` completos (en este intento no se corrió `make ci`: habría salido en rojo por
`test` y `skills-check`, como arriba).

## Intento 2: en verde

Aplicado el arreglo del intento 1 en `internal/app/skills_test.go`: un caso `strings.HasPrefix(texto, "<!--")` en el
`switch` de `enlazarDesdeElSkillMd`, antes del de la prosa, con su comentario y la frase del de la función. Medido en
esta sesión, por objetivos y en primer plano: `go test -count=1 ./internal/skills/` → 551 pasan;
`go test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/` → rojo antes del arreglo (las marcas sin pareja)
y 187 pasan después; `make skills-check` → ok en app, skills y evals; `make fmt-check lint` → 0 issues; `make test`
→ sin ningún FAIL; `make vuln schema-check goreleaser-check secrets mod-verify mod-tidy-check` → sin fugas, módulos
verificados y `go mod tidy -diff` sin diferencias; `make test-integration test-tiempos` → ok en todos los paquetes.
