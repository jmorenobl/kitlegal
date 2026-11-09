# Contrato: las variantes de las tres familias

FR-020 a FR-025; SC-007, SC-009. Research D9 a D13, D15, V1, V3, V6; data-model §1 a §3 y §6.

## 1. Dónde viven

```text
laboratorio/
├── variantes.yaml, variantes/, evals/                 # H39, sin cambios
├── normativa/                                         # H40: las cuatro de hoy, sin cambiar un byte
│   └── normativa-{control,remisiones-en-orden,cuerpo-corto,testigo-en-referencia}/skills/normativa/…
├── plazos-y-territorio/                               # NUEVO: control y las tres formas
│   └── plazos-y-territorio-{control,sin-condicion,condicion-por-orden,lista-con-verbo}/skills/plazos-y-territorio/
│       ├── SKILL.md
│       └── references/{administrativo,calculo_de_plazos,jerarquia_normativa,municipios}.md
└── jurisprudencia/                                    # NUEVO
    └── jurisprudencia-control/skills/jurisprudencia/
        ├── SKILL.md
        └── references/{consulta_preparada,documento_aportado}.md
```

Las dos `<skill>-control` nuevas se copian de `skills/<skill>` del árbol de partida, que es el de `v0.9.0` (research V1),
antes de cambiar ninguna skill: `cp -R skills/<skill>/. laboratorio/<skill>/<skill>-control/skills/<skill>/`. Las tres
variantes de `plazos-y-territorio` son su control con los cambios de §2 en el `SKILL.md`; sus cuatro referencias, copia
byte a byte de las de control. Ninguna `<skill>-producto` tiene directorio: es `skills/<skill>` en cada sondeo (FR-022).
Nada de esto lo empotra el binario, lo instala `kitlegal skills install`, lo regenera `make skills-sync` ni lo elige o
mide el job de evals (FR-025), y el tope de 11 000 caracteres no se aplica a ninguna variante versionada (FR-024).

## 2. Las tres formas de `plazos-y-territorio` (FR-023)

Cada una es `plazos-y-territorio-control` con estos tramos del `SKILL.md` (`diff -u`, cada `-`/`+` una línea entera) y
nada más: la línea de `jerarquia_normativa.md`, la del paquete, «Si no lo dice, **pregúntalo** y espera» en el paso 1,
el resto del cuerpo, el frontmatter y las referencias quedan como en control. Las órdenes nuevas siguen el ADR 0051,
punto 1: «lee», el enlace, cuándo y lo que trae.

### 2.1 `plazos-y-territorio-sin-condicion`

```diff
@@ -29,13 +29,12 @@
 <!-- fin de los paquetes -->
 - [jerarquia_normativa.md](references/jerarquia_normativa.md): para el nivel que regula algo, con el autonómico de la
   `comunidad` de `data`, y la norma que prevalece.
-- [municipios.md](references/municipios.md): al presentar un territorio, si falta el municipio o no sale con 0, y
-  para explicar su cobertura.
-- [calculo_de_plazos.md](references/calculo_de_plazos.md): en los municipios que cuentan, un plazo sin calcular y una
-  lectura fallida o con `redaccion_posterior`.
 
 ## Protocolo
 
+Antes de la primera orden o llamada, lee [municipios.md](references/municipios.md): qué hacer si falta el municipio o
+no sale con 0, qué trae `data`, la plantilla del territorio y qué dice su cobertura.
+
 Territorio: del 1 al 4, y el 5 y el 6 si va más allá. Plazos o recursos: el 7 y, solo si hay plazo, del 8 al 11.
 
 1. **El municipio**, de la conversación. Si no lo dice, **pregúntalo** y espera: no lo supongas, tampoco el de quien
@@ -54,7 +53,9 @@
    nada**, ni el municipio ni la fecha. Si el asunto está en la cobertura de su paquete, la línea `⚠ NO CUBIERTO:` de
    su fila, entera, **ni `plazos calcular`** ni ninguna fecha (regla 9). Si no dice la materia, **no la preguntes ni
    detengas la respuesta**: sigue con el 8, y al responder di que la fecha no vale si la ley regula el asunto aparte,
-   nombrando esos asuntos solo con la referencia. Un acto presunto no tiene fecha: ningún plazo.
+   nombrando esos asuntos solo con la referencia. Un acto presunto no tiene fecha: ningún plazo. Antes del 8, lee
+   [calculo_de_plazos.md](references/calculo_de_plazos.md): los municipios que cuentan, el plazo sin calcular y la
+   lectura en la fecha de inicio.
 8. **Situar el plazo**: la fecha de inicio (notificación o publicación; en el silencio, acuerdo de iniciación o
    entrada en el registro), que si falta **se pregunta**; y los municipios que cuentan, con los pasos 1 a 4: **el del
    interesado y el de la sede del órgano**, o en el contencioso-administrativo **el de la sede del juzgado**.
```

