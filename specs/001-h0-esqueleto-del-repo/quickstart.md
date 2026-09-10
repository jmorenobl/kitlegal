# Quickstart: validación de H0

**Fecha**: 2026-09-10 · **Spec**: [spec.md](./spec.md) · **Plan**: [plan.md](./plan.md)

Guía de validación de extremo a extremo del hito. Cada escenario corresponde a uno o varios criterios de
éxito del spec y **es ejecutable**: si el escenario no produce el resultado esperado, el hito no está
terminado.

No contiene código de implementación. Las formas y comportamientos concretos están en
[`contracts/cli-version.md`](./contracts/cli-version.md) y
[`contracts/make-targets.md`](./contracts/make-targets.md).

---

## Prerrequisitos

Una máquina limpia con exactamente dos cosas (SC-010):

- Go 1.21 o superior — `go version`
- `git` — `git --version`

Nada más. Ni `golangci-lint`, ni `govulncheck`, ni `gitleaks`, ni `lefthook`: los construye `go tool`
desde los módulos pinados en `tools/` la primera vez que se invocan (FR-042 d).

Tampoco hace falta instalar el parche exacto de Go, **ni importa cuál se tenga**, más nuevo o más viejo:
`go.mod` declara `toolchain go1.26.6` y el `Makefile` exporta `GOTOOLCHAIN` con ese valor, de modo que cada
orden se ejecuta con ese parche —el mismo que ejecuta la integración continua— y el go command lo descarga
y lo verifica solo si falta. Cualquier `go` ≥ 1.21 sirve. `make check-tools` comprueba que `go` y `git`
están y que el toolchain fijado es obtenible ([research.md D1 y D18](./research.md)).

```bash
git clone https://github.com/jmorenobl/kitlegal.git
cd kitlegal
```

> La **primera** ejecución compila las cuatro herramientas desde fuente y requiere red; tarda varios
> minutos. Las siguientes las sirve la caché de construcción de Go en segundos.

---

## Escenario 1 — Veredicto reproducible en local (US1 · SC-001)

```bash
make build
make test
make lint
make vuln
```

**Esperado**: las cuatro terminan con éxito, sin hallazgos y sin avisos. `make build` deja un ejecutable
en `bin/kitlegal`. `make test` ejecuta con detector de carreras y deja `coverage.out`.

---

## Escenario 2 — La orden agregada (US1 esc. 5 · SC-002)

```bash
git status --porcelain > /tmp/antes.txt
make ci
git status --porcelain > /tmp/despues.txt
diff /tmp/antes.txt /tmp/despues.txt
```

**Esperado**: `make ci` ejecuta sus ocho controles y termina con éxito; el `diff` es vacío — la orden que
emite el veredicto **no ha modificado nada** (M1 del contrato).

---

## Escenario 3 — El formato es un gate, no una corrección (US1 esc. 6 · FR-012)

```bash
# rompe el formato a propósito: solo espacios, sobre una línea que ya existe
sed -i.bak 's/^func main() {/func  main( )  {/' cmd/kitlegal/main.go && rm cmd/kitlegal/main.go.bak
git diff cmd/kitlegal/main.go     # confirma que el único cambio es de espaciado
make ci                           # debe FALLAR
git diff --stat cmd/kitlegal/main.go   # el fichero sigue como lo dejamos
make fmt                          # ahora sí lo corrige
git diff cmd/kitlegal/main.go     # vacío: `fmt` ha devuelto la línea a su forma canónica
make ci                           # vuelve a pasar
git checkout -- cmd/kitlegal/main.go   # deshacer (aquí ya es un no-op)
```

**Esperado**: el primer `make ci` falla nombrando `cmd/kitlegal/main.go` y **no** lo modifica; `make fmt`
sí lo corrige; el segundo `make ci` pasa. Si el primer `make ci` pasara, el control de formato no sería un
gate.

> **La mutación es de espaciado sobre código existente, y tiene que seguir siéndolo.** No vale añadir una
> declaración nueva al final del fichero (`printf 'package main\n\nfunc  sobra( )  {\n}\n' >> …`): eso
> introduce una segunda cláusula `package` —error de sintaxis, no defecto de formato— que `make fmt` no
> puede corregir, y aunque se escribiera sin ella, la función nueva quedaría sin usar y `unused` haría
> fallar `lint` en el segundo `make ci`. Con cualquiera de las dos variantes el escenario registraría un
> rojo falso: el segundo `make ci` no pasaría por un motivo que no es el formato.

