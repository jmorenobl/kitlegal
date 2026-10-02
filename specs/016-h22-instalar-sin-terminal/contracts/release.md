# Contrato: la release, el snapshot y los dos flujos

Requisitos: FR-002, FR-003, FR-006, FR-031, FR-040 a FR-043, FR-060 a FR-066, FR-068. Decisiones en
[../research.md](../research.md) D6, D7 y D9 a D12; lo que no se puede comprobar sin publicar, en S1 a S6. Este
contrato **sustituye**, en lo que nombra, a `specs/009-h19-instalar-sin-clonar/contracts/release.md` §2 a §6, que no se
edita (FR-080); lo que no nombra sigue como allí.

## 1. `.goreleaser.yaml`

Lo que cambia, y lo que `TestConfiguracionDeLaRelease` pasa a fijar (§7). Todo lo demás —`ldflags`, `sboms`, `signs`,
`homebrew_casks`, `scoops`, `nfpms`, `changelog`— queda igual (FR-006).

| Sección | Hoy | Con H22 | FR |
|---|---|---|---|
| `builds[0]` | sin `id` | `id: kitlegal`, delante de `main` | FR-006 |
| `universal_binaries` | no existe | una entrada: `id: kitlegal-universal`, `ids: [kitlegal]`, `replace: false` y `hooks.post` con una orden (abajo) | FR-001, FR-002, FR-006 |
| `archives[0]` | sin `ids` | `ids: [kitlegal]`, delante de `name_template` | FR-006 |
| `checksum.extra_files` | `./scripts/install.sh` | además `./dist/kitlegal.mcpb` y `./dist/kitlegal-plugin.zip`, en ese orden | FR-002 |
| `release.extra_files` | `./scripts/install.sh` | además `./dist/kitlegal.mcpb` y `./dist/kitlegal-plugin.zip`, en ese orden | FR-002 |

```yaml
universal_binaries:
  - id: kitlegal-universal
    ids: [kitlegal]
    replace: false
    hooks:
      post:
        - cmd: go run ./cmd/empaquetar piezas -version {{ .Version }} -macos {{ .Path }} -windows dist/kitlegal_windows_amd64_v1/kitlegal.exe -icono mcp/icon.png -salida dist
          output: true
```

Con esto (medido, research V6, V8, V11): `dist/` lleva `kitlegal.mcpb` y `kitlegal-plugin.zip`; `checksums.txt`, una
línea de cada uno; los archivos siguen siendo los seis de hoy; el universal queda en
`dist/kitlegal-universal_darwin_all/kitlegal` y no se publica ni entra en `checksums.txt`; el cask y el bucket no
cambian. `PUBLISHER_TOKEN` sigue nombrado solo en los dos `token`.

## 2. `Makefile`

| Objetivo | Receta | En `ci` | FR |
|---|---|---|---|
| `release` | sin cambios: `$(GORELEASER) release --snapshot --clean --skip=publish,sign,sbom`. Ahora deja además las dos piezas | no | FR-002 |
| `snapshot-check` | sin cambios: las comprobaciones nuevas son subpruebas de `TestSnapshot` (§3) | no | FR-060 a FR-064, FR-066 |
| `plugin-check` (nuevo) | `go test -count=1 -tags=snapshot -run '^TestPluginValido$$' .` | no | FR-065 |

`plugin-check` entra en `.PHONY` y en `make help`: «valida con claude plugin validate el plugin del snapshot y el
catálogo de su versión (requiere Claude Code y el dist/ de make release; fuera de ci)». La línea de ayuda de `release`
nombra las dos piezas. `ci` no cambia sus prerrequisitos. El `Makefile` sigue sin nombrar `PUBLISHER_TOKEN`.

## 3. `TestSnapshot`: seis subpruebas nuevas

En un fichero nuevo de la raíz, `snapshot_piezas_test.go`, con la etiqueta `snapshot`, registradas en la tabla de
`TestSnapshot` junto a las cuatro de hoy, que no cambian. Cada una falla nombrando lo que falta, sobra o difiere.

