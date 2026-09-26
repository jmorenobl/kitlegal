# 0019 · Distribución: el binario por gestor de paquetes, con las skills dentro

- **Estado**: aceptada
- **Fecha**: 2026-09-26
- **Hito**: H19, adelantado a la fase 1 (detrás de H6) por el disparador que fijó el ADR 0013. Enmienda el ADR 0001 en
  lo que dice de `scripts/boe` y el ADR 0008 en lo que dice de `scripts/`; sustituye la frase «hasta entonces
  `make install`» del ADR 0013.

## Contexto y problema

Hasta H6 la única forma de instalar kitlegal es clonar el repositorio y ejecutar `make install`: `go install` deja el
binario en `$GOBIN` y `scripts/instalar-skills.sh` enlaza cada `skills/<nombre>/` en `~/.claude/skills/`. La skill
instalada **es** el árbol de trabajo del clon, y su `scripts/boe` es un enlace que llega, a través de
`bin/instalado/kitlegal`, al binario instalado. Eso vale para quien desarrolla y para nadie más: exige clonar, tener Go,
y la skill que ve cualquier proyecto es la de la rama en la que esté el clon.

El 2026-09-20, con H6 en curso, la persona que lleva el proyecto preguntó cómo instalarlo en otro proyecto y
actualizarlo con cada versión, y fijó el objetivo: que alguien distinto de quien desarrolla lo instale **sin clonar**,
en Claude Code, Codex o Antigravity, con la menor fricción posible, y que lo instalado en un proyecto se versione con
el proyecto, como `.kitlegal/`. Ese es literalmente el disparador con el que el ADR 0013 adelanta el release: «en el
momento en que alguien distinto de quien desarrolla tenga que instalarlo».

Hay dos hechos que condicionan cualquier solución:

- El binario es la herramienta de las skills y no es opcional: el sobre de salida, la caché, la huella y los avisos de
  vigencia son lo que hace citable una respuesta (constitución, II y VIII). Una skill sin binario no es kitlegal.
- El binario pesa 19 MB (11 MB comprimido) por plataforma, y en macOS todo lo que se extrae de un archivo descargado
  con el navegador queda en cuarentena: un ejecutable sin firmar y sin notarizar no arranca.

## Opciones consideradas

1. **Seguir con el clon y `make install`**, documentándolo mejor. Rechazada: no resuelve nada de lo pedido; exige Go y
   git, y la skill instalada cambia con la rama del clon.
2. **Un bundle `.skill`/zip por skill** que se descarga e instala en cada host. Para llevar el binario hay dos formas:
   gordo por plataforma (11 MB, el usuario elige plataforma, y en macOS hay que firmar y notarizar con un Developer ID)
   o fino con un shim que descarga el binario en el primer uso (KB, esquiva la cuarentena porque `curl` no la aplica,
   pero introduce un descargador propio —superficie de cadena de suministro que hay que asegurar con digest fijado y
   firma— y una versión de skill que declara qué binario espera, es decir, dos versiones que pueden divergir).
   Rechazada por esto último: el producto tendría dos artefactos con versión propia.
3. **El binario por el gestor de paquetes de cada plataforma, con las skills dentro, y un applet que las instala.**
   Un solo artefacto y una sola versión; la confianza, la firma, el `PATH` y la actualización los pone el gestor de
   paquetes, que es lo que ya hace bien; y `kitlegal skills install` es el adaptador por host. Elegida.

Sobre dónde se instalan las skills se consideraron copias por host y un directorio neutro con enlaces. Se elige lo
segundo: **una sola fuente de verdad por ámbito** (`.agents/skills/`), que es el patrón del estándar Agent Skills, el
de la CLI `skills` y el que ya usa este repositorio (`.claude/skills/*` son enlaces a `.agents/skills/*`).

## Decisión

1. **Las skills viajan dentro del binario** (`go:embed` de `skills/*/SKILL.md` y `skills/*/references/`). No hay
   `scripts/` en ninguna skill: la tabla de comandos invoca `kitlegal <applet> <verbo> …` con el binario en el `PATH`.
   El despacho por `os.Args[0]` del ADR 0001 sigue en el kernel, pero las skills ya no dependen de él.
2. **El binario se instala con el gestor de paquetes de la plataforma**, y todo lo genera goreleaser en cada etiqueta:
   `install.sh` para `curl | sh` (macOS y Linux sin más requisitos: descarga la release de la plataforma, verifica el
   checksum, instala en `~/.local/bin` y no toca los ficheros de arranque del shell), tap de Homebrew (macOS y Linux),
   bucket de Scoop (Windows), `.deb` y `.rpm` adjuntos a la release (nfpm) y `go install`. Checksums,
   SBOM, firma keyless con cosign y atestación de procedencia, como preveía H19 (§3 del roadmap, «Cadena de suministro»).
   Sin repositorio apt propio ni winget hasta que el uso lo pida.
