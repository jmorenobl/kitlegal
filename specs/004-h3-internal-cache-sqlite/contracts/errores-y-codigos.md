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

Las **ocho** situaciones de SC-011 —que corresponden a las **nueve** filas marcadas ★ (1, 2, 4, 5, 6, 8,
9, 10, 11), porque un bullet de SC-011 puede agrupar varias filas: «ruta o variable inservibles» son las
filas 4, 5 y 6, y «ausencia… en solo lectura» y «directorio inexistente… en solo lectura» se resuelven
ambos en la fila 8— y las **seis** restantes (3, 7, 12, 13, 14, 15), que el diseño hace inevitables y
FR-033 obliga a clasificar. La caché no produce ninguna otra clase: nunca `no-encontrado` (3),
`limite-o-tos` (5) ni `identidad-humana` (6).

| # | Situación | Clase | Código | Requisito / decisión | Mensaje nombra | Se mide en |
|---|---|---|---|---|---|---|
| 1 ★ | clave vacía en `Get` o `Put` | `argumentos` | 2 | FR-011 | la operación | `TestClasesDeError` |
| 2 ★ | vigencia `<= 0` en `Put` (sin escribir) | `argumentos` | 2 | FR-010 | la vigencia recibida | `TestClasesDeError` |
| 3 | opción inválida: `ConDirectorio("")`, `ConReloj(nil)`, `ConRegistrador(nil)` | `argumentos` | 2 | D2 | la opción | `TestClasesDeError` |
| 4 ★ | `KITLEGAL_CACHE_DIR` presente y vacía, en cualquier modo | `argumentos` | 2 | FR-022 | la variable | `TestClasesDeError` |
| 5 ★ | la ruta (opción o variable) existe y no es un directorio, en cualquier modo | `argumentos` | 2 | FR-022 | origen y ruta | `TestClasesDeError` |
| 6 ★ | modo normal: el directorio no se puede crear (`MkdirAll` falla: con `ENOTDIR` porque el padre es un fichero, o por permisos) o no se puede escribir (fallo al crear `cache.db` cuando no existe, o `SQLITE_READONLY_DIRECTORY` al migrar u operar); o `cache.db` **ya existe** y es él el que no se deja abrir para escribir (permisos del propio fichero): misma clase, y el mensaje nombra el fichero y no el directorio (FR-035) | `argumentos` | 2 | FR-022 | origen y ruta (la del directorio o la del fichero, según cuál sea el culpable) | `TestClasesDeError` y `TestDirectorioNoCreable` (padre que es un fichero); `TestIntegracionDirectorioNoEscribible/normal-argumentos` (permisos del directorio); `TestIntegracionFicheroDenegado/normal-argumentos` (permisos del fichero) |
| 7 | ruta por omisión indeterminable (`os.UserHomeDir` falla) | `argumentos` | 2 | D3 (no enumerada en FR-033) | `HOME` y `KITLEGAL_CACHE_DIR` | `TestClasesDeError` |
| 8 ★ | solo lectura: entrada ausente o expirada, base sin esquema, `cache.db` inexistente o directorio inexistente. «Inexistente» es la regla del contrato de apertura §6: `Stat` falla con `fs.ErrNotExist` **o** con `syscall.ENOTDIR` (un componente de la ruta, el padre, es un fichero: el directorio no existe como directorio) | `fuente-no-disponible` | 4 | FR-015, FR-016, FR-018 | la clave y que la invocación es de solo lectura | `TestClasesDeError`; `TestDirectorioNoCreable` (padre fichero, solo lectura); `TestSoloLecturaSinBaseNoCreaNada`; kernel: `offline-ausente`, `offline-expirada`, `offline-sin-base`, `offline-directorio-inexistente` |
| 9 ★ | versión de esquema mayor que la conocida (normal); distinta de 0 y de la conocida (solo lectura); fichero intacto | `inesperado` | 1 | FR-027 | ruta, versión esperada y encontrada | `TestClasesDeError`, `TestVersionMayorQueLaConocida` |
| 10 ★ | el fichero existe y no es una base utilizable (`SQLITE_NOTADB` u otro error al leer el esquema); no se borra | `inesperado` | 1 | FR-028 | ruta | `TestClasesDeError`, `TestFicheroInutilizable` |
| 11 ★ | escritura en solo lectura (también sobre un cliente sin base); nada escrito | `inesperado` | 1 | FR-017 | la clave | `TestClasesDeError`; kernel: `offline-guardar` |
| 12 | solo lectura: acceso denegado al comprobar el directorio o `cache.db` (`ErrPermission`), **cualquier otro fallo de `Stat` que no sea «inexistente»** según la regla de la fila 8 (E/S…); `cache.db` que está pero **no se deja leer** (permisos del propio fichero, comprobado con `os.Open` antes de mirar `-wal`: el mensaje nombra el fichero y el acceso denegado, nunca el directorio, FR-035); o reapertura con `immutable=1` que vuelve a fallar (el mensaje nombra el fichero sin atribuir la causa a nada que no se haya comprobado) | `inesperado` | 1 | FR-015, FR-033 | ruta | `TestIntegracionDirectorioDenegado` (`Stat` denegado), `TestIntegracionFicheroDenegado/solo-lectura-inesperado` (el fichero no se deja leer), `TestIntegracionReaperturaInmutableFalla` (la reapertura inmutable vuelve a fallar: páginas estropeadas, `SQLITE_CORRUPT`) |
| 13 | solo lectura en directorio no escribible con `cache.db-wal` presente y `-shm` ausente: la primera consulta falla con `SQLITE_CANTOPEN` (14) —y con 1544 si el driver lo reportara así: la comprobación de `-wal` se dispara ante los dos códigos— | `inesperado` | 1 | D5 | ruta y los dos auxiliares | `TestIntegracionWALSinMemoriaCompartida` |
| 14 | fallo de E/S o del driver sobrevenido (disco lleno, error al cerrar, páginas de la tabla de entradas estropeadas: `SQLITE_CORRUPT` al leer o escribir); la base **bloqueada** por otra invocación más tiempo que la espera de 5 s (`SQLITE_BUSY` tras agotar `esperaAnteBloqueo`, contrato de apertura §4: el mensaje dice que está bloqueada y nombra la espera, nunca «inutilizable»); operación tras `Close` | `inesperado` | 1 | FR-031, FR-033, FR-004 | operación, ruta y clave si la hay; en el bloqueo, además la espera agotada | `TestClasesDeError` (tras `Close`; bloqueo agotado en `New` y `Put`; `Get` y `Put` sobre páginas de entradas estropeadas), `TestOperacionTrasCierre`, `TestBloqueoQueNoSeSueltaAgotaLaEspera`, `TestFalloDelControladorAlOperar` |
| 15 | contexto cancelado o vencido durante cualquier operación, `New` incluido, **también cuando vence durante la espera ante un bloqueo** (el motor no la interrumpe; el cliente mira el contexto entre tramos y termina en cuanto acaba el tramo en curso, contrato de apertura §4) **y cuando se cancela durante la migración**, la única escritura de `New` (la transacción se deshace entera y no queda ninguna versión registrada); la causa lleva el error del contexto además del que devolviera el driver | `fuente-no-disponible` | 4 | D9 (no enumerada en FR-033; coincide con H1 «plazo agotado» y H2 fila 2) | operación y clave si la hay | `TestClasesDeError` (contexto terminado antes, durante la espera ante bloqueo —cada llamada con su propio plazo— y durante la migración), `TestContextoCancelado`, `TestNewConLaBaseBloqueadaRespetaElContexto`, `TestPutConLaBaseBloqueadaRespetaElContexto`, `TestContextoCanceladoDuranteLaMigracion` |

