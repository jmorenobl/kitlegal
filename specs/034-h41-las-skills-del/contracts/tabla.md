# Contrato: la tabla de las tres familias y la lectura con `cd`

FR-032, FR-040 a FR-042; SC-006, SC-009. Research D15 a D17; data-model §4, §7, §8. Lo que no se nombra aquí es el
guion de H40 tal cual (`specs/032-h40-por-qu-e/contracts/tabla.md`), y el de H39 para su familia, byte a byte.

## 1. Cómo se lanza y de qué familia es

```sh
make laboratorio-tabla DIRECTORIO=<ruta>     # go run ./cmd/laboratorio tabla <ruta>
```

Sin red, modelo ni credencial. `<ruta>` es un sondeo si tiene `sondeo.json`; si no, se leen sus subdirectorios que lo
tienen, en orden de nombre. **La familia es la de la primera variante que lee** (sondeos y variantes en orden de nombre):
una de las tres del producto si es suya; si no, la de H39, con su comportamiento de hoy. En la tabla de una familia del
producto, un directorio de variante de otra familia —también de H39— es un defecto:
`laboratorio: <ruta> no es ninguna variante de <skill>` en la salida de error, código 1 y sin tabla (el texto de H40 para
`normativa`). Lo que el guion no sabe leer sigue la regla genérica del spec (casos límite): defecto `inesperado`
(ADR 0023), código 1, sin más casos; el plan no diseña, comprueba ni prueba nada para un sondeo manipulado a mano
(research D17, D20).

## 2. Qué cuenta, por sesión

Las reglas de H40 (su contrato, §2), con la skill y las referencias de la familia (data-model §8):

- **Turno**, **primera herramienta** de cada modo, **terminada**, **activa antes**: las de H40, con la skill de la
  familia. `territorio_resolver`, `plazos_calcular`, `cita_preparar`, `cita_cotejar` y las de `boe` son herramientas del
  registro, y sus órdenes, consultas: cuentan como primera herramienta como las de `normativa`.
- **Abre la referencia `<f>`**: un `Read` de `<skill>/references/<f>` o de una ruta que termina en
  `/<skill>/references/<f>`, entero o por tramos; o una orden de `Bash` que contiene `references/<f>` (H40); **o, en las
  familias del producto, una orden de `Bash` con lectura con `cd`** (FR-041; research D16):
  - partida en órdenes simples como hoy (`|`, `||`, `&&`, `;`, `&` y saltos de línea; palabras como en sh, con comillas y
    barra invertida), tiene una orden simple cuya primera palabra que no es una asignación es `cd` y cuya palabra
    siguiente, sin una barra final, es `<skill>/references` o termina en `/<skill>/references`;
  - y, en una orden simple posterior de la misma orden, una palabra que es exactamente `<f>` o `./<f>`.

  | Orden de `Bash` (familia `normativa`) | Abre `normas.md` |
  |---|---|
  | `cd /s/claude/skills/normativa/references && cat normas.md` | sí |
  | `cd /s/claude/skills/normativa/references/ && cat "normas.md"` | sí |
  | `cd "$HOME/x/normativa/references"; cat 'normas.md'` | sí |
  | `cd /s/claude/skills/normativa/references && head -n 40 ./normas.md` | sí |
  | `cd /s/claude/skills/normativa/references` y, en otro `Bash`, `cat normas.md` | no |
  | `cd /s/claude/skills/jurisprudencia/references && cat normas.md` | no |
  | `cd /s/claude/skills/normativa && cat normas.md` | no |
  | `cat normas.md && cd /s/claude/skills/normativa/references` | no |
  | `cd /s/claude/skills/normativa/references && cat otra/normas.md` | no |
  | `cd /s/claude/skills/normativa/references && cat /x/normas.md` | no |
  | `cd /s/claude/skills/normativa/references && cat mis-normas.md` | no |
  | `cd /s/claude/skills/normativa/references && cat normas.md.bak` | no |
  | `cat /s/claude/skills/normativa/references/normas.md` (H40) | sí |

  En la familia de H39, `cd …/articulos-del-boe/references && cat cierre.md` no la abre: se cuenta con su contrato
  (`specs/031-h39-c-omo-se/contracts/tabla.md` §2). `Grep` y `Glob` no abren nada. Abierta varias veces, cuenta una, con
  el turno de la primera.
- **Cumple**: `reglaTestigo`, como en H40; solo se escribe en `normativa-testigo-en-referencia`.

## 3. La tabla

Una fila por variante, modo, modelo y versión de Claude Code, en el orden de la declaración de la familia (data-model
§1), `orden` antes que `herramienta`. Columnas: las seis de H39; «Activan antes»; una por cada referencia de la tabla
(data-model §8), en orden de nombre; y, **solo en la de `normativa`**, «Cumplen la regla» (FR-040). Recuentos,
intervalos de Wilson al 95 %, turnos, la fila con `n` = 0 y el detalle —una fila por sesión, en el orden de las filas y,
dentro, por ruta—, como en H40. La celda de una referencia que la variante de la fila no tiene va **vacía**, no a cero,
en la fila y en el detalle. El nombre de cada sesión del detalle lleva la variante, el modo y la eval: de ahí sale el
desglose por eval que lee la persona (FR-053); la tabla no lo agrega (fuera de alcance).

Ejemplo, la tabla de `plazos-y-territorio` sobre las cinco sesiones fijas de contracts/aceptacion.md §1, una por
variante (Wilson como en H40):

