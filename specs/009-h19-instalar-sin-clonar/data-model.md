# Data model: H19 · Instalar sin clonar

Entidades y reglas del dominio `internal/core/instalacion` y de su adaptador `internal/disco`. Lo observable desde la
línea de órdenes está en [contracts/applet-skills.md](./contracts/applet-skills.md); el formato del manifiesto, en
[contracts/manifiesto.md](./contracts/manifiesto.md). Las decisiones que sostienen cada regla, en
[research.md](./research.md) (D1-D34).

## 1. Skill empotrada

| Campo | Tipo | Regla |
|---|---|---|
| `Nombre` | cadena | nombre del directorio `skills/<n>/` con `SKILL.md` en lo empotrado (FR-004); forma de nombre de skill (V44) |
| `Ficheros` | lista de `FicheroEmpotrado` | `SKILL.md` y todo fichero bajo `references/`, en orden de ruta |
| `FicheroEmpotrado.Ruta` | cadena | relativa al directorio de la skill, con `/` (`SKILL.md`, `references/normas.md`) |
| `FicheroEmpotrado.Contenido` | bytes | byte a byte lo del árbol al compilar (FR-003) |
| `FicheroEmpotrado.Huella` | cadena | `sha256:<hex>` de `Contenido` |

La lista de skills empotradas está ordenada por nombre. La construye `internal/app` desde `kitlegal.Skills()`; el
dominio la recibe ya leída (no importa `io/fs`, V26).

## 2. Ámbito

| Campo | Tipo | Local | Global (`-g`) | `--dir <ruta>` |
|---|---|---|---|---|
| `Clase` | enumerado | `local` | `global` | `dir` |
| `Raiz` | cadena | `""` (directorio de trabajo) | `$HOME` | — |
| `Neutro` | cadena | `.agents/skills` | `$HOME/.agents/skills` | `<ruta>` limpia, como se pasó |
| `Guardas` | lista de rutas | `.agents`, `.agents/skills` | `$HOME/.agents`, `$HOME/.agents/skills` | `<ruta>` |
| `ConHosts` | booleano | sí | sí | no |
| `Banderas` | cadena para las órdenes | `""` | `-g` | `--dir '<ruta tal como se pasó>'` |

Rutas: con `/` y como se alcanzan desde el directorio de trabajo (D11). El adaptador las convierte al separador del
sistema al tocar el disco. Lo que hay por encima de `Raiz` o de `<ruta>` no se comprueba (FR-027).

## 3. Entrada del disco (puerto `Disco`)

Lo que el dominio sabe de una ruta, siempre **sin seguir enlaces** (FR-028, D6):

| Campo | Cuándo | Significado |
|---|---|---|
| `Tipo` | siempre | `ausente`, `directorio` (real), `fichero` (regular), `enlace`, `otro` (tubería, socket, dispositivo…) |
| `Destino` | `enlace` | destino literal (`os.Readlink`) |
| `Resuelve` | `enlace` | si seguirlo llega a algo que existe (`os.Stat` sobre el enlace, sin leer nada); falso si cuelga o está en ciclo |

Operaciones del puerto: `Examinar(ruta) (Entrada, error)`, `Huella(ruta) (string, error)` (solo fichero regular:
`Lstat` → `Open` → `Stat` + `SameFile`; si no, error «no es un fichero regular»), `Leer(ruta) ([]byte, error)` (mismas
garantías; para el manifiesto) y `Nombres(ruta) ([]string, error)` (entradas de un directorio real, para las copias).
Un error de E/S que no es «no existe» es un fallo de la orden (exit 1), no un conflicto.

Puerto `Enlazador`: `Disponible(directorio string) (bool, error)` y `Enlazar(destino, ruta string) error` (D9). Puerto
`Escritor` (fase de aplicación): `CrearDirectorio(ruta)`, `EscribirFichero(ruta, contenido)` (atómico, `0o600`, D8),
`Retirar(ruta)` (fichero, enlace o directorio vacío) y `Enlazar` (el del `Enlazador`).

