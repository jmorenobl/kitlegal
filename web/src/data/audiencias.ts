// Los textos de las dos portadas: una URL por audiencia, cada una con su
// título, su descripción y sus ejemplos (ADR 0024). Las citas se nombran por su
// id de citas.yaml: el texto de la norma nunca se escribe aquí.

import { formatoEntero, numeroDeMunicipios } from "../lib/proyecto";

export type Tono = "sello" | "derogada" | "agente" | "vigente";

export interface Audiencia {
  ruta: string;
  titulo: string;
  descripcion: string;
  imagen: string;
  h1: string;
  subtitulo: string;
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
  casos: { icono: string; tono: Tono; etiqueta: string; pregunta: string; cita: string }[];
  cta: { titulo: string; texto: string };
}

const municipios = formatoEntero(numeroDeMunicipios());

export const despachos: Audiencia = {
  ruta: "/",
  titulo: "kitlegal: IA para despachos con la ley vigente del BOE",
  descripcion:
    "Skills para Claude Code, Codex o Antigravity que responden con el texto consolidado del BOE: artículo exacto, aviso de vigencia y fuente verificable.",
  imagen: "inicio",
  h1: "Usa IA en tu despacho sin miedo a normas derogadas ni artículos inventados.",
  subtitulo:
    "kitlegal conecta tu asistente de IA (Claude Code, Codex, Antigravity) con el texto consolidado del BOE. Cada respuesta llega con el artículo exacto, su vigencia y la fuente oficial para comprobarla.",
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
      etiqueta: "Territorio",
      valor: `${municipios} municipios`,
      texto: "Sitúa cualquier municipio del INE en su provincia, su comunidad y su régimen.",
      tono: "agente",
    },
    {
      etiqueta: "Software libre",
      valor: "EUPL-1.2",
      texto: "El programa corre en tu equipo: sin cuentas y sin servidores de kitlegal entre tú y el BOE.",
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
    },
  ],
  cta: {
    titulo: "Dota a tu despacho de la ley vigente, con su cita.",
    texto:
      "Menos riesgo de citar una norma derogada o un artículo que no dice lo que se cree. Software libre para juristas exigentes.",
  },
};

export const ciudadania: Audiencia = {
  ruta: "/ciudadania/",
  titulo: "kitlegal: consulta tus derechos con la ley vigente del BOE",
  descripcion:
    "Pregunta a tu asistente de IA por plazos, multas, alquiler o impuestos y recibe la respuesta con el texto vigente del BOE y la cita para comprobarla.",
  imagen: "ciudadania",
  h1: "Consulta tus derechos y trámites ante la Administración con la ley real en la mano.",
  subtitulo:
    "Pregunta a tu asistente de IA por plazos de recursos, multas, alquileres o impuestos sabiendo que responde con el texto vigente del BOE, y con la cita para que lo compruebes.",
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
      etiqueta: "Tu municipio",
      valor: `${municipios} municipios`,
      texto: "Sitúa tu municipio en su provincia y su comunidad con los datos del INE.",
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
  casos: [
    {
      icono: "badge",
      tono: "vigente",
      etiqueta: "Trabajo",
      pregunta: "¿Cuánto preaviso tengo que dar si dejo mi trabajo?",
      cita: "et-49",
    },
    {
      icono: "home",
      tono: "sello",
      etiqueta: "Alquiler y vivienda",
      pregunta: "¿Qué pasa si el casero tarda en devolverme la fianza?",
      cita: "lau-36",
    },
    {
      icono: "account-balance",
      tono: "agente",
      etiqueta: "Trámites con la Administración",
      pregunta: "Si la Administración no contesta a mi solicitud, ¿se entiende concedida?",
      cita: "lpacap-24",
    },
    {
      icono: "child-friendly",
      tono: "sello",
      etiqueta: "Impuestos y familia",
      pregunta: "¿Qué deducción por maternidad recoge el IRPF?",
      cita: "lirpf-81",
    },
  ],
  cta: {
    titulo: "Empieza a consultar tus derechos con la ley vigente.",
    texto:
      "Ten a mano la legislación del BOE en tus gestiones con la Administración, en el trabajo o con tu vivienda. Software libre y gratuito.",
  },
};
