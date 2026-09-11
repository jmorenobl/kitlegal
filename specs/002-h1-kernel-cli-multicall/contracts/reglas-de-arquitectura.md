# Contrato: reglas de arquitectura ejecutables

Las cinco reglas de dependencia de `docs/ROADMAP.md` §2 y de la constitución §IV, y **cómo se comprueba
cada una**. Es un contrato hacia quien añade código al proyecto —persona o agente—: quien intenta
saltárselas recibe un fallo, no un aviso en una revisión.

**Requisitos que lo definen**: FR-050 … FR-053, FR-035, FR-040. **Criterio**: SC-008.

---

## 1. Las cinco reglas

| # | Regla | Estado en H1 |
|---|---|---|
| R1 | `internal/core/**` no importa `internal/{source,httpx,cache,store,graph,render,cli,app}` **ni I/O de la biblioteca estándar** (`log`, `log/slog`, `os`, `io`, `net/http`, `database/sql`) | **Activa**: existe `internal/core/schema`, que importa solo `crypto/sha256`, `encoding/hex`, `encoding/json` y `time` |
| R2 | Solo `internal/httpx` importa `net/http` | **Activa y vacía**: `httpx` no existe, luego **nadie** puede importarlo |
| R3 | Solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | **Activa y vacía**: ninguno de los tres existe |
| R4 | Solo `internal/cli` y `cmd/` llaman a `os.Exit` | **Activa, y en forma más estricta que la regla**: `os.Exit` aparece **dos veces**, una en cada raíz de composición (`cmd/kitlegal/main.go` y el `main` del binario de e2e); ni siquiera `internal/cli` lo usa |
| R5 | Solo `internal/render` escribe en la salida estándar; los logs van a la de error | **Activa en su forma completa**: desaparece la excepción que H0 dio a `cmd/` para `fmt.Print*`; `os.Stdout` y `os.Stderr` solo se nombran en las dos raíces de composición, que los inyectan |

H0 acotó R5 por ruta como medida provisional (prohibir `fmt.Print*` bajo `internal/**`, dejando `cmd/`
libre) porque `internal/render` no existía. **H1 sustituye esa forma acotada por la regla completa**
(FR-052).

**R1 crece hacia la biblioteca estándar.** El dominio se declara «puro, sin I/O» desde la constitución
§IV; en H1 eso deja de ser una afirmación y pasa a ser una lista denegada. Sin ella, un `*slog.Logger` o
un `io.Writer` podrían instalarse en `internal/core` sin que ningún control dijera nada, porque no son
paquetes internos.

**Las dos raíces de composición.** Un `package main` no puede propagar un código de salida sin `os.Exit`
—retornar de `main` sale siempre con 0— ni inyectar descriptores sin nombrarlos. El binario de e2e
(`internal/app/testdata/kitlegal-e2e`) es un `package main` legítimo de este hito y es, línea por línea,
el mismo `main()` que el distribuido salvo el registro de applets: se le trata como lo que es, una raíz de
composición, y se le exceptúa **por ruta exacta**. No se exceptúa `internal/app/testdata/ejemplo`, que es
código de applet y sigue sujeto a las tres prohibiciones, ni `testdata/**` en general.

---

## 2. Dónde se comprueba cada una

Cada regla se comprueba en la capa más baja que puede verla, y las que son de importación se comprueban
**dos veces** en capas independientes: FR-051 exige que la garantía no dependa solo de la configuración
del lint.

| Regla | `depguard` (lint) | `forbidigo` (lint) | `internal/arch_test.go` (test) |
|---|---|---|---|
| R1 | ✔ lista `core` (internos + I/O de la biblioteca estándar) | — | ✔ grafo transitivo real |
| R2 | ✔ lista `red` | — | ✔ |
| R3 | ✔ lista `sql` | — | ✔ |
| R4 | — | ✔ `^os\.Exit$`, marca `R4:`, exceptuando solo las dos raíces de composición | — |
| R5 | — | ✔ `^fmt\.Print(\|f\|ln)$` (marca `R5:`, **sin excepción alguna**) y `^os\.Stdout$` / `^os\.Stderr$` (marca `R5-descriptores:`, exceptuando solo las dos raíces de composición) | — |

Las excepciones se escriben con `path` + `text` sobre la marca de la regla, de modo que exceptuar
`os.Exit` en una raíz de composición **no puede** desactivar por accidente la prohibición de `fmt.Print*`,
que lleva otra marca. La marca cumple además FR-053: el fallo nombra la regla violada.