**Directorio de la sonda** (`DS`): el directorio existente más próximo a `<Raiz>/.claude/skills`, sin subir por encima
de la raíz —`<Raiz>/.claude/skills`, si no `<Raiz>/.claude`, si no `<Raiz>`—, elegido por el dominio con lo que ya
examinó: `.claude/skills` y `.claude` sin seguir enlaces (tienen que ser directorios reales, §4.1); la raíz, en cambio,
se usa tal cual (FR-027), también si se alcanza por un enlace (un `HOME` enlazado), y cuenta como existente si
`Examinar(<Raiz>)` da `directorio` o un `enlace` que resuelve; en el ámbito local es el directorio de trabajo (`.`), que
existe siempre y no se examina. Es el sistema de ficheros donde vivirá la entrada de host. `Disponible(DS)` se pregunta
solo para decidir entre enlace y copia (una entrada de host que hay que crear, o una declarada `copia` que sigue siendo
un directorio real) y una vez por directorio en la invocación; el `Enlazador` del sistema crea y retira ahí un enlace de
prueba y deja el directorio con las mismas entradas y los mismos bytes (FR-048, FR-068). Un error de `Disponible` (la
sonda no se pudo retirar) es un fallo de la orden (exit 1), no un conflicto. Si no existe ninguno de los tres (solo con
`-g` y un `HOME` que no existe), no se pregunta y se predice `enlace`; la fase 2 aplica FR-024 (D9).

## 4. Conflictos de `install` (FR-040 a FR-043, FR-047)

Se comprueban **todos** antes de escribir nada. Cada entrada comprobada cae como mucho en una clase. El orden de
comprobación es el de esta sección; lo que queda por debajo de una entrada que no es un directorio real no se examina
(esa entrada es el conflicto).

### 4.1 Ámbito (antes que las skills)

| Comprobación | Conflicto | FR |
|---|---|---|
| Cada ruta de `Guardas` existe y no es `directorio` | `ruta que no es directorio` (esa ruta) | FR-027 |
| `<Neutro>/kitlegal.json` existe y es ilegible (manifiesto.md §3) | `manifiesto ilegible` | FR-035 |
| Con `--dir`: el manifiesto es legible y declara alguna entrada de host | `manifiesto con entradas de host` | FR-013 |
| Se enlaza en el host (§4.3) y `<Raiz>/.claude` existe y no es `directorio` (solo con `--host claude`; sin `--host` cuenta como ausente) | `ruta que no es directorio` (`.claude`) | FR-022, FR-023 |
| Se enlaza en el host y `<Raiz>/.claude/skills` existe y no es `directorio` | `ruta que no es directorio` (`.claude/skills`) | FR-026 |

Con un conflicto de las tres primeras filas no se sabe qué es de quién: se nombran los conflictos de ámbito y no se
examina ninguna skill.

### 4.2 Directorio neutro, por skill pedida

Sea `S` una skill pedida (empotrada), `D` su declaración en el manifiesto (si la hay) y `E = <Neutro>/<S>`.

| `E` es | `D` existe | Resultado |
|---|---|---|
| `ausente` | no | se instala |
| `ausente` | sí | se reponen sus ficheros (`actualizada`) |
| `directorio` | no | **carpeta ajena** |
| `directorio` | sí | se examinan sus ficheros (tabla siguiente) |
| `enlace` que resuelve | cualquiera | **enlace a otro sitio** |
| `enlace` que no resuelve | cualquiera | **enlace roto** |
| `fichero` u `otro` | cualquiera | **fichero** |

Ficheros de `S` (con `E` directorio y declarada). Para cada ruta `P` empotrada y cada ruta `Q` declarada, primero cada
**directorio intermedio** por debajo de `E` (p. ej. `references`): si existe y no es `directorio` → **ruta que no es
directorio**, y sus ficheros no se examinan. Después:

