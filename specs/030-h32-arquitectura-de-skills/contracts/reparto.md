# Contrato · El reparto de las skills (FR-001 a FR-009, FR-019, FR-020, FR-034)

Lo que entrega el hito en `skills/` y de dónde sale cada cosa de `c991808`. Los textos completos de los cuatro
cuerpos están en `borradores/` (`<nombre>.SKILL.md`), medidos en esta sesión; sus pasajes cambiados, con el de
`c991808` al lado, en [textos.md](textos.md).

## 1. Las skills de la entrega

| Skill | Clase | Capacidad | `kitlegal-applets` | `kitlegal-referencias` (generadas enteras) | Sale de `c991808` |
|---|---|---|---|---|---|
| `consulta-juridica` | entrada | ninguna: método, capacidades y cobertura | — | — | nueva |
| `normativa` | capacidad | normativa | `boe graph` | `normas` | `boe-legislacion` (renombrada) |
| `jurisprudencia` | capacidad | jurisprudencia | `cita` | — | `jurisprudencia` (mismo nombre) |
| `plazos-y-territorio` | capacidad | plazos y territorio | `territorio plazos` | `jerarquia_normativa` | `legal-core` (renombrada), sin lo administrativo |

Cada applet de capacidad (`boe`, `graph`, `cita`, `plazos`, `territorio`) está en el `kitlegal-applets` de exactamente
una skill de capacidad, y la de entrada no declara ninguno (FR-005). Plazos y territorio van juntos (research D2):
todo plazo con festivos locales resuelve antes sus municipios con `territorio`, y separarlos haría cargar dos skills en
cada pregunta de plazos («Design coherent units»).

**Nombres** (FR-007; research D1): la capacidad, en español, en minúsculas y con guiones, como sintagma nominal; la de
entrada, por lo que hace, la consulta. Ninguno contiene `boe`, `cendoj`, `eur`, `anthropic` ni `claude`.

**Nombres retirados** (FR-025), con las skills que heredan sus preguntas:

| Nombre retirado | Lo heredan |
|---|---|
| `boe-legislacion` | `normativa` |
| `legal-core` | `plazos-y-territorio` |

`jurisprudencia` no es un nombre retirado: la entrega la empotra.

## 2. Qué cubre la base y dónde queda cada cosa de `legal-core` (FR-019, FR-020)

| Lo de `legal-core` en `c991808` | Dónde queda en la entrega |
|---|---|
| `references/cobertura_por_materia.md` (generada de `data/cobertura.yaml`) | región `cobertura_por_materia` de `plazos-y-territorio/references/administrativo.md`, con la línea `⚠ NO CUBIERTO:` de cada fila (FR-012) |
| `references/recursos_y_plazos.md` (generada de `data/plazos.yaml`) | región `recursos_y_plazos` del mismo fichero |
| `references/leyes_vertebrales.md` (generada de `data/normas.yaml`) | región `leyes_vertebrales` del mismo fichero |
| `references/calculo_de_plazos.md`: «Lo que la ley regula aparte», «El acto presunto», «El silencio», «La reclamación del art. 24.2 LTAIBG» | prosa de `administrativo.md`: lo que solo pide el procedimiento común |
| `references/calculo_de_plazos.md`: «Los municipios que cuentan», «El plazo sin calcular», «La lectura en la fecha de inicio» | siguen en `plazos-y-territorio/references/calculo_de_plazos.md`: valen para cualquier plazo que cuente `plazos` |
| `references/jerarquia_normativa.md` (generada de `data/jerarquia.yaml`) | sigue en `plazos-y-territorio`, generada entera: la jerarquía vale para cualquier materia (FR-020) |
| `references/municipios.md` | sigue en `plazos-y-territorio` |
| el método («qué norma prevalece», qué regula cada nivel) | paso 5 de `plazos-y-territorio`, con «las leyes del paquete»; la clasificación de la pregunta, en `consulta-juridica` |
| la descripción de «derecho público» | `consulta-juridica` habla de cualquier materia; `plazos-y-territorio` dice qué plazos cuenta hoy (FR-004, FR-008) |

