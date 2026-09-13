<!-- Propuesta de cambio de H4. Esta sección la escribe la tarea que enlaza boe en el binario distribuido (T026); el resto del documento lo completan las tareas de cierre. -->

## Dependencias (constitución §V)

**Ninguna entrada nueva en `go.mod`** (FR-125). Lo que cambia es la **superficie del binario distribuido**: al
registrar `boe`, `cmd/kitlegal` enlaza `internal/source/boe`, `internal/httpx` e `internal/cache`, y con ellos doce
módulos de terceros que hasta H3 solo usaban esos paquetes y sus tests. Por eso se retiran
`TestElBinarioNoEnlazaHTTPX` y `TestElBinarioNoEnlazaCache`, que solo eran ciertos mientras ningún applet usara la red
ni la caché, y `modulosDelBinario` (`internal/arch_test.go`) pasa de seis a dieciocho módulos, cada uno justificado en
su línea; `TestDependenciasDelBinario` falla ante cualquier otro (FR-124, research.md D14, `docs/PENDIENTES.md`).

Medida sin red (`go list` consulta el módulo y la caché de módulos):

```sh
go list -deps -f '{{if .Module}}{{.Module.Path}}{{end}}' ./cmd/kitlegal | sort -u
go list -deps -f '{{.ImportPath}}|{{if .Module}}{{.Module.Path}}{{end}}|{{join .Imports ","}}' ./cmd/kitlegal
```

La primera da la lista; la segunda, para cada módulo, qué paquete de otro módulo lo importa, que es la columna «Lo
importa» de la tabla. Coinciden con lo que research.md D14 anticipó (S10): ni uno más ni uno menos.

| Módulo | Versión | Lo importa | Justificación |
|---|---|---|---|
| `github.com/temoto/robotstxt` | v1.1.2 | `internal/httpx` | Lista cerrada de §V (H2). Interpreta el `robots.txt` de cada sitio antes de la primera petición; una ruta denegada es el código 5 (constitución §I, FR-122) |
| `golang.org/x/time` | v0.16.0 | `internal/httpx` (`golang.org/x/time/rate`) | Lista cerrada de §V (H2). El ritmo por sitio, con el intervalo que fija la fila de la fuente en `docs/SOURCES.md` (FR-122) |
| `modernc.org/sqlite` | v1.58.0 | `internal/cache` | Lista cerrada de §V (H3, ADR 0002). SQLite sin cgo para la caché de las consultas (FR-090); sin él no hay binario con `CGO_ENABLED=0` |
| `modernc.org/libc` | v1.75.6 | `modernc.org/sqlite` | Entra con el controlador: es el entorno de C traducido a Go sobre el que corre SQLite, y no hay versión del controlador sin él (H3, `gates/pr-h3.md`) |
| `modernc.org/mathutil` | v1.7.1 | `modernc.org/libc` | Entra con `modernc.org/libc` |
| `modernc.org/memory` | v1.12.1 | `modernc.org/libc` | Entra con `modernc.org/libc` |
| `github.com/remyoudompheng/bigfft` | v0.0.0-20230129092748-24d4a6f8daec | `modernc.org/mathutil` | Entra con `modernc.org/mathutil` |
| `github.com/dustin/go-humanize` | v1.0.1 | `modernc.org/libc` | Entra con `modernc.org/libc` |
| `github.com/google/uuid` | v1.6.0 | `modernc.org/libc` | Entra con `modernc.org/libc` |
| `github.com/mattn/go-isatty` | v0.0.24 | `modernc.org/libc` | Entra con `modernc.org/libc` |
| `github.com/ncruces/go-strftime` | v1.0.0 | `modernc.org/libc` | Entra con `modernc.org/libc` |
| `golang.org/x/sys` | v0.47.0 | `modernc.org/sqlite`, `modernc.org/libc`, `modernc.org/memory`, `github.com/mattn/go-isatty` | Entra con el controlador y con lo que este arrastra, para las llamadas al sistema |

Los nueve que llegan con el controlador son exactamente los indirectos que H3 midió y justificó al fijar
`modernc.org/sqlite` (`specs/004-h3-internal-cache-sqlite/gates/pr-h3.md`, «Dependencias»); H4 no cambia ninguna
versión. Los seis de H1 (`alecthomas/kong`, `invopop/jsonschema` y lo que el segundo arrastra) siguen igual.
