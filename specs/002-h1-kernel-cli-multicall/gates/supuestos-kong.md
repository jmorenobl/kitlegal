# Supuestos de Kong — resultado de la comprobación (T002)

`research.md` [D24](../research.md) dejó cinco supuestos sobre
`github.com/alecthomas/kong` **sin verificar** porque la biblioteca no estaba en el módulo local y la
descarga quedó denegada en aquel entorno. T002 los comprueba contra la versión que `go get` trajo, antes
de escribir kernel alguno.

## Versión resuelta

| Dato | Valor |
|---|---|
| Módulo | `github.com/alecthomas/kong` |
| Versión | **v1.16.1** |
| Cómo se obtuvo | `go get github.com/alecthomas/kong` (descarga limpia; no estaba en `$GOMODCACHE`) |
| Estado en `go.mod` | requisito **directo** del módulo raíz |
| Dependencias que arrastra | **ninguna** (no añade ninguna entrada `// indirect`) |

## Veredicto

**Los cinco supuestos se cumplen. No se aplica ninguna contingencia.** En particular **S3, el
bloqueante, se cumple**: la ayuda integrada de Kong **no** se desactiva y la ayuda **no** se genera desde
el registro a falta de otra opción. Lo que T010 derivará del registro (`internal/app/ayuda.go`) sigue
siendo una decisión de diseño de FR-026, no una contingencia.

## Los cinco supuestos, uno por uno

Cada fila nombra el subcaso de `TestSupuestosDeKong`
([`internal/cli/supuestos_kong_test.go`](../../../internal/cli/supuestos_kong_test.go)) que lo comprueba.
El fichero monta el analizador con los escritores y la función de terminación sustituidos, de modo que el
hecho de que el test **retorne** es ya la comprobación de que ningún camino del análisis termina el
proceso por su cuenta.

| # | Supuesto | Resultado | Evidencia |
|---|---|---|---|
| **S1** | Hay un constructor que acepta gramática y opciones, y un método de análisis que recibe los argumentos ya separados de `os.Args[0]`. | ✅ **se cumple** | `func New(grammar any, options ...Option) (*Kong, error)` y `func (k *Kong) Parse(args []string) (*Context, error)`: la lista se pasa **sin** el nombre del programa. Subcaso `S1`. |
| **S2** | Los escritores de salida y de error que Kong usa para la ayuda y los mensajes son sustituibles. | ✅ **se cumple** | `func Writers(stdout, stderr io.Writer) Option`, que fija los campos públicos `Kong.Stdout` y `Kong.Stderr`. Subcaso `S2`. |
| **S3** | La función con la que Kong termina el proceso es sustituible, de modo que **ningún** camino invoque `os.Exit`. | ✅ **se cumple** — el bloqueante queda cerrado | `func Exit(exit func(int)) Option`, que fija el campo público `Kong.Exit` (por omisión `os.Exit`). Subcasos `S3` (dos). |
| **S4** | Las etiquetas cubren ayuda, valor por omisión, nombre, forma corta y variable de entorno, y hay soporte para `time.Duration`. | ✅ **se cumple** | Las cinco etiquetas se reflejan en el modelo (`Flag.Help`, `Flag.HasDefault`/`Default`, `Flag.Name`, `Flag.Short`, `Flag.Envs`) y `time.Duration` tiene decodificador nativo. Subcasos `S4` (dos). |
| **S5** | Un argumento o bandera desconocido produce un error distinguible, traducible a `ErrArgumentos`. | ✅ **se cumple** | El análisis devuelve `*kong.ParseError`, reconocible con `errors.As`, y el mensaje nombra el token concreto (`unknown flag --…`). Subcasos `S5` (dos). |

### Detalle de S2 — dónde acaba lo que Kong escribe

