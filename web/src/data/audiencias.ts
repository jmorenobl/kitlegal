// Los textos de las dos portadas: una URL por audiencia, cada una con su
// título, su descripción y sus ejemplos (ADR 0024 y 0034). La de la ciudadanía
// es la portada del sitio. Las citas se nombran por su id de citas.yaml: el
// texto de la norma nunca se escribe aquí.

import { formatoEntero, numeroDeMunicipios, REPOSITORIO } from "../lib/proyecto";

export type Tono = "sello" | "derogada" | "agente" | "vigente";

export interface Audiencia {
  ruta: string;
  // Cómo se nombra la audiencia en el selector de las portadas, en el pie y en
  // la imagen para compartir.
  nombre: string;
  icono: string;
  titulo: string;
  descripcion: string;
  imagen: string;
  // La fotografía de la cabecera (src/assets/ilustraciones/) y la parte de ella
  // que se conserva al recortarla, como clase de `object-position`.
  ilustracion: { nombre: string; alt: string; foco: string };
  insignia: string;
  h1: string;
  subtitulo: string;
  // El segundo botón de la cabecera de la portada: lo que esa audiencia hace
  // después de ver cómo funciona.
  secundario: { texto: string; destino: string; icono: string };
  metricas: { etiqueta: string; valor: string; texto: string; tono: Tono }[];
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
    antetitulo: string;
    titulo: string;
    texto: string;
    pilares: { icono: string; tono: Tono; fondo: string; titulo: string; texto: string; pie: string }[];
  };
  // `consulta` es el slug de la página de consultas.ts que desarrolla la
  // respuesta, si la hay.
  casos: { icono: string; tono: Tono; etiqueta: string; pregunta: string; cita: string; consulta?: string }[];
  // La invitación a escribir al buzón del proyecto: kitlegal no recoge datos de
  // uso (ADR 0027), así que lo que le falta a una audiencia solo se sabe si lo
  // cuenta.
  contacto?: {
    antetitulo: string;
    titulo: string;
    texto: string;
    peticiones: string[];
    // Lo que hoy no hace, y por qué: una frase por cosa. `detalle` enlaza la
    // explicación larga.
    todavia: string[];
    detalle: { texto: string; destino: string };
    aviso: string;
    asunto: string;
  };
  cta: { titulo: string; texto: string };
}

const municipios = formatoEntero(numeroDeMunicipios());