```text
# Sondeo de las variantes de plazos-y-territorio

| Variante | Modo | Modelo | Claude Code | Sin terminar | n | Activan antes | administrativo.md | calculo_de_plazos.md | jerarquia_normativa.md | municipios.md |
|---|---|---|---|---|---|---|---|---|---|---|
| plazos-y-territorio-control | orden | claude-sonnet-5-5 | 2.1.317 | 0 | 1 | 1 de 1 · 100,0 % [20,7–100,0 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 1 de 1 · 100,0 % [20,7–100,0 %] · turno 2 (2–2) |
| plazos-y-territorio-producto | herramienta | claude-sonnet-5-5 | 2.1.317 | 0 | 1 | 1 de 1 · 100,0 % [20,7–100,0 %] | 1 de 1 · 100,0 % [20,7–100,0 %] · turno 2 (2–2) | 1 de 1 · 100,0 % [20,7–100,0 %] · turno 3 (3–3) | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] |
| plazos-y-territorio-sin-condicion | orden | claude-sonnet-5-5 | 2.1.317 | 0 | 1 | 1 de 1 · 100,0 % [20,7–100,0 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 1 de 1 · 100,0 % [20,7–100,0 %] · turno 2 (2–2) | 0 de 1 · 0,0 % [0,0–79,3 %] | 1 de 1 · 100,0 % [20,7–100,0 %] · turno 2 (2–2) |
| plazos-y-territorio-condicion-por-orden | herramienta | claude-sonnet-5-5 | 2.1.317 | 0 | 1 | 1 de 1 · 100,0 % [20,7–100,0 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] |
| plazos-y-territorio-lista-con-verbo | orden | claude-sonnet-5-5 | 2.1.317 | 0 | 1 | 1 de 1 · 100,0 % [20,7–100,0 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 1 de 1 · 100,0 % [20,7–100,0 %] · turno 2 (2–2) | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] |

## Detalle

| Sesión | Terminada | Activa antes | administrativo.md | calculo_de_plazos.md | jerarquia_normativa.md | municipios.md |
|---|---|---|---|---|---|---|
| h41/plazos-y-territorio-control/orden/02-territorio-municipio-no-cubierto-01 | sí | sí | no | no | no | turno 2 |
| h41/plazos-y-territorio-producto/herramienta/05-plazo-reposicion-leganes-01 | sí | sí | turno 2 | turno 3 | no | no |
| h41/plazos-y-territorio-sin-condicion/orden/05-plazo-reposicion-leganes-01 | sí | sí | no | turno 2 | no | turno 2 |
| h41/plazos-y-territorio-condicion-por-orden/herramienta/11-plazo-reposicion-sin-leer-el-precepto-01 | sí | sí | no | no | no | no |
| h41/plazos-y-territorio-lista-con-verbo/orden/10-plazo-reposicion-anterior-a-la-lpac-01 | sí | sí | no | turno 2 | no | no |
```

En la de `normativa`, con un `sondeo.json` cuya lista tiene una referencia que la familia no declara (`nueva.md`), esa
columna va entre las de su orden de nombre y su celda va vacía en las filas de `normativa-control` y
`normativa-remisiones-en-orden` (contracts/aceptacion.md §1).

## 4. Uso, de fuera adentro

Las lee la persona una vez, para decidir la fusión (FR-052) y escribir el ADR de FR-053. En la aceptación, tres tablas
con diecinueve filas entre las tres: `normativa`, cinco filas de 16 columnas (de 450 a 650 bytes cada una con `n` = 40,
como en H40); `plazos-y-territorio`, diez filas de 11 columnas (unos 400 bytes); `jurisprudencia`, cuatro filas de 9
(unos 300). Y el detalle de las 760 sesiones, una línea por sesión de unos 140, 110 y 95 bytes: unos 28, 44 y 15 KB, con
las cabeceras unos 31, 48 y 17 KB, en línea con los 28,6, 20,5 y 15,6 KB de `evidencias/sondeos-20261106/` con cinco,
cuatro y cuatro filas. Su tamaño lo fijan las variantes, los modos y las sesiones pedidas, no el uso del producto ni nada
que se acumule. No es una señal: no avisa de nada, deja de leerse cuando la persona decide, y un sondeo nuevo da otras
tablas. Ninguna skill del producto la pide.

## 5. Controles

- **`TestTablaDeLasFamilias`** (`internal/evals/laboratorio_tabla_test.go`, en `make ci`; FR-032, FR-040, FR-041;
  SC-006): sobre las sesiones fijas de `internal/evals/testdata/laboratorio/familias/<skill>/sondeo/` de cada familia
  —los dos modos; en `plazos-y-territorio`, las cinco variantes; lecturas con `cd` que cuentan y que no; una referencia
  de la lista del producto que la familia no declara; una sesión sin terminar— la salida es, byte a byte, su
  `tabla.md`. Y el defecto de §1: un directorio de variante de otra familia en un sondeo del producto, con código 1, sin tabla y su mensaje, sobre
  `internal/evals/testdata/laboratorio/familias/mezcla/sondeo/h41/` (sin `tabla.md`).
- **`TestLecturaConCd`** (`internal/evals/laboratorio_tabla_test.go`; FR-041; SC-006): la tabla de §2, orden a orden,
  con la familia de `normativa`, y la de H39 con la lectura con `cd`, que no la abre.
- **`TestTablaDeNormativa`**, **`TestTablaDelLaboratorio`**, **`TestReglaTestigo`** y **`TestOrdenesDeBash`** (H39 y
  H40): sin cambiar ningún valor esperado.
- **Las suites congeladas de H39 y H40** (`h39-describe.txtar`, `h40-laboratorio-tabla.txtar`), en `TestEntregaDelHito`,
  sin cambiar ningún fichero suyo (FR-042).
- **El guion congelado de H41** `h41-laboratorio-tabla.txtar` (contracts/aceptacion.md §1).
