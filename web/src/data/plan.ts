// Los textos de /plan/, la página del plan para juristas y despachos. El estado
// y el orden de cada hito no se escriben aquí: salen de docs/roadmap/ al
// construir (lib/roadmap.ts). Aquí va lo que el roadmap dice en su lengua de
// trabajo, contado para quien no la conoce: cada frase resume la entrada del
// hito, su fase en docs/ROADMAP.md §4 o los ADR 0044 y 0059. La página no cita el
// texto de ninguna norma; nombra los preceptos en que se apoya el plan para que
// un jurista los revise.

export interface Etapa {
  // La fase de la cabecera de los ficheros de docs/roadmap/.
  fase: string;
  nombre: string;
  titulo: string;
  texto: string;
  // Lo que gana cada uno al terminar la etapa (docs/ROADMAP.md §4 y ADR 0059).
  ganan: { ciudadania: string; despacho: string; asesoria: string };
  // Lo que la etapa trae sin número de hito todavía (las verticales, del backlog).
  sinNumero?: { titulo: string; frase: string }[];
  // Qué fueron las entregas ya hechas de la etapa, si las hay. Sin ella, la
  // página da solo su número.
  hechas?: string;
}

// Las fases de docs/roadmap/ que no son etapas de la página: están hechas y se
// cuentan en «Lo que hace hoy». Un hito pendiente de cualquier otra fase sin
// etapa hace fallar la construcción.
export const fasesHechas = ["0", "terreno-y-memoria"];

export const etapas: Etapa[] = [
  {
    fase: "1",
    nombre: "Etapa 1",
    titulo: "Terminar la base, para cualquier materia",
    texto:
      "Las herramientas comunes a todas las materias: leer la norma, contar un plazo, situar el territorio y el destinatario, comprobar un documento y leer la doctrina. Cada una se prueba con dos casos reales antes de darla por buena.",
    ganan: {
      ciudadania: "Pregunta de cualquier materia y sabe qué está cubierto y qué no.",
      despacho: "Normativa europea leída y citada; lo que dice un contrato o una póliza sale del documento, también en PDF.",
      asesoria: "Los plazos de la revisión tributaria con su precepto, y la doctrina de la DGT y del TEAC leída.",
    },
    hechas:
      "Las trece primeras reorganizaron por capacidades y midieron las instrucciones del asistente, y declararon lo que cubre.",
  },
  {
    fase: "2",
    nombre: "Etapa 2",
    titulo: "Despachos y asuntos ante la administración",
    texto:
      "Entregas cortas, cada una útil desde el día en que se termina: lo que un despacho usa a diario y lo necesario para llevar un asunto ante cualquier administración, de consultar a actuar y a seguir.",
    ganan: {
      ciudadania: "Lleva su propio asunto: qué vence, el aviso del silencio y sus escritos listos para firmar.",
      despacho: "Plazos procesales, revisión de las citas de un escrito, el Constitucional y los asuntos de sus clientes.",
      asesoria: "Nada propio todavía: lo tributario que no entra en la etapa 1 sigue sin cubrir.",
    },
  },
  {
    fase: "3",
    nombre: "Etapa 3",
    titulo: "Áreas completas para profesionales",
    texto:
      "Cada área es un paquete de normas, plazos, doctrina y escritos sobre la base. Entran en el orden que pida el uso real, la fiscal primero.",
    ganan: {
      ciudadania: "El plan no prevé nada propio para la ciudadanía en esta etapa.",
      despacho: "Lo que pidan los despachos: jurisprudencia europea, doctrina sectorial o la cartera de asuntos.",
      asesoria: "Su área completa, y un paquete que junta fiscal, laboral y mercantil para una gestoría.",
    },
    sinNumero: [
      {
        titulo: "Fiscal",
        frase:
          "Sobre lo tributario de la etapa 1: el IVA y el impuesto sobre sociedades si no entraron antes, el calendario del contribuyente, los modelos, el sancionador, las haciendas forales y los escritos.",
      },
      { titulo: "Laboral", frase: "Estatuto de los Trabajadores, jurisdicción social, Seguridad Social y convenios colectivos." },
      { titulo: "Mercantil", frase: "Sociedades de capital, Ley Concursal, Código de Comercio y BORME." },
    ],
  },
  {
    fase: "transparencia",
    nombre: "Más adelante",
    titulo: "Lo que hacen las administraciones",
    texto:
      "Qué contratan y qué subvencionan, qué sociedades y administradores aparecen en el BORME y cómo van sus presupuestos. No resuelve la situación de nadie, y por eso va detrás.",
    ganan: {
      ciudadania: "Consultar qué contrata y qué subvenciona su ayuntamiento, y cómo va su presupuesto.",
      despacho: "Las sociedades y administradores del BORME, que sirven también al área mercantil.",
      asesoria: "Las mismas consultas, sin nada propio.",
    },
  },
];