| Subprueba | Falla si… | Requisitos |
|---|---|---|
| `dos-piezas` | falta `dist/kitlegal.mcpb` o `dist/kitlegal-plugin.zip`, o `checksums.txt` no lleva de cada uno una línea `<sha256>  <nombre>` con la huella del fichero | FR-060; SC-003 |
| `manifiesto-de-la-extension` | `manifest.json`, leído de forma estricta, lleva un campo que [data-model §3](../data-model.md) no nombra (`user_config` incluido) o le falta alguno; `manifest_version` no es `0.3`; un campo fijo no lleva su valor; un texto no es el de `internal/empaquetado`; `version` no es la que imprime `version` el binario del snapshot de la plataforma que ejecuta el test, sin su `v`; o `description` pasa de 120 caracteres | FR-061, FR-066; SC-004, SC-009 |
| `binarios-de-la-extension` | las entradas del `.mcpb` no son exactamente las cuatro; `server/kitlegal` no lleva el bit de ejecución; no es un universal de exactamente dos arquitecturas, una `amd64` y una `arm64` por su tipo de CPU (`debug/macho`); los bytes de alguna no son los del `kitlegal` de `kitlegal_darwin_<arquitectura>.tar.gz`; o `server/kitlegal.exe` no es el `kitlegal.exe` de `kitlegal_windows_amd64.zip` | FR-062; SC-005 |
| `icono-de-la-extension` | `icon.png` del `.mcpb` no es, byte a byte, `mcp/icon.png` del árbol, o no es un PNG de 512 × 512 px | FR-066; SC-009 |
| `servidor-de-la-extension` | con el `.mcpb` extraído en un directorio temporal cuyo nombre lleva espacios, el binario del snapshot de la plataforma del test puesto en `server/kitlegal` con un enlace simbólico, la orden de `mcp_config` con `${__dirname}` resuelto a ese directorio, sus argumentos (que tienen que ser `mcp` y `serve`) y `/` como directorio de trabajo, el cliente de `internal/mcp/mcptest` no completa el saludo o no lista las herramientas; o el conjunto de nombres y descripciones que lista no es el de `tools` del manifiesto; o no contiene las diez de hoy, escritas en el test | FR-061, FR-064; SC-004, SC-006 |
| `skills-del-plugin` | las entradas del plugin no son exactamente `.claude-plugin/plugin.json` y, bajo `skills/`, los ficheros que deja `kitlegal skills install` del binario del snapshot en un directorio vacío (sin su `kitlegal.json`), byte a byte; o `plugin.json`, leído de forma estricta, lleva un campo que [data-model §4](../data-model.md) no nombra (`mcpServers` incluido), o su versión o sus textos no son los del manifiesto | FR-063; SC-007 |

Cómo lo hacen, medido con el prototipo (research V17 a V19 y V32):

- Las herramientas se comparan como conjunto, por nombre, porque el orden de `tools/list` no es el del registro.
- El servidor se lanza con la orden como parámetro y `mcp` y `serve` como constantes, con `KITLEGAL_CACHE_DIR` y `HOME`
  en temporales. El binario de la plataforma se pone en `server/kitlegal` con un enlace simbólico y no copiándolo:
  escribir un ejecutable desde el proceso del test y lanzarlo a continuación da a veces «text file busy» en Linux
  cuando otra subprueba lanza un proceso a la vez, y las subpruebas de `TestSnapshot` corren en paralelo.
- La versión se compara con la que imprime el binario del snapshot de la plataforma del test: los seis se compilan
  con la misma (`builds[0].ldflags`), y `binarios-de-la-extension` ata los del `.mcpb` a los de los archivos.
- «Exactamente esas entradas» ya dice que el plugin no lleva `bin/`, `.mcp.json` ni ningún `.mcpb`.
- El paquete del paso se importa como `paso` (research D19).

## 4. `TestPluginValido` (`make plugin-check`)

En el mismo fichero y con la misma etiqueta, fuera de `TestSnapshot`:

1. extrae `dist/kitlegal-plugin.zip` en un directorio temporal;
2. escribe en otro `.claude-plugin/marketplace.json`, con lo que da `empaquetar catalogo` para la versión de
   `dist/metadata.json` y la huella de `kitlegal-plugin.zip` en `dist/checksums.txt`;
3. ejecuta `claude plugin validate .` en cada uno de los dos directorios, y falla, con lo que la orden escribió, si
   `claude` no está en el `PATH` o sale con un código distinto de 0.

No abre ninguna sesión con modelo ni usa ninguna credencial (FR-065). Ninguna tarea del run lo ejecuta: lo ejecuta el
trabajo `snapshot` de la propuesta de cambio (SC-008).

## 5. `.github/workflows/ci.yml`: trabajo `snapshot`

Los cuatro pasos de hoy y dos más, en este orden; los permisos (`contents: read`), el runner y la ausencia de secretos
y del token del flujo no cambian. El trabajo `ci` no se toca.

