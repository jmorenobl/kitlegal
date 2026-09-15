# T035 · intento 1 de 3 (2026-09-15): en verde, marcada

Base `234c6f0` (`feat(H5): T034`); el workflow commitea este intento como `feat(H5): T035`. Verificación: `go clean
-testcache` y `make ci` en primer plano, código 0, `ci: todos los controles en verde` y ningún `FAIL` (log en
`gates/ci.log`, ignorado).

## Qué cambia

Tests nuevos en `internal/skills/{sincronia,esquemas,normas,skill,comandos,frontmatter}_test.go`; `export_test.go`, con
un único punto de entrada (`CompilarEsquemaDeNormas`), porque todos los tests del paquete son de `skills_test`; y cuatro
reestructuraciones que no cambian lo que el paquete acepta ni lo que devuelve, en `normas.go` (dos), `comandos.go` y
`sincronia.go`. Ningún cambio en `testdata/`, `schemas/`, `codecov.yml` ni `.golangci.yml`, ni en ningún contrato del
feature; 0 `//nolint` y 0 `t.Skip` añadidos frente a `main`. El esquema en línea de `TestValidarDocumentoYAML` (una
constante del test) gana `propertyNames` en los elementos de `piezas`, sin cambiar ningún caso anterior.

Cada test nuevo escribe sus entradas como literales o en `t.TempDir()`, y los fallos de disco salen de la estructura de
ese directorio, no de permisos: un fichero donde va un directorio (`ENOTDIR`), una carpeta donde va el fichero del
esquema (`EISDIR`), un directorio con un fichero dentro donde va un fichero generado (`ENOTEMPTY` al retirarlo), un
enlace colgante donde va el directorio de una skill y un enlace que otra skill ya creó (`EEXIST`).

## Ramas cubiertas (líneas de la lista de `gates/tarea-T030.md`, sobre `234c6f0`)

| Test nuevo | Ramas | Entradas |
|---|---|---|
| `TestCompilarEsquema` (2) | `esquemas.go` 33, 45 | un JSON cortado y `"minLength": "tres"` |
| `TestValidarDocumentoYAML`, 5 casos nuevos | `esquemas.go` 198, 266, 388-391, 461, 500, 504 | `segundo-documento-mal-formado`, `nombre: [` tras `---` (198); `valor-que-no-se-decodifica-con-su-etiqueta`, `edad: !!int treinta` (266); `patron-de-clave-en-un-mapa-de-una-lista`, `- Alto: 4` en la segunda pieza (388-391 y 504: el recorrido baja por la lista, pasa por un escalar y la primera pieza no tiene la clave); `valor-que-llega-por-una-fusion`, `- <<: *etiquetas` (461 y 504); `valor-que-json-no-representa-tras-una-clave-repetida-por-un-alias`, `&talla talla: 3` y `*talla : {medida: .inf}` (500) |
| `TestCompilarEsquemaDeNormasDesdeUnaRuta` | `normas.go` 91, 96, tras la reestructuración 1 | el esquema publicado; una carpeta (`EISDIR`); un JSON con `"type": "texto"` en `rango` |
| `TestLeerNormas`, 2 casos nuevos | `normas.go` 186; la premisa de la reestructuración 2 | `campo-repetido-en-una-norma-por-un-alias`, `&titulo titulo:` y `*titulo :` en la misma norma: el recorrido del lector común no ve la repetición y el esquema valida el documento convertido, pero `yaml.v3` rechaza el campo repetido al leer la norma («field titulo already set in type skills.Norma»); `clave-de-texto-que-es-una-lista`, `? !!str [BOE-A-2015-10565]` |
| `TestLeerFrontmatter/clave-que-no-es-del-estandar` | `frontmatter.go` 244 | `user-invocable: true`: va a `Claves`, sin defecto de lectura |
| `TestListarYCargar/nombre-de-un-fichero` | `skill.go` 92 | `Cargar` con `notas.md`, un fichero del directorio de skills (`ENOTDIR`), que no es un `DefectoDeSkill` |
| `TestDescripcionDeVerbo/defectos/propiedades-cerradas-como-una-lista` | `comandos.go` 434, tras la reestructuración 3 | `"properties": {"entrada": {}]` |
| `TestRegenerarYComparar/tabla-con-banderas-distintas` | `sincronia.go` 293 | `uno leer` con otras banderas globales que `dos listar`: defecto de SKILL.md, sin nada escrito |
| `TestRegenerarSinPoderListar` (2) | `sincronia.go` 172, 181, 222, 262 | un fichero en `skills` y un fichero en `data`, que se lista porque alfa declara `normas` (`ENOTDIR`) |
| `TestCompararYEscribirSinBuscarLasDerivas` (3) | `sincronia.go` 404, 439, 524, 539, 544, 599, 601 | un fichero en `references/` y otro en `scripts/` de alfa; un fichero en el directorio de alfa tras regenerar (`ENOTDIR`); `Comparar` y `Escribir` dan el mismo error, y `Escribir` no escribe nada |
| `TestEscribirSinAplicarUnArreglo` (3) | `sincronia.go` 444, 475, 484 y 492 (uno solo tras la reestructuración 4), 488, 506 | `references/normas.md` como directorio con un fichero (`ENOTEMPTY`); el directorio de alfa como enlace colgante tras regenerar (`EEXIST` al crearlo); beta, que declara `uno` como alfa, como enlace al directorio de alfa tras regenerar (`EEXIST` al enlazar `scripts/uno`, que alfa acaba de crear) |