| Fichero | Estado en disco | Resultado |
|---|---|---|
| `Q` declarado y empotrado | `fichero` con la huella declarada | intacto: se reescribe si difiere de lo empotrado |
| `Q` declarado y empotrado | `fichero` con otra huella, o `enlace`, `directorio` u `otro` (no se abre ni se sigue) | **fichero editado** |
| `Q` declarado y empotrado | `ausente` | se escribe (FR-046) |
| `Q` declarado y ya no empotrado | `fichero` con la huella declarada | se retira (FR-046) |
| `Q` declarado y ya no empotrado | otra huella u otro tipo | **fichero editado** |
| `Q` declarado y ya no empotrado | `ausente` | se quita del manifiesto |
| `P` empotrado y no declarado | `ausente` | se escribe |
| `P` empotrado y no declarado | cualquier otra cosa | **fichero ajeno** |
| entrada no declarada que no es ninguna `P` | cualquiera | se deja intacta; no es conflicto (FR-047) |

### 4.3 Host `claude`, por skill pedida

**Se enlaza en esta ejecución** si el ámbito tiene hosts y: hay `--host claude` (FR-023), o `<Raiz>/.claude` es un
`directorio` real (FR-022). Sea `H = <Raiz>/.claude/skills/<S>`, `L = ../../.agents/skills/<S>` (FR-021) y `DH` la
entrada de host declarada (si la hay).

Si se enlaza:

| `H` es | `DH` | Resultado |
|---|---|---|
| `ausente` | ninguna o cualquiera | se crea el enlace `L` (o la copia si `Disponible(DS)` es falso, §3) |
| `enlace` con destino literal `L` (resuelva o no) | ninguna | se **adopta** y se declara `enlace` (FR-041) |
| `enlace` con destino literal `L` | `enlace` | sin cambios |
| `enlace` con destino literal `L` | `copia` | se declara `enlace` |
| `enlace` con otro destino, que resuelve | cualquiera | **enlace a otro sitio** |
| `enlace` con otro destino, que no resuelve | cualquiera | **enlace roto** |
| `directorio` | ninguna o `enlace` | **carpeta ajena** |
| `directorio` | `copia`, `Disponible(DS)` verdadero | se retira la copia y se crea `L` (FR-046), si **todos** sus ficheros están declarados e intactos y no contiene nada no declarado; si no, **fichero editado** o **fichero ajeno** (y **ruta que no es directorio** para un intermedio) y la copia no se toca |
| `directorio` | `copia`, `Disponible(DS)` falso | la copia se mantiene y se actualiza con las reglas del §4.2; sin nada que actualizar, la skill sale `sin cambios` (FR-045) |
| `fichero` u `otro` | cualquiera | **fichero** |

Si no se enlaza (sin `--host` y `.claude` ausente o que no es un directorio real): no se examina nada de host; cada
`DH` declarada se **quita del manifiesto** sin tocar el disco (FR-046) y la skill sale `actualizada`.

### 4.4 Lo que no se toca nunca

Las skills declaradas y no empotradas (FR-036) y las declaradas no pedidas (FR-034): ni se examinan ni se cambian, y
su entrada del manifiesto se copia byte a byte.

## 5. El plan y su orden de aplicación (FR-044, FR-045, D7)

Sin conflictos, el dominio devuelve un `Plan`:

| Campo | Contenido |
|---|---|
| `Skills` | por skill pedida: nombre, ruta presentada, `Estado` (`instalada` / `actualizada` / `sin cambios`) y enlaces resultantes con su modo previsto |
| `Retirar` | fase 1: ficheros declarados intactos que se reescriben o ya no se empotran; copias que pasan a enlace (ficheros y, después, sus directorios) |
| `Enlazar` | fase 2: `<Raiz>/.claude` y `<Raiz>/.claude/skills` si faltan (`CrearDirectorio`) y después las entradas de host que hay que crear, cada una con los ficheros de su recurso de copia |
| `Manifiesto` | fase 3: función de los modos que resultaron en la fase 2 → bytes canónicos del manifiesto final, o nada si no cambia |
| `Escribir` | fase 4: directorios que faltan y ficheros nuevos o reescritos del directorio neutro y de las copias |