// Un título corto y una frase por hito con número que no está hecho. La
// construcción falla si falta el de alguno de docs/roadmap/: un hito nuevo en la
// cola entra en la web con su frase.
export const enLlano: Record<string, { titulo: string; frase: string }> = {
  H38: {
    titulo: "Que se actualice solo",
    frase: "Al actualizar el programa, las instrucciones del asistente instaladas para todo el equipo se actualizan con él, salvo que las hayas modificado, y se avisa de que valen desde la conversación siguiente.",
  },
  H36: {
    titulo: "Plazos de cualquier materia",
    frase: "Cada forma de contar un plazo se describe con sus preceptos, empezando por los dos de hoy, y se prueba con el cómputo civil de la LEC, con agosto y del 24 de diciembre al 6 de enero inhábiles.",
  },
  H45: {
    titulo: "Materias nuevas sin tocar el programa",
    frase: "Una materia entra con sus normas, sus plazos y lo que la ley regula aparte, y cada cosa aparece en la referencia de su materia. Se prueba con lo administrativo y con una materia de ensayo, antes de la tributaria.",
  },
  H42: {
    titulo: "Lo tributario, como segunda materia",
    frase: "Los plazos del recurso de reposición y de la reclamación económico-administrativa contra los tributos del Estado. Fuera quedan las sanciones, los tributos locales y autonómicos propios, las aduanas, las haciendas forales, los modelos y los escritos.",
  },
  H34: {
    titulo: "Normativa europea",
    frase: "Las fuentes se conectan todas de la misma manera, y la normativa de EUR-Lex se lee y se cita como la española, tras comprobar sus condiciones de uso.",
  },
  H35: {
    titulo: "El documento que trae la persona",
    frase: "Lo que se dice de una póliza, un contrato o un convenio sale del documento, también en PDF. Un escaneado lo lee el asistente y lo marca como no comprobado.",
  },
  H44: {
    titulo: "Doctrina de la DGT y del TEAC",
    frase: "Una consulta de la DGT o una resolución del TEAC se cita leída, separada de la norma, y diciendo si es vinculante. Cómo se obtienen depende de sus condiciones de uso, aún por revisar.",
  },
  H37: {
    titulo: "Destinatarios y territorio",
    frase: "El destinatario de un escrito puede no ser una administración, y el territorio no es solo el municipio: partido judicial y vecindad civil, por ahora declarados no cubiertos.",
  },
  H43: {
    titulo: "La prueba de la etapa",
    frase: "Preguntas de cinco o seis materias en las que la norma se lee y se cita, ningún plazo sale de memoria y lo no cubierto se declara.",
  },
  "H9.1": {
    titulo: "Plazos procesales civiles y sociales",
    frase: "El último día de un plazo de la LEC o de la LRJS, con su precepto. En lo social, si la pregunta no dice la modalidad procesal, se pide antes de calcular, porque en algunas agosto y del 24 de diciembre al 6 de enero cuentan.",
  },
  H8: {
    titulo: "Revisar las citas de un escrito",
    frase: "Un escrito, un dictamen o la respuesta de otro asistente se revisan cita a cita, en la redacción vigente a su fecha, también en PDF.",
  },
  H27: {
    titulo: "El Tribunal Constitucional, desde el BOE",
    frase: "Una sentencia del Constitucional se lee del BOE sin que la persona tenga que traerla.",
  },
  H10: {
    titulo: "Seguimiento de expedientes",
    frase: "Lo presentado y lo notificado ante cualquier administración, lo que vence y el aviso del silencio, para un despacho o para un ciudadano.",
  },
  H11: {
    titulo: "Escritos listos para firmar",
    frase: "Solicitud de acceso a la información, recurso de alzada o de reposición y alegaciones, en texto y en Word, con la fundamentación citada y el plazo calculado.",
  },
  H15: {
    titulo: "Boletines y edictos",
    frase: "El boletín de la comunidad y el Tablón Edictal Único, empezando por la Comunidad de Madrid.",
  },
  H16: {
    titulo: "Ordenanzas municipales",
    frase: "La ordenanza aprobada y sus modificaciones, citando cada anuncio del boletín, sin presentarla como texto consolidado.",
  },
  H28: {
    titulo: "Doctrina publicada en el BOE",
    frase: "Dictámenes del Consejo de Estado y doctrina de la Fiscalía General y de la Abogacía del Estado, con su texto delante.",
  },
  H12: { titulo: "Comprobar las fuentes municipales", frase: "Qué publican la Plataforma de Contratación y la Base de Datos de Subvenciones, antes de programar nada." },
  H13: { titulo: "Contratación pública", frase: "Qué ha licitado y adjudicado un ayuntamiento, a quién y por qué importe." },
  H14: { titulo: "Subvenciones", frase: "Qué subvenciones convoca y concede un ayuntamiento, y a quién." },
  H17: { titulo: "BORME", frase: "Las sociedades y los administradores que aparecen en el BORME, con sus nombramientos y ceses." },
  H18: { titulo: "Presupuestos municipales", frase: "Lo presupuestado y lo liquidado por un ayuntamiento, con sus desviaciones." },
};

// Lo que queda en el backlog, sin número ni orden (ADR 0044, punto 2).
export const otrasMaterias =
  "Sin número ni fecha quedan otras materias, que entran cuando el uso las pida: civil, consumo y seguros, familia y el procedimiento sancionador de tráfico. Penal queda fuera mientras no se decida abrirlo.";

// Lo que kitlegal no hará, por decisión (ADR 0004, 0027 y 0036).
export const nunca = [
  {
    titulo: "Consultar el CENDOJ de forma automática",
    texto:
      "Su buscador cierra el paso a los programas, y no se sortea. La persona busca la sentencia con su navegador, con la búsqueda que le prepara el asistente, y la trae; ninguna se cita sin tenerla delante.",
  },
  {
    titulo: "Presentar o notificar en nombre de nadie",
    texto: "Todo lo que pide identidad termina en un fichero listo para que la persona lo firme y lo presente ella.",
  },
  {
    titulo: "Enviar tus consultas a un servidor de kitlegal",
    texto:
      "Ni preguntas, ni asuntos, ni datos de uso llegan a kitlegal, tampoco con consentimiento: el programa corre en tu ordenador. Lo que escribes a tu asistente sigue las condiciones de ese asistente.",
  },
];
