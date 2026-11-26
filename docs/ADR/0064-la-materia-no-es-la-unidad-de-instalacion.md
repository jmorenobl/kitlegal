# 0064 · La materia no es la unidad de instalación: un binario y un plugin con todas las materias, y el plugin `kitlegal` como contrato para terceros

- **Estado**: aceptada (2026-11-26)
- **Fecha**: 2026-11-26
- **Hito**: transversal, fuera de un run, como documentación. Sustituye, en `docs/roadmap/backlog/distribucion.md`, la
  línea de los packs en lo que tenía de instalación («`kitlegal skills install <pack>` instala un pack entero» y «un
  plugin por pack, generado desde `packs/`»). Precisa el ADR 0044, punto 5: un pack agrupa verticales, pero no se
  instala aparte. No cambia el ADR 0044, punto 3, ni los ADR 0062 y 0063. No cambia ningún contrato del binario.

## Contexto y problema

El 2026-11-26 se estudió una propuesta para partir el plugin `kitlegal` en tres partes:

- `kitlegal-base`, con el binario, el servidor MCP y las skills de capacidad;
- un plugin por materia (`kitlegal-fiscal`, `kitlegal-laboral`, `kitlegal-mercantil`…), cada uno con sus skills y
  dependiente de la base por `dependencies` de `plugin.json`;
- `kitlegal` 2.0.0, que no trae nada propio y depende de los demás.

Sus motivos eran que cada persona instale solo las materias que usa, para que se activen mejor las skills porque
compiten menos descripciones, que cada materia tenga su propia versión, y que haya un plugin por dominio, como en
Ventanilla. Se decidió por mérito de diseño, no por si encaja con los ADR: un ADR que fijara un mal diseño se
sustituiría. La segunda lectura la hizo Fable, sin la conclusión de la primera.

Lo que se encontró en el repositorio y en la documentación:

- **Una materia no es un artefacto que se pueda separar.** Sus tipos de plazo van empotrados en el binario
  (`data/datos.go`, `data/plazos.yaml`). Sus referencias son ficheros dentro de cada skill de capacidad, con una línea
  generada entre las marcas de los paquetes en su `SKILL.md` (ADR 0063). Su cobertura se genera dentro de
  `consulta-juridica` desde `data/cobertura.yaml`. Un plugin no puede añadir ficheros a la skill de otro plugin ni datos
  al binario de otro, así que la materia tendría que volver a ser una skill propia, que repite el protocolo de cita y
  de `⚠ NO CUBIERTO:` o encadena dos skills.
- **Más skills empeoran la activación, no la mejoran.** Con `normativa` sola, la eval 04 no la activó en 1 de 160
  sesiones; con las cuatro skills, en 14 de 160 (ADR 0055) y en 33 de 320 (ADR 0057). Lo arreglaron las descripciones
  cortas, con 0 de 800, no instalar menos. Hoy una materia añade cero descripciones y unos 70 caracteres a cada cuerpo
  (ADR 0063). Una skill por materia añadiría una descripción más por materia, y competiría con la skill de capacidad
  por la misma pregunta.
- **La cobertura dependería de lo instalado.** Hoy lo no cubierto es un dato del producto, generado y comprobado sin
  modelo. Repartido en plugins, lo inferiría el agente según lo que encuentre instalado. Una pregunta que cruza dos
  materias, hoy dos referencias dentro de una skill (una orden de leer medida en 38 a 40 de 40, ADR 0051 y 0053),
  pasaría a ser dos skills de dos plugins.
- **Las versiones independientes no lo serían.** Un plazo de una materia pide una release del binario, la cobertura se
  regenera en la base y el job de evals mide las skills juntas. Los rangos de versión crearían combinaciones que nadie
  mide; hoy cada release es una sola combinación, la medida.
