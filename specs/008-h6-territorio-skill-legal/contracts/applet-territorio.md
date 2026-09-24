# Contrato: el applet `territorio` y su verbo `resolver`

Qué se invoca, qué se devuelve, con qué códigos de salida y qué lo vigila. Hereda sin repetirlos el contrato del
applet (ADR 0005), el sobre y su huella (ADR 0006), la fecha de consulta (ADR 0015) y el ensayo (ADR 0011).

## 1. Invocación

```
kitlegal territorio resolver <consulta>
territorio resolver <consulta>          # por el enlace simbólico del multicall
```

- Un único argumento posicional, **obligatorio**: `Consulta string \`arg:"" name:"consulta" help:"…"\`` (FR-002).
  Sin `optional:""`, de modo que Kong lo exige y el esquema lo declara `required` (research.md V40:
  `internal/cli/describe.go:475-491`).
- **Consecuencia sobre `--describe`**: el análisis de la invocación corre **antes** que la decisión de describir
  (`internal/app/main.go:301-320`), así que `territorio resolver --describe`, sin el posicional, termina en 2 y no
  emite nada; el verbo se describe con su argumento —`territorio resolver <consulta> --describe`—, como lo invocan los
  guiones del e2e de `boe` y los quickstarts de H4 y H5 (research.md V40). Describir **no ejecuta** el verbo: el
  registro no se analiza (§7).
- Ningún verbo por omisión: nombrar `resolver` es obligatorio, como en `boe` (ADR 0005).
- El applet **no declara ninguna bandera**: hereda las ocho globales.
- El applet no abre ninguna conexión ni toca la caché en ningún camino (FR-043).

## 2. Procedencia del sobre

| Situación | `fuente` | `url` | `fecha_consulta` |
|---|---|---|---|
| Éxito | `kitlegal.territorio` | `kitlegal:applet/territorio` | la más antigua de las fechas de los ficheros que sostienen `data` (data-model §2.8) |
| Fallo decidido por el applet (ambiguo, no encontrado, código mal formado) | `kitlegal.territorio` | `kitlegal:applet/territorio` | la misma |
| Fallo anterior al applet (falta el argumento, bandera desconocida) | `kitlegal.cli` | `kitlegal:cli` | el reloj del montador |
| Fallo decidido por el applet bajo `--dry-run` (el kernel descarta el `Resultado` del applet y firma el fallo él) | `kitlegal.cli` | `kitlegal:cli` | el reloj del montador |

Consecuencia comprobable: **dos ejecuciones del mismo verbo con la misma consulta producen la misma salida byte a
byte**, con `--offline` y sin él (US1 escenarios 3 y 4, SC-001).

## 3. `data` del verbo

La forma exacta, sus claves y sus vocabularios están en [data-model.md](../data-model.md) §2.5, §2.3 y §2.4. Del
contrato forman parte, además:

1. **Ocho claves de primer nivel**, siempre las ocho: `municipio`, `codigo_ine`, `provincia`, `comunidad`, `dir3`,
   `regimen`, `boletines`, `cobertura` (FR-006).
2. **Ningún campo con `omitempty`**: lo que no hay va vacío, las listas vacías van como `[]`.
3. **`source` en cada dato** (FR-005): identificador de fila de `docs/SOURCES.md` o ruta de fichero de
   `data/territorio/`.
4. **`boletines`** lleva siempre el estatal y solo los niveles configurados; fuera del territorio configurado no
   aparece el nombre, el código ni la dirección de ningún boletín (FR-008, FR-021).
5. **`cobertura`** enumera sus tres aspectos siempre, y su vocabulario no tiene ningún valor que signifique «no
   existe» (FR-020, FR-022).
6. **DIR3 no verificado**: `dir3.codigo` y `dir3.source` vacíos y `cobertura.dir3` en `no-verificado`; nunca un código
   derivado presentado como registral (FR-023).

## 4. Códigos de salida

La tabla completa de entradas está en [data-model.md](../data-model.md) §2.6. Del contrato:

| Código | Cuándo | Clase del sobre de fallo |
|---|---|---|
| 0 | Municipio resuelto | — |
| 2 | Entrada mal formada, o nombre que corresponde a más de un municipio | `argumentos` |
| 3 | Municipio que no está en la relación (por nombre, o por código **bien formado**: provincia `01`-`52` y municipio `001`-`999`) | `no-encontrado` |
| 4, 5, 6 | **Nunca** (FR-016) | — |