Con `kong.Writers` inyectado, la ayuda de `--help` (396 bytes con la gramática de prueba) aparece **en el
escritor inyectado** y el de error queda vacío. Sin esa opción, `Kong.Stdout` y `Kong.Stderr` son los
descriptores del proceso (`*os.File`): el test comprueba que los inyectados **no son** esos, que es lo que
demuestra que nada de lo escrito llegó a ellos. Kong no tiene ningún otro sumidero.

La comprobación se hace así —comparando los dos sumideros de Kong— y no sustituyendo `os.Stdout` y
`os.Stderr` del proceso a propósito porque R5 de
[`contracts/reglas-de-arquitectura.md`](../contracts/reglas-de-arquitectura.md) prohíbe **nombrar** esos
dos símbolos fuera de las dos raíces de composición, **sin excepción para los tests**, y T016 lo hará
cumplir con `forbidigo`. Un test que los nombrara obligaría a abrirles una excepción que la regla no
contempla.

### Detalle de S3 — cómo se comporta `--help` con la terminación sustituida

Esto es lo que se observa, y hay una consecuencia que **T005 tiene que atender**:

1. Kong imprime la ayuda en el escritor de salida inyectado.
2. Llama a la función de terminación **una vez, con 0**, que el test anota en lugar de terminar.
3. Como esa función no termina nada, **el análisis continúa** y acaba devolviendo un `*kong.ParseError`
   por la gramática incompleta (`expected "repetir"` cuando falta el mandato; `expected "<texto>"` cuando
   falta su argumento).

Es decir: **tras pedir `--help`, el análisis devuelve a la vez la ayuda ya escrita y un error**. La
precedencia que exigen FR-026, FR-028 y la tercera aclaración del `clarify` —`--help` gana sobre todo,
texto para personas en la salida estándar y código 0— **no puede leerse del error devuelto**;
hay que decidirla antes, con el pre-escaneo de D25 que construye T003, o con la anotación de la llamada a
la terminación. No es una contingencia de D24: el supuesto se cumple —el control vuelve y nadie termina el
proceso—, pero es un hecho nuevo que el plan no preveía y que T005 no puede ignorar.

La obligación que esto deja escrita para el kernel: **`internal/cli` monta Kong siempre con
`kong.Exit(…)` y `kong.Writers(…)`**. Sin la primera, `Kong.Exit` es `os.Exit` por omisión y un `--help`
terminaría el proceso desde dentro de `internal/cli`, rompiendo R4 y FR-035 sin que ningún linter lo viera
—porque la llamada está en Kong, no en nuestro código—.

### Detalle de S4 — lo que se comprueba y lo que no

| Etiqueta | Cómo se comprueba |
|---|---|
| `help:"…"` | `Flag.Help` del modelo, y el texto aparece en la ayuda emitida |
| `default:"30s"` | `Flag.HasDefault` cierto, `Flag.Default` igual a `"30s"`, y el campo vale 30 s tras un análisis que no la menciona |
| `name:"asunto"` | `Flag.Name` |
| `short:"v"` | `Flag.Short` igual al carácter `v` |
| `env:"…"` | `Flag.Envs` lleva el nombre de la variable, y la ayuda emitida la muestra como `($KITLEGAL_ASUNTO_DE_PRUEBA)` |

**Lo que no se comprueba y por qué**: que Kong **lea** de verdad el valor de la variable de entorno. Eso
exigiría mutar el entorno del proceso, que obliga a renunciar a `t.Parallel()` y a un `//nolint` de
`paralleltest`. Lo comprobado es que la etiqueta se reconoce y queda cableada en el modelo y en la ayuda;
ninguna bandera global del kernel usa `env` (las ocho de FR-018 no tienen variable de entorno asociada:
`KITLEGAL_LOG` la lee `internal/cli/log.go` por su cuenta, D14), así que este hito no depende de esa
lectura.

`time.Duration` tiene decodificador nativo en Kong (`mapper.go`, `durationDecoder`). Queda comprobado en
los dos sentidos: `--timeout=1500ms` deja el campo en 1,5 s, y `--timeout=abc` devuelve un error que
nombra la bandera (`--timeout: expected duration but got "abc"`). **No hace falta** la contingencia de
declarar `--timeout` como cadena y analizarla con `time.ParseDuration`.

