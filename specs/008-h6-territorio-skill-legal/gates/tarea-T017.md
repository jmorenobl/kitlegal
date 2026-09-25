# T017 — marcada en el intento 2 tras redelimitarla en el intento 1

**Estado**: `[X]` (intento 2, 2026-09-22). El intento 2 heredó el árbol del intento 1 sin tocarlo y aplicó solo el
punto 1 de abajo en `probarNormasNombradas`: rojo antes (tres fallos donde el caso esperaba cinco, con la LEC y la
LOPDGDD ya en la tabla) y verde después. Antes de marcarla se volvió a comprobar, con `jq` sobre las siete búsquedas
grabadas (`.respuesta.cuerpo | fromjson`), que los siete identificadores, rangos y títulos de `data/normas.yaml`
coinciden byte a byte con lo grabado, y que las quince marcas `vertebral: true` son las quince filas de la tabla de
refs §1.4 y ninguna más. `make skills-sync` terminó con código 0 sin cambiar ningún fichero, y `make ci` en primer
plano terminó con «ci: todos los controles en verde».

## Intento 1: sin marcar, redelimitada por un caso negativo fuera de sus rutas

La implementación quedó entera en el árbol, dentro de las rutas declaradas, y el intento 2 la heredó.
`make ci` quedó en rojo por **un solo** subtest, `TestSkillsDelRepositorio/normas-nombradas`, en
`internal/app/skills_test.go`, que la línea de T017 no declaraba. La línea quedó redelimitada en `tasks.md` con esa
ruta; el arreglo se comprobó en verde sobre una copia desechable (abajo).

## Qué hizo el intento 2

1. Aplicar en `internal/app/skills_test.go`, dentro de `probarNormasNombradas` y **nada más**, este cambio:
   - en el texto que se añade al `SKILL.md` de la copia, `Ley 1/2000` → `Ley 0/2000`, `Ley Orgánica 3/2018` →
     `Ley Orgánica 0/2018`, `Real Decreto 1098/2001` → `Real Decreto 0/2001`, `Real Decreto-ley 8/2020` →
     `Real Decreto-ley 0/2020` y `Real Decreto Legislativo 7/2015` → `Real Decreto Legislativo 0/2015`; el
     `Real Decreto Legislativo 2/2004`, la norma que sí está en la tabla y no falla, se queda;
   - las mismas cinco sustituciones en los cinco `fmt.Sprintf(sinEntrada, …)` esperados;
   - al comentario de la función, tras «y una que sí tiene, que no falla.», añadir: «Las cinco llevan el número 0,
     que no lleva ninguna norma: la tabla crece con cada hito —en H6, con la LEC 1/2000 y la LOPDGDD 3/2018, que
     eran dos de estos ejemplos— y un ejemplo que pudiera entrar en ella dejaría el caso sin fallo.»
2. `rtk proxy make skills-sync` (debe terminar con código 0 y sin cambiar ningún fichero: `normas.md` ya está
   regenerado) y `make ci` en primer plano; con «ci: todos los controles en verde», marcar T017 `[X]` en el mismo
   turno.

No hace falta tocar nada más: el resto de la tarea ya está hecho y verificado.

## Por qué no quedó en verde

`probarNormasNombradas` fija que toda norma que nombra un `SKILL.md` empiece el título de una norma de
`data/normas.yaml`, y su caso negativo añade a una copia cinco normas «que la tabla no tiene». Dos de ellas eran
justo la **LEC** (Ley 1/2000) y la **LOPDGDD** (Ley Orgánica 3/2018), dos de las siete leyes vertebrales que esta
tarea tiene que añadir (FR-070). Con ellas en la tabla, esas dos ya no fallan y el caso ve tres fallos donde espera
cinco:

```
--- FAIL: TestSkillsDelRepositorio/normas-nombradas
    expected: […«Ley 1/2000»…, …«Ley Orgánica 3/2018»…, …«Real Decreto 1098/2001»…, …«Real Decreto-ley 8/2020»…, …«Real Decreto Legislativo 7/2015»…]
    actual  : […«Real Decreto 1098/2001»…, …«Real Decreto-ley 8/2020»…, …«Real Decreto Legislativo 7/2015»…]
```

No hay forma legítima de dejarlo en verde dentro de las rutas: los títulos se copian de las grabaciones (no se pueden
escribir de otro modo para esquivar el patrón) y las siete normas son obligatorias. El guardián de diff corre
**antes** de la verificación y rechaza cualquier fichero fuera de las rutas congeladas en `tarea-actual.json`, así que
el arreglo no puede ir en este intento. Es la misma clase que D28 y que lo que T015 encontró en
`TestLeerNormas/rango-no-admitido`: un caso negativo que usa como «ausente» un valor que el cambio de datos introduce.