En `valor-que-json-no-representa-tras-una-clave-repetida-por-un-alias`, la línea es la 3, la de la clave escrita, y no
la 4, la del valor que llega por el alias: es la regla de `lineaDe` (la línea del último nodo escrito de la ruta) sobre
la repetición por un alias que el recorrido del lector común no ve, y que ya cierran `LeerNormas` y `LeerFrontmatter`.
El test fija lo que el lector hace; no se cambia aquí.

### Por qué lo regenerado puede ser de antes del cambio

Tres casos (`skill-que-es-un-fichero`, `crear-el-directorio-de-un-enlace-colgante` y
`enlazar-lo-que-otra-skill-ya-enlazo`) cambian el árbol entre `Regenerar` y `Comparar` o `Escribir`. No es la carrera
que excluye T034: allí eran dos llamadas al sistema dentro de una misma función (listar y leer); aquí son dos funciones
públicas. `Comparar` y `Escribir` reciben lo regenerado como un dato (`Regenerado` es una estructura exportada) y buscan
las derivas en el árbol tal como está cuando se las llama, y su contrato nombra el error de una entrada que no se puede
consultar y el de un `references/` o un `scripts/` que no es un directorio. Sin ese cambio, las tres ramas no se
alcanzan: `Regenerar` no lista como skill ni un fichero ni un enlace.

## Reestructuraciones, con su rojo

1. **El esquema de normas se compila desde una ruta** (`normas.go`), como el esquema de eval en T034. La lectura y la
   compilación del esquema publicado estaban dentro del `sync.OnceValues`, con la ruta constante: sus dos errores (91 y
   96) no los podía dar ningún test. `compilarEsquemaDeNormas(ruta)` los comprueba desde la ruta que recibe (lee con
   `leerFichero`, como el resto del paquete, y el error de compilación nombra esa ruta); `esquemaDeNormas` la llama con
   la constante, y `LeerNormas` no cambia. Rojo: `undefined: compilarEsquemaDeNormas` (`build failed`) hasta el cambio.
2. **El identificador de una norma es el texto de su clave** (`normas.go`, `UnmarshalYAML`). `clave.Decode` sobre el
   identificador solo podía fallar con una clave que no fuera un escalar de texto ni un alias de uno, y ninguna llega:
   `UnmarshalYAML` solo se ejecuta al leer la tabla ya convertida a un mapa de texto y validada. Una sonda temporal
   (retirada antes de `make ci`) lo comprobó: una clave con etiqueta propia, nula, `!!binary` o `!!timestamp` da
   «normas: mapa con claves que no son texto»; `!!str` sobre una lista o un mapa, y un alias de una lista, hacen fallar
   la conversión («el documento YAML no se puede leer: …»). Ahora es `sinEnvoltorio(clave).Value`, la misma lectura que
   `lecturaDelFrontmatter.claves`. La premisa la fijan `clave-que-no-es-texto` (ya existía) y el nuevo
   `clave-de-texto-que-es-una-lista`; la resolución del alias, `identificador-repetido-por-un-alias`: un mutante que lee
   `clave.Value` sin resolverlo, sobre una copia desechable, hace fallar ese subtest y ningún otro de `TestLeerNormas`.
