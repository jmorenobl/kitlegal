# Contrato: errores tipados y códigos de salida de la caché

Cómo un fallo de `internal/cache` llega al código de salida sin que el paquete importe `internal/cli`, y
la tabla **cerrada** de situaciones. Decisiones en [research.md](../research.md) D3, D5, D8 y D9.

## 1. El puerto del dominio, sin cambios

`schema.ConClase` (H2) es el puerto: un error que lo implementa declara su clase y `cli.Clasificar` la
reconoce con `errors.As` después de los cinco sentinelas de H1. **H3 no toca `internal/core/schema` ni
`internal/cli`**: la caché produce tres de las seis clases existentes (FR-033) y no añade ninguna.

## 2. `cache.Error`

```go
type Error struct {
    Operacion string // «construir», «migrar», «leer», «escribir», «cerrar»
    Ruta      string // fichero o directorio implicado; vacío si no lo hay
    Origen    string // «opción ConDirectorio», «variable KITLEGAL_CACHE_DIR», «ruta por omisión»; vacío si no aplica
    Clave     string // clave implicada; vacía si no la hay
    Causa     error  // error de origen; Unwrap lo expone; no entra en el mensaje
    // clase, privada
}
func (e *Error) Error() string        // en español; nombra ruta, origen, clave, variable y versiones según la situación (FR-035)
func (e *Error) Unwrap() error        // Causa
func (e *Error) Clase() schema.Clase  // una de las tres de §3; un *Error nulo o a cero declara «inesperado»
```

Uso desde un adaptador, sin importar `internal/cli`:

```go
var fallo *cache.Error
if errors.As(err, &fallo) && fallo.Clase() == schema.ClaseFuenteNoDisponible { /* --offline sin entrada */ }
```

Envolver con `fmt.Errorf("…: %w", err)` conserva la clase (`errors.As` recorre `Unwrap`). El mensaje no
lleva la causa técnica (va al registrador a nivel `debug` y a `Unwrap`), ni traza (FR-035).

## 3. Tabla cerrada situación → clase → código

Las **ocho** situaciones de SC-011 (marcadas ★) y las siete que el diseño hace inevitables y FR-033
obliga a clasificar. La caché no produce ninguna otra clase: nunca `no-encontrado` (3), `limite-o-tos`
(5) ni `identidad-humana` (6).

| # | Situación | Clase | Código | Requisito / decisión | Mensaje nombra | Se mide en |
|---|---|---|---|---|---|---|
| 1 ★ | clave vacía en `Get` o `Put` | `argumentos` | 2 | FR-011 | la operación | `TestClasesDeError` |
| 2 ★ | vigencia `<= 0` en `Put` (sin escribir) | `argumentos` | 2 | FR-010 | la vigencia recibida | `TestClasesDeError` |
| 3 | opción inválida: `ConDirectorio("")`, `ConReloj(nil)`, `ConRegistrador(nil)` | `argumentos` | 2 | D2 | la opción | `TestClasesDeError` |
| 4 ★ | `KITLEGAL_CACHE_DIR` presente y vacía, en cualquier modo | `argumentos` | 2 | FR-022 | la variable | `TestClasesDeError` |
| 5 ★ | la ruta (opción o variable) existe y no es un directorio, en cualquier modo | `argumentos` | 2 | FR-022 | origen y ruta | `TestClasesDeError` |
| 6 ★ | modo normal: el directorio no se puede crear (`MkdirAll` falla: con `ENOTDIR` porque el padre es un fichero, o por permisos) o no se puede escribir (fallo al crear `cache.db`, o `SQLITE_READONLY_DIRECTORY` al migrar u operar) | `argumentos` | 2 | FR-022 | origen y ruta | `TestClasesDeError` y `TestDirectorioNoCreable` (padre que es un fichero); `TestIntegracionDirectorioNoEscribible/normal-argumentos` (permisos) |
| 7 | ruta por omisión indeterminable (`os.UserHomeDir` falla) | `argumentos` | 2 | D3 (no enumerada en FR-033) | `HOME` y `KITLEGAL_CACHE_DIR` | `TestClasesDeError` |
| 8 ★ | solo lectura: entrada ausente o expirada, base sin esquema, `cache.db` inexistente o directorio inexistente. «Inexistente» es la regla del contrato de apertura §6: `Stat` falla con `fs.ErrNotExist` **o** con `syscall.ENOTDIR` (un componente de la ruta, el padre, es un fichero: el directorio no existe como directorio) | `fuente-no-disponible` | 4 | FR-015, FR-016, FR-018 | la clave y que la invocación es de solo lectura | `TestClasesDeError`; `TestDirectorioNoCreable` (padre fichero, solo lectura); `TestSoloLecturaSinBaseNoCreaNada`; kernel: `offline-ausente`, `offline-expirada`, `offline-sin-base`, `offline-directorio-inexistente` |
| 9 ★ | versión de esquema mayor que la conocida (normal); distinta de 0 y de la conocida (solo lectura); fichero intacto | `inesperado` | 1 | FR-027 | ruta, versión esperada y encontrada | `TestClasesDeError`, `TestVersionMayorQueLaConocida` |
| 10 ★ | el fichero existe y no es una base utilizable (`SQLITE_NOTADB` u otro error al leer el esquema); no se borra | `inesperado` | 1 | FR-028 | ruta | `TestClasesDeError`, `TestFicheroInutilizable` |
| 11 ★ | escritura en solo lectura (también sobre un cliente sin base); nada escrito | `inesperado` | 1 | FR-017 | la clave | `TestClasesDeError`; kernel: `offline-guardar` |
| 12 | solo lectura: acceso denegado al comprobar el directorio o `cache.db` (`ErrPermission`), **cualquier otro fallo de `Stat` que no sea «inexistente»** según la regla de la fila 8 (E/S…), o error de apertura distinto de `SQLITE_READONLY_DIRECTORY` | `inesperado` | 1 | FR-015, FR-033 | ruta | `TestIntegracionDirectorioDenegado` |
| 13 | solo lectura en directorio no escribible con `cache.db-wal` presente y `-shm` ausente (`SQLITE_CANTOPEN`) | `inesperado` | 1 | D5 | ruta y los dos auxiliares | `TestIntegracionWALSinMemoriaCompartida` |
| 14 | fallo de E/S o del driver sobrevenido (`SQLITE_BUSY` tras la espera, disco lleno, error al cerrar); operación tras `Close` | `inesperado` | 1 | FR-033, FR-004 | operación, ruta y clave si la hay | `TestClasesDeError` (tras `Close`), `TestOperacionTrasCierre` |
| 15 | contexto cancelado o vencido durante cualquier operación, `New` incluido | `fuente-no-disponible` | 4 | D9 (no enumerada en FR-033; coincide con H1 «plazo agotado» y H2 fila 2) | operación y clave si la hay | `TestClasesDeError`, `TestContextoCancelado` |