**Por qué R4 y R5 solo en el lint**: son reglas de **símbolo**, no de importación. Todo paquete importa
`os`; lo prohibido es llamar a `os.Exit` o referenciar `os.Stdout`, y eso el grafo de dependencias no lo
ve. `forbidigo` sí, con el análisis de tipos activado —sin él, el filtro por paquete se ignora y un alias
de importación esquivaría el patrón.

**Por qué R1-R3 también en un test**: una configuración de lint se desactiva borrando tres líneas de YAML
o con un `//nolint`; un test que falla en `make ci` no. Además cubren cosas distintas: `depguard` ve las
importaciones declaradas en cada fichero; el test recorre el grafo **transitivo** real del módulo.

---

## 3. Qué se exige de cada comprobación

- **Falla, no avisa.** Una violación rompe la orden, no genera una nota.
- **Nombra la regla violada**, para que quien la provoque sepa qué ha roto sin leer la configuración.
- **Es demostrable**: para cada regla existe una comprobación que falla si la regla se retira o se viola
  (SC-008). Eso se ejerce rompiendo cada regla a propósito, una por una, **en una copia desechable del
  árbol fuera del repositorio**, y viendo fallar **las dos capas en R1, R2 y R3** y **el lint en R4 y R5**
  —las dos reglas de símbolo, por el motivo de §2— ([`quickstart.md`](../quickstart.md), escenario 8).
- **Cubre también los applets de ejemplo.** Los comodines de Go no descienden a `testdata`, así que los
  dos paquetes bajo `internal/app/testdata/` se enumeran explícitamente. Son la implementación de
  referencia que copiará cada applet posterior: no pueden estar por debajo del listón.
- **Las excepciones son mínimas y demostrablemente estrechas.** Las únicas que existen son las de R4 y
  R5-descriptores en las dos raíces de composición. Que no se desborden a los paquetes vecinos se
  comprueba introduciendo las mismas violaciones en `internal/app/testdata/ejemplo` —sobre la misma copia
  desechable, nunca sobre el árbol de trabajo, que ahí está protegido por el guardián de diff— y viendo
  fallar `make lint` ([`quickstart.md`](../quickstart.md), escenario 9).

---

## 4. Consecuencias para quien escribe código nuevo

| Si necesitas… | Hazlo en… | Porque… |
|---|---|---|
| Hacer una petición HTTP | `internal/httpx` (H2) | R2. Además **no existe ninguna llamada HTTP con método distinto de GET/HEAD** en todo el módulo (constitución §I) |
| Leer o escribir en SQLite | `internal/{cache,store,graph}` (H3, H12, H16) | R3 |
| Terminar el proceso | La raíz de composición del binario que escribas: `cmd/kitlegal/main.go` | R4. Todo lo demás devuelve `error` o `int` |
| Escribir algo que vea el usuario | `internal/render` | R5. Todo lo demás recibe un `io.Writer` o un `*slog.Logger` |
| Registrar un evento | El `*slog.Logger` que recibe tu método `Ejecutar` | Va siempre a la salida de error. **No** está dentro de `schema.Contexto`: el dominio no importa `log/slog` (R1) |
| Que un mensaje se vea siempre, pase lo que pase con el nivel de registro | El presentador, no `slog` | Un nivel puede ocultar un registro; para eso existe ([`banderas-y-exit-codes.md`](./banderas-y-exit-codes.md) §6) |
| Añadir lógica de dominio | `internal/core/**` | R1. Y sin importar nada que haga I/O, tampoco de la biblioteca estándar |

---

## 5. Qué NO cubren estas reglas en H1

- **Las reglas de fuente que el linter no ve** —límite de peticiones, User-Agent identificable, términos
  de uso— se juzgan en la capa 2 de la constitución (juez con rúbrica cerrada) y nacen con el primer
  adaptador, en H2/H4. En H1 no hay fuentes.
- **La prohibición de que un adaptador use el espacio de nombres reservado `kitlegal.` / `kitlegal:`**
  (FR-016): no hay adaptadores en H1; la comprobación mecánica nace en H4.
- **El test de que no existe ninguna llamada HTTP con método distinto de GET/HEAD**: `docs/ROADMAP.md` lo
  sitúa en H25, sobre `httpx`. En H1 se cumple trivialmente porque nadie puede importar `net/http`.