### Detalle de S5 — el error es distinguible por tipo

El análisis devuelve `*kong.ParseError` (con `Unwrap()` y `ExitCode()`), reconocible con `errors.As`, para
la bandera desconocida, el mandato desconocido, el valor con formato inválido y el argumento obligatorio
ausente. Los cuatro casos de FR-027 quedan cubiertos por un único tipo.

Dos observaciones que T005 usa:

- Kong **no escribe nada** al fallar: ni en el escritor de salida ni en el de error. El mensaje lo emite
  el presentador, que es lo que FR-040 exige. Solo imprime por su cuenta `Kong.Fatalf` y
  `Kong.FatalIfErrorf`, que son además los dos únicos otros sitios donde Kong llama a la terminación: **el
  kernel no llama a ninguno de los dos**.
- Al fallar, Kong **tampoco** llama a la función de terminación. El código de salida lo decide el kernel
  con `CodigoSalida`, no Kong con su `ParseError.ExitCode()`.

Como todo fallo del análisis es, por construcción, un fallo de la línea de órdenes, T005 traduce
`*kong.ParseError` a `ErrArgumentos` sin distinguir subtipos. La contingencia de S5 —clasificar como
`argumentos` todo error del análisis— resulta ser, de hecho, la implementación elegida, y no porque el
supuesto fallara.

## Los dos supuestos menores sobre herramientas

D24 dejaba además dos supuestos verificables en local. El segundo se comprueba en esta tarea; el primero
no, porque su sujeto no existe todavía:

| Supuesto | Estado |
|---|---|
| `golangci-lint` v2 no excluye `testdata/` por omisión | **Pendiente**: `internal/app/testdata/ejemplo` lo crea T014 y T013 enumera esos paquetes en `lint`, `fmt` y `fmt-check`. Se comprueba allí. |
| `go mod tidy -diff` y `go mod verify` siguen en verde con dependencias reales en `go.sum` | ✅ **se cumple**: `make ci` pasa con `kong` y `testify` en el módulo. |

## Divergencia de nomenclatura respecto a D24 y D8

La tabla de D8 de `research.md` nombra el sentinela de la clase `limite-o-tos` como `ErrLimitado`.
`tasks.md` lo nombra **`ErrLimiteOTos`**, y es el nombre implementado, por coherencia con
`schema.ClaseLimiteOTos`, que ya existe desde T001. Los otros cuatro sentinelas conservan el nombre de
D8. Es una divergencia de nombre, no de comportamiento: la clase, el código 5 y la clasificación son los
que fija D8.

## Comprobación mecánica de la traducción a códigos de salida

La otra mitad de T002 es [`internal/cli/errors.go`](../../../internal/cli/errors.go). Lo que D8 dejaba
como intención y aquí queda comprobado:

- La traducción `schema.Clase` → código de salida está en **un único** `switch` sobre `Clase`, **sin rama
  `default`** (`codigoDeClase`), y la rama por defecto vive fuera, en `Clasificar`, que devuelve
  `ClaseInesperado` antes de entrar en él.
- El linter `exhaustive` **falla de verdad** si falta una clase. Ejercido quitando a propósito la rama de
  `schema.ClaseLimiteOTos`:
  `internal/cli/errors.go:115:2: missing cases in switch of type schema.Clase: schema.ClaseLimiteOTos (exhaustive)`.
  Rama restaurada acto seguido. Esto importa porque `.golangci.yml` tiene
  `exhaustive.default-signifies-exhaustive: true`: **una rama `default` desactivaría el control**, que es
  el motivo por el que la tarea la prohíbe.
- Envolver con `fmt.Errorf("…: %w", …)` no cambia la clase, en uno y en dos niveles, ni cuando el texto de
  la envoltura nombra otra clase (FR-032).
