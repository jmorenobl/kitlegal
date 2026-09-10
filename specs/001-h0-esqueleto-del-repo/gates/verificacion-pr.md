# Verificación en la plataforma · escenario 11 (SC-005 · SC-006)

**Fecha**: 2026-09-10 · **Hito**: H0 · **Tarea**: T014 · **Rama**: `h0-esqueleto-del-repo`
**Intentos**: 1 (bloqueo detectado) · 2 (bloqueo reverificado; fixture del escenario corregido)

## Veredicto

**No ejecutado: falta un prerrequisito humano.** El repositorio **no tiene remoto en la plataforma**, de
modo que no hay push, no hay propuesta de cambio y `ci.yml` no se dispara. Los cuatro resultados y los dos
números que esta tarea debe registrar quedan **pendientes**, sin simular ninguno.

Lo que sí era verificable sin plataforma y sin red —la mitad local del control de lint que el escenario
mide— se ejecutó y está registrado abajo, junto con una discrepancia real entre el guion del escenario y su
propio fixture.

## Hecho que bloquea: no hay remoto

```console
$ git remote -v
            # sin salida: ningún remoto configurado
$ git branch -a
* h0-esqueleto-del-repo
  main        # solo ramas locales; ninguna remota
```

La herramienta de la plataforma sí está instalada y con sesión abierta (`gh auth status` → `github.com`,
cuenta `jmorenobl`, protocolo ssh), pero eso no crea el repositorio remoto. Sin `origin` no hay
`git push -u origin …`, y sin push no hay `gh pr create`.

**Por qué no se resuelve desde aquí.** Crear el repositorio en la plataforma es una acción hacia fuera y
una decisión de alcance (visibilidad, propietario, nombre, protecciones de rama); la propia tarea la
clasifica como prerrequisito humano y no como artefacto del repositorio. Se deja sin hacer a propósito.

## Los cuatro resultados y los dos números

| Qué | Criterio | Estado |
|---|---|---|
| PR sucia bloqueada por lint | SC-005 | **PENDIENTE** · no hay remoto donde abrirla |
| PR limpia en verde | SC-006 | **PENDIENTE** · no hay remoto donde abrirla |
| Subida de cobertura termina bien (`CODECOV_TOKEN`) | research.md D15 | **PENDIENTE** · depende del remoto y del secreto |
| `internal/core` no figura como fallo ni como pendiente bloqueante | research.md D15 | **PENDIENTE** · depende del remoto |
| Duración **en caliente** | < 3 min (SC-006) | **SIN MEDIR** |
| Duración **en frío** | se registra, sin umbral | **SIN MEDIR** |

Ninguna de estas casillas admite un valor obtenido en local: SC-005 y SC-006 hablan literalmente del
veredicto **de la propuesta de cambio**, y la duración es la del flujo en los ejecutores de la plataforma,
con su caché. Escribir aquí un número de `make ci` local sería medir otra cosa.

## Dependabot y la directiva `toolchain`: tampoco observable todavía

La pregunta «¿la actualización automática de dependencias propone también subir la directiva `toolchain`?»
solo se responde viendo ejecutarse a Dependabot, que es un servicio de la plataforma. Lo que consta en el
repositorio:

- `go.mod` declara `go 1.26.0` y `toolchain go1.26.6`.
- `.github/dependabot.yml` tiene seis entradas (módulo raíz, los cuatro módulos de herramienta y las
  acciones), todas `gomod`/`github-actions` semanales. Ninguna menciona la directiva `toolchain`.
- `CONTRIBUTING.md` §«Subir el parche de Go (directiva `toolchain`)» describe la subida como un
  procedimiento **manual** cuyo disparador es un hallazgo de `make vuln` en la biblioteca estándar.

Es decir: el repositorio **asume que la subida es manual**. Queda por confirmar en la plataforma si
Dependabot la propone además por su cuenta —y, si lo hiciera, si ese procedimiento manual sobra o se
solapa—. **No se da por buena ninguna de las dos hipótesis aquí.**