3. **La lectura de `properties` lee el cierre como un token más** (`comandos.go`). El bucle miraba con `PeekKind` si
   venía `}` y, al salir, leía ese `}` con un `ReadToken` aparte cuyo error (434) no podía darse: `jsontext` solo da `}`
   desde `PeekKind` si el estado admite el cierre (tras un nombre exige `:`: `decode.go` 354-356, `state.go` 403-411), y
   leer ese `}` solo desapila el objeto que ya comprobó la apertura (`state.go` 320-332; el espacio de nombres solo se
   invalida con `AllowDuplicateNames`, que no se usa). Ahora cada vuelta lee un token: el cierre sale del bucle y lo
   demás es el nombre de un miembro, con un único error de lectura. Lo fijan todos los documentos válidos de
   `TestDescripcionDeVerbo` y, para el error, `propiedades-cortadas` y el nuevo `propiedades-cerradas-como-una-lista`.
   `gates/tarea-T030.md` describe este bloque como la lectura del «primer token»: era la del cierre.
4. **`aplicar` crea el directorio una sola vez** (`sincronia.go`): los dos bloques iguales que creaban el directorio
   antes de enlazar (484) y antes de escribir (492) pasan a uno, antes de los dos. Lo fija
   `crear-el-directorio-de-un-enlace-colgante` (mutante 8 de abajo).

## Rojo de los tests de ramas que ya existían

Pasaron a la primera, porque cubren ramas ya escritas. Para comprobar que cada uno detecta su rama, 18 mutantes a la vez
sobre una copia desechable fuera del repositorio (`rsync` a `/tmp`, retirada al terminar); `go -C <copia> test -run
'^(TestListarYCargar|TestRegenerarSinPoderListar|TestRegenerarYComparar|TestCompararYEscribirSinBuscarLasDerivas|TestEscribirSinAplicarUnArreglo|TestCompilarEsquema|TestValidarDocumentoYAML|TestLeerFrontmatter|TestLeerNormas)$'`
falló en exactamente estos 19 subtests, sin pánicos:

| Mutante | Subtest que falla |
|---|---|
| 1. `Cargar` devuelve el error de consulta como un `DefectoDeSkill` | `TestListarYCargar/nombre-de-un-fichero` |
| 2. `Regenerar` ignora el error de `Listar` | `TestRegenerarSinPoderListar/directorio-de-skills-que-es-un-fichero` |
| 3. `declaracionDeLaSkill` ignora el error de `ValidarFrontmatter` | `TestRegenerarSinPoderListar/directorio-de-datos-que-es-un-fichero` |
| 4. `contenidoRegenerado` descarta el defecto de `RenderizarTabla` | `TestRegenerarYComparar/tabla-con-banderas-distintas` |
| 5. `entradasDeLaCarpeta` ignora el error de consulta | `TestCompararYEscribirSinBuscarLasDerivas/skill-que-es-un-fichero` |
| 6. `entradasDeLaCarpeta` toma una carpeta que no es un directorio por vacía | `…/references-que-es-un-fichero` y `…/scripts-que-es-un-fichero` |
| 7. `aplicar` ignora el error de retirar | `TestEscribirSinAplicarUnArreglo/retirar-un-directorio-con-contenido` |
| 8. `aplicar` ignora el error de crear el directorio | `TestEscribirSinAplicarUnArreglo/crear-el-directorio-de-un-enlace-colgante` |
| 9. `aplicar` ignora el error de enlazar | `TestEscribirSinAplicarUnArreglo/enlazar-lo-que-otra-skill-ya-enlazo` |
| 10. `CompilarEsquema` no da error con un contenido que no es JSON | `TestCompilarEsquema/no-es-json` |
| 11. `CompilarEsquema` no da error con un esquema que no compila | `TestCompilarEsquema/no-compila` |
| 12. `analizar` toma un segundo documento mal formado por el final | `TestValidarDocumentoYAML/segundo-documento-mal-formado` |
| 13. `normalizar` no da error si la conversión falla | `TestValidarDocumentoYAML/valor-que-no-se-decodifica-con-su-etiqueta` |
| 14. `mapasConLaClave` no baja por las listas | `TestValidarDocumentoYAML/patron-de-clave-en-un-mapa-de-una-lista` |
| 15. `lineaDe` da la línea 0 si un paso de la ruta no está escrito | `TestValidarDocumentoYAML/valor-que-llega-por-una-fusion` |
| 16. `hijoDe` da, desde un escalar, un nodo de la línea 99 | `TestValidarDocumentoYAML/valor-que-json-no-representa-tras-una-clave-repetida-por-un-alias` |
| 17. `frontmatter` descarta de `Claves` la clave que no es del estándar | `TestLeerFrontmatter/clave-que-no-es-del-estandar` |
| 18. `UnmarshalYAML` ignora el error de leer la norma | `TestLeerNormas/campo-repetido-en-una-norma-por-un-alias` |