### 2.2 `plazos-y-territorio-condicion-por-orden`

```diff
@@ -29,10 +29,6 @@
 <!-- fin de los paquetes -->
 - [jerarquia_normativa.md](references/jerarquia_normativa.md): para el nivel que regula algo, con el autonómico de la
   `comunidad` de `data`, y la norma que prevalece.
-- [municipios.md](references/municipios.md): al presentar un territorio, si falta el municipio o no sale con 0, y
-  para explicar su cobertura.
-- [calculo_de_plazos.md](references/calculo_de_plazos.md): en los municipios que cuentan, un plazo sin calcular y una
-  lectura fallida o con `redaccion_posterior`.
 
 ## Protocolo
 
@@ -41,7 +37,8 @@
 1. **El municipio**, de la conversación. Si no lo dice, **pregúntalo** y espera: no lo supongas, tampoco el de quien
    usa el kit.
 2. **Resolverlo** con `territorio_resolver` o `kitlegal territorio resolver <nombre o código INE> --json`, tal como lo
-   dio la persona. Sin herramienta ni binario, la regla 7.
+   dio la persona. Sin herramienta ni binario, la regla 7. Si falta el municipio o `territorio resolver` no sale con
+   `0`, lee [municipios.md](references/municipios.md): qué hacer en cada caso.
 3. **Responder con `data`**, la cobertura en su forma fija y de cuándo son los datos (`fecha_consulta`). **Ningún
    boletín que no esté en `boletines`.**
 4. **Sin el 0**: con el 2 o el 3 de un municipio, lo que dice la referencia de los municipios; con otro código, di que
@@ -57,15 +54,20 @@
    nombrando esos asuntos solo con la referencia. Un acto presunto no tiene fecha: ningún plazo.
 8. **Situar el plazo**: la fecha de inicio (notificación o publicación; en el silencio, acuerdo de iniciación o
    entrada en el registro), que si falta **se pregunta**; y los municipios que cuentan, con los pasos 1 a 4: **el del
-   interesado y el de la sede del órgano**, o en el contencioso-administrativo **el de la sede del juzgado**.
+   interesado y el de la sede del órgano**, o en el contencioso-administrativo **el de la sede del juzgado**. Si
+   alguno de los municipios cuenta, lee [calculo_de_plazos.md](references/calculo_de_plazos.md): cuáles cuentan y qué
+   se pasa a `plazos`.
 9. **Pedir el plazo**: el tipo, de la referencia del paquete; `plazos_calcular` o `kitlegal plazos calcular <tipo>
    <inicio> <municipios>... --json`, con los municipios del 8 en su orden. Pídelo aunque creas que no regía. **Con el 3
    y un `data.mensaje` que empieza por «el plazo de tipo», no hay fecha**: `⚠ PLAZO SIN CALCULAR:`, sin la línea del
-   plazo ni ninguna fecha de vencimiento, y sin el 10.
+   plazo ni ninguna fecha de vencimiento, y sin el 10. Con ese 3, lee
+   [calculo_de_plazos.md](references/calculo_de_plazos.md): por qué no hay fecha y qué se dice.
 10. **Leer cada precepto de `data.preceptos` en la fecha de inicio**, con `boe_articulo` y `fecha` o
     `kitlegal boe articulo <identificador> <id> --fecha <inicio> --json`, sin `kitlegal graph check`, y citarlo con
     `⚠ REDACCIÓN A FECHA:`. Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di aparte que no has podido leerlo, sin
-    afirmar qué dice.
+    afirmar qué dice. Con 4 o 5, lee [calculo_de_plazos.md](references/calculo_de_plazos.md): qué se dice del bloque
+    que falta. Con `redaccion_posterior`, lee [calculo_de_plazos.md](references/calculo_de_plazos.md): su línea de
+    redacción de hoy.
 11. **Responder**: `data.fecha` en la línea del plazo y ninguna otra fecha de vencimiento (regla 8); los festivos que
     la movieron, de `data.festivos`, con fecha, ámbito y, si es local, municipio; cada aviso de `data.avisos`, su
     `texto` tal cual; el alcance de los festivos, solo si lo dice un aviso.
```

