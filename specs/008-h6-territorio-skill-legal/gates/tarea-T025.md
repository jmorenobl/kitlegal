# T025 — marcada en el intento 2; en el intento 1, la guía y no el hito daba un resultado distinto del esperado

**Estado**: `[X]` (intento 2, 2026-09-22; abajo, «Intento 2»). El intento 2 ejecutó la guía **corregida** tal cual y
en orden, desde los prerrequisitos hasta la limpieza, sin ningún resultado distinto del esperado; midió la cobertura
sobre su propio `make ci` y escribió `gates/pr-h6.md`. Lo que sigue es la nota del intento 1, que se conserva porque
documenta las tres correcciones de `quickstart.md` y el motivo de cada una.

## Intento 1: sin marcar

La tarea manda detenerse sin marcarse ante cualquier resultado distinto del esperado, y el
escenario 11 dio uno. El producto está en verde: `make ci` pasa, la cobertura sobra en los dos umbrales y todo lo que
el hito entrega se comporta como dice la guía. Lo que falla son **tres textos de la propia guía** que no casan con los
datos congelados; los tres quedan corregidos en `quickstart.md` en este mismo intento, así que el siguiente la ejecuta
tal cual desde los prerrequisitos. **`gates/pr-h6.md` no se ha escrito**: lo escribe el intento que termine en verde.

Ejecutado en una sesión desatendida el **2026-09-22 entre las 11:40 y las 11:49 (hora de Madrid), sobre `2df74db`**
(`feat(H6): T024`), con el árbol limpio fuera de `specs/008-h6-territorio-skill-legal/`. Prerrequisitos, `make ci` y
escenarios 1 a 13 tal cual y en orden, con las formas `rtk proxy` de la tabla de la guía; después, la limpieza. Los
escenarios que siguen al 11 se ejecutaron igualmente, solo para diagnóstico —son de solo lectura o trabajan sobre
clones en la carpeta temporal—, para que el intento siguiente no tropiece con otra discrepancia pendiente de descubrir.

## La discrepancia: la sonda de §11 cuenta un comentario

`rtk proxy grep -c 'vertebral: true' data/normas.yaml` imprimió **`16`**; la guía espera quince. La línea 3 de
`data/normas.yaml` es el comentario de cabecera que documenta la marca —«# vertebral: true marca las quince leyes de la
tabla…»— y la expresión, sin anclar, casa también con él. Las marcas reales son **exactamente quince** (líneas 12, 19,
26, 33, 41, 56, 64, 72, 93, 101, 108, 115, 122, 130 y 138), y `TestNormasDelRepositorio/vertebrales`, que exige las
quince y ninguna más, está en verde en `make ci` y en la última orden del propio §11.

**Arreglo en la raíz, que es la sonda**: `rtk proxy grep -cE '^[[:space:]]+vertebral: true$' data/normas.yaml`, que casa
solo con el campo sangrado de cada norma; probada antes de escribirla, imprime `15`. El comentario de `data/normas.yaml`
es correcto y no se toca: está fuera del directorio del hito, T025 no lo declara, y quitarle la frase para que una
sonda floja acierte sería arreglar el síntoma.

## Otras dos expectativas que no casaban con los datos congelados

Ninguna es un defecto del hito, pero las dos obligaban a improvisar al ejecutar la guía tal cual:

1. **§6, el nombre ambiguo.** `Villanueva` da `código: 3` y clase `no-encontrado` («ningún municipio de la relación se
   llama "Villanueva"»): en la relación congelada no hay ningún municipio con ese nombre exacto. La guía lo preveía
   —«si `Villanueva` resultara no ser ambiguo, el caso ambiguo se toma del propio registro»—, pero su orden de reserva
   solo listaba nombres. Con el primero, `Arroyomolinos`, las dos órdenes dan lo esperado: `código: 2`, clase
   `argumentos` y «el nombre "Arroyomolinos" es el de 2 municipios de la relación; consulta uno por su código INE: 10023
   Arroyomolinos (Cáceres); 28015 Arroyomolinos (Madrid)» (los otros dos, `Cabanes` y `Campillo, El`, igual). **Arreglo**:
   las dos primeras órdenes de §6 toman ya el nombre del registro, con la misma técnica que el código ausente de §6 y
   las órdenes de §5 y §9 —el primer nombre oficial que la relación repite, ambiguo por construcción—, y la reserva
   desaparece. Probadas antes de escribirlas: `nombre de más de un municipio: Arroyomolinos` y lo de arriba.