`TestClasesDeError` (`internal/cache/errores_test.go`) construye el error real de cada fila que se puede
provocar sin permisos (1-11, 14, 15: **trece** subpruebas) y comprueba `cli.Clasificar` y
`cli.CodigoSalida`; las filas 12 y 13 dependen de permisos y se miden en los tests de integración
nombrados —tres para la 12, uno por cada forma de la situación, y uno para la 13—, que comprueban la
misma clase con las mismas funciones.

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
| fichero no escribible (fila 6, `cache.db` ya existe) | `caché: no se puede abrir "/ruta/cache.db" para escribir (opción ConDirectorio): acceso denegado` (el sufijo `: acceso denegado` solo cuando la causa es `fs.ErrPermission`) |
| fichero ilegible en solo lectura (fila 12) | `caché: no se puede leer "/ruta/cache.db" en solo lectura: acceso denegado` (mismo sufijo condicional) |
| bloqueo agotado al construir o migrar (fila 14) | `caché: "/ruta/cache.db" está bloqueada por otra invocación y la espera de 5s se agotó` |
| bloqueo agotado al leer o escribir (fila 14) | `caché: no se pudo escribir "prueba:http://…": "/ruta/cache.db" está bloqueada por otra invocación y la espera de 5s se agotó` |
| contexto terminado al construir o migrar (fila 15) | `caché: el contexto terminó antes de construir la caché en "/ruta/cache.db"` |
| versión ajena | `caché: "/ruta/cache.db" tiene el esquema en la versión 7 y este binario conoce la 1: no se modifica` |
| fichero inutilizable | `caché: "/ruta/cache.db" no es una base de datos utilizable; no se borra` |
| WAL sin memoria compartida (fila 13) | `caché: "/ruta/cache.db" tiene un registro de escritura "/ruta/cache.db-wal" y el directorio no permite crear "/ruta/cache.db-shm": no se puede leer en solo lectura` |
| ausencia en solo lectura | `caché: no hay entrada vigente para "prueba:http://…" y la invocación es de solo lectura (--offline)` |
| escritura en solo lectura | `caché: no se puede escribir "prueba:http://…" en una caché de solo lectura` |
| tras cerrar | `caché: cliente cerrado; no se puede leer "…"` |

La forma exacta la fija `TestErrorMensajes`; lo que este contrato exige es **qué nombra** cada uno. Qué
forma toma un error concreto lo decide el **constructor interno de su situación** (§2), no solo la clase
ni los campos exportados: dos situaciones pueden compartir clase, operación y ruta —la versión ajena y el
fichero inutilizable, o el fichero inutilizable y el WAL sin memoria compartida— y aun así decir cosas
distintas, y los detalles que no caben en los campos exportados (las dos versiones, los dos auxiliares)
viajan con ese constructor.