### 2.3 `plazos-y-territorio-lista-con-verbo`

```diff
@@ -29,10 +29,10 @@
 <!-- fin de los paquetes -->
 - [jerarquia_normativa.md](references/jerarquia_normativa.md): para el nivel que regula algo, con el autonómico de la
   `comunidad` de `data`, y la norma que prevalece.
-- [municipios.md](references/municipios.md): al presentar un territorio, si falta el municipio o no sale con 0, y
+- Lee [municipios.md](references/municipios.md): al presentar un territorio, si falta el municipio o no sale con 0, y
   para explicar su cobertura.
-- [calculo_de_plazos.md](references/calculo_de_plazos.md): en los municipios que cuentan, un plazo sin calcular y una
-  lectura fallida o con `redaccion_posterior`.
+- Lee [calculo_de_plazos.md](references/calculo_de_plazos.md): en los municipios que cuentan, un plazo sin calcular y
+  una lectura fallida o con `redaccion_posterior`.
 
 ## Protocolo
```

«una» pasa a la línea siguiente para no pasar de 120 columnas; el número de caracteres es el mismo.

### 2.4 Las órdenes de cada forma

| Forma | Referencia | Sitio | Condición | Lo que trae |
|---|---|---|---|---|
| sin condición | `municipios.md` | párrafo propio al principio del «Protocolo», antes de la primera orden o llamada | ninguna | qué hacer si falta o no sale con 0, `data`, la plantilla y la cobertura |
| | `calculo_de_plazos.md` | final del paso 7, «Antes del 8» | ninguna (antes del cálculo) | los municipios que cuentan, el plazo sin calcular y la lectura en la fecha de inicio |
| una condición por orden | `municipios.md` | paso 2, detrás de `territorio resolver` | falta el municipio **o** `territorio resolver` no sale con `0` | qué hacer en cada caso |
| | `calculo_de_plazos.md` | final del paso 8 | alguno de los municipios cuenta | cuáles cuentan y qué se pasa a `plazos` |
| | `calculo_de_plazos.md` | paso 9, detrás de la llamada | el 3 con «el plazo de tipo» | por qué no hay fecha y qué se dice |
| | `calculo_de_plazos.md` | paso 10, detrás de la lectura | sale con 4 o 5 | qué se dice del bloque que falta |
| | `calculo_de_plazos.md` | paso 10, detrás de la lectura | `redaccion_posterior` | su línea de redacción de hoy |
| lista con el verbo | `municipios.md` | lista de «Referencias», en su sitio | las de control | — (como control) |
| | `calculo_de_plazos.md` | ídem | las de control | — (como control) |

Por qué cada sitio, en research D11; a qué condiciones llegan las cuatro evals del sondeo en cada forma, en research D15.

## 3. Cuerpos y lo que movería H41.1 (FR-024)

Medidos en esta sesión sobre copias (research V3). El run no aplica ningún movimiento.

| Variante | Cuerpo | Pasa de 10 930 | Lo que movería H41.1 | Cuerpo tras ello |
|---|---|---|---|---|
| `plazos-y-territorio-sin-condicion` | 10 991 | sí | M1 | 10 882 |
| `plazos-y-territorio-condicion-por-orden` | 11 245 | sí | M1, M2, M3 | 10 891 |
| `plazos-y-territorio-lista-con-verbo` | 10 938 | sí | M1 | 10 829 |

Cada movimiento lleva su texto entero, al final de `references/calculo_de_plazos.md`, detrás de una línea en blanco,
con la regla del ADR 0052, punto 4 (research D13):

- **M1.** Del paso 10: «Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di aparte que no has podido leerlo, sin afirmar
  qué dice.» (105 caracteres). Llega la eval 11 (`sin_servir`), y en las tres formas la referencia se manda leer en ese
  caso (sin condición, antes del 8; con «Con 4 o 5»; con «una lectura fallida» de la lista).