`Aplicar(plan, escritor)` del dominio recorre las fases en orden a través del puerto `Escritor` —el adaptador solo
implementa las operaciones sueltas— y se para en el primer fallo (exit 1 con la operación y la ruta). Invariante que
prueba `TestFalloAMitadSeCompleta` con un disco en memoria que falla en la operación *n*, para cada *n*: volver a
planificar sobre el disco resultante da cero conflictos y, aplicado, el estado final. Con nada que cambiar, las listas están vacías y el
manifiesto no se escribe: el disco queda byte a byte igual (FR-033, FR-045).

Invariante de la predicción, que prueban `TestPlan` y `TestHallazgos` con un `Enlazador` sintético que admite enlaces
en un directorio de fuera del ámbito (el «temporal») y no en ninguno del ámbito: `Disponible` solo se pregunta por un
`DS` del ámbito (el test registra los directorios preguntados); el primer `install` deja `copia`; el segundo da
`sin cambios` en todas las skills con las listas del plan vacías y el disco byte a byte igual (FR-045); `--dry-run` da
la misma salida que la orden real en las dos ejecuciones (FR-048); y `doctor` no da el hallazgo «copia» (FR-069), así
que no propone una orden que no lo arreglaría (FR-066 (i)).

`--dry-run`: el mismo `Plan`, sin aplicar; `Skills` se presenta en `Resultado.Ensayo` (D13).

## 6. Hallazgos de `doctor` (FR-065 a FR-069)

Solo sobre skills declaradas **y empotradas**. Antes, las mismas comprobaciones de ámbito que §4.1 (filas 1-3): con
cualquiera de ellas, exit 1 nombrándola y ningún hallazgo. Sin manifiesto: exit 0 con `manifiesto` falso.

| # | Clase | Dónde y cuándo | Ruta del hallazgo | Orden (contracts/applet-skills.md §6) |
|---|---|---|---|---|
| 1 | fichero editado | fichero declarado (neutro o copia real) con otra huella o que no es fichero regular | la del fichero | `rm [-r] -- '<ruta>' && kitlegal skills install <S> …` (`-r` si es un directorio real) |
| 1 | fichero editado | fichero declarado que falta | la del fichero | `kitlegal skills install <S> …` |
| 2 | enlace colgando | `DH` declarada `enlace` con destino literal `L` que no resuelve | `H` | `kitlegal skills install <S> … --host claude` |
| 2 | enlace colgando | `<Neutro>/<S>` o un intermedio (neutro o copia) que es un enlace que no resuelve | la de ese enlace | `rm -- '<ruta>' && kitlegal skills install <S> …` |
| 3 | enlace a otro sitio | `DH` que ya no es lo declarado: enlace con otro destino, fichero u otro, directorio donde se declaró `enlace`, enlace donde se declaró `copia` | `H` | `rm [-r] -- '<ruta>' && kitlegal skills install <S> … --host claude` |
| 3 | enlace a otro sitio | `DH` que falta (o `.claude`/`.claude/skills` que ya no es un directorio real; no se lee a través) | `H` | `kitlegal skills install <S> … --host claude` |
| 3 | enlace a otro sitio | `<Neutro>/<S>` o un intermedio que es un enlace que resuelve, un fichero u otra entrada | la de esa entrada | `rm -- '<ruta>' && kitlegal skills install <S> …` |
| 4 | copia | `DH` declarada `copia` que sigue siendo un directorio real y `Disponible(DS)` es verdadero, con `DS` = `<Raiz>/.claude/skills` (§3) | `H` | `kitlegal skills install <S> … --host claude` |
| 5 | versión distinta | `version` del manifiesto distinta de la del binario (FR-077): **un** hallazgo y ninguno por skill | `<Neutro>/kitlegal.json` | `kitlegal skills install <todas las declaradas y empotradas> …` |
| 5 | versión distinta | si el manifiesto coincide: una skill declarada y empotrada con otra versión | `<Neutro>/<S>` | `kitlegal skills install <S> …` |