## Bloques que siguen sin cubrir en los seis ficheros

Unión de los dos perfiles del `make ci` de este intento, con la orden de `gates/tarea-T030.md` (líneas del árbol de
T035): 18 bloques, todos de una sentencia.

**Errores que solo puede dar una constante o el contrato de una biblioteca**, que no se fuerzan:

| Bloque | Qué es | Motivo |
|---|---|---|
| `esquemas.go:40` | error de `AddResource` en `CompilarEsquema` | `AddResource` (jsonschema v6.0.3, `compiler.go` 121-133) solo falla con una URL que no es absoluta, con la de un metaesquema o con una ya registrada; la URL es la constante `urlDelEsquema` y el compilador se crea en la línea anterior, así que no depende del contenido (sus otros dos errores, sí: 33 y 45, en `TestCompilarEsquema`) |
| `esquemas.go:324` | error de `Validate` que no es un `*jsonschema.ValidationError` | `Schema.Validate` (`validator.go` 15-50) convierte todo fallo en un `*ValidationError`, también un valor que no es de JSON (`kind.InvalidJsonValue`); la rama protege ese contrato: sin ella, otro tipo de error acabaría en un puntero nulo en lugar de en un error |
| `normas.go:129` | error de `esquemaDeNormas()` en `LeerNormas` | el valor de `sync.OnceValues` sobre la constante `rutaDelEsquemaDeNormas`; `TestEsquemaDeNormas` y `TestCompilarEsquemaDeNormasDesdeUnaRuta` exigen que el publicado compile, y sus dos causas ya se prueban desde una ruta |

**La normalización a JSON de lo que ya se comprobó**:

| Bloque | Qué es | Motivo |
|---|---|---|
| `esquemas.go:275` | error de `json.Marshal` del documento convertido | `sinRepresentacionJSON`, justo antes, rechaza lo único que `json.Marshal` no codifica de lo que `(*yaml.Node).Decode` da en un `any`: un mapa con claves que no son texto y un número infinito o NaN. La sonda temporal comprobó el resto de valores con forma propia: `!!binary /w==` (UTF-8 inválido, que `encoding/json` sustituye), `!!timestamp 2001-12-14`, `18446744073709551615` y un entero de 31 cifras se codifican sin error |
| `esquemas.go:280` | error de `jsonschema.UnmarshalJSON` sobre lo que acaba de codificar `json.Marshal` | su entrada es la salida de la línea anterior: un único documento JSON válido |

**Lecturas de lo que un listado o una consulta acaba de dar**, que sin permisos ni otro proceso no fallan (como en T034):

| Bloque | Qué es | Motivo |
|---|---|---|
| `skill.go:102` | leer `SKILL.md` en `Cargar` | `os.Lstat` acaba de darlo como fichero regular por la misma ruta |
| `sincronia.go:217` | error de `Cargar` que no es un `DefectoDeSkill`, en `regenerarSkill` | en `Regenerar` el nombre es el de una entrada que `os.ReadDir` acaba de dar como directorio (un enlace no lo es), así que la ruta de `SKILL.md` no pasa por ningún fichero ni enlace y su consulta solo puede dar «no existe»; su lectura es la de `skill.go:102`. La consulta que falla por la estructura la cubre `nombre-de-un-fichero`, fuera de `Regenerar` |
| `sincronia.go:320`, con su propagación en `sincronia.go:234` | leer `data/<referencia>.yaml` en `referenciasGeneradas` | solo se llama con el frontmatter sin defectos, y `ValidarFrontmatter` acaba de encontrar ese nombre entre los ficheros regulares que `os.ReadDir` da del directorio de datos |
| `sincronia.go:608` | `os.ReadDir` de `references/` o de `scripts/` en `entradasDeLaCarpeta` | `os.Lstat` acaba de darla como directorio |
| `sincronia.go:644` | leer un fichero generado en `arregloDelFichero` | `os.Lstat` acaba de darlo como fichero regular |
| `sincronia.go:682` | `os.Readlink` en `arregloDelEnlace` | `os.Lstat` acaba de darlo como enlace simbólico |

