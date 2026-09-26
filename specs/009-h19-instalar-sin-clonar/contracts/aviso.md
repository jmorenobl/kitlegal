# Contrato: el aviso de versión, sin red

Requisitos: FR-070 a FR-077; decisión en [../research.md](../research.md) D5 y D31.

## 1. Qué invocaciones lo buscan

Exactamente las que resuelven un applet registrado **distinto de `skills`** —por el primer argumento o por el nombre
de invocación— y lo ejecutan con un verbo, nombrado o por omisión, **una vez analizado** el verbo, termine como
termine: con éxito, con cualquier código, con el exit 2 de un error de argumentos que detecta el propio verbo (`kitlegal
boe articulo` sin norma), con `--dry-run` y con `--describe`.

No lo buscan: `version`; toda ayuda (`kitlegal --help`, `kitlegal boe --help`, `kitlegal boe articulo --help`); ni los
fallos anteriores a tener el applet con su verbo (sin applet, applet desconocido, argumentos tras `version`, applet sin
verbo nombrado ni por omisión). Tampoco `kitlegal skills …` (FR-074).

## 2. Qué manifiesto mira

Sin seguir enlaces en ninguna ruta de este apartado (FR-028):

1. `./.agents` en el directorio de trabajo: si no existe, paso 4. Si existe y no es un directorio real, **fin sin
   efecto**.
2. `./.agents/skills`: si no existe, paso 4. Si existe y no es un directorio real, **fin sin efecto**.
3. `./.agents/skills/kitlegal.json`: si no existe, paso 4. Si existe, **es el manifiesto** (y el global no se mira);
   si es ilegible ([manifiesto.md](./manifiesto.md) §3), **fin sin efecto**.
4. Si `HOME` no está definido o está vacío, **fin sin efecto** (nunca `/.agents/…` ni una ruta relativa). Si no, los
   pasos 1-3 sobre `$HOME/.agents`, `$HOME/.agents/skills` y `$HOME/.agents/skills/kitlegal.json`; si no hay manifiesto
   ahí, fin sin efecto.

## 3. Cuándo avisa

La forma de la versión del binario se comprueba **antes** que el §2: un binario de desarrollo no examina el disco en
absoluto. Avisa solo si la versión del binario tiene forma SemVer 2.0.0 (con o sin `v`) y, en el manifiesto encontrado,
la `version` de nivel superior **o** la de alguna skill declarada **que el binario empotra** es distinta de la del
binario con la regla de igualdad: quitar una `v` inicial a cada una y comparar byte a byte (FR-077). La versión de una skill no
empotrada no se compara (FR-036).

## 4. La línea

Exactamente una línea en la salida de error, por `Presentador.Aviso`:

```text
aviso: las skills instaladas son de kitlegal <instalada> y este binario es kitlegal <binario>; ejecuta: kitlegal skills install
```

con ` -g` al final cuando el manifiesto es el global. `<instalada>` es la `version` del manifiesto si difiere; si no,
la de la primera skill declarada y empotrada, por orden de nombre, que difiere. Las versiones del manifiesto no llevan
caracteres de control (manifiesto.md §1), así que la línea es siempre una.

## 5. Lo que nunca hace

Nunca escribe en la salida estándar, nunca cambia el código de salida ni la salida estándar (byte a byte), nunca abre
red (FR-072, FR-076), y si escribir la línea falla, el fallo no se propaga (research D5). Un binario de desarrollo —sin
forma SemVer: `dev`, un hash de `git describe`, la cadena vacía— no compara ni avisa (FR-073).