## 3. La skill de entrada (FR-001 a FR-004)

Texto completo en `borradores/consulta-juridica.SKILL.md` (cuerpo: 3 660 caracteres; descripción: unos 705). Cada punto
de FR-001, con el apartado que lo lleva:

| Punto | Apartado del cuerpo | Lo que dice |
|---|---|---|
| (1) qué tipo de pregunta es | «1. Qué tipo de pregunta es» | lo que dice una norma, una sentencia, hasta cuándo, el territorio, qué hacer |
| (2) qué capacidad y qué skill | «2. Qué capacidad la resuelve» | la tabla de cuatro filas: lo que pide → capacidad → skill |
| (3) qué cubre la base, por capacidad y por paquete | «3. Qué cubre la base» | una viñeta por capacidad y la lista de paquetes, generada entre `<!-- paquetes de la base: lo genera make skills-sync -->` y `<!-- fin de los paquetes de la base -->` ([paquetes.md](paquetes.md) §4) |
| (4) lo que no cubre, en prosa | «4. Lo que no cubre, en prosa» | sin forma fija nueva; solo la norma leída, citada `<forma legible> [<identificador>, bloque <id>]`; ningún plazo, órgano ni vía de memoria, ni los de otra materia, ni `kitlegal plazos calcular` para otra materia, con su porqué |
| FR-002 documentos y destinatarios | viñeta «Sin herramienta, no cubiertos» | escritos, recursos, reclamaciones, solicitudes; a quién se reclama, qué órgano resuelve, ante quién se presenta |
| FR-003 ninguna regla de capacidad | todo el cuerpo | no lleva ninguna línea `⚠` en un bloque `text`, ni protocolo de lectura, cálculo o cotejo: remite a la skill de cada capacidad |
| FR-004 cualquier materia | descripción | «de cualquier materia —un seguro, un contrato, una multa…—»; ninguna descripción de la entrega dice «derecho público» |

`⚠ NO CUBIERTO:` sigue siendo la de lo que la ley regula aparte (FR-011): la entrada lo dice en prosa («lo declara
`plazos-y-territorio` con su línea») y no lleva la línea.

## 4. Las invariantes de cada skill de capacidad (FR-006)

La frase que lleva cada una en su cuerpo (`borradores/`). «=» es que la frase es la de `c991808`.

| Invariante | `normativa` | `jurisprudencia` | `plazos-y-territorio` |
|---|---|---|---|
| no inventar contenido legal | regla 2, «Nunca inventar contenido legal» (=) | regla 1 (=) | regla 1 (sin `boe-legislacion`) |
| cada afirmación con su cita | paso 4, «Cada afirmación sobre una norma, con su cita de lo que `kitlegal boe` devolvió aquí» | paso 4 «Citar» y la introducción, «solo se cita la sentencia cuyo documento está en la conversación» (=) | regla 2, «Cada afirmación sobre una norma, con su identificador… con su cita» |
| distinguir ley y reglamento | paso 4, «Ley y reglamento» (=) | regla 4 (=) | regla 3 (=) |
| señalar la variación autonómica | regla 5, «si puede variar por normativa autonómica o foral, dilo» | regla 4 (=) | regla 4 (=) |
| declarar no cubierto lo no declarado, sin plazos, órganos ni vías de memoria | regla 2, «Lo que lo leído no da —un plazo, un órgano, una vía— no está cubierto: dilo, sin suplirlo de memoria» | regla 4, «Ningún artículo, plazo, órgano ni vía de memoria: aquí no están cubiertos», y el paso 7 | regla 9, «Un plazo, un órgano o una vía de otra materia no están cubiertos: dilo, sin darlos de memoria» |
| no aplicar el procedimiento común a lo que la ley regula aparte | regla 2, «ni con el procedimiento común» | regla 4 (=) | regla 9 (=) y paso 7 |
| no concluir que algo no existe | regla 1 (=) | regla 2, «Ni “existe” ni “no existe”» (=) | regla 5 (=) |
| no actuar en nombre de nadie | regla 4 (más corta) | regla 5 (=) | regla 6 (=) |
| sin herramienta ni binario, lo dice | regla 7, con `⚠ SIN CONSULTA AL BOE:` (=) | regla 6 (=) | regla 7, con `⚠ SIN CONSULTA AL BOE:` |

