# Contrato: applet `skills` (`install`, `list`, `doctor`)

Lo observable del applet desde la línea de órdenes. El comportamiento en disco (qué es conflicto, qué se escribe, qué
es hallazgo) está en [../data-model.md](../data-model.md) §4 a §7; el manifiesto, en [manifiesto.md](./manifiesto.md);
el aviso de los demás applets, en [aviso.md](./aviso.md). Requisitos: FR-010 a FR-077.

## 1. Invocación

```text
kitlegal skills install [skill…] [-g|--global] [--host claude] [--dir <ruta>]
kitlegal skills list    [-g|--global] [--dir <ruta>]
kitlegal skills doctor  [-g|--global] [--dir <ruta>]
```

- Tres verbos, **ninguno por omisión**: `kitlegal skills` sin verbo sale con 2 y la lista de verbos (el mismo
  comportamiento que `territorio`).
- Las ocho banderas globales se heredan del kernel (`--json`, `--timeout`, `--offline`, `--dry-run`, `--describe`,
  `--no-graph`, `--asunto`, `--verbose`). `--offline` no cambia nada: el applet nunca abre red (FR-016).
- Banderas propias: `-g`/`--global` (booleana), `--host <valor>` (solo `install`), `--dir <ruta>` (cadena tal cual; no
  se expande `~` ni se hace absoluta).
- `skill…`: argumentos posicionales opcionales de `install`. Sin ninguno, todas las skills empotradas.

## 2. Validación de la invocación (antes de tocar el disco)

En este orden; la primera que falla decide, y **ninguna lee ni escribe nada del disco**:

| # | Caso | Verbos | Código | Mensaje (contiene) |
|---|---|---|---|---|
| 1 | `-g` y `--dir` a la vez (con o sin `--host`, también con `--dry-run`) | los tres | 2 | `-g y --dir se excluyen` |
| 2 | `--host` junto a `--dir` | `install` | 2 | `--host no se combina con --dir` |
| 3 | `--host` con un valor distinto de `claude` | `install` | 2 | `el único host admitido es claude` |
| 4 | un nombre que no es de una skill empotrada (también una declarada en el manifiesto y no empotrada, FR-036) | `install` | 2 | `no es ninguna skill de este binario; skills disponibles: boe-legislacion, legal-core` |
| 5 | `-g` con `HOME` sin definir o vacío | los tres | 1 | `HOME no está definido o está vacío` |

Un nombre repetido cuenta una vez. Los errores 1-4 declaran la clase `argumentos`; el 5 declara la clase
`inesperado` (FR-012 fija el 1). Este orden es la precedencia de FR-052: un error de
argumentos gana con 2 aunque la invocación caiga además en el 5 o en cualquier exit 1 del ámbito (conflicto, hallazgo,
manifiesto ilegible, ruta que no es directorio, manifiesto con entradas de host), que no llegan a comprobarse.

## 3. Ámbito y rutas que se presentan

| Ámbito | Raíz | Directorio neutro | Manifiesto | Hosts | Ruta presentada de `<skill>` |
|---|---|---|---|---|---|
| local (por omisión) | el directorio de trabajo | `.agents/skills` | `.agents/skills/kitlegal.json` | `claude`, en `.claude/skills/<skill>` | `.agents/skills/<skill>` |
| global (`-g`) | `$HOME` | `$HOME/.agents/skills` | `$HOME/.agents/skills/kitlegal.json` | `claude`, en `$HOME/.claude/skills/<skill>` | `<HOME>/.agents/skills/<skill>` (absoluta; `HOME` se usa tal cual, limpio) |
| `--dir <ruta>` | — | `<ruta>` | `<ruta>/kitlegal.json` | ninguno | `<ruta>/<skill>` (relativa o absoluta, como se pasó) |

Toda ruta se escribe con `/`, limpia (`path.Clean`) y sin barra final; con `--dir` relativo, cuelga de la ruta tal como
se pasó. El ámbito local no busca ninguna raíz de proyecto ni sube de directorio (FR-011).

## 4. Salida correcta (el sobre)

Procedencia de applet calculado (FR-050): `"fuente": "kitlegal.skills"`, `"url": "kitlegal:applet/skills"`,
`fecha_consulta` del reloj del kernel. El `data` de cada verbo:

