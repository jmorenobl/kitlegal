# Contrato: los datos de la cobertura y las regiones que salen de ellos

FR-001 a FR-003, FR-006, FR-007, FR-011, FR-023; SC-001, SC-002, SC-008. Todo en `internal/skills`, que solo usa la
sincronía (`make skills-sync`, `make ci`): ningún applet cambia y el binario no lee `data/cobertura.yaml`.

## 1. `data/cobertura.yaml`

```yaml
# (cabecera: deja de decir que solo lleva lo regulado aparte; dice que es la lista de lo cubierto por capacidad y
# materia, el motivo de lo no declarado y lo que la ley regula aparte, y quién lee cada parte)
motivo_de_lo_no_declarado: fuera de lo que kitlegal cubre
cubierto:
  - capacidad: normativa
    alcance: <el de FR-001, fila 1>
  - capacidad: jurisprudencia
    alcance: <el de FR-001, fila 2>
  - capacidad: plazos
    alcance: <el de FR-001, fila 3>
    materia: administrativo
  - capacidad: territorio
    alcance: <el de FR-001, fila 4>
regulados_aparte:
  # sus 9 filas de hoy, sin cambiar un byte
```

Cada `alcance` va en una línea, como escalar YAML sin comillas (el de los plazos lleva una comilla invertida, que el
lector lee tal cual: research V15). `TestLeerCobertura` (subtest `repositorio`) lee el fichero del repositorio y exige el motivo y las cuatro
filas, campo a campo, con los textos de FR-001; las 9 de `regulados_aparte` las sigue fijando la región (§5).

## 2. `schemas/cobertura.yaml.json`

Se añaden, sin reformatear lo de hoy (research D5):

- en la raíz, `motivo_de_lo_no_declarado` en `required` y en `properties` (`{"type": "string", "minLength": 1}`), y
  `cubierto` en `properties` (`{"type": "array", "items": {"$ref": "#/$defs/cubierto"}}`);
- en `$defs`, `cubierto`: objeto con `additionalProperties: false`, `required: ["capacidad", "alcance"]`, `capacidad`
  con `enum` `["normativa", "jurisprudencia", "plazos", "territorio"]`, `alcance` con `minLength: 1` y `materia` con
  `{"type": "string"}`, sin patrón: si es o no un `vertical` lo dice solo la sincronía, con el defecto de §3 (research
  D4 y D5).

Los defectos (hojas del esquema, `internal/skills/esquemas.go`; textos de `jsonschema` v6.0.3, research V4), uno por
mutante de `TestLeerCobertura` sobre `coberturaDePrueba` con las dos claves nuevas:

| Mutante (FR-002) | Defecto |
|---|---|
| fila sin `capacidad` | `cubierto/0, línea <n>: missing property 'capacidad'` |
| fila sin `alcance` | `cubierto/0, línea <n>: missing property 'alcance'` |
| `capacidad: documentos` | `cubierto/0/capacidad, línea <n>: value must be one of 'normativa', 'jurisprudencia', 'plazos', 'territorio'` |
| fila con `normas: [BOE-A-2015-10565]` | `cubierto/0, línea <n>: additional properties 'normas' not allowed` |
| sin `motivo_de_lo_no_declarado` | `línea <n>: missing property 'motivo_de_lo_no_declarado'` |
| `motivo_de_lo_no_declarado: ""` | `motivo_de_lo_no_declarado, línea <n>: minLength: got 0, want 1` |

`<n>` es la línea del nodo en el documento del caso; el test fija la cifra. Los casos de hoy, sobre `regulados_aparte`,
no cambian sus mensajes.

## 3. La región de lo cubierto (`internal/skills/cubierto.go`)

En el `SKILL.md` que lleve sus dos marcas, cada una sola en su línea:

```text
<!-- lo que cubre la base: lo genera make skills-sync -->
- **Normativa**: <alcance de la fila 1>
- **Jurisprudencia**: <alcance de la fila 2>
- **Plazos**: <alcance de la fila 3>
- **Territorio**: <alcance de la fila 4>
<!-- fin de lo que cubre la base -->
```

Una línea por fila de `cubierto`, en su orden: `- **` + la capacidad con su primera letra en mayúscula + `**: ` + el
alcance. `materia` no se escribe. Dos filas de una capacidad dan dos líneas (la vertical de ensayo, §6). La sincronía la
regenera justo después de las regiones de los paquetes y antes de medir el cuerpo; `partirElCuerpo` no parte sus líneas
(research D16).

Defectos de la skill que lleva las marcas (`DefectoDeSkill.Error()`, «`<skill>: <defecto>`»); con cualquiera, la región
queda como está y la skill no lleva nada regenerado:

| Caso | Defecto |
|---|---|
| sin `data/cobertura.yaml` | `SKILL.md: la región lo que cubre la base sin data/cobertura.yaml`, la forma del de hoy de una región de una referencia sin su YAML |
| `data/cobertura.yaml` con defectos | los de su lector, con su fichero delante (`data/cobertura.yaml: cubierto/0, línea 12: …`), cada uno una vez en la skill aunque lo dé también otra región |
| una `materia` que no es el `vertical` de ninguna norma | `data/cobertura.yaml: la fila de lo cubierto de la capacidad «<capacidad>» lleva la materia «<materia>», que no es el vertical de ninguna norma de data/normas.yaml`, una por fila, en su orden |
| marcas que no van en pareja | el de hoy de `regionDeMarcas.buscar`, con el nombre de la región: `SKILL.md: las marcas de lo que cubre la base no van en pareja: <falla>` |

