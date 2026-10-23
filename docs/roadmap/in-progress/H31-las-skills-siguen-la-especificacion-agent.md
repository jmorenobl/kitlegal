---
id: H31
fase: "1"
depende_de: [H30]
---

# H31 · Las skills siguen la especificación Agent Skills: guía de autoría y gates

- **Objetivo**: que `boe-legislacion`, `legal-core` y `jurisprudencia` sean un `SKILL.md` corto que guía y unas
  referencias que el agente carga cuando las necesita, con la descripción en tercera persona, como piden la
  especificación Agent Skills (agentskills.io) y la guía de autoría de Anthropic, sin perder ninguna regla ni
  empeorar la activación ni las respuestas.
- **Entrega**: una propuesta de cambio con:
  1. `docs/SKILLS.md`, la guía de autoría del proyecto sobre las dos fuentes, con su fecha de consulta: qué va en el
     cuerpo y qué en una referencia, la forma de la descripción, los presupuestos con su porqué y la regla de
     `references/` (ADR 0046).
  2. Gates en `make ci` (`skills-check`), que sustituyen a «menos de 300 líneas», y cada uno sale en rojo con un
     mutante:
     - la validación de la especificación, en Go y sin red (sobre `ValidarFrontmatter`): la de hoy y, además,
       ninguna etiqueta XML en `name` ni en `description`, y ni `anthropic` ni `claude` en `name`;
     - el cuerpo de `SKILL.md` —sin el frontmatter y con la tabla de comandos generada— en menos de 500 líneas y como
       mucho 11 000 caracteres;
     - la descripción con como mucho 900 caracteres;
     - `references/` plano, con ficheros `.md`; cada referencia enlazada desde `SKILL.md` por su ruta relativa
       (`references/<fichero>.md`); ninguna enlazada desde otra referencia; y la de más de 100 líneas abre con un
       índice de sus apartados.
     Los caracteres se cuentan como runas.
  3. La descripción en tercera persona la juzga el juez con modelo de las evals (ADR 0037 y 0046), con la clase
     `habla_en_primera_o_segunda_persona`, en el job de evals y nunca en `make ci` ni en un paso del run:
     - la rúbrica, el esquema y los casos son los de `evidencias/adr-0046/`, idénticos byte a byte, con un control
       de identidad en `make ci`; la medida versionada es la de `evidencias/adr-0046/medida.json` (63 de 63 defectos
       marcados y 0 de 36 correctos, con `claude-opus-5-5` y Claude Code 2.1.289), y el job comprueba sin modelo,
       antes de abrir ninguna sesión, que corresponde y se cumple;
     - el mensaje de cada voto es el de `mensaje_de` en `evidencias/adr-0046/guiones/juez.py`; la orden del voto, la
       de `scripts/evals-voto.sh`; la frase citada se comprueba sin modelo; marca con tres votos unánimes, pedidos
       por orden;
     - juzga la descripción de cada skill cuando la elección del ADR 0045 incluye el cambio de una descripción, y
       reutiliza por huella el juicio de una descripción que no ha cambiado;
     - su umbral es 0 y hace fallar el job; el informe da, de cada descripción marcada, la frase y el motivo.
  4. La regla de `references/`: el protocolo se escribe a mano, sin cabecera; lo que sale de `data/` sigue generado,
     con su cabecera «generado desde» y declarado en `kitlegal-referencias`. `make skills-sync` no toca una
     referencia a mano, y una referencia con la cabecera que no es lo que genera su fila es deriva, como hoy. El
     conjunto de dependencias de H30 cuenta la referencia a mano en la parte (b), sin parte (c).
  5. Las tres skills reescritas con esa forma y con la descripción en tercera persona, con su versión subida en
     `CHANGELOG.md`, y `CONTRIBUTING.md` al día.
  El ADR 0046, la constitución 2.15.0 (principio VIII), `CLAUDE.md`, la DoD §1.10 y la fila de §3 de
  `docs/ROADMAP.md` y `evidencias/adr-0046/` están en `main` desde la propuesta de la entrada: el run los cumple, no
  los decide.
- **Alcance**:
  - En el cuerpo va lo que toda respuesta de la skill necesita: el protocolo, las reglas invariantes, las formas
    fijas que las evals comprueban sin modelo (las líneas `⚠`), la tabla de comandos y, por cada referencia, cuándo
    leerla. En una referencia, el detalle que solo pide una clase de pregunta.
  - La reescritura cambia la forma, no las reglas: ninguna regla de la skill desaparece ni aparece. `research.md`
    lleva la tabla de cada regla de la versión de `main` con su sitio en la nueva, y la revisión la contrasta.
  - Los controles de contenido que hoy leen `SKILL.md` —la prosa sin expresiones prohibidas, la línea sin consulta,
    las órdenes para PowerShell, las formas fechadas— leen la skill entera, `SKILL.md` y sus referencias a mano:
    mover un texto a una referencia no lo saca de ningún control.
  - Las órdenes de ejemplo de una referencia son de lo que la caché de las sesiones sirve sin red (ADR 0043), como
    las de `SKILL.md`.
- **Fuera de alcance**: los nombres de las skills y las skills por capacidad (H32); la cobertura en positivo y
  `⚠ NO CUBIERTO:` (H33); reglas nuevas en el protocolo; evals o casos nuevos de las skills; el umbral, la rúbrica y
  los casos de `afirma_lo_no_leido` y de las demás clases de hoy; la rúbrica, los casos y la medida del juez de las
  descripciones, que son los de `evidencias/adr-0046/` y solo los cambia una persona fuera de un run.
- **Controles**: un mutante de `make ci` por gate; un test de que `make skills-sync` conserva una referencia a mano y
  rehace una generada; un test de que el conjunto de dependencias incluye la referencia a mano; cada control de
  contenido falla con su regla movida a una referencia; el control de identidad del juez de las descripciones con
  `evidencias/adr-0046/`; tests del juicio de la descripción con votos simulados (marca con tres «si» y su frase, no
  con dos, nulo con una frase que no está), sin abrir ninguna sesión.
- **Aceptación**: `make ci` en verde con los gates nuevos sobre las tres skills; y el job de evals sobre la cabeza de
  la propuesta, que con las tres descripciones cambiadas mide las 45 evals en los dos modos y juzga las tres
  descripciones, en verde con los umbrales vigentes (ADR 0029, 0038 y 0045), sin nada reutilizado y sin ninguna
  descripción marcada. Antes de fusionar, una persona lanza la medida del juez de las descripciones con el código del
  job (etiqueta `evals-medir-juez`) y da 63 de 63 y 0 de 36.
- **Riesgos y dependencias**: la reescritura cambia lo que lee el modelo, y deciden las evals. La serie de cierres de
  cada skill empieza de cero, porque su clave es el árbol de `skills/<skill>` (ADR 0038). La primera release tras H31
  necesita la suite completa sobre el commit etiquetado (ADR 0045). Depende de H30.
