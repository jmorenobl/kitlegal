# Contrato: el manifiesto `kitlegal.json`

Un manifiesto por ámbito, en la raíz de su directorio neutro (FR-030). Es un contrato **interno**: lo lee y lo escribe
este binario, se versiona con el proyecto y no tiene esquema publicado en `schemas/` (spec, *Fuera de alcance*).
Requisitos: FR-030 a FR-036; decisión en [../research.md](../research.md) D10.

## 1. Forma

```json
{
  "skills": {
    "boe-legislacion": {
      "ficheros": {
        "boe-legislacion/SKILL.md": "sha256:…64 hexadecimales…",
        "boe-legislacion/references/normas.md": "sha256:…"
      },
      "hosts": {
        "claude": {"modo": "enlace", "ruta": ".claude/skills/boe-legislacion"}
      },
      "version": "v0.1.0"
    },
    "legal-core": {
      "ficheros": {"legal-core/SKILL.md": "sha256:…", "legal-core/references/jerarquia_normativa.md": "sha256:…",
                   "legal-core/references/leyes_vertebrales.md": "sha256:…"},
      "hosts": {
        "claude": {
          "ficheros": {".claude/skills/legal-core/SKILL.md": "sha256:…", "…": "…"},
          "modo": "copia",
          "ruta": ".claude/skills/legal-core"
        }
      },
      "version": "v0.1.0"
    }
  },
  "version": "v0.1.0"
}
```

| Clave | Tipo | Regla |
|---|---|---|
| `version` | cadena | versión del binario que lo escribió por última vez; no vacía, sin caracteres de control |
| `skills` | objeto | una entrada por skill declarada, por nombre (puede estar vacío) |
| `skills.<n>` | clave | nombre de skill: `a-z`, `0-9` y `-`, sin guion al principio, al final ni dos seguidos, 1-64 caracteres |
| `skills.<n>.version` | cadena | versión del binario que instaló la skill; mismas reglas que `version` |
| `skills.<n>.ficheros` | objeto | ruta relativa al directorio neutro → huella; cada ruta empieza por `<n>/` |
| `skills.<n>.hosts` | objeto | opcional (se omite sin entradas); única clave admitida: `claude` |
| `…hosts.claude.ruta` | cadena | exactamente `.claude/skills/<n>` (relativa a la raíz del ámbito) |
| `…hosts.claude.modo` | cadena | `enlace` o `copia` |
| `…hosts.claude.ficheros` | objeto | solo y obligatorio en `copia`: ruta relativa a la raíz del ámbito, que empieza por `.claude/skills/<n>/` → huella |

- **Rutas**: con `/`, relativas, sin componentes vacíos, `.` ni `..`, e iguales a su forma limpia. Ninguna absoluta
  (FR-032).
- **Huella**: `sha256:` seguido de 64 hexadecimales en minúscula, la de los bytes del fichero (la misma forma que el
  `hash` del sobre).
- Ni fechas, ni usuarios, ni máquinas, ni rutas absolutas (FR-032).

## 2. Escritura

- Serialización canónica: `encoding/json/v2` con `json.Deterministic(true)` (claves de objeto en orden), sangrado de dos
  espacios, sin escapar HTML y con un salto de línea final. Dos instalaciones de lo mismo con el mismo binario y los
  mismos modos dan manifiestos **byte a byte iguales** en cualquier ruta o máquina (FR-032).
- Solo se escribe si su contenido cambia (FR-033), y siempre de forma atómica (temporal en el mismo directorio y
  renombrado; research D8), en la fase 3 del orden de aplicación (data-model §5).
- Al escribir por un subconjunto de skills, las entradas de las demás —también las de skills no empotradas— se
  conservan byte a byte (FR-034, FR-036). La `version` de nivel superior pasa a ser la del binario que escribe.

## 3. Lectura

Un manifiesto es **ilegible** (FR-035) —para `install`, `list`, `doctor` y el aviso— si:

1. `kitlegal.json` existe y no es un fichero regular visto con `Lstat` (un enlace simbólico aunque apunte a un
   manifiesto válido, un directorio, una tubería…): no se abre;
2. no se puede leer;
3. no es JSON, tiene un miembro desconocido (`json.RejectUnknownMembers(true)`), un nombre repetido en cualquier objeto
   (lo rechaza `jsontext` por omisión) o datos tras el documento;
4. incumple cualquier regla de la tabla del §1.

Un manifiesto que no existe no es ilegible: el ámbito no tiene manifiesto.