### 4.1 `install` (exit 0) — FR-051

Lista de las skills pedidas, en orden de nombre:

```json
[
  {
    "nombre": "boe-legislacion",
    "ruta": ".agents/skills/boe-legislacion",
    "estado": "instalada",
    "enlaces": [{"host": "claude", "ruta": ".claude/skills/boe-legislacion", "modo": "enlace"}]
  },
  {"nombre": "legal-core", "ruta": ".agents/skills/legal-core", "estado": "sin cambios", "enlaces": []}
]
```

- `estado`: `instalada` (no estaba declarada), `actualizada` (cambian sus ficheros, sus entradas de host o su entrada del
  manifiesto, versión incluida) o `sin cambios`.
- `enlaces`: las entradas de host de la skill **tras** la orden, con `modo` `enlace` o `copia`; lista vacía sin hosts.
- Una skill declarada y no empotrada no aparece nunca (FR-036).

### 4.2 `list` (exit 0) — FR-060, FR-061

```json
{
  "directorio": ".agents/skills",
  "manifiesto": true,
  "version": "v0.1.0",
  "skills": [
    {"nombre": "boe-legislacion", "ruta": ".agents/skills/boe-legislacion", "version": "v0.1.0", "empotrada": true,
     "enlaces": [{"host": "claude", "ruta": ".claude/skills/boe-legislacion", "modo": "enlace"}]},
    {"nombre": "otra-skill", "ruta": ".agents/skills/otra-skill", "version": "v0.0.9", "empotrada": false, "enlaces": []}
  ]
}
```

Sin manifiesto: `"manifiesto": false`, `"version": null`, `"skills": []`, y nada en la salida de error. `list` no
escribe nada (FR-062).

### 4.3 `doctor` sin hallazgos (exit 0) — FR-067

```json
{"directorio": ".agents/skills", "manifiesto": true, "version": "v0.1.0", "version_del_binario": "v0.1.0", "hallazgos": []}
```

Sin manifiesto: `"manifiesto": false`, `"version": null`, `"hallazgos": []`, y nada en la salida de error. `doctor` no
deja ningún cambio en disco (FR-068).

### 4.4 Esquemas

`schemas/instalacion.json` publica las tres partes (`skills doctor`, `skills install`, `skills list`) desde
`--describe`; `version` admite `null` y las listas, vacías (FR-053). El fallo con exit 1 lo describe la rama `else` del
sobre, común a todos los applets.

## 5. Fallo con exit 1: conflictos, hallazgos y ámbitos que no se pueden leer — FR-052

El sobre de fallo del kernel, sin cambios: con `--json`, `{"ok": false, …, "data": {"clase": "inesperado", "mensaje":
"…"}}`; y el mismo `mensaje` en la salida de error. El `mensaje` es una cabecera y una línea por entrada:

```text
skills install: nada se ha creado ni cambiado; conflictos:
carpeta ajena: .agents/skills/boe-legislacion
enlace roto: .claude/skills/legal-core
```

```text
skills doctor: 2 hallazgos:
fichero editado: .agents/skills/legal-core/SKILL.md: rm -- '.agents/skills/legal-core/SKILL.md' && kitlegal skills install legal-core
versión distinta: .agents/skills/kitlegal.json: kitlegal skills install boe-legislacion legal-core
```

```text
skills list: manifiesto ilegible: .agents/skills/kitlegal.json
```

- Línea de conflicto: `<clase>: <ruta>`. Clases, literales: `carpeta ajena`, `fichero`, `enlace a otro sitio`, `enlace
  roto`, `fichero editado`, `fichero ajeno`, `ruta que no es directorio`, `manifiesto ilegible`, `manifiesto con
  entradas de host`.
- Línea de hallazgo: `<clase>: <ruta>: <orden>`. Clases, literales: `fichero editado`, `enlace colgando`, `enlace a otro
  sitio`, `copia`, `versión distinta`.
- Orden de las líneas: conflictos, por ruta comparada byte a byte y, a igual ruta, por el orden de la lista de clases
  anterior; hallazgos, como fija FR-066 (primero los que llevan `rm`, después los demás; dentro, por ruta y por número
  de clase de FR-065).