---

## Escenario 4 — `kitlegal version` (US3 · SC-003)

```bash
make build
./bin/kitlegal version
echo "código de salida: $?"
git rev-parse HEAD
```

**Esperado**: tres líneas con versión, commit y fecha; código de salida `0`; el commit impreso **idéntico**
al de `git rev-parse HEAD`. Ninguno de los tres valores es `dev`, `none` ni `unknown` (FR-004).

```bash
./bin/kitlegal
echo "código de salida: $?"     # 2
./bin/kitlegal inventado
echo "código de salida: $?"     # 2
```

---

## Escenario 5 — Cobertura por encima del umbral (SC-011)

```bash
make test
go tool cover -func=coverage.out | tail -1
```

**Esperado**: el total supera el 70 % declarado en `codecov.yml`, **sin exclusiones ni excepciones**. La
única línea razonablemente no cubierta es la de `main()`, que solo llama a `run` y propaga su código.

---

## Escenario 6 — Órdenes con el objeto ausente (SC-012 · FR-011)

```bash
make test-integration ; echo "-> $?"   # 0, ejecuta el comando real sobre 0 tests
make test-e2e          ; echo "-> $?"  # 0, anuncia que los aporta H1
make schema-check      ; echo "-> $?"  # 0, anuncia que los aportan H4 y H11
make skills-sync       ; echo "-> $?"  # 0, anuncia que los aporta H5
make release           ; echo "-> $?"  # ≠ 0, "sin configurar hasta H6 (.goreleaser.yaml)"
```

**Esperado**: ninguna es silenciosa; las cuatro primeras nombran el objeto ausente y el hito que lo
aportará; `release` falla explícitamente.

---

## Escenario 7 — Prerrequisitos ausentes (FR-010 · SC-010)

```bash
env PATH=/usr/bin:/bin make build
```

**Esperado**: `check-tools` falla con un mensaje que nombra `go`, la versión mínima (1.21) y de dónde
obtenerlo. No un error opaco del intérprete de órdenes.

El segundo caso es el toolchain fijado, no el binario ausente. Se provoca sobrescribiendo en la línea de
órdenes la variable de la que sale el pin, sin tocar `go.mod` ni instalar otro Go:

```bash
make check-tools GO_TOOLCHAIN=go1.26.99   # parche inexistente
```

**Esperado**: falla nombrando el toolchain pedido, la directiva `toolchain` de `go.mod` de la que sale y el
error literal del go command (`toolchain not available`). No un fallo a mitad de un control posterior.

Y la comprobación complementaria —que es la que hace útil el pin—: **el entorno no puede debilitarlo**, y
el parche que cada quien tenga instalado es indiferente.

```bash
make check-tools                        # pasa
GOTOOLCHAIN=local make check-tools      # pasa igualmente
GOTOOLCHAIN=auto  make check-tools      # pasa igualmente
go version                              # puede ser cualquier go >= 1.21, no tiene que ser go1.26.6
make build && go version -m bin/kitlegal | head -2   # el binario sí lleva go1.26.6
```

**Esperado**: las tres invocaciones pasan, porque la asignación del `Makefile` prevalece sobre la variable
heredada del entorno. Que `go version` no diga `go1.26.6` **no** es un fallo: lo que tiene que decir
`go1.26.6` es el toolchain con el que se construye y se analiza, y eso se lee en el binario producido.
Si alguna de las tres fallara por la versión de Go instalada, el pin no estaría haciendo su trabajo
([research.md D1](./research.md)).

---

## Escenario 8 — Ganchos de pre-commit (FR-022)

Los dos bloques de órdenes de este escenario se ejecutan **en la misma sesión de intérprete**: el segundo
deshace anclado a la revisión que anota el primero.