Las banderas de ámbito (`…`) van siempre; `--host claude` va si alguna de las skills nombradas tiene `DH` declarada y
nunca con `--dir`. Por debajo de una entrada que no es un directorio real no hay hallazgos de la clase 1. `doctor`
compara versiones también con un binario de desarrollo (la excepción de FR-073 es solo del aviso). Orden de la lista:
primero los que llevan `rm`, después los demás; dentro de cada grupo, por ruta byte a byte y, a igual ruta, por número
de clase. Invariantes que prueba `TestOrdenesDeDoctorArreglan`: con un único hallazgo, su orden lo hace desaparecer sin
producir otro; con varios, ejecutarlas en ese orden deja `doctor` limpio (FR-066 (i) y (ii)), siempre que las rutas no
contengan las excepciones que FR-066 enumera.

## 7. Aviso

`Aviso(disco, home, versionDelBinario, empotradas) (linea string, hay bool)`: la búsqueda y la regla de
[contracts/aviso.md](./contracts/aviso.md). Puro sobre el puerto `Disco`; nunca devuelve error: todo lo que no permite
decidir es «sin efecto».

## 8. Versiones

| Función | Regla |
|---|---|
| `FormaSemVer(v)` | `v` opcional + la gramática de SemVer 2.0.0 (núcleo `X.Y.Z` sin ceros a la izquierda, pre-release e identificadores de construcción) |
| `MismaVersion(a, b)` | quitar **una** `v` inicial a cada una y comparar byte a byte (`v0.1.0` = `0.1.0`; `0.1.0` ≠ `0.1.0+abc`) |

## 9. Errores y clases

| Error del dominio | Clase declarada (`schema.ConClase`) | Código |
|---|---|---|
| invocación inválida (contracts/applet-skills.md §2, filas 1-4) | `argumentos` | 2 |
| `-g` sin `HOME` | `inesperado` | 1 |
| `Conflictos` (lista ordenada de `{clase, ruta}`) | `inesperado` | 1 |
| `Hallazgos` (lista ordenada de `{clase, ruta, orden}`) | `inesperado` | 1 |
| `AmbitoIlegible` en `list`/`doctor` (`{clase, ruta}`) | `inesperado` | 1 |
| fallo de E/S (examinar, escribir o retirar la sonda del creador de enlaces, §3) | ninguna (lo no previsto) | 1 |

Ningún `panic` en rutas de usuario; `FuzzLeerManifiesto` lo comprueba sobre el lector del manifiesto.

## 10. Datos de salida (Go → JSON)

| Tipo | Campos (clave JSON) |
|---|---|
| `SkillInstalada` | `nombre`, `ruta`, `estado`, `enlaces` ([]`Enlace`) |
| `Enlace` | `host`, `ruta`, `modo` |
| `Listado` | `directorio`, `manifiesto` (bool), `version` (*string, `nullable`), `skills` ([]`SkillListada`) |
| `SkillListada` | `nombre`, `ruta`, `version`, `empotrada` (bool), `enlaces` ([]`Enlace`) |
| `Diagnostico` | `directorio`, `manifiesto` (bool), `version` (*string, `nullable`), `version_del_binario`, `hallazgos` ([]`Hallazgo`) |
| `Hallazgo` | `clase`, `ruta`, `orden` |

Listas vacías se serializan `[]`, nunca `null`.
