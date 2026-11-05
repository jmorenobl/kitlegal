// Los textos de la portada, la de juristas y despachos: su título, su
// descripción y sus ejemplos (ADR 0024 y 0034). Las citas se nombran por su id
// de citas.yaml: el texto de la norma nunca se escribe aquí.

import { REPOSITORIO } from "../lib/proyecto";

export interface Audiencia {
  ruta: string;
  // Cómo se nombra la audiencia en la imagen para compartir.
  nombre: string;
  titulo: string;
  descripcion: string;
  imagen: string;
  h1: string;
  subtitulo: string;
  // El botón principal de la cabecera de la portada: lo que esa audiencia hace
  // (instalar, en la de despachos). «Ver cómo funciona» va a su lado como enlace.
  secundario: { texto: string; destino: string; icono: string };
  comparativa: {
    consulta: string;
    respuesta: string;
    marcas: string[];
    insignia: string;
    alertaTitulo: string;
    alertaTexto: string;
    pieMalo: string;
    cita: string;
  };
  garantias: {
    titulo: string;
    texto: string;
    // La página que cuenta cómo se comprueba lo que se promete, si la hay.
    enlace?: { texto: string; destino: string };
    // Las garantías. Destacadas y a la par, los dos frenos de un despacho: la
    // de las sentencias (el miedo es la jurisprudencia inventada) y el secreto
    // profesional (dónde están los datos: lo que corre en el equipo y lo que
    // no). Debajo, leer la norma y comprobar su vigencia.
    sentencias: { titulo: string; texto: string };
    secreto: { titulo: string; texto: string };
    norma: { titulo: string; texto: string };
    vigencia: { titulo: string; texto: string };
  };
  // La etiqueta es la materia del caso, en neutro: el color de la portada es para
  // el estado de una norma (vigente, derogada), no para una categoría.
  // `consulta` es el slug de la página de consultas.ts que desarrolla la
  // respuesta, si la hay.
  casos: { icono: string; etiqueta: string; pregunta: string; cita: string; consulta?: string }[];
  // La invitación a escribir al buzón del proyecto: kitlegal no recoge datos de
  // uso (ADR 0027), así que lo que le falta a una audiencia solo se sabe si lo
  // cuenta.
  contacto?: {
    titulo: string;
    texto: string;
    peticiones: string[];
    // Lo que todavía no hace y está en el roadmap o en el backlog
    // (docs/roadmap/), sin fechas: una frase por cosa.
    enMarcha: string[];
    // Lo que no hará por decisión, con su porqué plegado. `detalle` enlaza la
    // explicación larga.
    descartado: { resumen: string; explicacion: string };
    detalle: { texto: string; destino: string };
    aviso: string;
    asunto: string;
  };
  cta: { titulo: string; texto: string };
}