- `list` y `doctor` salen con 1 y **una** línea `skills <verbo>: <clase>: <ruta>` ante un manifiesto ilegible (también
  un `kitlegal.json` que es un enlace), una ruta del ámbito que no es directorio (FR-027) o, con `--dir`, un manifiesto
  con entradas de host (FR-013); sin dar skills ni hallazgos.
- Un fallo de escritura tras la comprobación sale con 1 y el mensaje `skills install: <operación> <ruta>: <error del
  sistema>` (FR-044).
- La sonda del creador de enlaces (research D9; data-model §3) se hace en el directorio de host del propio ámbito y se
  retira; si no se puede retirar, `install` (también con `--dry-run`) y `doctor` salen con 1 y el mensaje `skills
  <verbo>: retirar la sonda <ruta>: <error del sistema>`, sin conflictos ni hallazgos.

## 6. Órdenes de `doctor` — FR-066

Una sola línea de shell POSIX por hallazgo:

```text
[rm [-r] -- '<ruta>' && ]kitlegal skills install <skill>… [-g | --dir '<ruta del --dir>'] [--host claude]
```

- `rm -- '<ruta>' && …` para: fichero editado con huella distinta o que ya no es un fichero regular; enlace a otro
  sitio cuando en la ruta hay algo; enlace colgando del directorio neutro o de dentro de una copia. `rm -r` solo si la
  entrada es un directorio real; nunca `-f`.
- Sin `rm`: fichero editado que falta; enlace a otro sitio cuando la entrada de host falta; enlace colgando de una
  entrada de host con el destino literal de FR-021; copia; versión distinta.
- Skills que nombra: la del hallazgo; en «versión distinta» del manifiesto, todas las declaradas y empotradas, por
  orden de nombre (ninguna si no hay). Nunca una no empotrada.
- `--host claude` si y solo si alguna de las skills que nombra tiene una entrada de host declarada. Nunca con `--dir`.
- Comillas simples con la `'` escrita `'\''`. Ruta de `rm`: la del hallazgo, como se alcanza desde el directorio de
  trabajo (§3).
- La ruta de `--dir` va en la palabra siguiente a la bandera, `--dir '<ruta>'`, salvo si empieza por `-`: entonces va
  en la misma palabra, `--dir='<ruta>'`, porque el análisis de la invocación leería `'-raro'` como otra bandera y la
  orden saldría con 2 después de que `rm` hubiera retirado la entrada, sin cumplir la garantía (i) de FR-066.

## 7. `--dry-run` en `install` — FR-048

Sin conflictos: exit 0, **nada en la salida estándar** y, en la salida de error, la línea del kernel seguida de una
por skill pedida (el prefijo `--dry-run: se habría pedido ` lo pone el kernel):

```text
--dry-run: no se ha ejecutado nada; se habría ejecutado el applet "skills", el verbo "install", con los argumentos […]
--dry-run: se habría pedido instalar boe-legislacion en .agents/skills/boe-legislacion: instalada; enlace .claude/skills/boe-legislacion (enlace)
--dry-run: se habría pedido instalar legal-core en .agents/skills/legal-core: sin cambios
```

Con conflictos: la línea del kernel y el mismo fallo, mensaje y exit 1 que sin la bandera. En los dos casos, 0 cambios
en disco (ni directorios vacíos). `list` y `doctor` no cambian nada con o sin la bandera.

El modo que se predice (`enlace` o `copia`) sale de la misma sonda que usa la orden real, hecha en el directorio real
más próximo a `.claude/skills` dentro del ámbito, que es el sistema de ficheros donde se enlazaría; la sonda crea y
retira un enlace de prueba y deja ese directorio con las mismas entradas y los mismos bytes (research D9).

## 8. Códigos de salida

| Código | Cuándo |
|---|---|
| 0 | `install` sin conflictos; `list`; `doctor` sin hallazgos (también sin manifiesto) |
| 1 | conflictos de `install` (también con `--dry-run`); hallazgos de `doctor`; ámbito ilegible en `list`/`doctor`; `-g` sin `HOME`; fallo de escritura; sonda del creador de enlaces que no se pudo retirar |
| 2 | argumentos inválidos (§2) y los del kernel (bandera desconocida, verbo que falta) |

`skills` no emite operaciones de grafo (FR-054) y no decide nunca 3, 4, 5 ni 6 (el 4 del plazo agotado lo pone el
kernel).
