# Data model · H23

Las entidades del hito, con el paquete en el que viven y el requisito que las pide. Los nombres de los tipos y de las
funciones exportadas son los que fija el plan; los que no se exportan son orientativos.

## 1. Identificadores (`internal/core/ids`)

| Tipo | Forma | Operaciones | Requisito |
|---|---|---|---|
| `ECLI` | `ECLI:ES:<órgano>:<año>:<número>`: `ECLI`, `ES`, órgano `[A-Z][A-Z0-9]{0,6}`, año `[0-9]{4}`, número `[A-Z0-9.]{1,25}` | `AnalizarECLI(entrada) (ECLI, error)`; `String()`; `Organo()` | FR-003, FR-015 |
| `ROJ` | `[A-Z]+( [A-Z]+)* [0-9]+/[0-9]{4}` | `AnalizarROJ(entrada) (ROJ, error)`; `String()` | FR-004 |

- El error es el de siempre del paquete: de la clase `argumentos`, con la entrada y el motivo. Un ECLI bien formado de
  otro país lo dice («no es español»).
- Nada se recorta ni se normaliza. No hay lista de órganos ni de siglas (spec, «Fuera de alcance»).
- `FuzzECLI` y `FuzzROJ`, con corpus en `internal/core/ids/testdata/fuzz/`: ninguna entrada entra en
  pánico, y lo aceptado casa con la forma de arriba y vuelve igual por `String()` (FR-113).

## 2. La referencia y la consulta (`internal/source/cendoj`)

| Entidad | Campos | Reglas |
|---|---|---|
| `Referencia` | forma (`ecli`, `roj` o `resolucion`), valor, fecha (`AAAA-MM-DD` o vacía) | la construye `NuevaReferencia(ecli, roj, resolucion, fecha)`, que aplica la validación de [contracts/cita-resolver.md §1](./contracts/cita-resolver.md) y devuelve un error de la clase `argumentos`. `DelTribunalConstitucional()` dice si es un ECLI de órgano `TC` |
| `ConsultaResolver` | `Referencia` | es la `core.Consulta` del verbo `resolver`; su vigencia (`TTL`), 30 días |

## 3. Lo que se entrega (`internal/source/cendoj`)

| Entidad | Campos (clave JSON) | Reglas |
|---|---|---|
| `Resolucion` | `ecli`, `roj`, `organo`, `fecha`, `numero_resolucion`, `numero_recurso`, `ponente`, `url` | `ecli`, `roj`, `fecha` y `url` nunca vacíos; `fecha`, `AAAA-MM-DD`; `url`, absoluta. Sin resumen ni texto (FR-010) |
| `Entrega` | `resoluciones` (0 a 10), `pagina_completa`, `cobertura` | es el `data` del verbo y su `Salida` para `--describe`; `resoluciones` nunca es nula |
| `Cobertura` | `cubierto` \| `tribunal-constitucional-no-cubierto` | vocabulario cerrado, declarado en el esquema (FR-015) |

Transiciones de una consulta: validada → (Tribunal Constitucional → entrega sin resoluciones) | (caché vigente →
entrega o «no encontrado») | (fuente → clasificada → entrega, «no encontrado», `limite-o-tos` o
`fuente-no-disponible`). Solo la entrega y el «no encontrado» que vienen de la fuente se guardan.

## 4. La caché de la fuente

| Campo | Valor |
|---|---|
| clave | `cendoj.jurisprudencia\|1\|resolver\|<forma>\|<valor>\|<fecha>` |
| contenido | `fecha_consulta` (RFC 3339 con nanosegundos), `url` y uno de `datos` (una `Entrega`) o `no_encontrado` (el mensaje) |
| vigencia | 30 días |

Usa el puerto `core.Cache` y el adaptador de H3 como están (FR-040, FR-041).

## 5. El grafo (`internal/core/grafo`, `internal/source/cendoj/grafo.go`)

- Constante nueva `TipoResolucion = "Resolucion"` y las claves de sus datos: `DatoECLI`, `DatoROJ`, `DatoOrgano`,
  `DatoFecha`, `DatoNumeroResolucion`, `DatoNumeroRecurso`, `DatoPonente`, `DatoURL`.
- Por cada `Resolucion` de una entrega, un `schema.Nodo{ID: <ecli>, Tipo: TipoResolucion, Datos: …}` en
  `Resultado.Grafo`, con `Vigencia` de 30 días. Ni aristas ni textos (FR-042).
- `internal/graph` y el applet `graph` no cambian (FR-043).

