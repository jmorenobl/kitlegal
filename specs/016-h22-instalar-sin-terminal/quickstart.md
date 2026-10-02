# Quickstart · H22 · Instalar sin terminal

Guía de validación del hito, con el árbol ya implementado. Cada escenario se ejecuta tal cual desde la raíz del
repositorio, en una misma terminal (§2 a §4 comparten la variable `trabajo`). Ninguno toca la red, el índice ni el
historial de git; lo que escriben fuera de un directorio temporal es `dist/`, que está en `.gitignore`, y §9 lo retira.
Hacen falta `go`, `make`, `unzip` y `jq`.

Las órdenes y las salidas de §2 a §5 son las ejecutadas con el prototipo del plan (research, cabecera); los nombres de
test son los que fija [plan.md](./plan.md).

## 1. Los tests del paso, en `make ci` (US3, US5)

```sh
go test -count=1 ./internal/empaquetado/
go test -count=1 -run '^TestHerramientasAnunciadas$' ./internal/app/
go test -count=1 -run '^TestElBinarioNoEnlazaElPaso$' ./internal/
go test -count=1 -run '^TestConfiguracionDeLaRelease$' .
```

Las cuatro terminan con `ok`. La primera ejecuta `TestPiezas`, `TestPiezasReproducibles`, `TestPiezasSinEntrada`,
`TestIcono`, `TestDescripcionCorta`, `TestCatalogo` y `TestEjecutar` (FR-004, FR-005, FR-015, FR-016, FR-067, FR-070).

## 2. El paso a mano, dos veces sobre los mismos binarios (US3.2; FR-004, FR-010, FR-020)

```sh
trabajo=$(mktemp -d)
printf 'binario de macOS de prueba' > "$trabajo/macos"
printf 'binario de Windows de prueba' > "$trabajo/windows.exe"
mkdir "$trabajo/a" "$trabajo/b"
go run ./cmd/empaquetar piezas -version 0.0.0-prueba -macos "$trabajo/macos" -windows "$trabajo/windows.exe" -icono mcp/icon.png -salida "$trabajo/a"
go run ./cmd/empaquetar piezas -version 0.0.0-prueba -macos "$trabajo/macos" -windows "$trabajo/windows.exe" -icono mcp/icon.png -salida "$trabajo/b"
cmp "$trabajo/a/kitlegal.mcpb" "$trabajo/b/kitlegal.mcpb" && cmp "$trabajo/a/kitlegal-plugin.zip" "$trabajo/b/kitlegal-plugin.zip" && echo "iguales byte a byte"
unzip -Z1 "$trabajo/a/kitlegal.mcpb"
unzip -Z1 "$trabajo/a/kitlegal-plugin.zip"
unzip -p "$trabajo/a/kitlegal.mcpb" manifest.json | jq -r '.version, (.tools | length), (.description | length)'
```

Esperado: `iguales byte a byte`; las cuatro entradas de la extensión (`manifest.json`, `icon.png`, `server/kitlegal`,
`server/kitlegal.exe`); las seis del plugin (`.claude-plugin/plugin.json` y los cinco ficheros de `skills/`); y
`0.0.0-prueba`, `10` y `71`.

## 3. Sin una entrada, el paso falla y la nombra (US3.3; FR-005)

```sh
go run ./cmd/empaquetar piezas -version 0.0.0-prueba -macos "$trabajo/macos" -windows "$trabajo/no-existe.exe" -icono mcp/icon.png -salida "$trabajo/a"
echo "código: $?"
```

Esperado, en la salida de error, `empaquetar: falta el binario de Windows: open …/no-existe.exe: no such file or
directory`, la línea `exit status 1` de `go run` y `código: 1`.

## 4. El catálogo de una versión (US2.1; FR-030)

```sh
huella=$( (sha256sum "$trabajo/a/kitlegal-plugin.zip" 2>/dev/null || shasum -a 256 "$trabajo/a/kitlegal-plugin.zip") | cut -d ' ' -f 1)
go run ./cmd/empaquetar catalogo -version 0.4.0 -sha256 "$huella" -salida "$trabajo/marketplace.json"
jq -r '(.plugins | length), .plugins[0].source.source, .plugins[0].source.url, .plugins[0].version' "$trabajo/marketplace.json"
rm -rf "$trabajo"
```

