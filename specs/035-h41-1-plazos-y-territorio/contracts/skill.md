# Contrato: `skills/plazos-y-territorio` tras H41.1

FR-001 a FR-004, FR-012; SC-002, SC-005. Research V2 a V4, D2, D3; data-model §1 y §2.

## 1. Lo que es

`skills/plazos-y-territorio/` es `laboratorio/plazos-y-territorio/plazos-y-territorio-sin-condicion/skills/plazos-y-territorio/`
con el movimiento M1 y nada más: los mismos cinco ficheros, el `SKILL.md` con un cambio (§2) y
`references/calculo_de_plazos.md` con una línea en blanco y una línea más al final (§3). La región de los paquetes es la
que escribe `make skills-sync` desde `data/`, como hoy. Ninguna variante del laboratorio cambia (FR-006).

Respecto a v0.9.0 (`plazos-y-territorio-control`), el `SKILL.md` lleva los tramos de
`specs/034-h41-las-skills-del/contracts/variantes.md` §2.1, palabra por palabra y con sus saltos de línea —salen de
«Referencias» las viñetas de `municipios.md` y `calculo_de_plazos.md`; el «Protocolo» empieza con el párrafo de
`municipios.md` y una línea en blanco; el paso 7 termina con la orden de `calculo_de_plazos.md`— y el cambio de §2.

## 2. M1 en el `SKILL.md` (FR-003; research D2)

`diff` de `-sin-condicion` a la skill del producto, y nada más:

```diff
68,69c68
<     `⚠ REDACCIÓN A FECHA:`. Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di aparte que no has podido leerlo, sin
<     afirmar qué dice.
---
>     `⚠ REDACCIÓN A FECHA:`.
```

El paso 10 queda así, y el 11 empieza en la línea siguiente:

```text
10. **Leer cada precepto de `data.preceptos` en la fecha de inicio**, con `boe_articulo` y `fecha` o
    `kitlegal boe articulo <identificador> <id> --fecha <inicio> --json`, sin `kitlegal graph check`, y citarlo con
    `⚠ REDACCIÓN A FECHA:`.
11. **Responder**: …
```

## 3. M1 en `references/calculo_de_plazos.md` (FR-003; research D3)

El fichero de v0.9.0 entero (2 760 bytes, 36 líneas) y, detrás, estos bytes, que lo dejan en 2 869:

```text
⏎
Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di aparte que no has podido leerlo, sin afirmar qué dice.⏎
```

(`⏎` es un salto de línea.) Las últimas tres líneas del fichero:

```text
  diría el régimen transitorio de la norma que la cambió, que no has leído. Dilo así.

Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di aparte que no has podido leerlo, sin afirmar qué dice.
```

`administrativo.md`, `jerarquia_normativa.md` y `municipios.md` son byte a byte las de control.

## 4. Lo que queda como en v0.9.0 (FR-002, FR-012)

El frontmatter, con la descripción; la viñeta de `jerarquia_normativa.md` («- [jerarquia_normativa.md](…): para el
nivel que regula algo, con el autonómico de la `comunidad` de `data`, y la norma que prevalece.»), sin «lee»; la región
de los paquetes; «Si no lo dice, **pregúntalo** y espera» en el paso 1; y el resto del cuerpo, salvo lo de §1 y §2.

## 5. Medidas

| Qué | Hoy (v0.9.0) | Tras el hito |
|---|---|---|
| Cuerpo del `SKILL.md`, caracteres (V2) | 10 930 | **10 882** |
| Cuerpo con la línea de la vertical de ensayo de H32 (+70) | 11 000 | **10 952** |
| `SKILL.md`, bytes | 12 161 | 12 113 |
| `calculo_de_plazos.md`, bytes | 2 760 | 2 869 |
| `municipios.md`, bytes | 3 944 | 3 944 |
| Frases del cuerpo con un enlace a una referencia, fuera de los paquetes | 3 | 3 (contracts/controles.md §2.2) |

## 6. Uso, de fuera adentro

- **Quién la lee.** El agente de quien usa el kit, con la skill instalada por `kitlegal skills install`: el cuerpo, una
  vez por pregunta, al activar la skill; `municipios.md`, una vez por pregunta a la que llega la skill, antes de la
  primera orden o llamada (la orden de §1); `calculo_de_plazos.md`, una vez por pregunta que llega al final del paso 7;
  la referencia del paquete y `jerarquia_normativa.md`, como en v0.9.0.
- **Ejemplo, por pregunta.** «¿Qué comunidad y qué boletines tiene Leganés?»: cuerpo + `municipios.md`, 12 113 +
  3 944 = 16 057 bytes, y una orden, `kitlegal territorio resolver Leganés --json`. «¿Hasta cuándo puedo recurrir en
  reposición una resolución de Leganés notificada el 3 de marzo?»: además `calculo_de_plazos.md` (2 869) y
  `administrativo.md` (9 930), 28 856 bytes; las órdenes de v0.9.0 —`territorio resolver` por municipio que cuenta, una
  `plazos calcular` y una `boe articulo --fecha` por precepto de `data.preceptos`—, ni una más ni una menos.
- **Tamaño con meses de uso.** Fijo. Cientos de normas y miles de bloques consultados no tocan ningún fichero de la
  skill: viven en la caché y el grafo de quien la usa, no en la skill. El cuerpo solo crece con los paquetes que declare
  `data/` (una línea de unos 70 caracteres por paquete en la región), y su tope de 11 000 lo hacen cumplir
  `TestSkillsDelRepositorio` y, con la vertical de ensayo, `TestVerticalDeEnsayo`. Las referencias solo cambian cuando
  un hito las cambia.
- **Invocaciones de `kitlegal` por pregunta.** Las de v0.9.0: el hito no añade ni quita ninguna orden ni llamada; solo
  cambia dónde, y sin qué condición, se mandan leer dos referencias.
- **Señales.** Ninguna: ni el cuerpo ni las referencias dependen de la consulta ni caducan. Las únicas señales del hito
  son los defectos de los controles (contracts/controles.md §5), que lee quien cambia la skill o el laboratorio una vez
  por `make ci` o `make skills-check`, una línea por fichero o remisión que difiere, y que dejan de darse cuando lo
  cambiado vuelve a ser lo declarado.