**Consultas dentro de un directorio que ya se consultó**:

| Bloque | Qué es | Motivo |
|---|---|---|
| `sincronia.go:635`, con sus propagaciones en `sincronia.go:551` y `sincronia.go:560` | `os.Lstat` de `SKILL.md` o de `references/normas.md` que falla con algo que no es «no existe» | `arreglosDeLaSkill` consulta antes `references/` y `scripts/` del mismo directorio: si la ruta hasta él pasa por un fichero o por un bucle de enlaces, el error sale ahí (`skill-que-es-un-fichero`); si el directorio no existe, `SKILL.md` tampoco; y `references/` es un directorio o no existe, así que `references/normas.md` está o no está. El nombre de la referencia es el de un generador conocido: hoy, solo `normas` |
| `sincronia.go:673`, con su propagación en `sincronia.go:575` | `os.Lstat` de `scripts/<applet>` | igual, con `scripts/` consultada antes, para todo nombre de applet sin separador de ruta. Con separador (véase *Observación*), la consulta sí fallaría con un fichero en su lugar, pero un test así fijaría como resultado un caso en que la sincronía no converge |

**La escritura de un fichero generado**:

| Bloque | Qué es | Motivo |
|---|---|---|
| `sincronia.go:498` | `os.WriteFile` en `aplicar` | la búsqueda de derivas consultó la ruta (no existe, es un fichero regular o no lo es y se acaba de retirar) y `crearDirectorio` acaba de dejar su directorio. Entre las dos, solo cambian el árbol los arreglos anteriores de la misma escritura, que dejan ficheros llamados `SKILL.md` o `normas.md`, enlaces llamados como un applet y directorios llamados como una skill, `references` o `scripts`, y retiran lo que ya consultaron: ninguno deja un directorio donde luego se escribe un fichero. Solo lo hacen fallar, además de los permisos, un sistema de ficheros de solo lectura o lleno y otro proceso, cadenas de enlaces entre skills sustituidas tras regenerar que hacen pasar la ruta del fichero por un enlace que la propia escritura crea o rehace: un applet registrado con el nombre de un fichero generado, o un enlace de `scripts/` cuyo destino nuevo lleva un enlace colgante con ese nombre. Un test así fijaría el orden en que `Escribir` aplica los arreglos de dos skills que comparten directorio a través de enlaces que la sincronía no define, no la escritura |

## Observación, sin cambio

El registro del binario rechaza un nombre de applet vacío, con espacios o que empieza por `-` (`internal/app/registro.go`,
`ErrNombreInvalido`), y `ValidarFrontmatter` solo exige que cada applet de `kitlegal-applets` esté registrado y no se
repita. Un applet con separador de ruta (`a/b`) se enlazaría en `scripts/a/b`, y en la comparación siguiente `scripts/a`
sería una entrada sobrante que `Escribir` no puede retirar: la sincronía no convergería. Hoy el único applet registrado
es `boe`. Validar la forma del nombre en la sincronía no está especificado y no se implementa aquí; queda anotado para
quien lo decida.

## Cifras

- `internal/skills`: 98,5 % (1169/1187 sentencias), frente a 95,8 % (1134/1184) sobre `234c6f0`, con tres sentencias
  más en total (las reestructuraciones). Total: 96,8 % en el perfil unitario (`-func`; 5677/5866) y 97,3 % en el de
  integración (`-func`; unión 5706/5866). Los demás árboles dan las cifras de T034. Tabla de `gates/pr-h5.md`, fechada
  por este commit.
- Unión local con la orden de `gates/tarea-T030.md`: 18 bloques de una sentencia en los seis ficheros de
  `internal/skills` (antes 53 líneas de Codecov) y los 16 de `internal/evals` que dejó T034. Es la medida local; la de
  `codecov/patch` la lee T030 en la plataforma.
- Linter fijado por el repositorio sin hallazgos, `dupl` y `misspell` incluidos.

`gates/tarea-T030.md` no cambia: su plan para el intento siguiente de T030 sigue valiendo.