```bash
make hooks
antes=$(git rev-parse HEAD)   # ancla del deshacer: el estado exacto de partida
# un cambio de contenido que sobrevive al formateo…
printf '\n// Comprobación del gancho de pre-commit.\n' >> cmd/kitlegal/main.go
# …y, sobre él, el mismo defecto de espaciado del escenario 3
sed -i.bak 's/^func main() {/func  main( )  {/' cmd/kitlegal/main.go && rm cmd/kitlegal/main.go.bak
git add cmd/kitlegal/main.go
git commit -m "test: comprobar el gancho"
git show -p HEAD -- cmd/kitlegal/main.go   # el commit: solo la línea de comentario
git diff HEAD -- cmd/kitlegal/main.go      # vacío: lo confirmado es lo que hay en el árbol
```

**Esperado**: el gancho ejecuta `make fmt`, devuelve `func main() {` a su forma canónica, vuelve a preparar
el fichero y el commit sale adelante conteniendo **solo** la línea de comentario, ya formateada — sin el
espaciado roto. El gancho **no** es la autoridad final: la CI ejecuta los mismos controles.

> **Por qué hacen falta las dos mutaciones.** El defecto de espaciado es lo que el gancho tiene que
> corregir, pero por sí solo no basta: al formatearlo, el fichero vuelve a ser idéntico a `HEAD`, el índice
> se queda sin cambios y `git commit` aborta («nothing to commit»), con lo que no habría commit que
> inspeccionar. La línea de comentario aporta el cambio de contenido que sobrevive al formateo. Y por lo
> mismo que en el escenario 3, ninguna de las dos puede ser una declaración nueva: rompería la sintaxis o
> caería en `unused`, y entonces el gancho fallaría en `make fmt` o en `make lint-fast` y el commit no
> saldría por un motivo ajeno a lo que se está comprobando.

```bash
git reset --soft "$antes"          # retira el commit de prueba y NO toca índice ni árbol
git restore --source="$antes" --staged --worktree -- cmd/kitlegal/main.go   # solo el fichero mutado
git rev-parse HEAD                                   # debe imprimir el mismo valor que $antes
git diff --stat "$antes" -- cmd/kitlegal/main.go     # vacío: el fichero vuelve a su contenido de partida
git diff --cached --stat "$antes" -- cmd/kitlegal/main.go   # vacío: el índice también
unset antes
```

> **El deshacer va anclado a `$antes`, nunca a `HEAD~1`, y es quirúrgico, nunca `--hard`.** Dos exigencias
> distintas:
>
> 1. **Anclado, no relativo.** El `HEAD~1` a ciegas asume que el commit de prueba llegó a existir; si el
>    gancho lo impidió —que es justo uno de los desenlaces que este escenario puede producir—, borraría el
>    commit de la tarea anterior y sus ficheros versionados. La pareja `reset --soft "$antes"` +
>    `restore --source="$antes"` es idempotente: deja el historial y `cmd/kitlegal/main.go` en el estado de
>    partida tanto si el commit se creó como si no. Si la variable no estuviera definida (sesión distinta),
>    las órdenes fallan sin tocar nada y hay que recuperar la revisión de partida antes de repetir.
> 2. **Quirúrgico, no `--hard`.** `git reset --hard "$antes"` descartaría **toda** modificación sin
>    confirmar de **cualquier** fichero versionado, no solo el commit de prueba. Bajo el workflow `hito` eso
>    es destructivo: el paso que abre cada tarea escribe `gates/tarea-actual.json` y
>    `gates/tareas-intentos.json` —ficheros versionados, confirmados después por el paso de commit— y en
>    este punto están modificados y sin confirmar. Un `--hard` los devolvería al contenido de la tarea
>    anterior, y el guardián de diff leería entonces el id y las rutas de esa tarea anterior y rechazaría
>    `lefthook.yml` y `tools/lefthook/` como ficheros fuera de rutas. `--soft` mueve solo `HEAD`; el
>    `git restore` acotado por ruta toca solo el fichero que este escenario mutó.
>
> Por lo mismo, la comprobación final va **acotada a `cmd/kitlegal/main.go`**: mientras el directorio del
> feature tenga cambios versionados sin confirmar —y bajo el workflow siempre los tiene durante una tarea—,
> ni `git diff --stat "$antes"` a secas ni `git status --porcelain` estarán vacíos, y perseguir ese vacío
> con un `git checkout .` o un `git clean` destruiría el trabajo de la propia tarea. Lo que este escenario
> garantiza, y lo único que hay que comprobar, es que `HEAD` vuelve a `$antes` y que el fichero mutado
> vuelve a su contenido de partida.
>
> Terminar sin deshacer tampoco es opción: `cmd/kitlegal/main.go` no está declarado en T003 y el guardián
> de diff rechazaría la tarea.