El mensaje del error del applet llega literal al `mensaje` del sobre de fallo y a la salida de error
(research.md V17). Para el nombre ambiguo, el mensaje lleva **todos** los candidatos en la forma fija
`<código INE> <nombre> (<provincia>)`, separados por `; ` y ordenados por código INE (FR-014, research.md D14).

## 5. Banderas globales

| Bandera | Efecto en `territorio` |
|---|---|
| `--json` | Presenta el sobre; sin ella, la forma de tabla del presentador |
| `--offline` | **Ninguno**: el applet no consulta nada, así que devuelve exactamente lo mismo (FR-009) |
| `--dry-run` | El applet se ejecuta igual y no rellena `Ensayo`; el kernel escribe su línea en la salida de error. Una consulta que se resuelve no emite sobre y termina con 0; una que el applet rechaza emite, con `--json`, el sobre de fallo firmado por el kernel (`kitlegal.cli`, `kitlegal:cli`, el reloj; §2) y termina con su código, 2 o 3 (research.md D7, V15) |
| `--timeout` | El plazo de la operación; no hay operación que lo agote |
| `--describe` | Emite el esquema de entrada y salida del verbo |
| `--no-graph`, `--asunto`, `--verbose` | Heredados, sin efecto propio en H6 (el grafo es H7) |

## 6. Esquema publicado

- Fichero: `schemas/municipio.json` —el nombre es el de la **entidad**, como `norma.json` y `bloque.json`
  (research.md D25)—, con `$id` `https://ventanillalegal.es/schemas/municipio.json`, título `territorio · municipio`
  y una parte `$defs.resolver` con el `$id` `https://ventanillalegal.es/schemas/municipio.json/resolver`, en la forma
  canónica del contrato de H4 (`internal/app/esquemas_test.go:58-61`, `:142`).
- No confundirlo con `schemas/territorio-municipios.yaml.json`, que valida el **fichero de datos**: el sufijo
  `.yaml.json` marca los esquemas de `data/`, aquí y en `normas.yaml.json` y `eval.yaml.json`.
- Se genera con la receta existente: `TestEsquemasPublicados` y su bandera `-actualizar-esquemas`. **No se escribe a
  mano.**
- Antes hay que parametrizar por applet la tabla `ficherosDeEsquemas` y los nueve puntos donde `boe` está literal
  (research.md V8, D16): la tabla conserva `nombre`, `entidad` y `verbos`, y gana el applet, que es lo que compone el
  título `<applet> · <entidad>` y el nombre de la parte publicada.
- `make schema-check` falla si lo publicado no coincide con lo que `--describe` emite (FR-007, FR-092).

## 7. Composición

```go
// internal/app/territorio.go
func AppletTerritorio(fuentes territorio.Fuentes) Applet
```

- La raíz de producción (`RegistroDeProduccion`) pasa las fuentes embebidas del paquete `data`; el binario de e2e, las
  mismas (los datos no dependen del entorno, FR-056).
- El registro se analiza **una sola vez por applet** y solo cuando hace falta (`sync.OnceValues` capturado en el valor
  del applet, no una variable de paquete): `--help` y `--describe` no lo analizan.
- Unas fuentes que no analizan o que no pasan la integridad son un **defecto de composición**: error inesperado
  (código 1), no un código de usuario. No puede ocurrir en el binario publicado, porque los ficheros viajan dentro y
  `make ci` los valida contra su esquema (FR-056).

## 8. Qué lo vigila

| Control | Test | En `make ci` |
|---|---|---|
| Forma del sobre y de `data` | `TestResolverDevuelveElTerritorio` (`internal/app/territorio_test.go`) | sí |
| Salida real contra el esquema publicado | `TestSalidaDeTerritorioContraSchemas` | sí |
| `--describe` sin deriva | `TestEsquemasPublicados`, `TestEsquemasCubrenTodosLosVerbos` | sí (`schema-check` el primero; `test` el segundo) |
| Códigos 0, 2 y 3, y que nunca hay 4, 5 ni 6 | `TestCodigosDeTerritorio` | sí |
| Igualdad byte a byte por nombre, por código y con `--offline` | `territorio-matriz.txtar` | sí (`test-e2e` dentro de `test`) |
| Matriz territorial completa | `territorio-matriz.txtar` (Leganés, Tordesillas, foral, ambiguo) | sí |
| Que ningún boletín no configurado aparece | `territorio-matriz.txtar` + `TestSalidaSinBoletinesNoConfigurados` | sí |
| Que el applet no abre conexiones | `TestArquitectura` R2 (no importa `net/http`) y el e2e sin reproducción | sí |