```yaml
      - name: Instalar Claude Code
        env:
          VERSION_DE_CLAUDE_CODE: 2.1.284
        run: npm install -g "@anthropic-ai/claude-code@${VERSION_DE_CLAUDE_CODE}"

      - name: Validar el plugin y el catálogo
        run: make plugin-check
```

`VERSION_DE_CLAUDE_CODE` es la de `jobs.evals.env` de `evals.yml`: `make ci` falla si las dos difieren (§7).

## 6. `.github/workflows/release.yml`

Tres trabajos: `publicar`, `humo` y `catalogo`. Sigue disparándolo solo una etiqueta `v*`, sin permisos a nivel de
flujo.

### 6.1 `publicar`

Sus pasos, sus permisos y sus dos tokens no cambian. `subject-path` de la atestación gana dos líneas,
`dist/kitlegal.mcpb` y `dist/kitlegal-plugin.zip`: nueve sujetos (FR-003). goreleaser ejecuta el paso por el gancho de
§1, sin ningún paso nuevo en el trabajo.

### 6.2 `humo`

Los seis pasos de hoy, sin tocar, y seis más detrás (FR-040 a FR-042), con las mismas reglas: ninguna acción —no obtiene
el código (FR-043)—, `shell: bash`, `working-directory: ${{ runner.temp }}`, `set -euo pipefail` como primera orden, y
los permisos y el `GH_TOKEN` de hoy. Los cuerpos son los medidos (research V20).

**7 · Descargar la extensión, el plugin y los archivos de sus binarios**

```bash
set -euo pipefail
gh release download "$GITHUB_REF_NAME" --repo jmorenobl/kitlegal --pattern kitlegal.mcpb --pattern kitlegal-plugin.zip --pattern kitlegal_darwin_amd64.tar.gz --pattern kitlegal_darwin_arm64.tar.gz --pattern kitlegal_windows_amd64.zip
```

**8 · Comprobar las huellas de la extensión, del plugin y de los tres archivos**

```bash
set -euo pipefail
awk '$2 == "kitlegal.mcpb" || $2 == "kitlegal-plugin.zip" || $2 == "kitlegal_darwin_amd64.tar.gz" || $2 == "kitlegal_darwin_arm64.tar.gz" || $2 == "kitlegal_windows_amd64.zip"' checksums.txt > huellas.txt
if [ "$(wc -l < huellas.txt)" -ne 5 ]; then
  echo "checksums.txt no lleva la huella de la extensión, del plugin y de los tres archivos" >&2
  exit 1
fi
sha256sum --check --strict huellas.txt
```

**9 · Verificar la atestación de la extensión y del plugin**

```bash
set -euo pipefail
gh attestation verify kitlegal.mcpb --repo jmorenobl/kitlegal
gh attestation verify kitlegal-plugin.zip --repo jmorenobl/kitlegal
```

**10 · Comprobar el manifiesto de la extensión**

```bash
set -euo pipefail
unzip -q kitlegal.mcpb -d extension
if ! jq -e --arg version "${GITHUB_REF_NAME#v}" '.manifest_version == "0.3" and .name == "kitlegal" and .version == $version and .homepage == "https://kitlegal.es" and .license == "EUPL-1.2" and .icon == "icon.png" and .server.type == "binary" and .server.entry_point == "server/kitlegal" and .server.mcp_config == {"command": "${__dirname}/server/kitlegal", "args": ["mcp", "serve"], "platform_overrides": {"win32": {"command": "${__dirname}/server/kitlegal.exe"}}} and .compatibility == {"platforms": ["darwin", "win32"]} and (has("user_config") | not)' extension/manifest.json > /dev/null; then
  echo "el manifiesto de kitlegal.mcpb no lleva la versión ${GITHUB_REF_NAME#v} o sus campos fijos" >&2
  exit 1
fi
```

**11 · Comprobar que los binarios de la extensión son los de los archivos**