export const despachos: Audiencia = {
  ruta: "/",
  nombre: "Para despachos y juristas",
  titulo: "IA para juristas: cada cita, comprobada en el BOE | kitlegal",
  descripcion:
    "Tu asistente de IA cita el texto consolidado del BOE, con el artículo exacto y su vigencia. Hoy, en Claude. Software libre y sin cuotas.",
  imagen: "despachos",
  h1: "La IA que ya usas en tu despacho, con cada artículo comprobado en el BOE antes de citarlo.",
  subtitulo:
    "No es otra plataforma: kitlegal conecta tu asistente de IA con el texto consolidado del BOE, y cada respuesta llega con el artículo exacto y la fuente oficial. kitlegal es gratis y corre en tu equipo; el asistente se paga aparte.",
  secundario: { texto: "Instalar kitlegal", destino: "/instalar/", icono: "download" },
  comparativa: {
    consulta: "¿En qué plazo tiene que resolver la Administración si la norma del procedimiento no fija ninguno?",
    respuesta:
      "«Con carácter general, el plazo para resolver es de seis meses, según la Ley 30/1992 de Régimen Jurídico de las Administraciones Públicas.»",
    marcas: ["seis meses", "Ley 30/1992"],
    insignia: "Norma derogada · plazo erróneo",
    alertaTitulo: "Dos errores en una frase",
    alertaTexto:
      "La Ley 30/1992 está derogada. Y seis meses es el máximo que puede fijar la norma del procedimiento (art. 21.2 de la Ley 39/2015); si no fija ninguno, el plazo es de tres meses (art. 21.3), como cita la respuesta con kitlegal.",
    pieMalo: "Sin fuente",
    cita: "lpacap-21",
  },
  garantias: {
    titulo: "Garantías que da el programa, no el modelo",
    texto:
      "Lo que no debe hacerse de memoria lo hace un programa determinista, y el modelo razona sobre lo que el programa le devuelve.",
    enlace: { texto: "Cómo se comprueba, con sus resultados y sus errores", destino: "/como-se-evalua/" },
    norma: {
      titulo: "Cada artículo, leído en el BOE",
      texto:
        "Del texto consolidado, al responder: con su identificador, el bloque, la fecha de consulta y la huella del texto.",
    },
    vigencia: {
      titulo: "La vigencia, antes de citar",
      texto: "Si el BOE marca la norma como derogada o con la vigencia agotada, lo dice primero.",
    },
    sentencias: {
      titulo: "Ninguna sentencia sin su documento",
      texto: "kitlegal no busca sentencias ni cita ninguna de memoria: solo cita la que traes, tras comprobar que es la que se pide.",
    },
    secreto: {
      titulo: "Tus asuntos no salen de tu equipo",
      texto:
        "kitlegal es un programa que corre en tu ordenador: lee los datos abiertos del BOE y guarda las normas consultadas en tu disco, sin cuentas ni servidores de kitlegal por medio. Lo que escribes a tu asistente sigue las condiciones de ese asistente.",
    },
  },
  casos: [
    {
      icono: "gavel",
      etiqueta: "Contratación pública",
      pregunta: "¿Cuál es el límite de un contrato menor?",
      cita: "lcsp-118",
    },
    {
      icono: "receipt-long",
      etiqueta: "Derecho tributario",
      pregunta: "¿Cuándo prescribe una deuda tributaria?",
      cita: "lgt-66",
      consulta: "prescripcion-deudas-hacienda",
    },
    {
      icono: "account-balance",
      etiqueta: "Régimen local",
      pregunta: "¿Qué atribuciones corresponden al Pleno municipal?",
      cita: "lbrl-22",
    },
    {
      icono: "schedule",
      etiqueta: "Procedimiento administrativo",
      pregunta: "¿Qué efecto tiene el silencio en un procedimiento iniciado a solicitud?",
      cita: "lpacap-24",
      consulta: "silencio-administrativo",
    },
  ],
  contacto: {
    titulo: "¿Qué le falta para tu despacho?",
    texto:
      "kitlegal ya sirve para consultar y citar la legislación consolidada del BOE. Estamos trabajando para que sirva al trabajo diario de un despacho y, como no recoge ningún dato de uso, la única forma de saber qué necesitas es que nos lo cuentes.",
    peticiones: [
      "Una consulta que tu asistente no supo responder, o que respondió mal.",
      "Una fuente que echas de menos: jurisprudencia, consultas de la DGT, convenios colectivos, boletines autonómicos.",
      "Lo que te frena para usarlo: la instalación, la confidencialidad, la forma de la respuesta.",
    ],
    enMarcha: [
      "Redactar escritos.",
      "Leer por sí mismo las sentencias del Tribunal Constitucional y las del Supremo que anulan una disposición, que se publican en el BOE.",
      "Consultar la jurisprudencia europea.",
    ],
    descartado: {
      resumen: "Consultar el CENDOJ de forma automática.",
      explicacion:
        "La jurisprudencia del Tribunal Supremo, la Audiencia Nacional, los tribunales superiores de justicia y las audiencias provinciales está en el CENDOJ, del Consejo General del Poder Judicial, cuyas condiciones de uso no permiten consultarlo de forma masiva ni automatizada. kitlegal las respeta: tú buscas la sentencia con tu navegador, con la consulta que te prepara tu asistente.",
    },
    detalle: { texto: "Por qué, con detalle", destino: `${REPOSITORIO}#y-las-sentencias` },
    aviso:
      "No incluyas datos de clientes ni de asuntos. Leemos tu mensaje solo para responderte y para decidir qué mejorar.",
    asunto: "kitlegal para despachos",
  },
  cta: {
    titulo: "Dota a tu despacho de la ley vigente, con su cita.",
    texto: "Gratis, en tu equipo y sin cuentas: pruébalo con la próxima consulta que tengas que citar.",
  },
};

// La audiencia principal de la web de kitlegal: despachos, juristas y agentes.
export const audiencias = [despachos];