---

## Escenario 9 — Detección de secretos (SC-009 · US2 esc. 3)

El token de prueba **se genera en el momento y no aparece escrito en esta guía**, por dos motivos
independientes: un literal aquí sería un secreto detectable dentro de un fichero versionado (nota final del
escenario), y los literales «de manual» —cadenas como `0123456789abcdefghijklmnopqrstuvwxyz`— caen en la
lista global de *stopwords* de gitleaks, que descarta todo hallazgo cuyo secreto contenga una de ellas: la
regla casaría, pero no se reportaría nada, `make secrets` pasaría y el escenario registraría un **rojo
falso** para SC-009.

```bash
# 36 caracteres alfanuméricos aleatorios: cumplen `ghp_[0-9a-zA-Z]{36}`, superan la
# entropía mínima de la regla y no pueden contener una stopword. El valor no se imprime.
tok=$(LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 36)
printf '\n// Token de prueba del escenario 9, se restaura al final: ghp_%s\n' "$tok" >> cmd/kitlegal/main.go
unset tok
make secrets      # debe FALLAR
make ci           # debe FALLAR también, porque secrets forma parte de ci
git restore -- cmd/kitlegal/main.go
git diff --stat -- cmd/kitlegal/main.go   # vacío: el token no queda en el árbol
```

> **La comprobación va acotada al fichero mutado**, por lo mismo que el deshacer del escenario 8: durante
> T013 el directorio del feature tiene cambios versionados sin confirmar (el registro de la validación que
> se está escribiendo, más el estado de la tarea), así que `git diff --stat` a secas nunca estará vacío
> bajo el workflow `hito`. Lo que hay que demostrar aquí es que el token no sobrevive en
> `cmd/kitlegal/main.go`.

**Esperado**: ambas fallan, y `make ci` falla **en `secrets`**, después de haber pasado `fmt-check`,
`lint`, `test`, `vuln` y `schema-check`. Que falle también `make ci` es lo que hace cierta SC-004: el
veredicto local predice el de la PR. Si `make secrets` pasara, hay que restaurar el fichero, generar otro
token y repetir —nunca sustituirlo por un literal fijo—; si volviera a pasar, el fallo está en la
configuración del control, no en el escenario.

> **Por qué una línea de comentario y no una declaración.** Un `const t = "ghp_…"` añadido al final del
> fichero rompe el escenario dos veces antes de llegar a `secrets`: gofmt exige una línea en blanco entre
> declaraciones de distinto tipo, así que `fmt-check` —el primer control de `ci`— fallaría por formato, y
> aunque se añadiera ya formateado, la constante quedaría sin usar y `unused` haría fallar `lint`. En
> ambos casos `make ci` moriría en un control anterior y el escenario no demostraría nada sobre la
> detección de secretos. El comentario es inerte para el formato y para el lint —como el del escenario
> 8— y gitleaks analiza el texto del fichero, no su árbol sintáctico.

> **Ningún fichero versionado puede contener un token detectable.** `make secrets` es
> `gitleaks dir . --redact --no-banner`: recorre **todo** el árbol, `specs/` incluido, y H0 no crea
> `.gitleaksignore` («Notas» de `tasks.md`). Un token literal en esta guía, en `README.md` o en los
> registros de `gates/` dejaría `make ci` en rojo de forma permanente y no por un control, sino por el
> propio documento. De ahí que el token se genere al vuelo, que `--redact` mantenga su valor fuera de la
> salida de gitleaks y que T013 lo excluya del registro de la validación.

---

## Escenario 10 — Dependencias saneadas (US2 esc. 4)

```bash
go mod edit -require=github.com/stretchr/testify@v1.11.1
make mod-tidy-check   # debe FALLAR señalando la diferencia
make ci               # debe FALLAR también
go mod edit -droprequire=github.com/stretchr/testify
make mod-tidy-check   # vuelve a pasar
```

---

## Escenario 11 — Criterio de aceptación literal del hito (SC-005 · SC-006)

