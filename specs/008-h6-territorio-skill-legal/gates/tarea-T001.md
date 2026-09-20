# T001 — el esquema queda cerrado por el ejecutor; la relación del INE la escribe la persona en la pausa (intento 2)

**Estado**: marcada `[X]` con `make ci` en verde. `schemas/territorio-municipios.yaml.json` está escrito y verificado;
`data/territorio/municipios.yaml` **no existe todavía** y lo escribe una persona en la pausa humana que este mismo
commit dispara. Abajo, «Lo que hace la persona en la pausa».

## Por qué el intento 1 no podía llegar nunca a la pausa

El intento 1 dejó la tarea `[ ]` porque la relación del INE no estaba en la máquina. Esa lectura es un callejón sin
salida por la mecánica del propio bucle (`.specify/workflows/hito/workflow.yml`):

- la pausa humana de FR-045 es `gate_humano_datos`, y solo se dispara si `clasificar_datos` ve un fichero nuevo bajo
  `schemas/` en `git diff <base> HEAD`, es decir, **ya commiteado** (research.md V9; contrato de datos §4: «la pausa se
  dispara porque la tarea añade también el esquema bajo `schemas/`»; tasks.md, «Ocho provocan pausa humana —T001, …:
  material existente modificado o esquema nuevo»);
- `estado_cierre` solo commitea una tarea marcada `[X]`; una tarea `[ ]` con nota nueva es «redelimitada» y no se
  commitea;
- por tanto, dejar T001 sin marcar hasta que exista `municipios.yaml` significa que la pausa en la que la persona debe
  escribirlo (contrato de datos §4, pasos 1 y 4; D17) no llega jamás: el bucle gasta los tres intentos y se detiene en
  `siguiente_tarea` («intentos agotados») sin pausa y sin fichero.

El material tampoco puede obtenerlo el ejecutor por ninguna vía legítima: ni descarga (FR-043, ADR 0017, «Cero red en
todo el bucle») ni memoria («Nada escrito de memoria»: ningún código INE ni dígito de control). La única vía que el
diseño prevé es la persona en la pausa, y la pausa exige el commit del esquema.

**Decisión de este intento**: la parte del ejecutor —el esquema, verificado— cierra la tarea; la parte de la persona
—la relación congelada— entra en la pausa que este cierre dispara. Es el mismo patrón del manifiesto de grabaciones de
H4 (la tarea `[datos]` se marca con el manifiesto y las grabaciones se confirman en su pausa). La cláusula «si el
material no está disponible se detiene sin marcarse» se lee como lo que es: una guarda para que el ejecutor no escriba
ningún dato del INE, no como una condición que haga imposible la pausa.

**Nada pasa en vacío si la persona no escribe el fichero**: T002 copia de `municipios.yaml` las columnas `comunidad` y
`provincia` y, sin él, se detiene sin marcarse; T007 lo embebe con `//go:embed` (error de compilación si falta) y
`TestTerritorioDelRepositorio` exige los cuatro ficheros y las 19 comunidades.

## Lo que hace la persona en la pausa

Fuera del repositorio (contrato de datos §4, pasos 1 y 4):

1. Descargar `diccionario26.xlsx` por la dirección de la fila `ine.municipios` de `docs/SOURCES.md`
   (<https://www.ine.es/daco/daco42/codmun/diccionario26.xlsx>; 8.132 municipios con dígito de control; referencia
   2026-02-04, la que ya declara esa fila).
2. Escribir `data/territorio/municipios.yaml` con **una línea por municipio**, en la forma de data-model §3.1, con
   `fecha: "2026-02-04"` y `source: ine.municipios`:

   ```yaml
   fecha: "2026-02-04"
   source: ine.municipios
   municipios:
     "28074": {dc: "8", nombre: "Leganés", provincia: "28", comunidad: "13"}
   ```

   La clave es el código INE de cinco cifras (provincia + municipio, **sin** el dígito de control); `dc` es el dígito
   de control oficial de la hoja; `provincia` son las dos primeras cifras de la clave; `comunidad` el código de dos
   cifras de la comunidad autónoma de la hoja.
