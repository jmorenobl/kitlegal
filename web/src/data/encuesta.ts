// La encuesta de /plan/, para que un jurista diga si el plan es correcto y qué
// necesita. Es anónima: no pide nombre, correo ni ningún dato de clientes.
//
// Es la única fuente de las preguntas. `node web/scripts/encuesta.mjs` escribe
// con ellas el guion de Google Apps Script que crea el formulario de Google, y
// el guion devuelve lo que va en encuesta-formulario.json: la dirección a la
// que se envían las respuestas, el campo de cada pregunta y la huella de estas
// preguntas. Si se cambia una pregunta o una opción, la huella deja de
// coincidir, la construcción falla y hay que volver a crear el formulario.
//
// Sin importaciones: el guion de la encuesta lo lee con Node, sin Astro.

// huellaDe da la huella de las preguntas que guarda encuesta-formulario.json.
// La calculan igual el guion y la página: sha256 del JSON de los bloques.
export async function huellaDe(b: Bloque[]): Promise<string> {
  const datos = new TextEncoder().encode(JSON.stringify(b));
  const resumen = await crypto.subtle.digest("SHA-256", datos);
  return `sha256:${Array.from(new Uint8Array(resumen), (x) => x.toString(16).padStart(2, "0")).join("")}`;
}

export type Tipo = "una" | "varias" | "texto";

export interface Pregunta {
  clave: string;
  tipo: Tipo;
  titulo: string;
  ayuda?: string;
  opciones?: string[];
}

export interface Bloque {
  titulo: string;
  texto: string;
  preguntas: Pregunta[];
}

// Sin JavaScript, una respuesta marcada no se desmarca: la última opción es la
// de quien no quiere opinar.
const VALORACION = ["Correcto", "Con matices", "Incorrecto", "No lo sé / no es mi área"];
const ACUERDO = ["De acuerdo", "Con matices", "En desacuerdo", "No lo sé / no es mi área"];
const SIN_DATOS = "No incluyas datos de clientes ni de asuntos.";