Se comprueba con **dos propuestas de cambio desechables**, que no se integran en `main`.

Tres precondiciones que hacen falsable el criterio y son fáciles de romper por descuido:

1. **Las dos ramas salen de la rama del hito**, no una de otra. Si `prueba-limpia` se creara estando en
   `prueba-forbidigo`, heredaría `internal/prueba/p.go` y fallaría en lint: SC-006 quedaría registrado
   como rojo sin que nada esté mal. De ahí el `git checkout h0-esqueleto-del-repo` antes de cada rama, y de
   ahí que la comprobación del árbol vaya **acotada a `internal/`**: bajo el workflow `hito` el directorio
   del feature siempre tiene cambios sin confirmar, así que un `git status --porcelain` a secas nunca sale
   vacío y no dice nada sobre lo que aquí importa, que es que `internal/prueba/p.go` no ha viajado.
2. **Hay que abrir la propuesta de cambio.** `ci.yml` se dispara por `pull_request` y por push a la rama
   principal (FR-025); un `git push` a una rama lateral no ejecuta ningún flujo. Sin PR abierta no hay
   veredicto de plataforma que medir, y SC-005/SC-006 hablan literalmente de la propuesta de cambio.
3. **Las ramas desechables no arrastran nada más que su propio fichero.** Cada commit de prueba prepara
   **solo el fichero que le corresponde** —`git add internal/prueba/p.go` en la sucia, `git add CHANGELOG.md`
   en la limpia—, nunca `git add -A`. Bajo el workflow `hito`, cuando este escenario se ejecuta el
   directorio del feature tiene cambios versionados sin confirmar (`gates/tarea-actual.json`,
   `gates/tareas-intentos.json` y cualquier avance de `gates/verificacion-pr.md`); un `git add -A` los
   confirmaría **en la rama desechable**, y al volver a la rama del hito revertirían a `HEAD` y el
   `git branch -D` final se llevaría el único ejemplar del registro, dejando además el commit de T014
   etiquetado con el id de la tarea anterior. Por lo mismo, **el registro `gates/verificacion-pr.md` se
   escribe en la rama del hito y solo después de cerrar las propuestas y borrar las dos ramas**. (Como
   alternativa equivalente, crear cada rama en un `git worktree add` temporal: así el árbol de la rama del
   hito no cambia en ningún momento.)

**PR sucia** — debe quedar bloqueada:

```bash
git checkout h0-esqueleto-del-repo
git checkout -b prueba-forbidigo
mkdir -p internal/prueba
printf 'package prueba\n\nimport "fmt"\n\nfunc P() { fmt.Println("no") }\n' > internal/prueba/p.go
make lint     # debe FALLAR con forbidigo
git add internal/prueba/p.go                                  # solo este fichero, nunca `git add -A`
git commit --no-verify -m "test: comprobar forbidigo"         # el gancho local no es lo que se mide aquí
git push -u origin prueba-forbidigo
gh pr create --draft --base h0-esqueleto-del-repo --head prueba-forbidigo \
  --title "test: comprobar forbidigo (desechable)" \
  --body "PR desechable del escenario 11 de H0. No se integra."
gh pr checks prueba-forbidigo --watch   # debe terminar en fallo
```

**Esperado**: el control de lint falla, en local y en la PR, señalando `fmt.Println` en
`internal/prueba/p.go`. El mismo `fmt.Println` en `cmd/` **no** falla: la regla está acotada por ruta
(FR-014).

> **Por qué `--no-verify` en este commit, y solo en este.** El gancho de pre-commit instalado en T003
> ejecuta `make lint-fast`, que es exactamente el control que este fichero está diseñado para infringir: sin
> `--no-verify` el commit no saldría, no habría push ni propuesta de cambio, y SC-005 —que habla del
> veredicto **de la plataforma**— quedaría sin medir. La autoridad que este escenario mide es la CI, no el
> gancho; que el gancho también lo detecte ya está comprobado en el escenario 8 y anotado en
> `CONTRIBUTING.md`. El commit de la rama limpia **no** lleva `--no-verify`: allí el gancho debe pasar.

**PR limpia** — debe pasar en menos de 3 minutos:

```bash
git checkout h0-esqueleto-del-repo   # imprescindible: no partir de prueba-forbidigo
git checkout -b prueba-limpia
git status --porcelain -- internal/   # vacío: la rama no hereda internal/prueba/p.go
# un cambio trivial que no viola nada: una línea en CHANGELOG.md
printf '\n<!-- comprobación desechable del escenario 11 de H0 -->\n' >> CHANGELOG.md
git add CHANGELOG.md                 # solo este fichero, nunca `git add -A`
git commit -m "docs: comprobar el flujo"   # sin `--no-verify`: aquí el gancho debe pasar
git push -u origin prueba-limpia
gh pr create --draft --base h0-esqueleto-del-repo --head prueba-limpia \
  --title "docs: comprobar el flujo (desechable)" \
  --body "PR desechable del escenario 11 de H0. No se integra."
gh pr checks prueba-limpia --watch   # debe terminar en verde
```

**Esperado**: todos los controles pasan. **Anotar la duración del flujo en dos escenarios** y dejar ambos
números en la PR del hito (obligación que *Assumptions* del spec traslada al plan):

| Escenario | Cómo provocarlo | Número |
|---|---|---|
| **En caliente** (el normal) | Caché de dependencias y herramientas poblada | debe ser < 3 min (SC-006) |
| **En frío** | Invalidar la caché del flujo (cambiar la clave o borrarla desde la interfaz de Actions) | se registra, sin umbral |

Si el escenario en caliente superara los 3 minutos, se corrige el **diseño del flujo** —partir `make ci`
en jobs paralelos, cada uno invocando una orden del `Makefile`— nunca el criterio
([research.md D12](./research.md)).

En esta misma PR limpia se comprueban dos cosas más de la cobertura, ambas de plataforma y no de
repositorio ([research.md D15](./research.md)):

- El paso de subida a Codecov **termina bien**. Con `fail_ci_if_error: true` y sin el secreto
  `CODECOV_TOKEN` dado de alta, el flujo fallaría por la subida y no por un control: hay que dar de alta el
  secreto antes, no rebajar la bandera.
- El estado del componente `internal/core` **no** aparece como fallo ni como pendiente que bloquee la
  fusión, dado que en H0 su `paths` no casa con ningún fichero.

Al terminar, **cerrar las dos propuestas de cambio sin integrarlas y borrar ambas ramas**, local y remota:

```bash
gh pr close prueba-forbidigo --delete-branch
gh pr close prueba-limpia    --delete-branch
git checkout h0-esqueleto-del-repo
git branch -D prueba-forbidigo prueba-limpia
```

**Y solo entonces**, ya de vuelta en la rama del hito y con las dos ramas borradas, se escribe el registro
`gates/verificacion-pr.md` con los dos números y los cuatro resultados. Mientras las ramas desechables
existan, cualquier cosa escrita en el directorio del feature corre el riesgo de acabar confirmada en una de
ellas y desaparecer con el `git branch -D`: los números se anotan fuera del repositorio (o se releen de las
propuestas antes de cerrarlas) y se vuelcan al registro al final.

---

## Escenario 12 — Documentación fundacional (US4 · SC-008)

```bash
head -2 LICENSE                    # Apache License, Version 2.0
grep -c "Unreleased" CHANGELOG.md  # ≥ 1
ls docs/ADR/                       # los cuatro ADR
```

Y una lectura, no un comando: `README.md` y `CONTRIBUTING.md` deben bastar para que alguien que llega por
primera vez construya el proyecto y ejecute todos sus controles sin preguntar a nadie (SC-010). Si al
seguir esta guía hubo que consultar algo que no está en esos dos ficheros, falta en ellos.

---

## Resumen de trazabilidad

| Escenario | Criterios de éxito |
|---|---|
| 1 | SC-001 |
| 2 | SC-002 |
| 3 | SC-002, SC-007 (formato) |
| 4 | SC-003 |
| 5 | SC-011 |
| 6 | SC-012 |
| 7 | SC-010 |
| 8 | SC-007 (pre-commit) |
| 9 | SC-004, SC-009 |
| 10 | SC-004, SC-009 |
| 11 | SC-005, SC-006 |
| 12 | SC-008, SC-010 |

SC-004 se demuestra de forma acumulada: los escenarios 3, 9 y 10 muestran que un control que falla en la
PR falla también en local, porque es literalmente la misma orden.