Esperado: `1`, `archive`, `https://github.com/jmorenobl/kitlegal/releases/download/v0.4.0/kitlegal-plugin.zip` y
`0.4.0`. La última orden retira el temporal de §2 a §4.

## 5. El snapshot deja las dos piezas, con sus huellas (US1.1, US3.1; FR-002, FR-006; SC-003)

```sh
make release
ls dist/kitlegal.mcpb dist/kitlegal-plugin.zip
grep -E '  (kitlegal\.mcpb|kitlegal-plugin\.zip)$' dist/checksums.txt
ls dist/*.tar.gz dist/*.zip
```

Esperado: los dos ficheros; dos líneas de `checksums.txt`, una de cada uno; y, en el último listado, los seis archivos
de hoy más `dist/kitlegal-plugin.zip`, sin ningún `kitlegal_darwin_all.tar.gz`. Escribe `dist/` (lo vacía antes) y
nada versionado. En el equipo del plan tardó 18 s.

## 6. Las comprobaciones del snapshot (US1.2 a US1.6, US5; FR-060 a FR-064, FR-066; SC-003 a SC-007, SC-009)

```sh
make snapshot-check
```

Esperado: `ok` en sus dos órdenes. La primera es `TestSnapshot`, con las cuatro subpruebas de hoy y las seis nuevas
(`dos-piezas`, `manifiesto-de-la-extension`, `binarios-de-la-extension`, `icono-de-la-extension`,
`servidor-de-la-extension`, `skills-del-plugin`); la segunda, los guiones `instalador-`. Necesita el `dist/` de §5.
Para ver los nombres:

```sh
go test -count=1 -tags=snapshot -run '^TestSnapshot$' -v . | grep -E '^ +--- '
```

## 7. El binario distribuido es el de antes (US3.6; FR-001, FR-080)

```sh
make schema-check
make skills-check
go test -count=1 -run '^TestHerramientasDelServidor$' ./internal/app/
```

Esperado: las tres en verde, sin haber tocado lo que comprueban.

## 8. El veredicto del repositorio (US3.4; FR-068; SC-011)

```sh
make ci
```

Esperado: `ci: todos los controles en verde`, con `goreleaser check` incluido. `make vuln`, dentro de `make ci`,
necesita la red de las herramientas de Go.

## 9. Limpieza

```sh
rm -rf dist
git status --short
```

Esperado: `git status --short` no enseña nada que no enseñara antes de §1.

## 10. Lo que ejecutan la persona o el workflow, y ninguna tarea

- **`make plugin-check`** (SC-008): necesita Claude Code en el `PATH` y el `dist/` de §5. Lo ejecuta el trabajo
  `snapshot` de la propuesta de cambio, con la versión de `evals.yml`; ninguna sesión del run ejecuta `claude`.
- **`release.yml`**: lo dispara la etiqueta que empuja una persona (ADR 0020). Ahí se ejecutan por primera vez la
  atestación de los dos ficheros, las seis comprobaciones nuevas de `humo` y el trabajo `catalogo`. Antes, la persona da
  a `PUBLISHER_TOKEN` permiso de escritura sobre `jmorenobl/kitlegal-plugins` (FR-032).
- **La prueba de SC-002**, tras publicar una release: en un Mac distinto del que compiló, sin kitlegal y sin abrir una
  terminal, descargar `kitlegal.mcpb` y `kitlegal-plugin.zip` de la release con un navegador; abrir el primero con
  doble clic e instalarlo; subir el segundo en *Customize > Plugins > Add > Upload plugin*, en el modo de chat; y, en
  una conversación nueva y sin carpeta, preguntar «¿qué dice el art. 21 de la Ley 39/2015?». Esperado:
  `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`. Se anota además si macOS bloquea el binario, cómo llega
  una versión nueva de cada pieza, si la app admite el marketplace y qué responde Claude en la web con el plugin y sin la
  extensión.
- **El cierre**: empujar la rama, abrir la propuesta de cambio y medir la CI y las evals lo hace el workflow tras la
  revisión final.
