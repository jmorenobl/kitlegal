# T034 · intento 1 de 3 (2026-09-15): en verde, marcada

Base `c4c7613` (`feat(H5): T033`); el workflow commitea este intento como `feat(H5): T034`. Verificación: `go clean
-testcache` y `make ci` en primer plano, código 0, `ci: todos los controles en verde` y ningún `FAIL` (log en
`gates/ci.log`, ignorado).

## Qué cambia

Tests nuevos en `internal/evals/{trazas,sesion,preparar,formato}_test.go`, dos reestructuraciones en `trazas.go` y
`formato.go` y una frase de data-model §9 (regla 5). Ningún cambio en `testdata/`, `schemas/`, los casos que fija el
contrato del job §9 (`TestLeerTrazas`, `TestLeerSesion` y los demás siguen iguales), `codecov.yml` ni `.golangci.yml`;
0 `//nolint` y 0 `t.Skip` añadidos frente a `main`. Cada test nuevo escribe sus entradas como literales en `t.TempDir()`,
y los fallos de disco salen de la estructura de ese directorio —un fichero donde se espera una carpeta da `ENOTDIR`; una
carpeta donde se espera un fichero, `EISDIR`—, no de permisos.

## Ramas cubiertas (líneas de la lista de `gates/tarea-T030.md`, sobre `c4c7613`)

| Test nuevo | Ramas | Entradas |
|---|---|---|
| `TestLeerTrazasConDefectos` (21 subtests) | `trazas.go` 452, 457, 519, 521, 528, 582, 608, 627, 637, 641, 654, 659, 671, 691, 723, 743, 745, 770, 788, 799, 804, 814, 852, 882, 887-889, 908-910, 981, 1022, 1061-1063, 1085 | una entrada `strace.log` y una carpeta `t.2001`; `t.99999999999999999999`; una línea sin salto final y otra tras la final; `+++ exited with 256 +++` y con 20 cifras; `clone(…) = 99999999999999999999`; `execve` con argv y entorno sin leer (`0x…`), con `\q` en la ruta, con el argv abreviado (`...`), sin `, ` entre dos cadenas y con `\xb`; `connect` con la dirección sin leer, `AF_INET` sin `sin_addr` y con `203.0.113.256`, `AF_UNIX` con `sun_path=@"…"` y con `\400`, y `AF_NETLINK` en una invocación; un hilo creado por dos líneas y dos hilos que se crean entre sí sin descender del raíz |
| `TestLeerTrazasSinInvocaciones` | `trazas.go` 346, 646, 964, 1001 | un proceso de `bash` con solo su línea final (research.md V53) y otro con `execve("/usr/bin/env", [], …) = 0` |
| `TestDecodificarCadena` (11) | `trazas.go` 707, 728 (y, en la unidad, 723, 743 y 745) | escapes válidos (de una letra, octal de una a tres cifras, `\377`, hexadecimal) e inválidos: barra final, `\x4`, `\xg1`, `\q`, `\400` |
| `TestLeerSesionConMensajesSinSuForma` (12) | `sesion.go` 228, 245, 263, 271, 291, 295, 308, 312, 333, 337, 344 | sin `type`; `init` con `model` numérico, sin `model` y sin `claude_code_version`; `assistant` con `content` de texto y sin `content`; `tool_use` de `Skill` sin `input` y sin `input.skill`; `result` con `is_error` de texto, sin `subtype`, sin `is_error` y `success` sin `result` |
| `TestLeerSesionConVariosMensajesSystem` | `sesion.go` 267 | `init`, `system` con `subtype` `compact_boundary` y un segundo `init` con otro modelo: cuentan el modelo y la versión del primero |
| `TestFaltaSinOrigenesODeOtroPunto` (2) | `preparar.go` 76, 98 | `Falta` sin orígenes y con un punto que no es de data-model §7.1 |
| `TestPrepararConConjuntosQueNoSeCopian` (2) | `preparar.go` 140, 311, 316-318 | un conjunto que es un fichero (`ENOTDIR`) y un conjunto con una carpeta |
| `TestPrepararSesionSinPoderLeerOEscribir` (3) | `preparar.go` 218, 247, 251 | el fichero de la eval como directorio de evals; `eval.txt` y `pregunta.txt` que son carpetas, tras preparar la caché con las grabaciones de H4 |
| `TestPrepararYComprobarSinDirectorioTemporal` | `preparar.go` 132-134, 166-168 | `TMPDIR` en un fichero: `os.MkdirTemp` falla con `ENOTDIR`. Sin `t.Parallel`, porque `t.Setenv` cambia el entorno del proceso; los `t.TempDir()` se crean antes |
| `TestCompilarEsquemaDeEvalQueNoSirve` (2) | `formato.go` 84, 89, tras la reestructuración 2 | una carpeta en la ruta del esquema y un JSON con `"type": "entero"`, que no compila |