- **M2** (solo `-condicion-por-orden`). Del paso 9: «**Con el 3 y un `data.mensaje` que empieza por «el plazo de
  tipo», no hay fecha**: `⚠ PLAZO SIN CALCULAR:`, sin la línea del plazo ni ninguna fecha de vencimiento, y sin el 10.»;
  la orden que la sigue pasa a «Con el 3 y un `data.mensaje` que empieza por «el plazo de tipo», lee
  [calculo_de_plazos.md](references/calculo_de_plazos.md): por qué no hay fecha y qué se dice.». Llega la eval 10.
- **M3** (solo `-condicion-por-orden`). Del paso 8: «: **el del interesado y el de la sede del órgano**, o en el
  contencioso-administrativo **el de la sede del juzgado**», que la referencia ya dice en «Los municipios que cuentan»;
  el paso queda en «… y los municipios que cuentan, con los pasos 1 a 4.». Llegan las evals 05, 10 y 11, por «si alguno
  de los municipios cuenta», a su rama administrativa; a la del contencioso, la eval 09 del job, que el sondeo no usa
  (research D13: sin M3, la variante queda unos 80 caracteres por encima).

## 4. Controles

- **`TestVariantesDeControl`** (`internal/evals/laboratorio_familias_test.go`; FR-020, FR-022; SC-007, SC-009), por
  familia del producto:
  - la carpeta `laboratorio/<skill>/` tiene un directorio por variante versionada de la familia y nada más, y ninguno
    se llama `<skill>-producto`;
  - los ficheros de `<skill>-control` son, ruta a ruta, los de la tabla de huellas SHA-256 del test, escrita al copiar
    las dos nuevas (17 ficheros: 9 de `normativa`, 5 de `plazos-y-territorio`, 3 de `jurisprudencia`); falla con la
    ruta del que difiere, falta o sobra;
  - los nombres de las variantes de las cuatro familias no se repiten.
- **`TestVariantesDeNormativa`** (H40; FR-021): sus subtests de las variantes declaradas, sus remisiones, las
  referencias y los nombres, sin cambiar ningún valor esperado. Salen «control es skills/normativa» (FR-020: deja de
  compararse con `skills/normativa`; lo cubre la huella) y «directorios de la familia» (lo hace `TestVariantesDeControl`
  para las tres).
- **`TestVariantesDePlazosYTerritorio`** (`internal/evals/laboratorio_familias_test.go`; FR-023; SC-009), por cada
  variante de §2, con su declaración (data-model §6):
  1. el «después» de cada sustitución, normalizado como en H40 (`textoDelHito`), es el texto de §2;
  2. la declaración se aplica sobre `plazos-y-territorio-control` y el directorio de la variante es, ruta a ruta y byte
     a byte, el resultado; falla con la ruta del primer fichero que difiere, falta o sobra;
  3. cada sustitución es del `SKILL.md`, y, quitados de los dos cuerpos los segmentos de `skills.Frases`
     (contracts/remisiones.md §6) con un enlace a `references/municipios.md` o a `references/calculo_de_plazos.md`, el
     de la variante y el de control tienen los mismos segmentos en el mismo orden; falla con el primer segmento que
     difiere.
- **`TestLaboratorioFueraDelProducto`** (H39 y H40, generalizado; FR-025; SC-009): los nombres que no pueden ser un
  tramo de una ruta de lo que el binario empotra, de lo que deja `kitlegal skills install --dir`, de una skill que
  ejecute el job ni de una carpeta de `evals/` son los de las skills de H39 y los de los directorios de variante de las
  carpetas de las tres familias del producto; ninguna ruta de `laboratorio/` tiene un tramo `<skill>-producto` ni mide
  (`rutaQueMide`). Lo demás, como hoy.
- **`TestSkillsDelProductoDesdeControl`** (contracts/remisiones.md §8).

## 5. Uso, de fuera adentro

Las lee el agente de una sesión del sondeo: el `SKILL.md` de su variante una vez, al activar la skill, y cada referencia
cuando la abre, que es lo que se mide; en la aceptación, 760 sesiones, una variante por sesión. Su tamaño es fijo —§3 y
los ficheros de control— y no crece con el uso del producto: cientos de normas y miles de bloques consultados no las
tocan. No son una señal. Ninguna skill del producto ni el binario que se distribuye las leen.