2. **§9, el registro de verificación.** La guía esperaba en `gates/verificacion-dir3.md` una fila del municipio «con
   entidades locales menores». No la hay, y es una decisión humana ya tomada: la pausa de T003 (`gates/tarea-T003.md`,
   «Resuelto en la pausa») dejó ese caso **sin ejemplo nombrado** porque el REL no publica volcado de entidades
   inferiores al municipio y nombrarlo sin fuente sería escribirlo de memoria, y lo dio por cubierto de hecho por la
   comprobación exhaustiva de los 8.132 municipios. **Arreglo**: la expectativa de §9 describe lo que el registro trae y
   remite a esa resolución. No relaja ningún control —ningún test ni umbral cambia—; deja escrito que es un pendiente de
   FR-046 y SC-008, y **el cuerpo de la publicación tiene que declararlo en *Pendientes*** para que quien fusione lo vea.

Los tres arreglos solo tocan `specs/008-h6-territorio-skill-legal/quickstart.md`, que está en el directorio del hito y
entre las rutas de esta tarea. Ningún fichero fuera de `specs/` cambia.

## Todo lo demás, conforme

| Escenario | Resultado en el intento 1 |
|---|---|
| Prerrequisitos | `go version go1.27.1 darwin/arm64`; rama `008-h6-territorio-skill-legal`; `make check-tools` sin salida y en verde; solo `fin del estado`; carpeta temporal recreada; `bin/kitlegal` compilado; `código: 0` |
| 1 · `make ci` | En verde de 11:41:06 a 11:42:11: `0 issues.`; todos los paquetes con tests en `ok` en los dos perfiles; `govulncheck` «No vulnerabilities found.» (y 1 vulnerabilidad en un módulo requerido que el código no llama, la de H5.1); `schema-check` y `skills-check` en `ok`; `no leaks found`; `all modules verified` en la raíz y en los cuatro módulos de herramientas; `tidy -diff` sin salida; `ci: todos los controles en verde` |
| 2 · cubierto | Sobre con `ok` verdadero, `kitlegal.territorio`, `kitlegal:applet/territorio`, `fecha_consulta` `2026-02-04T00:00:00Z` y `hash` `sha256:`; ocho claves; cobertura `configurado`, `configurado`, `verificado`; `estatal BOE`, `autonomico BOCM`, `provincial BOCM` con su motivo; `L01280745` y `comun`; `código: 0` |
| 3 · determinismo | `las cuatro salidas son idénticas byte a byte` |
| 4 · no cubierto | `no-configurado` en los dos boletines; solo `estatal`; Valladolid, Castilla y León, `comun`; recuento `0`; `código: 0` |
| 5 · foral | `comunidad foral: 15 · municipio: Abáigar`, `foral`, `no-configurado` |
| 6 · códigos | Ver arriba el ambiguo; `Municipio Que No Existe` → 3; `01999` → 3 `no-encontrado`; `99999` → 2 («la provincia "99" no está entre 01 y 52»); `2807` → 2; `280746` → 2 («el dígito de control recibido es "6" y el oficial es "5"»); ningún 4, 5 ni 6 |
| 7 · e2e | `ok`; con `-v`, `--- PASS: TestEntregaDelHito/territorio-matriz` |
| 8 · contrato | `schema-check` en `ok`; `territorio resolver` y `consulta`; `territorio · municipio` y `resolver`; sin posicional, `argumentos inválidos: expected "<consulta>"` y `código: 2`; en el clon, `schemas/municipio.json: el fichero no es la serialización canónica de sus partes` y `código: 2` |
| 9 · datos | Los nueve subtests en `PASS`; cabecera `fecha: "2026-02-04"` y `source: ine.municipios`; 8132 municipios; 19 comunidades; 18 sin `boletines`; `todos los municipios tienen DIR3 verificado`; integridad rota en el clon: «el municipio 01001 declara la comunidad 99 (…/99.yaml) y su provincia 01 es de la comunidad 16 (…/16.yaml)», `código: 2`; pliegue roto: «el municipio 01001, "ØAlegría-Dulantzi", se pliega a "øalegria dulantzi", con "ø" fuera del alfabeto del pliegue», `código: 2` |
| 10 · ids y fuzz | `ok`; trece entradas de corpus en cada objetivo; `FuzzCodigoINE` 3 761 523 ejecuciones y `FuzzCodigoDIR3` 4 005 877 en 30 s, los dos `PASS` y sin hallazgos |
| 11 · skill | Frontmatter con `kitlegal-applets: territorio` y las dos referencias; 170 líneas; las dos cabeceras «generado desde…»; `../../../bin/instalado/kitlegal`; **16 con la sonda vieja, 15 con la anclada**; solo `references/normas.md` (7 inserciones) en `skills/boe-legislacion`; los tres tests en `ok`; en el clon, `skills-sync` lo deja limpio y la edición a mano falla con «legal-core: references/jerarquia_normativa.md: contenido-distinto», `código: 2` |
| 12 · evals | Las tres evals; `formato`, `conjunto`, `conjunto-legal-core`, `normas-conocidas`, `grabado` y `cobertura-del-esquema` en `PASS` (también `avisos-del-esquema` y `avisos-de-la-skill`, de H5.1); `5c11d1d feat(H6): T020` antes que `7800adf feat(H6): T022`; solo `fin del diff` |
| 13 · documentación | `CHANGELOG.md:26` y `:281` (*De H6*); `legal-core` en `README.md` y 3 veces en `CONTRIBUTING.md`, y `data/territorio/` en los dos; la fila `mpt.rel` con 2026-09-21; `0` y `código de grep: 1` |
| Limpieza | Carpeta borrada; solo `fin del estado` |