## 6. El formulario (`internal/httpx`)

| Entidad | Qué es | Requisito |
|---|---|---|
| `ConFormulario(direccion)` | la dirección a la que el cliente puede enviar un formulario | FR-030 |
| `Peticion.Campos` | `map[string]string`: los campos de un envío | FR-030, FR-034 |
| `Consulta` | lo que comparte cookies: `Cliente.Consulta()` y `Consulta.Pedir` | FR-032, FR-033 |
| grabación con cuerpo | `peticion.cuerpo` en el fichero; nombre con `_c_<cuerpo>` | FR-034 |
| `httpxtest.Sitio` | sitio local de prueba que apunta lo que recibe; solo para tests | FR-021 (su control) |

Detalle en [contracts/httpx-formulario.md](./contracts/httpx-formulario.md).

## 7. La fuente y su composición

| Dónde | Qué |
|---|---|
| `internal/source/cendoj` | `NombreDeLaFuente`, `IntervaloEntrePeticiones`, `Fuente` (`core.Source`: `Name`, `Fetch`, `TTL`, `Terms`), `Nueva` con `ConCliente`, `ConCache` y `ConRegistrador` —como `boe`—, `OpcionesDeRed()` y `OpcionesDeReproduccion()` (las de `internal/httpx` con las que se pide: fuente y formulario; en red, además un solo intento) |
| `internal/app/cita.go` | `AppletCita(DependenciasDeCita)`, `DependenciasDeCita{Cliente, Cache}`, `DependenciasDeCitaDeRed()` (un `Ritmo` de 5 s por dependencias); el verbo `resolver` y sus argumentos; la declaración del Tribunal Constitucional y la salida legible |
| `internal/app/registro.go` | `AppletCita(DependenciasDeCitaDeRed())` en `RegistroDeProduccion` |
| `internal/app/ejemplo/kitlegal-e2e/main.go` | `cita` sobre la reproducción de `reproduccion/cendoj.jurisprudencia`, con el reloj del arnés |

## 8. El manifiesto de grabación (`internal/source/cendoj/testdata/grabaciones.json`)

`fuente` y `consultas`, cada una con una referencia (`ecli`, o `roj`, o `resolucion` con `fecha`) y `para`
([contracts/fuente-cendoj-y-grabacion.md §4](./contracts/fuente-cendoj-y-grabacion.md)). Lo lee el mismo código para
grabar y para comprobar que cada consulta tiene su grabación (`TestGrabacionesCompletas`).

## 9. El formato de eval y el informe (`internal/evals`, `schemas/eval.yaml.json`)

| Entidad | Campos nuevos | Requisito |
|---|---|---|
| `ComandoEsperado` | `ECLI`, `ROJ`, `Resolucion`, `Fecha`, `NoEncontrado` (`ecli`, `roj`, `resolucion`, `fecha`, `no_encontrado`) | FR-070, FR-072, FR-074 |
| `Eval` | `Sentencias []SentenciaEsperada{ECLI, ROJ}`, `SentenciaNoComprobada`, `SinSentencias`, `Direcciones` | FR-070 a FR-072 |
| `Consulta` (necesaria) | `NoEncontrado`: termina con 3 | FR-074 |
| `ResultadoDeEval` | `sentencias_encontradas`, `sentencias_ausentes`, `linea_de_sentencia_no_comprobada`, `sentencias_que_sobran`, `direcciones_encontradas`, `direcciones_ausentes`, `ecli_sin_resolver`; ninguna se escribe si va vacía | FR-071, FR-080, FR-081 |
| `Umbral` | sin campos nuevos; un nombre nuevo, `cita_sin_resolver:<modelo>:<modo>` | FR-080 |
| reglas del conjunto | `ReglasDeJurisprudencia()` | FR-075 |

Funciones nuevas: `ExtraerSentencias(respuesta)`, `ExtraerSentenciaNoComprobada(respuesta)`,
`ECLISinResolver(pregunta, respuesta, textos)` y la constante `GrabacionesDeCendoj`.

## 10. Lo que no cambia

`internal/cache`, `internal/graph`, `internal/mcp`, `internal/render`, `internal/empaquetado` y `cmd/` —salvo las
listas literales de sus tests que enumeran applets y herramientas—, el contrato
`Applet`, el sobre, las clases de error y los códigos, `skills/boe-legislacion/`, `skills/legal-core/`,
`docs/SOURCES.md`, `evidencias/`, `scripts/workflow/` y la constitución.