**Por qué el número 0 y no otros ejemplos reales**: cualquier norma real puede entrar en la tabla en un hito futuro
(el Real Decreto 1098/2001, reglamento de contratos, es candidato claro en cuanto llegue PLACSP o una vertical) y
repetiría el fallo. Ninguna norma del BOE lleva el número 0, así que el caso queda fijo mientras la tabla crece. Se
cambian las cinco, no solo las dos rotas, por esa misma razón.

## Lo que ya está en el árbol (dentro de las rutas)

| Fichero | Cambio |
|---|---|
| `data/normas.yaml` | las siete normas nuevas y `vertebral: true` en las quince de la tabla de refs §1.4 (ocho existentes + siete nuevas), en ninguna más (LIRPF, ET y LRJPAC sin marca); comentario de cabecera |
| `internal/skills/normas.go` | campo `Vertebral bool` (`yaml:"vertebral"`) |
| `internal/skills/referencias.go` | la **tabla de generadores** (`generadoresDeReferencias`: `normas`, `leyes_vertebrales`, `jerarquia_normativa`), cada fila con su YAML de datos, su título, sus encabezados y su cuerpo; la cabecera «generado desde data/<fichero>, no editar», el título y la fila de encabezados salen de la fila; `RenderizarLeyesVertebrales` y `RenderizarJerarquia` nuevas; `RenderizarNormas` da los mismos bytes que antes |
| `internal/skills/sincronia.go` | el `switch` por nombre desaparece: cada referencia se genera con su fila desde el YAML que la fila declara; los defectos de un YAML de datos van **una sola vez** aunque salgan de él dos referencias |
| `internal/skills/frontmatter.go` | `defectosDeLasReferencias` consulta la tabla: sin fila, «"x" sin generador conocido» (antes lo añadía `sincronia.go`); con fila y sin su YAML, «"x" sin data/<fichero>» con el fichero de la fila; `extensionDeLosDatos` retirada |
| `skills/boe-legislacion/references/normas.md` | regenerado **solo** con `make skills-sync`: siete filas nuevas, nada más (la cabecera y el resto, byte a byte iguales) |
| tests (`*_test.go` de los `.go` declarados) | ver abajo |

### Identificadores, títulos y rangos: de las búsquedas grabadas

Leídos con `jq` de las siete búsquedas grabadas en T015 (`testdata/evals/boe.legislacion-consolidada/…query…`), no de
`refs/` ni de `gates/tarea-T015.md`; coinciden con lo esperado, así que no hubo motivo para detenerse:

| Norma | Identificador | Rango grabado | Título grabado |
|---|---|---|---|
| CC | `BOE-A-1889-4763` | Real Decreto | Real Decreto de 24 de julio de 1889 por el que se publica el Código Civil. |
| LJCA | `BOE-A-1998-16718` | Ley | Ley 29/1998, de 13 de julio, reguladora de la Jurisdicción Contencioso-administrativa. |
| LEC | `BOE-A-2000-323` | Ley | Ley 1/2000, de 7 de enero, de Enjuiciamiento Civil. |
| LOPJ | `BOE-A-1985-12666` | Ley Orgánica | Ley Orgánica 6/1985, de 1 de julio, del Poder Judicial. |
| LGS | `BOE-A-2003-20977` | Ley | Ley 38/2003, de 17 de noviembre, General de Subvenciones. |
| LOPDGDD | `BOE-A-2018-16673` | Ley Orgánica | Ley Orgánica 3/2018, de 5 de diciembre, de Protección de Datos Personales y garantía de los derechos digitales. |
| LGP | `BOE-A-2003-21614` | Ley | Ley 47/2003, de 26 de noviembre, General Presupuestaria. |

`TestIdentificadoresDeLasNormas` las resuelve todas contra el manifiesto. Abreviaturas y materias son redacción
propia (no son datos de la fuente): ninguna abreviatura choca con las cinco de `abreviaturasDelHito` y ninguna materia
nueva es «tributos», así que las reglas del conjunto de `boe-legislacion` no cambian de resultado.

### Forma de las dos referencias nuevas (para T022)

- `leyes_vertebrales.md`: cabecera `data/normas.yaml`, título «Leyes vertebrales», los mismos encabezados, orden y
  escapes que `normas.md`, solo las marcadas.
- `jerarquia_normativa.md`: cabecera `data/jerarquia.yaml`, título «Jerarquía normativa», tabla
  `| Nivel | Boletín | Tipos de norma, de mayor a menor rango |` con una fila por nivel en el orden del documento, y
  la sección «## Reglas de interpretación» con una línea ``- `<código>`: <enunciado>`` por regla.

