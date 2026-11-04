# Contrato: el arnés del sondeo con las variantes de `normativa`

FR-008, FR-010 a FR-013; SC-006. Research D2, D7, D8, D14; data-model §1, §4, §5. Lo que no se nombra aquí es el arnés
de H39 tal cual (`specs/031-h39-c-omo-se/contracts/arnes.md`).

## 1. Cómo se lanza

La misma orden, los mismos siete argumentos y la misma cadena que en H39 (`Makefile` →
`scripts/evals-sondeo-llavero.sh laboratorio` → `scripts/evals-laboratorio.sh` →
`go test -tags evals … -run '^TestSondeoDelLaboratorio$' ./internal/evals/` → `sondearElLaboratorio`). Lo lanza una
persona con la credencial del sondeo; ni `make ci`, ni un gancho, ni el job, ni un paso del workflow (FR-012). Ninguna
receta cambia.

Las cinco filas de la aceptación (FR-040) son dos lanzamientos, uno por cada combinación de modos, y una tabla que los
junta (research D8):

```sh
make evals-laboratorio VARIANTES=normativa-control MODOS=orden,herramienta SESIONES=40 \
  MODELO=claude-sonnet-5-5 CLAUDE_CODE=<versión> CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h40/a
make evals-laboratorio \
  VARIANTES=normativa-remisiones-en-orden,normativa-cuerpo-corto,normativa-testigo-en-referencia MODOS=herramienta \
  SESIONES=40 MODELO=claude-sonnet-5-5 CLAUDE_CODE=<versión> CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h40/b
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h40
```

80 + 120 = 200 sesiones, diez por eval en cada variante y modo.

## 2. Argumentos

Todos los errores van juntos, uno por línea, cada uno con el nombre del argumento de `make`, y con cualquiera de ellos no
se abre ninguna sesión, no se construye nada y no se crea `DIRECTORIO` (FR-010, FR-011). Cambia solo `VARIANTES`:

| Argumento | Vale si | Mensaje si no |
|---|---|---|
| `VARIANTES` | nombres de una variante de H39 (`laboratorio/variantes.yaml`) o de `normativa` (data-model §1), separados por comas, sin repetir | `VARIANTES: «<v>» no es ninguna variante del laboratorio` / `VARIANTES: «<v>» se repite` (los de H39, sin cambios) |
| `VARIANTES`, la familia | todos los nombres que valen son de una sola familia | `VARIANTES: se mezclan variantes del laboratorio de H39 y variantes de normativa` (una línea, tras las de los nombres) |
| cada par variante × modo | el modo es de la variante: `normativa-control`, los dos; las otras tres, solo `herramienta` | `VARIANTES y MODOS: <v> no se mide en modo <m>` (el de H39) |

`MODOS`, `SESIONES` (múltiplo de 4: las dos familias tienen cuatro evals), `MODELO`, `CLAUDE_CODE`, `CONCURRENCIA` y
`DIRECTORIO`, y detrás la credencial y `claude --version`, exactamente como en H39 (FR-011). Ejemplo, en la salida de
error, de `VARIANTES=control,normativa-cuerpo-corto,normativa-otra MODOS=orden SESIONES=6 …`:

```text
VARIANTES: «normativa-otra» no es ninguna variante del laboratorio
VARIANTES: se mezclan variantes del laboratorio de H39 y variantes de normativa
VARIANTES y MODOS: normativa-cuerpo-corto no se mide en modo orden
SESIONES: «6» no es un múltiplo de 4 mayor que 0
```

Con nombres de las dos familias, los pares de cada una se comprueban con su declaración, primero los de H39.

## 3. Lo que prepara y abre, con una familia de `normativa`

Lo de H39 (contracts/arnes.md §3 de H39) con tres diferencias, todas de la familia (data-model §1):

1. **Skills**: en el temporal, `skills/<variante>/` es una copia de `laboratorio/normativa/<variante>/skills/`; cada
   sesión de la variante tiene en `claude/skills/` solo `normativa`, la de su variante, y ninguna otra skill (FR-008).
2. **Evals**: el directorio de evals del reparto es `evals/normativa`, con el que se prepara cada sesión como en el job
   (las consultas de todas sus evals, servidas por lo grabado y con los proxies que rechazan toda petición; ADR 0043), y
   el plan de cada par son `SESIONES/4` sesiones de cada una de `04-lgt-prescripcion.yaml`,
   `05-trlrhl-impuestos-municipales.yaml`, `15-irpf-rendimientos-por-materia.yaml` y
   `22-lcsp-contrato-menor-en-2019.yaml`, en ese orden, de nombre `<eval sin .yaml>-<nn>` (FR-010). Si una de las cuatro
   no está en `evals/normativa`, el arnés falla antes de construir nada, con su nombre (no es un error de uso: sale el
   registro de `go test`).