## Reestructuraciones, con su rojo

1. **El código de la línea final va de 0 a 255** (`trazas.go`; data-model §9, regla 5). `formaDeSalida` admitía
   `[0-9]{1,3}`, así que el error de `strconv.Atoi` (528 y 582) solo podía venir de esa constante y no se daba nunca, y
   `+++ exited with 300 +++` pasaba como un código que `strace` no escribe: escribe el estado de salida, de 0 a 255.
   Ahora la forma toma todas las cifras y el código se comprueba: uno que no cabe en un entero y uno mayor que 255 son
   dos defectos de línea con su texto, en lugar de un código falso o de «no es ninguna de las formas de línea». Rojo
   antes del cambio, con los demás subtests en verde: `codigo-final-mayor-que-255` («An error is expected but got nil»)
   y `codigo-final-que-no-cabe-en-un-entero` (el error decía «no es ninguna de las formas de línea de la traza»).
2. **El esquema del formato de eval se compila desde una ruta** (`formato.go`). La lectura y la compilación del esquema
   publicado estaban dentro del `sync.OnceValues`, con la ruta constante: sus dos errores (84 y 89) no los podía dar
   ningún test. `compilarEsquemaDeEval(ruta)` los comprueba desde la ruta que recibe (lee con `leerFichero`, como el
   resto del paquete, y el error de compilación nombra esa ruta); `esquemaDeEval` la llama con la constante, y
   `LeerEval` no cambia. Rojo: `undefined: compilarEsquemaDeEval` hasta el cambio.

## Sondas de las ramas que ya existían

Los demás tests pasaron a la primera, porque cubren ramas ya escritas; cada uno exige el texto de su rama. Para
comprobar que detectan su rama y no otra cosa, seis mutantes a la vez sobre una copia desechable fuera del repositorio
(`rsync` a `/tmp`, retirada al terminar), cada uno anulando una rama: la comprobación de fichero regular de `leerHilos`,
la de hilos sueltos de `atribuir`, la de hilo creado dos veces de `anotarCreaciones`, la de `message.content` de
`leerAssistant`, la de fichero regular de `copiarGrabaciones` y el error de escribir `eval.txt` de `PrepararSesion`.
`go -C <copia> test -run …` falló en exactamente seis subtests, uno por mutante: `hilo-que-es-una-carpeta`,
`hilos-que-no-descienden-del-raiz`, `hilo-creado-dos-veces`, `assistant-sin-content`, `grabacion-que-es-una-carpeta` y
`eval-txt-que-es-una-carpeta`.

## Bloques que siguen sin cubrir en los seis ficheros

Unión de los dos perfiles del `make ci` de este intento, con la orden de `gates/tarea-T030.md` (líneas del árbol de
T034): 16 bloques, 15 de una sentencia y uno sin sentencias.

**Errores que solo puede dar una constante**, que no se fuerzan:

| Bloque | Qué es | Motivo |
|---|---|---|
| `trazas.go:289` | error de `app.RegistroDeProduccion()` en `nuevoInterprete` | solo falla si el applet `boe` del binario tiene un nombre o unos verbos inválidos; es la raíz de composición de `internal/app`, y `TestRegistroDeProduccion` impide publicarlo |
| `trazas.go:300` | error de `kong.New` sobre `struct{ cli.Globales }` | solo falla si las etiquetas de `cli.Globales` están mal; `internal/cli/globales_test.go` construye la misma gramática con `require.NoError` |
| `trazas.go:297` (sin sentencias) | el cuerpo de `kong.Exit(func(int) {})` | `kong.New` no termina el proceso; la sustitución es la misma defensa que la del analizador del kernel (`internal/cli/parse.go`), y `globales_test.go` falla si se llamara |
| `trazas.go:233` y `trazas.go:263` | propagación de ese error en `LeerTrazas` e `InterpretarInvocacion` | los tres anteriores; cada test de las dos funciones construye el intérprete y fallaría con el error |
| `preparar.go:268` | error de `Registro.Registrar(app.AppletBoe(…))` en `registroDeBoe` | las reglas de `Registrar` miran solo el nombre y los verbos del applet `boe`, que no dependen del directorio de reproducción ni de las opciones de la caché; `RegistroDeProduccion` registra el mismo applet con las mismas reglas |
| `preparar.go:146` y `preparar.go:174` | propagación en `Preparar` y `ComprobarSinRed` | el anterior |
| `formato.go:109` | error de `esquemaDeEval()` en `LeerEval` | el valor de `sync.OnceValues` sobre `schemas/eval.yaml.json`; `TestEsquemaDeEval` exige que no falle, y sus dos causas ya se prueban en `compilarEsquemaDeEval` |
| `informe.go:230` | error de `json.Marshal` del `Informe` | tipo fijo de cadenas, enteros, booleanos, listas y punteros a entero, sin mapas, interfaces, números en coma flotante ni ciclos; el único error que admitiría, UTF-8 inválido, lo evita `jsontext.AllowInvalidUTF8` |

Pasar el registro o la gramática como parámetros solo serviría para que un test inyectara uno que falla: forzaría la
rama con un doble, no con una entrada que el paquete reciba, y las propagaciones seguirían igual.

**Fallos de disco que ninguna estructura de directorio provoca**, sin permisos ni carreras:

| Bloque | Qué es | Motivo |
|---|---|---|
| `trazas.go:486` | leer el fichero de un hilo en `leerHilo` | `leerHilos` ya comprobó con `os.ReadDir` que la entrada es un fichero regular `t.<n>` (una carpeta u otro tipo es la rama de la línea 458, que cubre `hilo-que-es-una-carpeta`) y entre el listado y la lectura no media nada |
| `conjunto.go:103` | leer una eval en `leerEntrada` | igual: la entrada ya es un fichero regular con la forma de nombre (lo contrario lo cubre `TestLeerConjunto`) |
| `preparar.go:322` | leer una grabación en `copiarGrabaciones` | igual: la carpeta es la rama de 316-318, que cubre `grabacion-que-es-una-carpeta` |
| `preparar.go:326` | escribir en el temporal de las grabaciones | el destino es el directorio que `Preparar` acaba de crear con `os.MkdirTemp`, y en él solo escribe esta función, ficheros regulares con los nombres de un listado |
| `preparar.go:351` | `os.RemoveAll` del temporal | borra lo que la propia función creó y llenó de ficheros regulares |
| `informe.go:496` | retirar `informe.md` tras no poder escribir `informe.json` | `informe.md` es el fichero que la función acaba de escribir en ese mismo destino |

Solo los hacen fallar los permisos (que la tarea excluye y que como root no fallan), un sistema de ficheros de solo
lectura o lleno, o que otro proceso cambie el directorio entre dos llamadas. La única estructura determinista encontrada
para las tres lecturas —pasar el directorio con un tramo `enlace/..`, que el sistema resuelve por el destino del enlace y
`filepath.Join` limpia léxicamente— haría que el listado y la lectura miraran directorios distintos: ejercitaría esa
diferencia entre las dos resoluciones de una ruta que ningún llamador da (el guion compone las rutas sin `..`), no la
rama.

## Cifras

- `internal/evals`: 99,0 % (1428/1443 sentencias), frente a 94,8 % (1365/1440) sobre `536359c`. Total: 96,3 % en el
  perfil unitario (`-func`; 5642/5863) y 96,8 % en el de integración (`-func`; unión 5671/5863). Tabla de
  `gates/pr-h5.md`, fechada por este commit.
- Unión local con la orden de `gates/tarea-T030.md`: 16 bloques de una línea en los seis ficheros (antes 82 líneas de
  Codecov) y los 53 de `internal/skills`, que son de T035. Es la medida local; la de `codecov/patch` la lee T030 en la
  plataforma.
- Linter fijado por el repositorio sin hallazgos, `dupl` y `misspell` incluidos (`misspell` marcó «abstracto» →
  *abstraction* en un comentario de un test nuevo, reescrito antes de `make ci`).

`gates/tarea-T030.md` no cambia: su plan para el intento siguiente de T030 sigue valiendo.