Ningún cuerpo ni referencia de una skill de capacidad remite a otra skill de la entrega (FR-006): las cuatro
remisiones a `boe-legislacion` de `legal-core` y su referencia `calculo_de_plazos.md`, y la de `jurisprudencia`
(regla 4), se sustituyen por la orden `kitlegal boe …`, que es un applet y no una skill, o se quitan
([textos.md](textos.md) §3 y §4).

## 5. Las formas fijas de cada skill, en un bloque `text` de su `SKILL.md` (FR-006, FR-013, FR-018)

| Forma | `normativa` | `jurisprudencia` | `plazos-y-territorio` |
|---|---|---|---|
| `⚠ SIN CONSULTA AL BOE:` | sí | — | sí, la misma línea |
| `⚠ REDACCIÓN A FECHA:` · `⚠ REDACCIÓN DE HOY:` | sí | — | sí, las mismas líneas (con `<id>`, no `<bloque>`) |
| `⚠ SIN REDACCIÓN VIGENTE:` · `⚠ REDACCIÓN MODIFICADA:` | sí | — | — |
| avisos de la regla 3 (`⚠ NORMA DEROGADA:`…) | en código en línea, como hoy | — | — |
| `⚠ SENTENCIA NO COMPROBADA:` | — | sí | — |
| `⚠ PLAZO:` · `⚠ PLAZO SIN CALCULAR:` · `⚠ NO CUBIERTO:` (dos) · los cuatro avisos de `plazos` | — | — | sí |

`consulta-juridica` no lleva ninguna. Las formas compartidas son, línea a línea, las mismas: lo comprueba el control
`formas-compartidas` ([forma-de-cita.md](forma-de-cita.md) §2).

Lo que la suite de aceptación busca (FR-006, `capacidades.txtar`): por cada fila «sí» de la tabla, una línea de un
bloque `text` del `SKILL.md` de esa skill que, sin los blancos del principio, empieza por la etiqueta. Las etiquetas de
aviso de la regla 3 de `normativa` no son líneas enteras —la frase la da el binario— y siguen como hoy en código en
línea (FR-018): las comprueba `avisos-de-la-skill` en el `SKILL.md`, no la suite.

## 6. Cada eval de hoy: dónde va, qué skills activa su pregunta y qué formas da (FR-005, FR-033, FR-034)

Cada fichero pasa con `git mv` a la carpeta de la skill que hereda su pregunta, con el mismo nombre. «Se activan»: las
skills que la pregunta pide; la de entrada se activa además en toda pregunta legal (no la mide ninguna eval de
capacidad), y la eval solo decide sobre la skill de su carpeta y sobre las de `no_se_activan`. «Formas»: las líneas
`⚠` que la eval comprueba, de la skill que las da.