`TestClasesDeError` (`internal/cache/errores_test.go`) construye el error real de cada fila que se puede
provocar sin permisos (1-11, 14, 15: **trece** subpruebas) y comprueba `cli.Clasificar` y
`cli.CodigoSalida`; las filas 12 y 13 dependen de permisos y se miden en los tests de integración
nombrados, que comprueban la misma clase con las mismas funciones.

## 4. Lo que el kernel hace con estos errores (H1/H2, sin cambios)

```
1. errors.Is contra los cinco sentinelas de cli (ninguno lo usa la caché)
2. errors.As contra schema.ConClase → la clase de cache.Error → codigoDeClase
3. si nada casa → inesperado (1)
```

Bajo `--timeout`, `conPlazoAgotado` del kernel envuelve cualquier error en `cli.ErrFuenteNoDisponible`
cuando el contexto venció: la fila 15 y el kernel dicen lo mismo (código 4).

## 5. Ausencia frente a fallo, por modo

| Modo | `Get` de una clave sin entrada vigente | Código de salida si el adaptador lo propaga |
|---|---|---|
| normal | `nil, false, nil` (FR-013): no es un fallo; el adaptador pide a la fuente | no cambia (FR-014) |
| solo lectura | `nil, false, *Error{clase: fuente-no-disponible}` (FR-016) | **4** |

## 6. Mensajes (en español, sin traza)

| Situación | Forma del mensaje |
|---|---|
| ruta inservible | `caché: la variable KITLEGAL_CACHE_DIR apunta a "/ruta", que existe y no es un directorio` |
| directorio no escribible | `caché: no se puede escribir en el directorio "/ruta" (opción ConDirectorio)` |
| versión ajena | `caché: "/ruta/cache.db" tiene el esquema en la versión 7 y este binario conoce la 1: no se modifica` |
| fichero inutilizable | `caché: "/ruta/cache.db" no es una base de datos utilizable; no se borra` |
| ausencia en solo lectura | `caché: no hay entrada vigente para "prueba:http://…" y la invocación es de solo lectura (--offline)` |
| escritura en solo lectura | `caché: no se puede escribir "prueba:http://…" en una caché de solo lectura` |
| tras cerrar | `caché: cliente cerrado; no se puede leer "…"` |

La forma exacta la fija `TestErrorMensajes`; lo que este contrato exige es **qué nombra** cada uno.
