# Contrato: el arnés del sondeo con las tres familias del producto

FR-022, FR-030 a FR-033; SC-007, SC-008. Research D10, D14, D15, V9; data-model §1, §3, §4. Lo que no se nombra aquí es
el arnés de H40 tal cual (`specs/032-h40-por-qu-e/contracts/arnes.md`).

## 1. Cómo se lanza

La misma orden, los mismos siete argumentos y la misma cadena que en H39 y H40 (`Makefile` →
`scripts/evals-sondeo-llavero.sh laboratorio` → `scripts/evals-laboratorio.sh` → `go test -tags evals … -run
'^TestSondeoDelLaboratorio$' ./internal/evals/` → `sondearElLaboratorio`). Lo lanza una persona con la credencial del
sondeo; ni `make ci`, ni un gancho, ni el job, ni un paso del workflow (FR-031). Ninguna receta cambia
(`TestRecetasDelSondeo`).

Las diecinueve filas de la aceptación (FR-033) son cuatro invocaciones y tres tablas, una por familia:

```sh
V=$(claude --version | awk '{print $1}')
make evals-laboratorio VARIANTES=normativa-control,normativa-producto MODOS=orden,herramienta SESIONES=40 \
  MODELO=claude-sonnet-5-5 CLAUDE_CODE=$V CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41/normativa/a
make evals-laboratorio VARIANTES=normativa-remisiones-en-orden MODOS=herramienta SESIONES=40 \
  MODELO=claude-sonnet-5-5 CLAUDE_CODE=$V CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41/normativa/b
make evals-laboratorio VARIANTES=plazos-y-territorio-control,plazos-y-territorio-producto,plazos-y-territorio-sin-condicion,plazos-y-territorio-condicion-por-orden,plazos-y-territorio-lista-con-verbo \
  MODOS=orden,herramienta SESIONES=40 MODELO=claude-sonnet-5-5 CLAUDE_CODE=$V CONCURRENCIA=5 \
  DIRECTORIO=../kitlegal-sondeos/h41/plazos-y-territorio
make evals-laboratorio VARIANTES=jurisprudencia-control,jurisprudencia-producto MODOS=orden,herramienta \
  SESIONES=40 MODELO=claude-sonnet-5-5 CLAUDE_CODE=$V CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41/jurisprudencia
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41/normativa
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41/plazos-y-territorio
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41/jurisprudencia
```

160 + 40 + 400 + 160 = 760 sesiones, diez por eval en cada variante y modo.

## 2. Argumentos

Todos los errores van juntos, uno por línea, cada uno con el nombre del argumento de `make`; con cualquiera no se abre
ninguna sesión, no se construye nada y no se crea `DIRECTORIO` (FR-030). Cambia solo `VARIANTES`:

| Argumento | Vale si | Mensaje si no |
|---|---|---|
| `VARIANTES` | nombres de variantes de alguna de las cuatro familias (H39 y las tres de data-model §1), separados por comas, sin repetir | `VARIANTES: «<v>» no es ninguna variante del laboratorio` / `VARIANTES: «<v>» se repite` (H39) |
| `VARIANTES`, la familia | todos los nombres que valen son de una sola familia | `VARIANTES: se mezclan variantes de <a>, … y variantes de <z>`, una por familia pedida, en el orden de las familias; la de H39 es «del laboratorio de H39» (research D14) |
| cada par variante × modo | el modo es de la variante (data-model §1): `<skill>-control`, `<skill>-producto` y las tres de `plazos-y-territorio`, los dos; las tres de H40 sin control, `herramienta` | `VARIANTES y MODOS: <v> no se mide en modo <m>` (H39) |

`MODOS`, `SESIONES` (un múltiplo de 4 mayor que 0: las cuatro familias tienen cuatro evals en su plan), `MODELO`,
`CLAUDE_CODE`, `CONCURRENCIA` y `DIRECTORIO`, y detrás la credencial y `claude --version`, como en H40. Ejemplo, en la
salida de error, de `VARIANTES=normativa-producto,jurisprudencia-control,plazos-y-territorio-otra MODOS=orden SESIONES=6
…`:

```text
VARIANTES: «plazos-y-territorio-otra» no es ninguna variante del laboratorio
VARIANTES: se mezclan variantes de normativa y variantes de jurisprudencia
SESIONES: «6» no es un múltiplo de 4 mayor que 0
```

Con `VARIANTES=control,normativa-control`, el mensaje de la mezcla es el de H40, byte a byte:
`VARIANTES: se mezclan variantes del laboratorio de H39 y variantes de normativa`.

## 3. Lo que prepara y abre, con una familia del producto

Lo de H40 (su contrato, §3), con la familia pedida:

1. **Skills.** En el temporal, `skills/<variante>/` tiene una sola skill, con el nombre de la de la familia:
   - de una variante versionada, la copia de `laboratorio/<skill>/<variante>/skills/` (como en H40);
   - de `<skill>-producto`, la copia de `skills/<skill>` del árbol que lanza el sondeo, en
     `skills/<skill>-producto/<skill>/`, con `os.CopyFS` (FR-022; research V9).
   Cada sesión de una variante tiene en `claude/skills/` esa skill y ninguna otra (FR-030).