## Cobertura medida sobre ese `make ci`

Con `go tool cover -func` (total del perfil) y, por paquete, la suma de sentencias del perfil contando cada bloque una
vez:

| Umbral | Exigido | Perfil unitario | Perfil de integración |
|---|---|---|---|
| Global | ≥ 70 % | **96,8 %** (6786/7007) | **97,3 %** (6815/7007; `go tool cover -func`: 97,2 %) |
| `internal/core/**` | ≥ 85 % | **98,6 %** (583/591) | 98,6 % |
| `internal/core/ids` | — | 100,0 % (64/64) | 100,0 % |
| `internal/core/territorio` | — | 100,0 % (446/446) | 100,0 % |
| `internal/core/schema` | — | 90,1 % (73/81), sin cambios desde H5.1 | 90,1 % |

El paquete `data` sale con 0/19: `Comunidades()` la ejercen los tests de `internal/app` a través de
`FuentesEmbebidas` (`registro.go:193`, ocho pasadas en el perfil), pero el perfil de `go test ./...` solo atribuye a
cada paquete lo que ejecutan sus propios tests, y `data` no tiene. Cabe de
sobra en el umbral global; conviene decirlo en el cuerpo para que no se lea como código sin probar.

## Lo que tiene que hacer el intento 2

1. Leer esta nota y ejecutar la guía **corregida** desde los prerrequisitos, tal cual y en orden; si todo da lo
   esperado, medir de nuevo la cobertura sobre su propio `make ci`.
2. Escribir `gates/pr-h6.md` con la plantilla del ritual y, en *Pendientes*, además de la regla de SC-015: el caso con
   entidades locales menores de FR-046 y SC-008 sin ejemplo nombrado (arriba, punto 2), y que la guía se corrigió en
   este cierre (las tres correcciones, con su motivo).