## Lo que sí quedó verificado en local (sin red)

La mitad local del control que el escenario mide: el lint prohíbe la escritura directa a la salida estándar
bajo `internal/**` y la permite bajo `cmd/**` (FR-014). Fixture creado y **borrado** al terminar; el árbol
quedó como estaba.

**Bajo `internal/` falla:**

```console
$ make lint
internal/prueba/p.go:7:12: use of `fmt.Println` forbidden because "la salida se emite por el escritor
que recibe la función; en H1 solo internal/render escribe en stdout" (forbidigo)
1 issues:
* forbidigo: 1
make: *** [lint] Error 1
```

**El mismo `fmt.Println` bajo `cmd/kitlegal/` no falla:**

```console
$ make lint      # con fmt.Println("no") dentro de run(), en cmd/kitlegal/main.go
0 issues.
```

La acotación por ruta de `.golangci.yml` (`path-except: ^internal/`) hace lo que dice. Ambos cambios se
revirtieron: `git checkout -- cmd/kitlegal/main.go` y borrado de `internal/prueba/`.

## Discrepancia encontrada: el fixture del escenario enmascara su propio mensaje

El fixture que el guion del escenario 11 dicta literalmente

```bash
printf 'package prueba\n\nimport "fmt"\n\nfunc P() { fmt.Println("no") }\n' > internal/prueba/p.go
```

**no produce el hallazgo de `forbidigo` que el apartado «Esperado» anuncia.** Produce dos de `revive`:

```console
$ make lint
internal/prueba/p.go:1:1: package-comments: should have a package comment (revive)
internal/prueba/p.go:5:1: exported: exported function P should have comment or be unexported (revive)
2 issues:
* revive: 2
```

**Causa, aislada hasta el final.** El procesador `uniq-by-line` de golangci-lint, activo por omisión, deja
**un único hallazgo por línea**. El `exported` de `revive` cae en la línea 5 (`func P() { … }`) y el de
`forbidigo` también (columna 12 de esa misma línea): el primero gana y el segundo se descarta. Comprobado
por bisección con configuraciones de un solo uso, todas fuera del árbol versionado:

| Configuración | Resultado |
|---|---|
| Solo `forbidigo`, sin exclusiones | `forbidigo: 1` — la regla casa |
| Solo `forbidigo`, **con** el `path-except: ^internal/` del proyecto | `forbidigo: 1` — la exclusión por ruta no es la causa |
| `default: standard` + `forbidigo` | `forbidigo: 1` |
| `default: standard` + `forbidigo` + **`revive`** | `revive: 2`, `forbidigo` desaparece ← reproducido |
| Lo anterior + `issues.uniq-by-line: false` | `forbidigo: 1` **y** `revive: 2` ← confirmado |

Añadir al fixture el comentario de paquete y el de la función exportada —dejando la escritura prohibida
como único defecto— devuelve el hallazgo esperado, y es así como se obtuvo la salida de la sección
anterior.

**Qué significa para SC-005 y qué no.**

- El control **no está roto** y no se ha tocado nada de `.golangci.yml`: `forbidigo` bloquea lo que debe
  bloquear bajo `internal/**`.
- La PR sucia **seguiría quedando bloqueada** aun con el fixture literal, porque el lint termina en error
  igualmente (código 1, 2 hallazgos). SC-005 —«la propuesta queda bloqueada por lint»— se cumpliría.
- Lo que **no** se cumpliría es la frase del «Esperado» del escenario: el fallo **no** señalaría
  `fmt.Println`, sino dos comentarios ausentes. Un escenario que se da por bueno leyendo «falla el lint»
  sin mirar qué regla falló dejaría `forbidigo` sin comprobar nunca, que es precisamente lo que FR-014
  quiere asegurar.