```bash
set -euo pipefail
campo() { od -An -tu1 -j "$1" -N 4 extension/server/kitlegal | awk 'NF == 4 { print (($1 * 256 + $2) * 256 + $3) * 256 + $4 }'; }
if [ "$(campo 0)" -ne 3405691582 ] || [ "$(campo 4)" -ne 2 ]; then
  echo "server/kitlegal no es un binario universal de dos arquitecturas" >&2
  exit 1
fi
vistas=""
for base in 8 28; do
  case "$(campo "$base")" in
    16777223) arquitectura=amd64 ;;
    16777228) arquitectura=arm64 ;;
    *) echo "server/kitlegal lleva una arquitectura que no es amd64 ni arm64" >&2; exit 1 ;;
  esac
  hasta=$(($(campo $((base + 8))) + $(campo $((base + 12)))))
  del_universal=$(head -c "$hasta" extension/server/kitlegal | tail -c "$(campo $((base + 12)))" | sha256sum | cut -d ' ' -f 1)
  del_archivo=$(tar -xzOf "kitlegal_darwin_${arquitectura}.tar.gz" kitlegal | sha256sum | cut -d ' ' -f 1)
  if [ "$del_universal" != "$del_archivo" ]; then
    echo "la arquitectura ${arquitectura} de server/kitlegal no es el binario de kitlegal_darwin_${arquitectura}.tar.gz" >&2
    exit 1
  fi
  vistas="${vistas}${arquitectura} "
done
if [ "$vistas" != "amd64 arm64 " ] && [ "$vistas" != "arm64 amd64 " ]; then
  echo "server/kitlegal no lleva una arquitectura amd64 y una arm64: ${vistas}" >&2
  exit 1
fi
de_la_extension=$(sha256sum < extension/server/kitlegal.exe | cut -d ' ' -f 1)
del_archivo=$(unzip -p kitlegal_windows_amd64.zip kitlegal.exe | sha256sum | cut -d ' ' -f 1)
if [ "$de_la_extension" != "$del_archivo" ]; then
  echo "server/kitlegal.exe no es el binario de kitlegal_windows_amd64.zip" >&2
  exit 1
fi
```

La cabecera del universal son enteros de 32 bits en big-endian: la magia (`0xcafebabe`, 3 405 691 582), el número de
arquitecturas y, desde el byte 8, veinte bytes por arquitectura —CPU, subtipo, desplazamiento, tamaño y alineación—;
16 777 223 y 16 777 228 son los tipos de CPU de `amd64` y de `arm64` (research V3). `head -c … | tail -c …` y no
`tail | head`, que bajo `pipefail` hace fallar el paso; `awk 'NF == 4'`, porque algún `od` añade una línea en blanco
(research V20).

**12 · Arrancar el servidor y comprobar que lista las herramientas del manifiesto**

```bash
set -euo pipefail
mkfifo respuestas
cache=$(mktemp -d)
{
  exec 4< respuestas
  printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"humo","version":"0"}}}'
  IFS= read -r saludo <&4
  printf '%s\n' '{"jsonrpc":"2.0","method":"notifications/initialized"}' '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'
  IFS= read -r lista <&4
  printf '%s\n' "$lista" > lista.json
} | KITLEGAL_CACHE_DIR="$cache" ./kitlegal mcp serve > respuestas
jq -r '.result.tools[].name' lista.json | sort > anunciadas.txt
jq -r '.tools[].name' extension/manifest.json | sort > del-manifiesto.txt
if [ ! -s anunciadas.txt ] || ! cmp anunciadas.txt del-manifiesto.txt; then
  echo "las herramientas que lista el servidor no son las de tools del manifiesto" >&2
  exit 1
fi
```

`./kitlegal` es el binario de Linux de la release, que el paso 4 de hoy ya extrajo. El cliente lee cada respuesta de
la tubería con nombre antes de enviar lo siguiente y antes de cerrar la entrada del servidor, sin esperas ni procesos
en segundo plano; si el servidor termina sin responder, `read` falla y el paso sale en rojo.

### 6.3 `catalogo` (nuevo)

```yaml
  catalogo:
    needs: [humo]
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - name: Obtener el código de la etiqueta
        uses: actions/checkout@v7
        with:
          persist-credentials: false

      - name: Instalar Go
        uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
          cache: false

      - name: Componer el catálogo de la etiqueta
        shell: bash
        env:
          GH_TOKEN: ${{ github.token }}
        run: |
          set -euo pipefail
          gh release download "$GITHUB_REF_NAME" --repo jmorenobl/kitlegal --pattern checksums.txt --dir "$RUNNER_TEMP"
          huella=$(awk '$2 == "kitlegal-plugin.zip" { print $1 }' "$RUNNER_TEMP/checksums.txt")
          go run ./cmd/empaquetar catalogo -version "${GITHUB_REF_NAME#v}" -sha256 "$huella" -salida "$RUNNER_TEMP/marketplace.json"

      - name: Publicar el catálogo
        shell: bash
        env:
          GH_TOKEN: ${{ secrets.PUBLISHER_TOKEN }}
        run: |
          set -euo pipefail
          destino=repos/jmorenobl/kitlegal-plugins/contents/.claude-plugin/marketplace.json
          contenido=$(base64 < "$RUNNER_TEMP/marketplace.json" | tr -d '\n')
          if anterior=$(gh api "$destino" --jq .sha); then
            cuerpo=$(jq -n --arg message "kitlegal $GITHUB_REF_NAME" --arg content "$contenido" --arg sha "$anterior" '{message: $message, content: $content, sha: $sha}')
          else
            cuerpo=$(jq -n --arg message "kitlegal $GITHUB_REF_NAME" --arg content "$contenido" '{message: $message, content: $content}')
          fi
          gh api --method PUT "$destino" --silent --input - <<< "$cuerpo"
```

