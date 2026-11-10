# Data model: H41.1

No hay datos nuevos en `data/`, ni esquemas, ni estado en disco que el binario cree. Las entidades son ficheros
versionados de la skill y del laboratorio, y las tablas de dos tests de `make ci`. Research D1 a D4 y D12; contracts/.

## 1. Skill del producto `plazos-y-territorio`

`skills/plazos-y-territorio/`: `SKILL.md` y `references/{administrativo,calculo_de_plazos,jerarquia_normativa,municipios}.md`.
La empotra el binario (`skills.go`, `//go:embed skills/*/SKILL.md skills/*/references/*`), la instala `kitlegal skills
install` y el sondeo la mide como `plazos-y-territorio-producto` sin copia versionada.

- Es `plazos-y-territorio-sin-condicion` con M1 (contracts/skill.md §1 a §3): fuera de lo que genera `make
  skills-sync`, ruta a ruta y byte a byte.
- Cuerpo del `SKILL.md`: 10 882 caracteres; ≤ 11 000 también con la línea de un paquete más (10 952 con la vertical de
  ensayo).
- Frontmatter: el de `plazos-y-territorio-control`.
- Remisiones del plan (§4): tres, en orden.

## 2. Variantes del laboratorio de `plazos-y-territorio`

`laboratorio/plazos-y-territorio/plazos-y-territorio-{control,sin-condicion,condicion-por-orden,lista-con-verbo}/skills/plazos-y-territorio/`.
No cambian un byte (FR-006). `-control` es la base del frontmatter y de las huellas; `-sin-condicion`, la base de la
skill del producto (§3). `plazos-y-territorio-producto` sigue sin directorio.

## 3. Declaración del producto (`internal/evals/laboratorio_familias_test.go`)

`declaracionDelProducto{desde string; cambios []cambioDelCuerpo; movimientos []movimientoALaReferencia}`, una por
skill del producto en `declaracionesDelProducto`:

| Campo | Qué es | Regla |
|---|---|---|
| `desde` | variante versionada de la familia sobre la que se aplica; vacío, la de control | nuevo; solo lo lleva `plazos-y-territorio`, con `plazos-y-territorio-sin-condicion` |
| `cambios[i].antes` / `.despues` | líneas enteras del `SKILL.md` de la base y las que las sustituyen | `antes` está una sola vez en el `SKILL.md` de la base, con los anteriores aplicados (si no, error con su número) |
| `movimientos[i]` | `texto` que sale del cuerpo, `referencia` a la que llega, `enLaReferencia` (bytes que se le añaden al final) | `texto` y `enLaReferencia` son iguales con los blancos juntados; `texto` está en el cuerpo de la base y no en el del producto |

La de `plazos-y-territorio`: un cambio (M1 en el paso 10) y un movimiento (a `calculo_de_plazos.md`), con los bytes de
contracts/controles.md §1.1. Las de `normativa` y `jurisprudencia`, como en H41.

## 4. Tablas de `TestRemisionesDelRepositorio` (`internal/skills/remisiones_test.go`)

| Tabla | Clave → valor | Tras el hito |
|---|---|---|
| `fallosAdmitidosDelProducto` (nueva) | skill de `skills/` → fallos de la regla que el test admite, en el orden en que los da `fallosDeLaRegla` | `plazos-y-territorio` → uno, el de `jerarquia_normativa.md` sin «lee»; las demás skills no están: admiten cero |
| `remisionesDelPlan` | skill → frases del cuerpo con un enlace a una referencia, fuera de la región de los paquetes, en orden | gana `plazos-y-territorio`, con tres (contracts/controles.md §2.2) |
| `fallosDeControl` | skill → fallos de la regla con su `<skill>-control` | no cambia |
| `fueraDeLasRemisiones` | skill → hito que la incluye | sale |

## 5. El sondeo de H41.1 (fuera del run)

Lo lanza la persona (US4) con el arnés de H41 sin cambiar: dos variantes (`plazos-y-territorio-control`,
`plazos-y-territorio-producto`) × dos modos × 40 sesiones = 160, diez por eval (02, 05, 10 y 11). `make
laboratorio-tabla` imprime cuatro filas, una por variante y modo, con las columnas de «Activan antes» y de cada
referencia (`administrativo.md`, `calculo_de_plazos.md`, `jerarquia_normativa.md`, `municipios.md`) y el detalle por
sesión. La tabla va a `evidencias/sondeos-<fecha>/`; no hay estado nuevo.
