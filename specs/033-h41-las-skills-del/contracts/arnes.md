# Contrato: el arnés del sondeo con las tres familias del producto

FR-030 a FR-033; SC-006, SC-007. Research D9 a D12, V7, V8; data-model §1, §3, §4. Lo que no se nombra aquí es el
arnés de H40 tal cual (`specs/032-h40-por-qu-e/contracts/arnes.md`).

## 1. Cómo se lanza

La misma orden, los mismos siete argumentos y la misma cadena que en H39 y H40 (`Makefile` →
`scripts/evals-sondeo-llavero.sh laboratorio` → `scripts/evals-laboratorio.sh` → `go test -tags evals … -run
'^TestSondeoDelLaboratorio$' ./internal/evals/` → `sondearElLaboratorio`). Lo lanza una persona con la credencial del
sondeo; ni `make ci`, ni un gancho, ni el job, ni un paso del workflow (FR-031). Ninguna receta cambia
(`TestRecetasDelSondeo`).

Las trece filas de la aceptación (FR-032) son cuatro lanzamientos y tres tablas, una por familia:

```sh
V=<versión de claude --version>
make evals-laboratorio VARIANTES=normativa-control,normativa-producto MODOS=orden,herramienta SESIONES=40 \
  MODELO=claude-sonnet-5-5 CLAUDE_CODE=$V CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41/normativa/a
make evals-laboratorio VARIANTES=normativa-remisiones-en-orden MODOS=herramienta SESIONES=40 \
  MODELO=claude-sonnet-5-5 CLAUDE_CODE=$V CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41/normativa/b
make evals-laboratorio VARIANTES=plazos-y-territorio-control,plazos-y-territorio-producto MODOS=orden,herramienta \
  SESIONES=40 MODELO=claude-sonnet-5-5 CLAUDE_CODE=$V CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41/plazos-y-territorio
make evals-laboratorio VARIANTES=jurisprudencia-control,jurisprudencia-producto MODOS=orden,herramienta \
  SESIONES=40 MODELO=claude-sonnet-5-5 CLAUDE_CODE=$V CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41/jurisprudencia
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41/normativa
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41/plazos-y-territorio
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41/jurisprudencia
```

160 + 40 + 160 + 160 = 520 sesiones, diez por eval en cada variante y modo.

## 2. Argumentos

Todos los errores van juntos, uno por línea, cada uno con el nombre del argumento de `make`; con cualquiera no se abre
ninguna sesión, no se construye nada y no se crea `DIRECTORIO` (FR-030). Cambia solo `VARIANTES`:

| Argumento | Vale si | Mensaje si no |
|---|---|---|
| `VARIANTES` | nombres de variantes de alguna de las cuatro familias (H39 y las tres de data-model §1), separados por comas, sin repetir | `VARIANTES: «<v>» no es ninguna variante del laboratorio` / `VARIANTES: «<v>» se repite` (H39) |
| `VARIANTES`, la familia | todos los nombres que valen son de una sola familia | `VARIANTES: se mezclan variantes de <a>, … y variantes de <z>`, una por familia pedida, en el orden de las familias; la de H39 es «del laboratorio de H39» (research D11) |
| cada par variante × modo | el modo es de la variante: `<skill>-control` y `<skill>-producto`, los dos; las tres de H40 sin control, `herramienta` | `VARIANTES y MODOS: <v> no se mide en modo <m>` (H39) |

`MODOS`, `SESIONES` (múltiplo de 4: las cuatro familias tienen cuatro evals en su plan), `MODELO`, `CLAUDE_CODE`,
`CONCURRENCIA` y `DIRECTORIO`, y detrás la credencial y `claude --version`, como en H40. Ejemplo, en la salida de error,
de `VARIANTES=normativa-producto,jurisprudencia-control,plazos-y-territorio-otra MODOS=orden SESIONES=6 …`:

```text
VARIANTES: «plazos-y-territorio-otra» no es ninguna variante del laboratorio
VARIANTES: se mezclan variantes de normativa y variantes de jurisprudencia
SESIONES: «6» no es un múltiplo de 4 mayor que 0
```

Con `VARIANTES=control,normativa-control`, el mensaje de la mezcla es el de H40, byte a byte:
`VARIANTES: se mezclan variantes del laboratorio de H39 y variantes de normativa`.

## 3. Lo que prepara y abre, con una familia del producto

Lo de H40 (contracts/arnes.md §3 de H40), con la familia pedida:

1. **Skills.** En el temporal, `skills/<variante>/` tiene una sola skill, con el nombre de la de la familia:
   - de una variante versionada, la copia de `laboratorio/<skill>/<variante>/skills/` (como en H40);
   - de `<skill>-producto`, la copia de `skills/<skill>` del árbol que lanza el sondeo, en
     `skills/<skill>-producto/<skill>/`, con `os.CopyFS` (FR-022; research D10).
   Cada sesión de una variante tiene en `claude/skills/` esa skill y ninguna otra (FR-030).