**Corregido en el intento 2.** El intento 1 tuvo que dejar esto anotado sin arreglarlo porque
`quickstart.md` no figuraba entre sus rutas declaradas; el intento 2 **sí** lo declara, así que el guion del
escenario 11 queda corregido en el propio `quickstart.md`:

- el `printf` del fixture ahora lleva comentario de paquete y de función exportada, de modo que la escritura
  prohibida es su **único** defecto;
- el paso pasa de `# debe FALLAR con forbidigo` a `# debe FALLAR, y el hallazgo debe ser forbidigo: 1`;
- el apartado «Esperado» incorpora una nota que explica el enmascaramiento por `uniq-by-line` y advierte de
  que **no basta con que el lint falle: hay que leer qué regla falló**.

Reproducido y verificado en local antes y después del cambio, sin red y sin tocar `.golangci.yml`:

```console
$ # fixture literal del intento 1 (el que el guion dictaba)
$ make lint
internal/prueba/p.go:1:1: package-comments: should have a package comment (revive)
internal/prueba/p.go:5:1: exported: exported function P should have comment or be unexported (revive)
2 issues:
* revive: 2          # ← ni rastro de forbidigo: el «Esperado» no se cumplía

$ # fixture corregido, ya en quickstart.md
$ make lint
internal/prueba/p.go:7:12: use of `fmt.Println` forbidden because "la salida se emite por el escritor
que recibe la función; en H1 solo internal/render escribe en stdout" (forbidigo)
1 issues:
* forbidigo: 1       # ← el hallazgo que FR-014 quiere asegurar
```

El fixture se borró al terminar (`internal/prueba/` inexistente, `git status --porcelain -- internal/`
vacío). Con esto, cuando exista el remoto el escenario 11 se ejecuta tal cual está escrito y mide de verdad
lo que dice medir.

## Qué hace falta para cerrar esta tarea

Dos prerrequisitos humanos, en este orden:

1. **Crear el repositorio en la plataforma y darlo de alta como `origin`** de este clon. Sin esto no hay
   propuesta de cambio y `ci.yml` —que se dispara por `pull_request` y por push a `main`— no se ejecuta
   nunca. Conviene que la rama base de las PR de prueba sea `h0-esqueleto-del-repo`, como dice el guion.
2. **Dar de alta el secreto `CODECOV_TOKEN`** en el repositorio. El paso de subida de `ci.yml` lleva
   `fail_ci_if_error: true`; sin el secreto el flujo fallará por la subida y no por un control, y el
   veredicto de SC-006 no significaría nada.

Hecho eso, se ejecuta el escenario 11 tal cual está escrito —dos ramas desechables salidas **cada una** de
la rama del hito, `git add` de su único fichero, `--no-verify` solo en la sucia— y se rellenan arriba los
cuatro resultados y los dos números, ya con las propuestas cerradas sin integrar y las dos ramas borradas
en local y en el remoto.

## Lo que no se hizo, a propósito

- **No se creó el repositorio remoto.** Es una acción hacia fuera y una decisión de alcance ajena al hito.
- **No se rebajó `fail_ci_if_error`** ni ninguna otra bandera para que el flujo pasara sin el secreto. El
  comentario de `ci.yml` ya advierte que la ausencia del secreto se resuelve dándolo de alta.
- **No se inventó ningún resultado ni ninguna duración.** Las seis casillas siguen vacías.
- **No se modificó `.golangci.yml`** al encontrar la discrepancia: el defecto está en el fixture del guion,
  no en la configuración.

## Estado del árbol

El único cambio que esta tarea aporta es este registro. Comprobado ruta a ruta al terminar:
`git diff --stat -- cmd/ internal/ .golangci.yml` vacío, `internal/prueba/` inexistente y ningún fichero
sin seguimiento. No se exige un árbol globalmente limpio: bajo el workflow `hito` nunca lo está, porque
`gates/tarea-actual.json` y `gates/tareas-intentos.json` son cambios sin confirmar del propio bucle.