3. **`kitlegal skills install | list | doctor`** es un applet como los demás (sobre, `--json`, `--dry-run`,
   `--describe`), con este contrato:
   - **Local por defecto**: instala en `./.agents/skills/<skill>/`, en el directorio de trabajo, que es lo que se
     versiona con el proyecto. `-g`/`--global` instala en `~/.agents/skills/<skill>/`.
   - **Los hosts son enlaces**, nunca copias: `--host claude` crea el enlace **relativo**
     `.claude/skills/<skill> -> ../../.agents/skills/<skill>` en el mismo ámbito (`./.claude/` o `~/.claude/`). Sin
     `--host`, además del directorio neutro enlaza en los hosts que detecte en ese ámbito por su directorio de
     configuración (`.claude/` presente) y los nombra en la salida; `--host claude` fuerza el enlace aunque `.claude/`
     no exista. Antigravity y Codex leen `.agents/skills/` directamente y no necesitan enlace; si la verificación
     pendiente dice otra cosa, ganan un valor de `--host`, no una copia.
   - En Windows, si el enlace no se puede crear, se copia y el manifiesto lo anota como copia, para que `doctor` lo
     reconozca y lo sustituya cuando pueda.
   - **Un manifiesto por ámbito**, en la raíz del directorio neutro, con la versión del binario que instaló y la huella
     de cada fichero: `install` solo sobrescribe lo que instaló él, es idempotente, y ante una entrada con el nombre de
     una skill que no es suya —una carpeta ajena, un fichero, un enlace a otro sitio o roto— la nombra, falla y no
     crea ni cambia nada (la regla que ya tenía `instalar-skills.sh`). `doctor` detecta ficheros editados, enlaces
     colgando, copias donde cabría un enlace y manifiestos más viejos que el binario.
   - **Aviso sin red**: cuando cualquier applet arranca y encuentra en su ámbito un manifiesto de skills más viejo que
     el binario, escribe una línea en stderr que nombra las dos versiones y la orden que lo arregla. Nunca en stdout,
     nunca cambia el código de salida, y nada consulta la red para saber si hay una versión nueva.
4. **`make install` queda como bucle de desarrollo**: `go install` y `kitlegal skills install -g --host claude`, que es
   lo que necesita el job de evals. La etiqueta y la publicación de la release siguen siendo humanas (constitución,
   capa 3; ADR 0007), y el repositorio —o al menos el tap, el bucket y las releases— tiene que ser público antes de la
   primera etiqueta.

## Consecuencias

- Para quien lo usa: dos órdenes (`curl -fsSL …/install.sh | sh` o `brew install jmorenobl/tap/kitlegal`, y
  `kitlegal skills install`), actualizar es repetir la primera —o `brew upgrade`— y la segunda, y el binario avisa
  cuando toca. `make install` no es una forma de instalar kitlegal: es el bucle de quien desarrolla y del job de evals. Lo instalado en un proyecto entra en su git.
- La forma de invocación en `SKILL.md` cambia de `scripts/boe …` a `kitlegal boe …`. Las evals miden que las skills
  siguen resolviendo lo mismo: sus `comandos` se declaran por applet, no por ruta.
- Desaparecen `scripts/instalar-skills.sh`, `bin/instalado/`, los enlaces que generaba `skills-sync` y los tests que
  los cubrían; los sustituyen los del applet `skills`.
- La constitución (principio VIII y la restricción sobre el estándar Agent Skills) deja de exigir `scripts/` como
  symlinks al binario; `CLAUDE.md` actualiza «Multicall» y «Skills sin código».
- Editar `skills/<nombre>/SKILL.md` en el clon ya no se ve en vivo en `~/.claude/skills/`: hay que reinstalar. Es el
  precio de que lo instalado sea siempre lo empotrado en el binario, y lo cubre `make install`.
- El release deja de cerrar la fase 3: desde H19 cualquier hito puede publicar, y `CHANGELOG.md` cierra su primera
  sección en `v0.1.0`.

## Pendiente de verificar (antes de fijarlo en código)

- Que Codex y Antigravity leen `.agents/skills/` en el proyecto y `~/.agents/skills/` en global, y cuál es el
  directorio global propio de cada uno. Claude Code lee `.claude/skills/` y `~/.claude/skills/`, comprobado.
- Los requisitos de la primera publicación en winget (revisión manual de Microsoft) y en `homebrew-core`
  (notoriedad), que quedan fuera hasta que haya demanda.