- **Solo serviría a un canal.** `kitlegal skills install`, Antigravity, Codex y el `.mcpb` suelto no tienen
  dependencias entre plugins (ADR 0019 y 0035): habría dos formas de que entre una materia, una por canal.
- **La audiencia principal es la que peor lo recibiría.** La documentación de dependencias de Claude Code
  (code.claude.com/docs/en/plugins/dependencies, leída el 2026-11-26) dice que un plugin activado desde claude.ai carga
  como `<name>@synced`, sin marketplace, y que hay que decir a la persona que active también la dependencia: no se
  instala sola. Actualizar fuera de la actualización automática pide `claude plugin update` y `/reload-plugins`, que se
  usan desde la terminal. Y la app de escritorio ya se ha apartado de esa documentación antes (ADR 0035: la fuente
  `archive` y `mcpServers` por dirección).

## Opciones consideradas

1. **Un plugin por materia sobre una base**, la propuesta. Rechazada por lo anterior.
2. **Un binario y un plugin con todas las materias**, como hoy, con un pack que solo agrupa. Elegida.
3. **Variantes completas del mismo release por perfil** (por ejemplo, una gestoría con fiscal, laboral y mercantil),
   cada una con su binario y sus skills, que no se instalan juntas. Aplazada: es la salida si un día sobra alguna
   materia, pero ninguna medida lo pide hoy.

## Decisión

1. **Un binario y un plugin, con todas las materias, en todos los canales**: el plugin `kitlegal`,
   `kitlegal skills install`, Antigravity, Codex y el `.mcpb`.
2. **Una materia entra como datos y referencias** en las skills de capacidad (ADR 0044, punto 3), nunca como una skill
   ni un plugin propios.
3. **Un pack agrupa verticales para un tipo de profesional, pero no se instala aparte**: sirve para decir qué materias
   pide cada profesional, en la web y en la documentación, y para ordenar sus evals y su cobertura.
4. **El servidor MCP solo lo declara el plugin `kitlegal`.** Ningún otro plugin de kitlegal lo declara, para que sus
   herramientas no aparezcan dos veces.
5. **El plugin `kitlegal` es un contrato para terceros.** Su nombre y los nombres de sus herramientas
   (`<applet>_<verbo>`) no cambian sin una versión mayor. Un plugin de otro, por ejemplo de Ventanilla, puede depender
   de él con `dependencies`.
6. **Seleccionar materias solo se vuelve a plantear con una medida** en la que, con muchos paquetes, la activación o
   la lectura de la referencia correcta empeore. Incluso entonces, la salida es la opción 3, no plugins sueltos.
7. **La segunda materia real trae una eval que cruza dos materias**, como la tributación en el IRPF de una
   indemnización por despido, en cuanto haya dos paquetes que la respondan.

## Consecuencias

**A favor**

- Instalar sigue siendo un paso y lo cubierto sigue siendo un dato fijo de cada release, medido entero por el job de
  evals.
- Cualquier canal recibe lo mismo.
- Ventanilla y otros terceros saben qué pueden dar por estable.

**En contra, y asumido**

- Quien solo usa una materia recibe las demás. Hoy le cuestan unos 70 caracteres por materia en el cuerpo de cada
  skill de capacidad y cero descripciones.
- La razón de la activación extrapola: nadie ha medido kitlegal con más de cuatro skills ni con muchos paquetes. Por
  eso el punto 6 deja la puerta abierta a una medida.
- Renombrar el plugin o una herramienta pasa a pedir una versión mayor y un aviso a quien dependa de ellos.

**Qué cambiaría esta decisión**

- Una medida con muchos paquetes en la que la activación o la lectura caigan (punto 6).
- Materias que escriban terceros a su propio ritmo: pediría que el binario lea datos de materia de fuera de lo
  empotrado, un puerto nuevo con su propio problema de procedencia.
- Que la app de escritorio instale sola las dependencias y deje a la skill de un plugin usar el servidor de otro.
  Quitaría la última razón del contexto, no las demás.