3. Confirmarlo en la rama del hito (`git add data/territorio/municipios.yaml && git commit`) y aprobar la pausa. T002
   arranca desde ese commit.

**Comillas obligatorias** (comprobado con el lector común, no supuesto): el lector `skills.ValidarDocumentoYAML`
(`go.yaml.in/yaml/v3`) resuelve un `fecha: 2026-02-04` sin comillas como `!!timestamp` y lo normaliza a
`"2026-02-04T00:00:00Z"`, que **no** casa con `^[0-9]{4}-[0-9]{2}-[0-9]{2}$`; una clave `28074:` sin comillas es un
entero y se rechaza como «mapa con claves que no son texto»; un `dc: 8` sin comillas es un número y se rechaza
(«got number, want string»). Con comillas, todo casa. El ejemplo de data-model §3.1 escribe la fecha sin comillas
porque ilustra la forma, no el literal. Lo mismo afecta a `fecha` de `dir3.yaml`, `estado.yaml` y los 19 ficheros de
comunidad (pausas de T002 y T003).

## Verificación del esquema (este intento)

El esquema no lo lee ningún control hasta T007, así que se verificó con un **arnés temporal**
(`internal/skills/territorio_arnes_temporal_test.go`, borrado antes de `make ci`; patrón de T018/T020 de H5) sobre el
lector común, `skills.CompilarEsquema` + `skills.ValidarDocumentoYAML`:

- **Verde**: compila con las aserciones de formato activas y acepta el ejemplo de §3.1 entrecomillado, leído a un tipo
  con las cinco columnas.
- **Rechazos**, cada uno en su ruta y su línea: fecha sin comillas (timestamp) y sin forma; raíz sin `fecha`, sin
  `source`, sin `municipios` o con clave de más; `source` vacío; clave de cuatro y de seis cifras, con letras y sin
  comillas; `municipios` como lista; `dc` de dos cifras, vacío y como número; `nombre` vacío; `provincia` y `comunidad`
  de una y de tres cifras; columna que falta, columna de más, fila que no es mapa; clave repetida en la raíz.
- **Mutantes del esquema** (once copias relajadas fuera del repositorio, una restricción quitada en cada una): cada
  copia hace caer exactamente los casos que vigilan esa restricción y ningún otro —`additionalProperties` de la fila y
  de la raíz, `propertyNames.pattern`, los `pattern` de `fecha`, `dc`, `provincia` y `comunidad`, los `minLength` de
  `nombre` y `source`, y los `required` de la fila y de la raíz—.

| Clave | Forma |
|---|---|
| raíz | `additionalProperties: false`; obligatorias `fecha`, `source`, `municipios` |
| `fecha` | texto con `pattern` `^[0-9]{4}-[0-9]{2}-[0-9]{2}$` |
| `source` | texto no vacío (el identificador lo comprueba contra `docs/SOURCES.md` el test de T007) |
| `municipios` | objeto con `propertyNames.pattern` `^[0-9]{5}$` y cada valor `#/$defs/municipio` |
| `municipio` | `additionalProperties: false`; obligatorias `dc` (`^[0-9]$`), `nombre` (no vacío), `provincia` y `comunidad` (`^[0-9]{2}$`) |

`make ci` en primer plano tras borrar el arnés: **en verde** (exit 0; `schema-check` solo compara los esquemas de los
verbos publicados y `skills-check` solo `data/normas.yaml`).

## Notas

- El intento 1 afirmó que `~/Downloads` estaba vacío; en esta sesión ese directorio devuelve «Operation not permitted»
  (permiso de macOS, con y sin sandbox), así que no se pudo comprobar ni desmentir. `~/Desktop`, `~/Documents`, el árbol
  de trabajo y los ficheros ignorados no contienen ningún volcado del INE.