`data/normas.yaml` lo lee el mismo lector perezoso que usan las regiones de las referencias, y lo que este dé con él es
lo de hoy: ningún caso nuevo.

## 4. El motivo fijo en la prosa

En cada skill con plantilla (§5), con el `SKILL.md` regenerado: si el texto, con cada tramo de blancos —saltos de línea
incluidos— reducido a un espacio, no contiene `«` + el motivo de `data/cobertura.yaml` (reducido igual) + `»`:

`<skill>: SKILL.md: la prosa no dice el motivo de lo no declarado, «fuera de lo que kitlegal cubre»`

Sin `data/cobertura.yaml`: `<skill>: SKILL.md: la plantilla ⚠ NO CUBIERTO: sin data/cobertura.yaml`; con defectos, los
del fichero (§3). En el repositorio lo dicen §4 de `consulta-juridica` (partido en dos líneas: «fuera» al final de una y
«de lo que kitlegal cubre» al principio de la siguiente) y «Cómo se presenta un plazo» de `plazos-y-territorio`.

## 5. El generador de `cobertura_por_materia` (`internal/skills/regiones.go`)

- **La plantilla** de una skill: la única línea de los bloques `text` de su `SKILL.md` que, sin sus blancos del
  principio, empieza por `⚠ NO CUBIERTO: ` y lleva `<motivo>`. Con otra cantidad, la skill no tiene plantilla.
- **Los motivos** de una referencia con la región: las líneas de sus bloques `text`, sin sus blancos del principio, que
  llevan `<precepto>` y no empiezan por `⚠`; una con `<norma>` (la de la fila con norma) y otra sin ella.
- **Cada fila**, una por entrada de `regulados_aparte` cuyo precepto es de una norma con el `vertical` del paquete de la
  referencia, en su orden: `| <procedimiento> | <línea> |`, con la línea según data-model §2. Con los datos del
  repositorio, la de los tributos es, byte a byte, la del spec (historia 3, escenario 1).
- **Defectos**, cada uno con `<skill>: references/<ref>.md: ` delante; con cualquiera, la referencia no se regenera:

| Caso | Defecto |
|---|---|
| región en una skill sin plantilla | `región cobertura_por_materia en una skill sin la plantilla ⚠ NO CUBIERTO: en SKILL.md` |
| región en una referencia sin un motivo de cada | `región cobertura_por_materia en una referencia sin sus dos motivos` |
| un marcador desconocido en la plantilla | `la plantilla ⚠ NO CUBIERTO: de SKILL.md lleva <marcador>, un marcador que la tabla no conoce` |
| un marcador desconocido en un motivo | `un motivo de la plantilla lleva <marcador>, un marcador que la tabla no conoce` |

  Los marcadores desconocidos, cada uno una vez, en el orden en que aparecen: primero los de la plantilla, después los de
  los motivos. Con la región en una skill sin plantilla, no se buscan motivos.
- **Las filas sin región** (`filasSinRegion`, sin cambios): su defecto de hoy, en cada skill con plantilla, que en el
  repositorio son `consulta-juridica` y `plazos-y-territorio`.

`docs/SKILLS.md` (tabla de los defectos de la región) cambia «las dos formas» por «la plantilla» en sus filas y gana la
de los motivos.

## 6. La vertical de ensayo (`TestVerticalDeEnsayo`)

- `internal/skills/testdata/vertical-de-ensayo/data/cobertura.yaml` gana, delante de `regulados_aparte`:

  ```yaml
  cubierto:
    - capacidad: plazos
      alcance: los de ensayo del molde de los paquetes.
      materia: ensayo
  ```

- `internal/skills/testdata/vertical-de-ensayo/skills/plazos-y-territorio/references/ensayo.md` gana, delante de su
  región, la frase y el bloque `text` con los dos motivos de FR-005, con los mismos bytes que `administrativo.md`.
- El test: `anadirEntradas` inserta las entradas de cada clave al final del bloque de esa clave en la copia;
  `partirPorLasMarcas` conoce las marcas de §3; `comprobarLoSincronizado` exige en la región de `consulta-juridica` la
  línea `- **Plazos**: los de ensayo del molde de los paquetes.`, la última, detrás de la del territorio. La
  vertical entra sin un `.go` ni un cambio en un `SKILL.md` fuera de las marcas, y el cuerpo de `plazos-y-territorio`
  con su enlace mide 10 756 caracteres.

## 7. Uso, de fuera adentro

| Salida | Quién la lee y cuándo | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| Región de lo cubierto | el agente, al cargar `consulta-juridica`: una vez por sesión, en su paso 3 | 4 líneas, 755 caracteres, 763 bytes (la de los plazos, 311; la del territorio, 117; research V9). Crece una línea por fila que traiga un paquete; con meses de uso, cientos de normas y miles de bloques consultados viven en la caché y el grafo, no aquí | no es una señal: cambia con los datos, en una release o con un paquete |
| Tabla `cobertura_por_materia` | el agente de `plazos-y-territorio`, en su paso 7, una vez por pregunta de plazo de un asunto del paquete, y copia una fila | 9 filas, 3128 bytes (hoy 3047); la de los tributos, 429. Crece con las filas de `regulados_aparte` de los paquetes, no con el uso | una fila deja de darse cuando su materia la cubre un paquete y la entrada sale de `regulados_aparte` |
| Defectos de §2 a §5 | quien mantiene kitlegal, cuando `make skills-sync` o `make ci` salen en rojo, una vez por ejecución | una línea por defecto, de 60 a 220 bytes; con los datos del repositorio, ninguno | cuando los datos y las skills vuelven a ser lo declarado |
