# Contrato: interfaz de línea de órdenes de `kitlegal` en H0

**Fecha**: 2026-09-10 · **Requisitos**: FR-002, FR-003, FR-004, FR-037, FR-038, FR-039, FR-040 ·
**Criterios**: SC-003

Este es el **único contrato observable del binario en H0**. Todo lo demás —multicall por `os.Args[0]`,
banderas globales, sobre de salida, `--describe`, `--json`— es alcance de H1 y **no** forma parte de este
contrato.

---

## Superficie

```
kitlegal version
```

Un único verbo. Sin banderas. Sin subverbos.

---

## `kitlegal version`

**Salida**: stdout, texto plano, exactamente tres líneas terminadas en `\n`.

```
kitlegal <version>
commit: <commit>
fecha:  <fecha>
```

| Marcador | Contenido | Ejemplo |
|---|---|---|
| `<version>` | Valor de `main.version`, inyectado en construcción | `0.0.0-dev-3-g7b51ac8` |
| `<commit>` | SHA-1 completo de `git rev-parse HEAD` de la revisión construida | `7b51ac8f3c2e9d1a04b6…` |
| `<fecha>` | Instante de construcción en UTC, RFC 3339 | `2026-09-10T09:12:33Z` |

**stderr**: vacío.

**Código de salida**: `0`.

**Invariantes**

- C1 — Los tres datos aparecen siempre. Un binario construido con `make build` o `make install` nunca
  muestra los valores por defecto del código (`dev`, `none`, `unknown`) — FR-004.
- C2 — `<commit>` coincide, carácter a carácter, con la salida de `git rev-parse HEAD` en la revisión desde
  la que se construyó — SC-003.
- C3 — La salida es texto plano. **No** es el sobre `{ok, fuente, url, fecha_consulta, hash, data}`: H0 no
  emite ningún dato legal, así que no hay cita que resolver (*Assumptions* del spec).
- C4 — El comando no abre ninguna conexión de red ni escribe en ningún fichero — FR-037.

---

## Cualquier otra invocación

```
kitlegal
kitlegal <verbo-desconocido>
kitlegal version extra
```

**Salida**: stderr, una línea de uso que nombra el verbo disponible. stdout vacío.

```
uso: kitlegal version
```

**Código de salida**: `2`.

**Motivo**: de la tabla de códigos estables del proyecto, `2` es «args». Devolver `0` ante un verbo
inexistente contradiría de facto FR-039; devolver `1` introduciría un código fuera de la tabla. La
decisión y sus alternativas están en [research.md D7](../research.md).

---

## Códigos de salida

H0 usa dos de los seis códigos estables del proyecto y **no contradice** ninguno (FR-039):

| Código | Significado | ¿Lo produce H0? |
|---|---|---|
| 0 | ok | Sí, en `kitlegal version` |
| 2 | args | Sí, ante verbo ausente o desconocido |
| 3 | no encontrado | No (H1 en adelante) |
| 4 | fuente no disponible | No (H3/H4) |
| 5 | rate-limited / TOS | No (H2) |
| 6 | requiere identidad humana | No (fase 6) |

---

## Forma interna que hace verificable el contrato

```go
func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int
```

- `run` **nunca** llama a `os.Exit` ni entra en pánico: devuelve el código (FR-038).
- `os.Exit` aparece exactamente una vez en todo el repositorio, en `main()`, respetando la regla «solo
  `internal/cli` y `cmd/` llaman a `os.Exit`» (constitución §IV).
- Toda la salida va por los escritores recibidos, nunca por `os.Stdout`/`os.Stderr` directamente desde
  `run`, lo que permite a `main_test.go` ejercer el contrato completo con búferes en memoria.

**Cobertura del contrato por tests** (`cmd/kitlegal/main_test.go`, tabla con `t.Parallel()`):

| Caso | Comprueba |
|---|---|
| `version` con valores inyectados | Las tres líneas, los tres valores, stderr vacío, código 0 (C1, C3) |
| sin argumentos | Línea de uso en stderr, stdout vacío, código 2 |
| verbo desconocido | Línea de uso en stderr, stdout vacío, código 2 |
| `version` con argumento sobrante | Código 2 |

C2 no se comprueba desde el test unitario (el test no conoce el commit de construcción): se verifica en
[quickstart.md](../quickstart.md) comparando la salida del binario con `git rev-parse HEAD`.

---

## Compatibilidad hacia H1

Cuando H1 introduzca `internal/cli` (Kong), el despacho multicall y el sobre de salida, `version`
cambiará de sitio y probablemente de forma. Este contrato **no** promete estabilidad más allá de H0: es
el comportamiento de un hito de fundación, no una API pública. `pkg/legalkit` y su compromiso de
compatibilidad son H24.