2. **Evals.** El directorio de evals del reparto es `evals/<skill>`, con el que se prepara cada sesión como en el job
   (research V7), y el plan de cada par son `SESIONES/4` sesiones de cada una de las cuatro evals de la familia
   (data-model §1), en ese orden, de nombre `<eval sin .yaml>-<nn>`. Si una no está en `evals/<skill>`, el arnés falla
   antes de construir nada, con su nombre (H40).
3. **Servidor.** En modo herramienta, `kitlegal mcp serve` del árbol, el de hoy, para todas las variantes: ninguna cambia
   las instrucciones ni las descripciones de los parámetros.
4. **`sondeo.json`.** El de H39; si el sondeo pide `<skill>-producto`, con `referencias_del_producto`: los nombres de los
   ficheros de `references/` de la copia del punto 1, en orden de nombre (data-model §4; FR-033). Solo los nombres: ni
   la copia ni una huella de la skill se guardan en `DIRECTORIO`.

Lo demás, como en H40: `orden` antes que `herramienta`; las variantes en el orden de la declaración de la familia; tras
un reparto con sesiones sin abrir por el límite de uso, ninguno más. El run no graba nada ni escribe en `evidencias/`
(FR-031): las sesiones quedan en `DIRECTORIO`, fuera del repositorio en el ejemplo de §1.

## 4. Salida y uso

La de H39, sin cambios de forma. La del primer lanzamiento de §1:

```text
Sondeo del laboratorio en /Users/…/kitlegal-sondeos/h41/normativa/a, con claude-sonnet-5-5 y Claude Code 2.1.317.
- normativa-control, orden: 40 abiertas.
- normativa-control, herramienta: 40 abiertas.
- normativa-producto, orden: 40 abiertas.
- normativa-producto, herramienta: 40 abiertas.
La tabla: make laboratorio-tabla DIRECTORIO=/Users/…/kitlegal-sondeos/h41/normativa/a
```

La lee la persona al terminar cada lanzamiento: una línea por par pedido, de unos 45 bytes, más dos; como mucho cuatro en
`plazos-y-territorio` y `jurisprudencia` y siete en `normativa` (sus dos variantes nuevas en dos modos y las tres de H40
en uno). No crece con el uso del
producto —cientos de normas y miles de bloques consultados no la cambian—, solo con las variantes y modos pedidos. No es
una señal: no avisa de nada y no se repite. Ninguna skill del producto la pide. Con un error de uso, la salida de error es
la lista de §2, una línea de 50 a 130 bytes por argumento que no vale, que deja de darse cuando la orden vale.

`sondeo.json`, de 55 a 256 bytes (data-model §4), lo lee el guion de la tabla una vez por sondeo.

## 5. Controles

- **`TestArnesDeLasFamilias`** (`internal/evals/laboratorio_sondeo_test.go`, en `make ci`, sin modelo; FR-022, FR-030,
  FR-033, SC-006, SC-007): `sondearElLaboratorio` con la preparación sin construir y el sustituto de `claude` de
  `TestArnesDelLaboratorio`, que anota sus invocaciones.
  - Con cada petición inválida —una variante que no existe (`plazos-y-territorio-otra`), `normativa-producto` y
    `jurisprudencia-control` juntas, `plazos-y-territorio-control` y `control` (H39) juntas, `SESIONES=6`—: 0
    sesiones, nada construido, `DIRECTORIO` sin crear y su mensaje de §2.
  - Con `VARIANTES=<skill>-control,<skill>-producto MODOS=orden,herramienta SESIONES=4`, una por familia: una sesión de
    cada una de sus cuatro evals en cada variante y modo, en `DIRECTORIO/<variante>/<modo>/<eval>-01/`, y ninguna más.
  - Con `VARIANTES=normativa-remisiones-en-orden MODOS=herramienta SESIONES=8`: dos de cada eval.
  - En cada sesión, `claude/skills/` tiene solo la skill de la familia, y sus ficheros son, ruta a ruta y byte a byte, los
    de `laboratorio/<skill>/<variante>/skills/<skill>/` en una variante versionada y los de `skills/<skill>` del árbol en
    `<skill>-producto`; en modo orden no hay `servidor.json`, y en modo herramienta arranca `kitlegal mcp serve`.
  - `DIRECTORIO/sondeo.json` tiene `referencias_del_producto` igual a los nombres de `skills/<skill>/references`, en
    orden de nombre, si se pidió `<skill>-producto`, y no la tiene si no.
- **`TestArnesDeNormativa`**, **`TestArnesDelLaboratorio`**, **`TestRecetasDelSondeo`** y
  **`TestServidorDelLaboratorio`** (H39 y H40): sin cambiar ningún valor esperado.
- **`TestSondeoDelLaboratorio`** (`-tags evals`): la entrada de la persona; no corre en `make ci`.