export const ciudadania: Audiencia = {
  ruta: "/",
  nombre: "Para ciudadanos y trámites",
  icono: "person-outline",
  titulo: "kitlegal: tus derechos y plazos, con la ley vigente del BOE",
  descripcion:
    "No necesitas otra IA: kitlegal hace que tu asistente lea el texto vigente del BOE antes de contestarte por un plazo, tu alquiler o tu trabajo. Gratis.",
  imagen: "inicio",
  ilustracion: {
    nombre: "portada-ciudadania",
    alt: "Una persona de espaldas, sentada a la mesa de la cocina con un portátil y un sobre.",
    foco: "object-[72%_50%]",
  },
  insignia: "Legislación española · La ley vigente, con su cita",
  h1: "Pregunta por tu plazo, tu alquiler o tu trabajo y recibe la ley vigente, con su cita.",
  subtitulo:
    "kitlegal hace que tu asistente de IA lea el texto vigente del BOE antes de contestarte: sabrás qué artículo lo dice, si la norma sigue en vigor y dónde comprobarlo. Es gratis y funciona en tu ordenador.",
  secundario: { texto: "Ver consultas con su cita", destino: "/consultas/", icono: "menu-book" },
  metricas: [
    {
      etiqueta: "Nada sin cita",
      valor: "La ley, palabra por palabra",
      texto: "Cada respuesta se apoya en el texto publicado en el BOE, con la norma y el artículo.",
      tono: "sello",
    },
    {
      etiqueta: "Vigencia",
      valor: "Sin leyes caducadas",
      texto: "Si una norma ya no está en vigor, tu asistente te lo avisa antes de nada.",
      tono: "derogada",
    },
    {
      etiqueta: "Precio",
      valor: "Gratis",
      texto: "Software libre: se instala y se usa sin pagar y sin registrarse.",
      tono: "agente",
    },
    {
      etiqueta: "Privacidad",
      valor: "Sin cuentas",
      texto: "kitlegal no guarda tus preguntas ni las envía a ningún servidor propio: solo consulta el BOE.",
      tono: "vigente",
    },
  ],
  comparativa: {
    consulta: "«Me han notificado una resolución y quiero presentar recurso de alzada: ¿qué plazo tengo?»",
    respuesta: "«Tiene tres meses para presentar el recurso de alzada, según la Ley 30/1992 de procedimiento administrativo.»",
    marcas: ["tres meses", "Ley 30/1992"],
    insignia: "Norma derogada · plazo erróneo",
    alertaTitulo: "Un error que cuesta el recurso",
    alertaTexto:
      "La Ley 30/1992 está derogada. Y contra una resolución expresa el plazo es de un mes: quien espere a los tres meses llega tarde y la resolución queda firme.",
    pieMalo: "Recurso fuera de plazo",
    cita: "lpacap-122",
  },
  garantias: {
    antetitulo: "Por qué fiarte",
    titulo: "Tres comprobaciones que hace un programa, no la memoria de la IA",
    texto:
      "Leer la ley, comprobar que sigue en vigor y situar tu municipio no se dejan a lo que la IA recuerde: lo hace un programa, siempre de la misma manera. Tu asistente razona sobre lo que el programa le devuelve.",
    pilares: [
      {
        icono: "menu-book",
        tono: "sello",
        fondo: "bg-panel",
        titulo: "Lee la ley en el BOE",
        texto:
          "Tu asistente no contesta de memoria: kitlegal descarga el texto oficial y actualizado de la norma, y la respuesta se apoya en esa redacción.",
        pie: "Texto consolidado oficial",
      },
      {
        icono: "history-toggle-off",
        tono: "derogada",
        fondo: "bg-derogada-fondo",
        titulo: "Te avisa si ya no está en vigor",
        texto:
          "Si el BOE marca una norma como derogada o con la vigencia agotada, tu asistente te lo dice antes de citarla.",
        pie: "Antes de citar",
      },
      {
        icono: "location-city",
        tono: "agente",
        fondo: "bg-panel",
        titulo: "Sabe de qué municipio hablas",
        texto: `Sitúa cualquiera de los ${municipios} municipios de España en su provincia y su comunidad, y te dice qué boletines oficiales tiene configurados y cuáles todavía no.`,
        pie: `${municipios} municipios del INE`,
      },
    ],
  },
  casos: [
    {
      icono: "home",
      tono: "sello",
      etiqueta: "Alquiler y vivienda",
      pregunta: "Mi casero no me devuelve la fianza. ¿Qué dice la ley?",
      cita: "lau-36",
      consulta: "devolucion-fianza-alquiler",
    },
    {
      icono: "account-balance",
      tono: "agente",
      etiqueta: "Trámites con la Administración",
      pregunta: "Presenté una solicitud y la Administración no contesta. ¿Se entiende concedida?",
      cita: "lpacap-24",
      consulta: "silencio-administrativo",
    },
    {
      icono: "badge",
      tono: "vigente",
      etiqueta: "Trabajo",
      pregunta: "Quiero dejar mi trabajo. ¿Cuánto preaviso tengo que dar?",
      cita: "et-49",
      consulta: "preaviso-baja-voluntaria",
    },
    {
      icono: "receipt-long",
      tono: "sello",
      etiqueta: "Hacienda",
      pregunta: "¿Cuánto tiempo tiene Hacienda para reclamarme una deuda?",
      cita: "lgt-66",
      consulta: "prescripcion-deudas-hacienda",
    },
  ],
  cta: {
    titulo: "Empieza a preguntar con la ley vigente delante.",
    texto:
      "Ten a mano la legislación del BOE en tus gestiones con la Administración, en el trabajo o con tu vivienda. Software libre y gratuito.",
  },
};