Lo que garantiza (FR-031, FR-032):

- Solo se ejecuta si `humo` sale en verde: si no, el catálogo sigue apuntando a la etiqueta anterior.
- No puede escribir en este repositorio (`contents: read`); `PUBLISHER_TOKEN` solo está en el entorno del paso que
  publica; `humo` sigue sin secretos.
- La versión es la etiqueta sin `v`; la dirección, la de la release de esa etiqueta; la huella, la que `checksums.txt`
  publicado da para `kitlegal-plugin.zip`. Sin esa línea, el paso que compone falla (la huella vacía no tiene su forma).
- Solo escribe `.claude-plugin/marketplace.json`, y lo sustituye entero.
- Si no puede escribir, `gh api` sale con 1 y el trabajo queda en rojo, sin afectar a `publicar` ni a `humo`. Volver a
  ejecutarlo escribe lo mismo.

Que `gh api` se comporte así no se pudo ejecutar (research S2); sus banderas sí están en su ayuda (V22).

## 7. `TestConfiguracionDeLaRelease`

Sigue leyendo `.goreleaser.yaml`, los flujos fuera de sus trabajos y cada trabajo de forma estricta. Cambia a la vez
que lo que fija, en la misma tarea, porque con la configuración nueva y el test de hoy `make ci` no pasa (research
V24).

| Subprueba | Qué fija de nuevo | FR |
|---|---|---|
| `plataformas` | `builds[0].id` es `kitlegal` | FR-006 |
| `universal` (nueva) | `universal_binaries` tiene una sola entrada, con `id: kitlegal-universal`, `ids: [kitlegal]`, `replace` escrito y falso, ningún gancho `pre` y un solo gancho `post`, con la orden de §1 carácter a carácter y `output: true` | FR-001, FR-002, FR-006, FR-068 |
| `archivos` | `archives[0].ids` es `[kitlegal]` | FR-006, FR-068 |
| `checksums`, `publicacion` | los tres `extra_files`, en orden | FR-002, FR-068 |
| `objetivos-del-makefile`, `ayuda` | la receta y la línea de ayuda de `plugin-check`; `ci` sigue sin él | FR-065 |
| `flujo-de-la-release` | los trabajos son `publicar`, `humo` y `catalogo` | FR-031, FR-068 |
| `atestacion` | los nueve sujetos | FR-003, FR-068 |
| `humo` | doce pasos; de cada uno de los seis nuevos, las órdenes de §6.2 como líneas enteras, y que `bash -n` lo lee | FR-040 a FR-043, FR-068 |
| `catalogo` (nueva) | `needs: [humo]`, `ubuntu-latest`, `permissions` exactamente `contents: read`, sin entorno de trabajo; cuatro pasos en orden —`actions/checkout` con `persist-credentials: false`, `actions/setup-go` con `go.mod` y sin caché, componer y publicar—; el entorno de componer es `GH_TOKEN` con el token del flujo y el de publicar, `GH_TOKEN` con `secrets.PUBLISHER_TOKEN`; las órdenes de §6.3 como líneas enteras; `bash -n` | FR-031, FR-068 |
| `tokens-de-la-publicacion` | de todos los flujos, `PUBLISHER_TOKEN` se nombra en dos sitios: el entorno del paso de goreleaser en `publicar` y el del paso que publica el catálogo; `release.yml` no lee ningún secreto fuera de esos dos entornos | FR-031, FR-068 |
| `trabajo-de-snapshot` | los seis pasos de §5 en orden, y `VERSION_DE_CLAUDE_CODE` igual a la de `jobs.evals.env` de `evals.yml` | FR-065 |

`probarSecretoDelPublicador` no cambia: en `.goreleaser.yaml`, `PUBLISHER_TOKEN` sigue solo en los dos `token`.