3. Datos ya comprobados para el cuerpo: `go.mod` y `go.sum` no cambian frente a `main` («Dependencias: ninguna
   nueva»); el único módulo que pasa a enlazarse en el binario es `go.yaml.in/yaml/v3`, con su motivo escrito en
   `modulosDelBinario` de `internal/arch_test.go` (T009): `internal/core/territorio` analiza con él los ficheros
   congelados de `data/territorio/`, entra por `internal/app` en cuanto existe el applet, es la biblioteca de YAML que
   el repositorio fijó en H5, y v4 no tiene versión estable (research.md D15).

## Intento 2: en verde y marcada

Ejecutado en una sesión desatendida el **2026-09-22 entre las 11:55 y las 12:03 (hora de Madrid), sobre `2df74db`**
(`feat(H6): T024`, el mismo commit que el intento 1), con el árbol limpio fuera de `specs/008-h6-territorio-skill-legal/`
(solo `quickstart.md`, con las tres correcciones del intento 1 sin commitear todavía, y los ficheros de estado del
workflow). Prerrequisitos, `make ci` y escenarios 1 a 13 tal cual y en orden, con las formas `rtk proxy` de la tabla de
la guía, y después la limpieza. **Ningún resultado distinto del esperado.**

- Las tres correcciones se confirmaron sobre la guía en ejecución: la sonda anclada de §11 imprime `15`; las dos
  órdenes de §6 toman `Arroyomolinos` del registro y dan `código: 2`, clase `argumentos` y el mensaje con los dos
  candidatos; y el registro de verificación de §9 trae lo que la expectativa corregida describe.
- Todo lo demás dio lo mismo que en el intento 1, con las cifras de fuzz de esta campaña: `FuzzCodigoINE` 3 791 151
  ejecuciones y `FuzzCodigoDIR3` 3 929 765 en 30 s, los dos `PASS`, y el árbol limpio después. La tabla completa de
  resultados está en `gates/pr-h6.md`, sección *Evidencia*.
- `make ci` en verde de 11:55:25 a 11:56:24 (`ci: todos los controles en verde`); la cobertura medida sobre su
  perfil coincide con la del intento 1 —global **96,8 %** (6786/7007) y **97,3 %** (6815/7007) en integración;
  `internal/core/**` **98,6 %** (583/591); `ids` y `territorio` al 100 %—, porque el árbol de código no cambió entre
  los dos.
- **`gates/pr-h6.md` escrito** con la plantilla del ritual (objetivo, alcance, dependencias, controles añadidos,
  evidencia, decisiones y pendientes), «Dependencias: ninguna nueva» y el motivo de `go.yaml.in/yaml/v3`, las medidas
  fechadas por `2df74db`, y en *Pendientes* la regla de SC-015, el caso con entidades locales menores de FR-046 y
  SC-008 sin ejemplo nombrado y las tres correcciones de la guía con su motivo. Además, cuatro pendientes que salieron
  al preparar el cuerpo y que no cambian ningún control: la vulnerabilidad `GO-2026-5970` en un módulo requerido que el
  código no llama (heredada de `main`, como en H5.1), las dos ramas de error sin test que anotó T016, dos frases de
  `CONTRIBUTING.md` heredadas de `main` que sitúan el release en H6 cuando ADR 0013 lo lleva a H19, y un ejemplo con
  el dígito `8` de memoria en `research.md:298`.
- Ficheros escritos por este intento: `gates/pr-h6.md` (nuevo) y esta nota; `quickstart.md` queda como lo dejó el
  intento 1. Ningún fichero fuera del directorio del hito cambia; `bin/` se borró en la limpieza y `coverage.out` y
  `coverage-integration.out` son los de `make ci`, ignorados por git.
- Verificación determinista al terminar: `make ci` en primer plano otra vez, en verde, y la tarea marcada `[X]` en el
  mismo turno.