| Eval de `c991808` | En la entrega | Se activan | Formas (skill) |
|---|---|---|---|
| `boe-legislacion/01-lpac-articulo-21` a `10-et-vacaciones` | `normativa/` (mismos nombres) | `consulta-juridica`, `normativa` | ninguna línea `⚠`: la cita `[<identificador>, bloque <id>]` (`normativa`) |
| `boe-legislacion/11-no-activa-programacion` | `normativa/` | ninguna de capacidad | — |
| `boe-legislacion/12-no-activa-acuerdo-entre-amigos` | `normativa/` | ninguna de capacidad | — |
| `boe-legislacion/13` a `17` (por materia) | `normativa/` | `consulta-juridica`, `normativa` | la cita (`normativa`) |
| `boe-legislacion/18-lrjpac-norma-derogada` | `normativa/` | `consulta-juridica`, `normativa` | `⚠ NORMA DEROGADA:` (`normativa`) |
| `boe-legislacion/19` y `20` (redacción cambiada) | `normativa/` | `consulta-juridica`, `normativa` | `⚠ REDACCIÓN MODIFICADA:` (`normativa`) |
| `boe-legislacion/21-sin-binario-ni-servidor` | `normativa/` | `consulta-juridica`, `normativa` | `⚠ SIN CONSULTA AL BOE:` (`normativa`) |
| `boe-legislacion/22` y `23` (en una fecha) | `normativa/` | `consulta-juridica`, `normativa` | las de la redacción en una fecha que pide su `redaccion_fechada` (`normativa`) |
| `boe-legislacion/expresiones-prohibidas.yaml` y `juez/` | `normativa/` | — | — |
| `jurisprudencia/01` a `10` | sin cambios | `consulta-juridica`, `jurisprudencia` | `⚠ SENTENCIA NO COMPROBADA:` donde lo pide su `sentencias` (`jurisprudencia`) |
| `legal-core/01` y `02` (territorio) | `plazos-y-territorio/` | `consulta-juridica`, `plazos-y-territorio`; no `normativa` (`no_se_activan: [normativa]`) | la cobertura en su bloque `text` (`plazos-y-territorio`) |
| `legal-core/03-no-activa-receta-de-cocina` | `plazos-y-territorio/` | ninguna (`no_se_activan: [normativa]`) | — |
| `legal-core/04-sin-binario-ni-servidor` | `plazos-y-territorio/` | `consulta-juridica`, `plazos-y-territorio`; no `normativa` | `⚠ SIN CONSULTA AL BOE:` (`plazos-y-territorio`) |
| `legal-core/05`, `06`, `09`, `11` (plazo) | `plazos-y-territorio/` | `consulta-juridica`, `plazos-y-territorio` (lee los preceptos con `kitlegal boe articulo --fecha`; `normativa` puede activarse) | `⚠ PLAZO:`, `⚠ REDACCIÓN A FECHA:` y los avisos de `plazos` que pida (`plazos-y-territorio`) |
| `legal-core/07`, `08`, `12` (cobertura) | `plazos-y-territorio/` | `consulta-juridica`, `plazos-y-territorio` | `⚠ NO CUBIERTO:`, la línea de su fila (`plazos-y-territorio`) |
| `legal-core/10-plazo-reposicion-anterior-a-la-lpac` | `plazos-y-territorio/` | `consulta-juridica`, `plazos-y-territorio` | `⚠ PLAZO SIN CALCULAR:` (`plazos-y-territorio`) |

Lo que cambia dentro de los ficheros (FR-033): `no_se_activan: [boe-legislacion]` → `[normativa]` en
`plazos-y-territorio/01` a `04`; en `07`, `08` y `12`, `no_cubierto` gana `precepto` (FR-014; [evals.md](evals.md) §3);
y el comentario de cabecera de `expresiones-prohibidas.yaml`. Nada más.

Las dos evals nuevas de `consulta-juridica` ([evals.md](evals.md) §2):

| Eval nueva | Se activan | Formas |
|---|---|---|
| `consulta-juridica/01-poliza-tras-el-fallecimiento` (FR-035) | `consulta-juridica`, `normativa`; no `jurisprudencia` | ninguna: la cita `[BOE-A-1980-22501, bloque aveintidos]` y la prosa de lo no cubierto; prohibida `plazos calcular` |
| `consulta-juridica/02-una-sola-capacidad` (FR-036) | `consulta-juridica` y `normativa` (`se_activan`); no `jurisprudencia` ni `plazos-y-territorio` | la cita `[BOE-A-2015-11430, bloque a38]` |
