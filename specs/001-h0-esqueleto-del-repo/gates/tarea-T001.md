# T001 — resolución

**Fecha**: 2026-09-10 · **Intento**: 1 de 3 (tras el rearranque del bucle) ·
**Estado de la tarea**: `[X]` — `make ci` en verde, exit 0, sin mutar el árbol.

Este fichero sustituye a la nota de bloqueo del run anterior. Se conserva porque documenta la escalada
que exigía el §4 de la constitución y cómo se cerró.

## Qué bloqueaba y cómo se desbloqueó

El run anterior dejó T001 sin marcar con siete de los ocho controles de `make ci` en verde: `secrets`
fallaba por falsos positivos en documentación de terceros versionada antes del hito, y la corrección
especificada (`.gitleaksignore` en la raíz, huella a huella) no estaba entre las rutas declaradas de la
tarea. Era un problema de delimitación, no de implementación, así que se escaló en lugar de esquivarlo.

Resuelto fuera de la tarea, en los tres commits previos a este run:

1. `1d2239e` — el extractor de rutas del workflow admite dotfiles de raíz sin extensión, de modo que
   `.gitleaksignore` **se puede declarar**.
2. `00494e5` — se elimina la copia duplicada `agent/` del paquete de skills. Los hallazgos pasan de
   cuatro a dos.
3. `73ee4e1` — decisión humana registrada: H0 entrega `.gitleaksignore` con las dos huellas
   justificadas; `tasks.md`, `plan.md`, `research.md` D10 y `data-model.md` quedan alineados.

Con la ruta ya declarada en `gates/tarea-actual.json`, T001 crea el fichero con las dos huellas y su
comentario justificativo, sin desactivar ninguna regla ni excluir ninguna otra ruta.

## Cambio adicional de este run

`go.mod` llevaba una directiva `ignore agent` que el run anterior había añadido para sacar del módulo los
ficheros Go de ejemplo del árbol duplicado `agent/`. Ese árbol ya no existe (`00494e5`) y `.agents/`, al
empezar por punto, es invisible para el go command, así que la directiva quedaba muerta. Se retira:
`go.mod` queda con la ruta de módulo, `go 1.26.0`, `toolchain go1.26.6` y ninguna dependencia de producto,
que es exactamente lo que fija T001.

## Veredicto control a control (`make ci`, exit 0)

| Control | Orden | Resultado |
|---|---|---|
| Formato | `fmt-check` | ✅ sin diferencias |
| Lint (22 linters de FR-013 + forbidigo, gosec incluido) | `lint` | ✅ `0 issues` |
| Tests con detector de carreras | `test` | ✅ `ok`, cobertura **83,3 %** (umbral global de FR-029: 70 %) |
| Vulnerabilidades | `vuln` | ✅ `No vulnerabilities found.` |
| Esquemas | `schema-check` | ✅ marcador verde, exit 0 |
| Secretos | `secrets` | ✅ `no leaks found` |
| Integridad de módulos | `mod-verify` | ✅ raíz + los tres módulos de herramienta descubiertos por glob |
| Dependencias saneadas | `mod-tidy-check` | ✅ sin diferencias |

Comprobado que `make ci` **no muta el árbol**: el estado de `git status --porcelain` es idéntico antes y
después de ejecutarla.

Comprobado además fuera de `ci`: `build` e `install` (mismos LDFLAGS, `CGO_ENABLED=0`), el binario
construido imprime los tres datos inyectados y termina con 0, `test-integration` (etiqueta `integration`,
conjunto vacío, éxito), `test-e2e` y `skills-sync` (marcador que nombra el objeto ausente y el hito que lo
aporta), `lint-fast`, `release` (exit ≠ 0, FR-011), `help` (las 20 órdenes) y el camino de fallo de
`check-tools` (`make check-tools GO_TOOLCHAIN=go1.26.99` falla nombrando el toolchain, la directiva de
`go.mod` y el error literal del go command). Los tres modfiles de herramienta están saneados
(`go mod tidy -diff` sin diferencias en cada uno).

`hooks` no se puede ejecutar todavía: necesita `tools/lefthook/`, que aporta T003. `mod-verify` ya lo
cubrirá sin tocar el `Makefile`, porque descubre los modfiles por glob sobre `tools/*/go.mod`.

## Las dos huellas excluidas

Ninguna es un secreto: son líneas del bloque «Bad» de un documento que enseña **qué no hacer** con las
credenciales, marcadas `// DON'T` en el propio fichero, dentro de las skills de agente instaladas
(`samber/cc-skills-golang`, en `.agents/`, versionadas en el commit `812a7d2`, anterior al hito).

| Huella (`Fingerprint`) |
|---|
| `.agents/skills/golang-security/references/secrets.md:generic-api-key:31` |
| `.agents/skills/golang-security/references/secrets.md:generic-api-key:37` |

Se anotan solo las huellas. Los valores literales **no** se transcriben: este fichero también lo recorre
`gitleaks dir .`, y copiarlos convertiría el registro en un hallazgo nuevo.