2. **Evals.** El directorio de evals del reparto es `evals/<skill>`, con el que se prepara cada sesión como en el job,
   y el plan de cada par son `SESIONES/4` sesiones de cada una de las cuatro evals de la familia (data-model §1), en ese
   orden, de nombre `<eval sin .yaml>-<nn>`. Si una no está en `evals/<skill>`, el arnés falla antes de construir nada,
   con su nombre (H40).
3. **Servidor.** En modo herramienta, `kitlegal mcp serve` del árbol, el de hoy, para todas las variantes.
4. **`sondeo.json`.** El de H39; si el sondeo pide `<skill>-producto`, con `referencias_del_producto`: los nombres de los
   ficheros de `references/` de la copia del punto 1, en orden de nombre (data-model §4; FR-032). Solo los nombres: ni
   la copia ni una huella de la skill se guardan en `DIRECTORIO`.

Lo demás, como en H40: `orden` antes que `herramienta`; las variantes en el orden de la declaración de la familia; tras
un reparto con sesiones sin abrir por el límite de uso, ninguno más. El run no graba nada ni escribe en `evidencias/`
(FR-031).

## 4. Salida y uso

La de H39, sin cambios de forma. La de la tercera invocación de §1:

```text
Sondeo del laboratorio en /Users/…/kitlegal-sondeos/h41/plazos-y-territorio, con claude-sonnet-5-5 y Claude Code 2.1.317.
- plazos-y-territorio-control, orden: 40 abiertas.
- plazos-y-territorio-control, herramienta: 40 abiertas.
- plazos-y-territorio-producto, orden: 40 abiertas.
…
- plazos-y-territorio-lista-con-verbo, herramienta: 40 abiertas.
La tabla: make laboratorio-tabla DIRECTORIO=/Users/…/kitlegal-sondeos/h41/plazos-y-territorio
```

La lee la persona al terminar cada invocación: una línea por par pedido, de unos 45 a 65 bytes, más dos; como mucho
diez en `plazos-y-territorio`, siete en `normativa` y cuatro en `jurisprudencia`. No crece con el uso del producto
—cientos de normas y miles de bloques consultados no la cambian—, solo con las variantes y modos pedidos. No es una
señal: no avisa de nada y no se repite. Ninguna skill del producto la pide. Con un error de uso, la salida de error es la
lista de §2, una línea de 50 a 130 bytes por argumento que no vale, que deja de darse cuando la orden vale.

`sondeo.json`, de 55 a 256 bytes (data-model §4), lo lee el guion de la tabla una vez por sondeo.

## 5. Controles

- **`TestArnesDeLasFamilias`** (`internal/evals/laboratorio_sondeo_test.go`, en `make ci`, sin modelo; FR-022, FR-030,
  FR-032, FR-033; SC-007, SC-008): `sondearElLaboratorio` con la preparación sin construir y el sustituto de `claude`
  de `TestArnesDelLaboratorio`, que anota sus invocaciones.
  - Con cada petición inválida —una variante que no existe (`plazos-y-territorio-otra`), `normativa-remisiones-en-orden`
    con `MODOS=orden`, `normativa-producto` y `jurisprudencia-control` juntas, `plazos-y-territorio-control` y
    `control` (H39) juntas, `SESIONES=6` y `SESIONES=0`—: 0 sesiones, nada construido, `DIRECTORIO` sin crear y su
    mensaje de §2.
  - Con las cuatro invocaciones de §1 (sus `VARIANTES` y `MODOS`), con `SESIONES=4`: una sesión de cada una de las
    cuatro evals de la familia en cada variante y modo pedidos, en `DIRECTORIO/<variante>/<modo>/<eval>-01/`, y ninguna
    más (16, 4, 40 y 16). Con `VARIANTES=normativa-remisiones-en-orden MODOS=herramienta SESIONES=8`: dos de cada eval.
  - En cada sesión, `claude/skills/` tiene solo la skill de la familia, y sus ficheros son, ruta a ruta y byte a byte,
    los de `laboratorio/<skill>/<variante>/skills/<skill>/` en una variante versionada y los de `skills/<skill>` del
    árbol en `<skill>-producto` —en `plazos-y-territorio-producto`, también los de `plazos-y-territorio-control`—; en
    modo orden no hay `servidor.json`, y en modo herramienta arranca `kitlegal mcp serve`.
  - `DIRECTORIO/sondeo.json` tiene `referencias_del_producto` igual a los nombres de `skills/<skill>/references`, en
    orden de nombre, si se pidió `<skill>-producto`, y no la tiene si no.
- **`TestArnesDeNormativa`**, **`TestArnesDelLaboratorio`**, **`TestRecetasDelSondeo`** y
  **`TestServidorDelLaboratorio`** (H39 y H40): sin cambiar ningún valor esperado.
- **`TestSondeoDelLaboratorio`** (`-tags evals`): la entrada de la persona; no corre en `make ci`.

Las cuatro invocaciones se prueban con `SESIONES=4` y no con 40 porque cada sesión se prepara como en el job (caché y
grafo de la eval): 40 por par serían 760 preparaciones en cada `make ci` para comprobar la misma división n/4.