export const bloques: Bloque[] = [
  {
    titulo: "Sobre ti",
    texto: "Para saber desde dónde se mira el plan. No pedimos nada que te identifique.",
    preguntas: [
      {
        clave: "perfil",
        tipo: "una",
        titulo: "¿Dónde trabajas?",
        opciones: [
          "Despacho individual",
          "Despacho de 2 a 10 personas",
          "Despacho de más de 10 personas",
          "Asesoría o gestoría",
          "Empresa",
          "Administración pública",
          "Otro",
        ],
      },
      {
        clave: "areas",
        tipo: "varias",
        titulo: "¿En qué áreas trabajas?",
        opciones: [
          "Administrativo",
          "Civil",
          "Consumo y seguros",
          "Familia",
          "Fiscal y tributario",
          "Laboral",
          "Mercantil",
          "Penal",
          "Tráfico",
          "Otra",
        ],
      },
      {
        clave: "primero",
        tipo: "una",
        titulo: "De lo que viene, ¿qué te sería más útil primero?",
        opciones: [
          "Plazos procesales civiles y sociales",
          "Revisar las citas de un escrito",
          "Sentencias del Tribunal Constitucional",
          "Seguimiento de expedientes",
          "Escritos listos para firmar",
          "Normativa europea",
          "Doctrina de la DGT y del TEAC",
          "Ordenanzas y boletines",
        ],
      },
    ],
  },
  {
    titulo: "Lo que hace hoy",
    texto: "Cómo se comporta ya el asistente.",
    preguntas: [
      {
        clave: "cendoj",
        tipo: "una",
        titulo:
          "Las sentencias no se buscan de forma automática: el asistente prepara la búsqueda, la persona la hace en el CENDOJ con su navegador y trae el documento, y nada se dice de una sentencia sin tenerla delante. ¿Es un modo de trabajo aceptable para un despacho?",
        opciones: ACUERDO,
      },
      {
        clave: "aparte",
        tipo: "una",
        titulo:
          "La LPAC remite a su normativa específica (DA 1ª.2) o a procedimientos propios (art. 112.2 y 112.4) estos casos: la aplicación de los tributos y su revisión; la Seguridad Social y el desempleo; los sancionadores tributario y aduanero, del orden social, de tráfico y de extranjería; la extranjería y el asilo; las impugnaciones que sustituyen al recurso de alzada (art. 112.2) y las reclamaciones económico-administrativas (art. 112.4). En todos ellos el asistente no aplica los plazos del procedimiento común y lo declara no cubierto. ¿La lista es completa y correcta?",
        opciones: VALORACION,
      },
      {
        clave: "festivos",
        tipo: "una",
        titulo:
          "Un plazo administrativo se calcula con los festivos nacionales, autonómicos y locales del municipio del interesado y del de la sede del órgano (art. 30.6 LPAC), y avisa cuando el territorio no está cubierto. ¿Es correcto?",
        opciones: VALORACION,
      },
      { clave: "nota-hoy", tipo: "texto", titulo: "¿Algo que matizar de lo que hace hoy?", ayuda: SIN_DATOS },
    ],
  },
  {
    titulo: "En qué se apoya lo que viene",
    texto: "Lo que el plan da por hecho de cada precepto. Te pedimos que confirmes la lectura.",
    preguntas: [
      {
        clave: "lec",
        tipo: "una",
        titulo:
          "En el proceso civil son inhábiles, además de los sábados, domingos y festivos, los días de agosto y del 24 de diciembre al 6 de enero (art. 130.2 LEC). El día del vencimiento cuenta hasta las veinticuatro horas, los plazos por meses se cuentan de fecha a fecha y, si el último día es inhábil, se prorroga al siguiente hábil (art. 133 LEC). ¿Es correcto?",
        opciones: VALORACION,
      },
      {
        clave: "lrjs",
        tipo: "una",
        titulo:
          "En lo social, agosto y los días del 24 de diciembre al 6 de enero son inhábiles, salvo en las modalidades procesales que enumera el art. 43.4 LRJS (el despido, entre otras), en las que esos días cuentan. Si la pregunta no dice la modalidad, el asistente la pide y no calcula. ¿Es correcto?",
        opciones: VALORACION,
      },
      {
        clave: "lgt",
        tipo: "una",
        titulo:
          "Lo tributario empieza por la revisión en vía administrativa: el recurso de reposición (art. 223 LGT) y la reclamación económico-administrativa (art. 235 LGT). Las sanciones, las aduanas, las haciendas forales, el calendario del contribuyente y los modelos quedan para después. ¿Es razonable empezar por ahí?",
        opciones: ACUERDO,
      },
      {
        clave: "dgt",
        tipo: "una",
        titulo:
          "La contestación a una consulta tributaria escrita vincula a los órganos de la Administración tributaria encargados de la aplicación de los tributos, en los términos del art. 89 LGT, y lo que no es vinculante se dice. La doctrina se cita siempre separada de la norma y de la jurisprudencia. ¿Es correcto?",
        opciones: VALORACION,
      },
      {
        clave: "escritos",
        tipo: "una",
        titulo:
          "Los primeros escritos: la solicitud de acceso a la información (art. 17 LTAIBG), el recurso de alzada o el potestativo de reposición (arts. 112 y 121 a 124 LPAC), según la notificación diga si el acto pone fin a la vía administrativa (art. 40.2), y las alegaciones (arts. 76 y 82 LPAC), con el plazo de audiencia que fije el órgano, entre diez y quince días (art. 82.2). Se firman como interesado o como representante (art. 5 LPAC) y pueden presentarse en el Registro Electrónico General de la AGE, dirigidos a cualquier administración (art. 16.4 LPAC). ¿Es correcto?",
        opciones: VALORACION,
      },
      {
        clave: "ordenanzas",
        tipo: "una",
        titulo:
          "Las ordenanzas se publican en el Boletín Oficial de la Provincia (art. 70.2 LRBRL), y las fiscales, íntegras, en el de la provincia o, en su caso, en el de la comunidad autónoma uniprovincial (art. 17.4 TRLRHL). En una comunidad uniprovincial, como Madrid, el plan busca todas las ordenanzas en el boletín autonómico. El asistente presentará la ordenanza aprobada y sus modificaciones citando cada anuncio, sin decir que es un texto consolidado. ¿Es correcto?",
        opciones: VALORACION,
      },
      {
        clave: "tc",
        tipo: "una",
        titulo:
          "Todas las sentencias del Tribunal Constitucional se publican en el BOE, y por eso pueden leerse sin que la persona traiga el documento. ¿Es correcto?",
        opciones: VALORACION,
      },
      { clave: "nota-apoyos", tipo: "texto", titulo: "¿Algo que corregir de estos preceptos?", ayuda: SIN_DATOS },
    ],
  },
  {
    titulo: "Prioridades",
    texto: "Decisiones del plan, no de la ley. Aquí cuenta tu criterio de profesional.",
    preguntas: [
      {
        clave: "orden",
        tipo: "una",
        titulo:
          "Primero se termina la base, con lo tributario, y después llega lo que usa un despacho a diario: plazos procesales, revisión de escritos y el Constitucional. ¿Te parece el orden adecuado?",
        opciones: ACUERDO,
      },
      {
        clave: "falta",
        tipo: "una",
        titulo: "¿Lo previsto para despachos en la etapa 2 cubre lo que un despacho pequeño necesita a diario?",
        opciones: ACUERDO,
      },
      {
        clave: "nunca",
        tipo: "una",
        titulo:
          "Lo que nunca hará: consultar el CENDOJ de forma automática, presentar o notificar en nombre de nadie y enviar las consultas a un servidor de kitlegal. ¿De acuerdo con esos límites?",
        opciones: ACUERDO,
      },
      { clave: "general", tipo: "una", titulo: "En conjunto, ¿das el visto bueno al plan?", opciones: ACUERDO },
      {
        clave: "comentario",
        tipo: "texto",
        titulo: "¿Qué falta, qué sobra o qué cambiarías?",
        ayuda: SIN_DATOS,
      },
    ],
  },
];