- Quedan fuera del repositorio, sin borrar porque hacerlo exigía aprobación en headless: las once copias mutantes en
  `/tmp/kitlegal-arnes-T001/` y el registro de esta verificación en `/tmp/kitlegal-T001-ci.log`. No afectan a nada.
- Mejora de proceso, para la lista de mejoras (no se aplica dentro del hito): toda tarea `[datos]` cuya entrega la
  completa una persona en la pausa debería decirlo en su línea con la secuencia explícita —«el ejecutor marca la tarea
  con el esquema; la persona añade el fichero en la pausa y lo confirma en la rama»— en lugar de la cláusula «se detiene
  sin marcarse», que en T001 produce el bloqueo descrito arriba. Afecta igual a T002, T003 y T015 de este hito.

## Resuelto en la pausa (2026-09-20)

`data/territorio/municipios.yaml` **escrito**, desde la fuente declarada y sin nada de memoria.

- Origen: <https://www.ine.es/daco/daco42/codmun/diccionario26.xlsx>, la dirección de la fila `ine.municipios` de
  `docs/SOURCES.md`. SHA-256 del fichero descargado:
  `07f8e8d64eba73fe9d425196fe82f4e09a9886a287abd650e88ca8d98dd49052` (306.819 bytes; `Last-Modified` 2026-03-13).
- Lectura: la hoja se abrió con la biblioteca estándar (`zipfile` + `ElementTree`, un `.xlsx` es un zip de XML),
  fuera del repositorio y sin versionar. Ni el producto ni `make ci` necesitan nada nuevo.
- `fecha: "2026-02-04"` y `source: ine.municipios`, como fija esta nota; las claves ordenadas por código INE.

### S1 y S5 quedan resueltos

La hoja trae **exactamente** las cinco columnas que S1 suponía: `CODAUTO`, `CPRO`, `CMUN`, `DC`, `NOMBRE`
(cabecera en la fila 2; 8.134 filas − 2 de cabecera = **8.132 municipios**, la cifra que declara `docs/SOURCES.md`).
Comprobado además, sobre el fichero generado y con el lector real (`skills.CompilarEsquema` +
`skills.ValidarDocumentoYAML`, en un `_test.go` temporal borrado antes de `make ci`): valida contra
`schemas/territorio-municipios.yaml.json`; la `provincia` de cada fila son las dos primeras cifras de su clave;
**52 provincias, cada una con una sola comunidad** (el control de coherencia de data-model §2.1 punto 6 pasa); y
**19 comunidades**, las que exige `TestTerritorioDelRepositorio`. S5 también se confirma: 13 Madrid, 15 Navarra,
16 País Vasco, 18 Ceuta, 19 Melilla. `make ci` en verde con el fichero dentro.

### Aviso para T002 y T003: el dígito de control de Leganés es **5**, no 8

La fila real del INE es `13 | 28 | 074 | 5 | Leganés`. Los artefactos del hito traen `8` en cinco sitios, escrito de
memoria antes de tener la hoja, y el juez del plan ya avisó de que un ejemplo incoherente puede «acabar copiado en un
fixture». **Ninguno de estos valores se copia a `data/`, a `testdata/` ni a un test**; hay que corregirlos en la tarea
que toque cada fichero:

| Sitio | Dice | Es |
|---|---|---|
| `research.md:128`, `data-model.md:226` (ejemplo de `municipios.yaml`) | `dc: "8"` | `dc: "5"` |
| `research.md:295` (D11, `digito_de_control`) | `"8"` | `"5"` |
| `research.md:298`, `data-model.md:241` (DIR3 de Leganés) | `L01280748` | lo que verifique T003 |
| `contracts/identificadores-ine-y-dir3.md:88` (casos de forma) | `L01280748` | solo ilustra la **forma**, no el municipio |

El DIR3 correcto **no se deduce aquí**: S2 supone que es `L` + el número de inscripción del REL y esa derivación es
justo lo que verifica la pausa de T003. Si el dígito del REL coincide con el del INE, Leganés sería `L01280745`; el
valor que entre en `dir3.yaml` sale del volcado del REL, no de este cálculo.