3. **Servidor**: en modo herramienta, `servidor.json` arranca `kitlegal mcp serve`, la orden de hoy, con el `kitlegal`
   del árbol: ninguna variante de `normativa` cambia las instrucciones ni las descripciones de los parámetros (FR-008).

Lo demás, como en H39: `kitlegal` y `laboratorio` del árbol en el temporal; `DIRECTORIO/sondeo.json` con el modelo y la
versión; `DIRECTORIO/<variante>/<modo>/<sesión>/` con la transcripción de cada sesión, que es lo que lee la tabla
(FR-012); `orden` antes que `herramienta`; las variantes en el orden de su declaración; y tras un reparto con sesiones
sin abrir por el límite de uso, ninguno más. El run no graba nada ni escribe en `evidencias/` (FR-013): las sesiones
quedan en `DIRECTORIO`, fuera del repositorio en el ejemplo de §1.

## 4. Salida y uso

La de H39, sin cambios de forma:

```text
Sondeo del laboratorio en /Users/…/kitlegal-sondeos/h40/b, con claude-sonnet-5-5 y Claude Code 2.1.317.
- normativa-remisiones-en-orden, herramienta: 40 abiertas.
- normativa-cuerpo-corto, herramienta: 40 abiertas.
- normativa-testigo-en-referencia, herramienta: 40 abiertas.
La tabla: make laboratorio-tabla DIRECTORIO=/Users/…/kitlegal-sondeos/h40/b
```

La lee la persona al terminar cada lanzamiento: en los dos de la aceptación, 2 y 3 líneas de unos 60 bytes, más dos. No
crece con el uso del producto —cientos de normas y miles de bloques consultados no la cambian—, solo con las variantes y
modos pedidos, como mucho cuatro líneas en una familia de `normativa`. No es una señal: no avisa de nada y no se repite.
Ninguna skill del producto la pide.

Con un error de uso, la salida de error es solo la lista de §2: una línea por argumento que no vale, de unos 60 a 130
bytes, que la persona lee para corregir la orden; deja de darse cuando la orden vale. Tampoco depende del uso del
producto.

## 5. Controles

- **`TestArnesDeNormativa`** (`internal/evals/laboratorio_sondeo_test.go`, en `make ci`, sin modelo; SC-006, FR-008,
  FR-010, FR-011): `sondearElLaboratorio` con la preparación sin construir y el sustituto de `claude` de
  `TestArnesDelLaboratorio`. Sus dos casos válidos son los dos lanzamientos de la aceptación (§1) con menos sesiones,
  así que entre los dos piden los cinco pares variante × modo que se miden; su caso inválido de modo pide los tres que
  no se miden.
  - Con `VARIANTES=normativa-otra`; con
    `VARIANTES=normativa-remisiones-en-orden,normativa-cuerpo-corto,normativa-testigo-en-referencia MODOS=orden`, cuyo
    mensaje son tres líneas `VARIANTES y MODOS: <v> no se mide en modo orden`, una por variante en el orden de su
    declaración; con `SESIONES=6`; con `VARIANTES=control,normativa-control`; con la versión de Claude Code distinta; y
    con `MODELO=sonnet`: 0 sesiones, nada construido, `DIRECTORIO` sin crear y, cada uno, su mensaje de §2 o el de H39.
  - Con `VARIANTES=normativa-control MODOS=orden,herramienta SESIONES=4` (el primer lanzamiento): una sesión de cada
    una de las cuatro evals de §3 en cada modo, en `DIRECTORIO/normativa-control/orden/<eval>-01/` y
    `DIRECTORIO/normativa-control/herramienta/<eval>-01/`, y ninguna más.
  - Con `VARIANTES=normativa-remisiones-en-orden,normativa-cuerpo-corto,normativa-testigo-en-referencia
    MODOS=herramienta SESIONES=8` (el segundo): dos sesiones de cada una de las cuatro evals en cada variante, en
    `DIRECTORIO/<variante>/herramienta/<eval>-0{1,2}/`, y ninguna más.
  - En cada sesión de los dos casos válidos, `claude/skills/` tiene solo `normativa`, y su `SKILL.md` es, byte a byte,
    el de su variante; en modo orden no hay `servidor.json`; en modo herramienta, `servidor.json` arranca
    `kitlegal mcp serve` del temporal, sin `laboratorio servidor` (la comprobación por sesión de
    `TestArnesDelLaboratorio`).
- **`TestArnesDelLaboratorio`**, **`TestRecetasDelSondeo`** y **`TestServidorDelLaboratorio`** (H39): sin cambiar ningún
  valor esperado.
- **`TestSondeoDelLaboratorio`** (`-tags evals`): la entrada de la persona; no corre en `make ci`.