### Tests

- `referencias_test.go`: `TestRenderizarNormas` sin cambio; `TestRenderizarLeyesVertebrales` (`solo-las-marcadas`,
  `todas-marcadas-dan-las-filas-de-las-normas`, `ninguna-marcada`, `no-cambia-las-normas-que-recibe`);
  `TestRenderizarJerarquia` (`bytes-exactos`, `en-el-orden-en-que-llegan`).
- `normas_test.go`: `TestLeerNormas/con-vertebral` y `/vertebral-que-no-es-booleano`; `TestNormasDelRepositorio`
  gana `vertebrales`, que compara con `ElementsMatch` las marcadas con los quince identificadores de la tabla (no pasa
  en vacío: sin marcas, falla con las quince como sobrantes de la lista esperada).
- `frontmatter_test.go`: `TestLeerFrontmatter/referencias-que-no-se-llaman-como-sus-datos`; `TestValidarFrontmatter`
  con una referencia conocida cuyo YAML no se llama como ella (`leyes_vertebrales` desde `data/normas.yaml`, válida y
  sin su YAML), una sin generador aunque exista un YAML de su nombre (`tributos`) y una conocida cuyo YAML es un
  directorio (`jerarquia_normativa`).
- `sincronia_test.go`: alfa declara las tres referencias —dos del mismo YAML— y el árbol trae `data/jerarquia.yaml`;
  `escribir-sincroniza` con las tres referencias byte a byte; derivas `marca-vertebral-sin-regenerar` (solo cambia
  `leyes_vertebrales.md`) y `jerarquia-sin-regenerar`; defectos `referencia-sin-su-yaml-de-datos`,
  `jerarquia-con-una-clave-desconocida` y `norma-con-vertical`, que ahora demuestra que el defecto del YAML compartido
  va una sola vez.
- Rojo antes del código: el paquete de test no compilaba (`Vertebral`, `RenderizarLeyesVertebrales`,
  `RenderizarJerarquia` no existían); con el código y sin los datos, solo `TestNormasDelRepositorio/vertebrales` en
  rojo; con los datos, los ocho tests nombrados por la tarea en verde (comprobado con `-v`).

## Verificación

- **Árbol real** (`make ci` en primer plano): formato, lint (0 issues), todos los paquetes en verde salvo
  `internal/app` por `TestSkillsDelRepositorio/normas-nombradas`; `internal/skills` al 98,2 %.
- **Copia desechable** (`rsync` del árbol a `/tmp`, con el cambio del punto 1 aplicado a su `skills_test.go`):
  `make ci` → «ci: todos los controles en verde»; `make skills-sync` → código 0 y `normas.md` idéntico (`cmp`) al del
  árbol. La copia ya está retirada.
- **Mutantes** sobre la copia, cada uno restaurado después: sin la deduplicación por YAML, falla
  `TestRegenerarYComparar/norma-con-vertical` (el defecto sale dos veces); `leyes_vertebrales` apuntando a la fila de
  las normas, fallan `escribir-sincroniza`, `datos-sin-regenerar` y `marca-vertebral-sin-regenerar`; `vertebral: true`
  en la LIRPF, falla `TestNormasDelRepositorio/vertebrales` nombrando `BOE-A-2006-20764` como sobrante.
- `misspell` marcó «columnas» suelta (ya conocido desde H4/T007): el campo se llama `encabezados` y los comentarios
  dicen «encabezados» o «cada columna»; sin tocar `.golangci.yml`.

## Para la revisión final del hito

- **plan.md**, inventario de tests: `internal/app/skills_test.go` figura solo en los pasos 20 y 21; tras esta
  redelimitación también lo toca T017 (paso 16), solo en `probarNormasNombradas`. `tasks.md` ya lo dice en su línea y
  en las dos notas que enumeran los ficheros compartidos.
- Lección, gemela de la de T015 con D29: una tarea que **añade datos** a una tabla tiene que buscar los casos
  negativos que usan como «ausente» algo de lo que añade (`grep` de los títulos o rangos nuevos en los `_test.go`) y
  declararlos. Ni el plan ni el juez de tareas lo vieron aquí.
- `make skills-sync` ejecuta `TestSkillsDelRepositorio` entero después de escribir: si un subtest ajeno a la
  generación falla, la orden termina en 2 aunque haya regenerado bien. Aquí regeneró `normas.md` y salió en 2 por
  `normas-nombradas`; no es un defecto de la orden, pero conviene saberlo al leer su código de salida.