export const despachos: Audiencia = {
  ruta: "/despachos/",
  nombre: "Para despachos y abogados",
  icono: "balance",
  titulo: "IA para abogados: cada cita, comprobada en el BOE | kitlegal",
  descripcion:
    "No cambies de IA: kitlegal hace que la app de Claude que ya usas cite el texto consolidado del BOE, con el artículo exacto y su vigencia. Gratis, sin cuotas.",
  imagen: "despachos",
  ilustracion: {
    nombre: "portada-despachos",
    alt: "Un escritorio de despacho con un portátil, un libro abierto con un marcapáginas ámbar y una lámpara.",
    foco: "object-[78%_50%]",
  },
  insignia: "Para despachos en España · Software libre y gratuito",
  h1: "La IA que ya usas en tu despacho, con cada artículo comprobado en el BOE antes de citarlo.",
  subtitulo:
    "No es otra plataforma ni otra suscripción: kitlegal conecta tu asistente de IA (la app de escritorio de Claude, o Claude Code) con el texto consolidado del BOE. Cada respuesta llega con el artículo exacto, su vigencia y la fuente oficial para contrastarla. Gratis, y en tu equipo.",
  secundario: { texto: "Cuéntanos qué necesita tu despacho", destino: "#contacto", icono: "mail" },
  metricas: [
    {
      etiqueta: "Nada sin cita",
      valor: "Norma, bloque y huella",
      texto: "Cada artículo llega con su identificador del BOE, la fecha de consulta y una huella del texto.",
      tono: "sello",
    },
    {
      etiqueta: "Vigencia",
      valor: "Aviso antes de citar",
      texto: "Si el BOE marca una norma como derogada o con la vigencia agotada, tu asistente lo dice primero.",
      tono: "derogada",
    },
    {
      etiqueta: "Alcance",
      valor: "Legislación, no sentencias",
      texto: "Lee la legislación consolidada del BOE. Hoy no consulta ni comprueba jurisprudencia.",
      tono: "agente",
    },
    {
      etiqueta: "Gratis y libre",
      valor: "EUPL-1.2",
      texto: "El programa corre en tu equipo: sin cuotas, sin cuentas y sin servidores de kitlegal entre tú y el BOE.",
      tono: "vigente",
    },
  ],
  comparativa: {
    consulta: "«¿En qué plazo tiene que resolver la Administración si la norma del procedimiento no fija ninguno?»",
    respuesta:
      "«Con carácter general, el plazo para resolver es de seis meses, según la Ley 30/1992 de Régimen Jurídico de las Administraciones Públicas.»",
    marcas: ["seis meses", "Ley 30/1992"],
    insignia: "Norma derogada · plazo erróneo",
    alertaTitulo: "Dos errores en una frase",
    alertaTexto:
      "La Ley 30/1992 está derogada. Y seis meses es el máximo que puede fijar la norma del procedimiento, salvo que una ley o el Derecho de la Unión Europea prevean más: si no fija ninguno, el plazo es de tres meses.",
    pieMalo: "Sin fuente",
    cita: "lpacap-21",
  },
  garantias: {
    antetitulo: "Arquitectura de confianza",
    titulo: "Tres garantías que da el programa, no el modelo",
    texto:
      "Lo que no debe hacerse de memoria lo hace un programa determinista: leer la norma, comprobar su vigencia y situar el territorio. El modelo razona sobre lo que el programa le devuelve.",
    pilares: [
      {
        icono: "menu-book",
        tono: "sello",
        fondo: "bg-panel",
        titulo: "Cotejo contra el BOE",
        texto:
          "Tu asistente no responde de memoria: kitlegal descarga el texto consolidado oficial, lo guarda en una caché de tu equipo y la respuesta se apoya en esa redacción.",
        pie: "Texto consolidado oficial",
      },
      {
        icono: "history-toggle-off",
        tono: "derogada",
        fondo: "bg-derogada-fondo",
        titulo: "Avisos de vigencia",
        texto:
          "kitlegal lee los metadatos del BOE y avisa si una norma está derogada, si su vigencia se ha agotado o si su consolidación no ha terminado.",
        pie: "Antes de citar",
      },
      {
        icono: "location-city",
        tono: "agente",
        fondo: "bg-panel",
        titulo: "Territorio real",
        texto: `Sitúa los ${municipios} municipios del INE en su provincia y su comunidad, con el código INE y el DIR3 del ayuntamiento, distingue régimen común y foral y dice qué boletines tiene configurados y cuáles no.`,
        pie: `${municipios} municipios del INE`,
      },
    ],
  },
  casos: [
    {
      icono: "gavel",
      tono: "sello",
      etiqueta: "Contratación pública",
      pregunta: "¿Cuál es el límite de un contrato menor?",
      cita: "lcsp-118",
    },
    {
      icono: "receipt-long",
      tono: "vigente",
      etiqueta: "Derecho tributario",
      pregunta: "¿Cuándo prescribe una deuda tributaria?",
      cita: "lgt-66",
      consulta: "prescripcion-deudas-hacienda",
    },
    {
      icono: "account-balance",
      tono: "agente",
      etiqueta: "Régimen local",
      pregunta: "¿Qué atribuciones corresponden al Pleno municipal?",
      cita: "lbrl-22",
    },
    {
      icono: "schedule",
      tono: "sello",
      etiqueta: "Procedimiento administrativo",
      pregunta: "¿Qué efecto tiene el silencio en un procedimiento iniciado a solicitud?",
      cita: "lpacap-24",
      consulta: "silencio-administrativo",
    },
  ],
  contacto: {
    antetitulo: "Hecho con quien lo usa",
    titulo: "¿Qué le falta para tu despacho?",
    texto:
      "kitlegal ya sirve para consultar y citar la legislación consolidada del BOE. Estamos trabajando para que sirva al trabajo diario de un despacho y, como no recoge ningún dato de uso, la única forma de saber qué necesitas es que nos lo cuentes.",
    peticiones: [
      "Una consulta que tu asistente no supo responder, o que respondió mal.",
      "Una fuente que echas de menos: jurisprudencia, consultas de la DGT, convenios colectivos, boletines autonómicos.",
      "Lo que te frena para usarlo: la instalación, la confidencialidad, la forma de la respuesta.",
    ],
    todavia: [
      "No redacta escritos todavía. Lo hará: está en el plan.",
      "No consulta jurisprudencia. La del Tribunal Supremo, la Audiencia Nacional, los tribunales superiores de justicia y las audiencias provinciales está en el CENDOJ, del Consejo General del Poder Judicial, y sus condiciones de uso no permiten consultarlo de forma masiva ni automatizada. Mientras no cambien, kitlegal no lo hará.",
      "Lo que sí puede llegar, por vías que lo permiten: las sentencias del Tribunal Constitucional y las del Supremo que anulan una disposición, que se publican en el BOE, y la jurisprudencia europea.",
    ],
    detalle: { texto: "Por qué, con detalle", destino: `${REPOSITORIO}#y-las-sentencias` },
    aviso:
      "No incluyas datos de clientes ni de asuntos. Leemos tu mensaje solo para responderte y para decidir qué mejorar.",
    asunto: "kitlegal para despachos",
  },
  cta: {
    titulo: "Dota a tu despacho de la ley vigente, con su cita.",
    texto:
      "Menos riesgo de citar una norma derogada o un artículo que no dice lo que se cree. Software libre y gratuito, para juristas exigentes.",
  },
};

// Las audiencias en el orden en que se ofrecen: la portada del sitio, primero.
export const audiencias = [ciudadania, despachos];
