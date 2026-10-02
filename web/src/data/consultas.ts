// Las páginas de /consultas/: una pregunta tal como se escribe en un buscador y
// lo que dice la norma vigente, punto por punto. Cada afirmación va con la cita
// de citas.yaml que la sostiene, y la página la muestra al lado: aquí no se
// escribe el texto de ninguna norma (ADR 0024), solo lo que la cita dice con
// otras palabras. Lo que la norma citada no dice va en `limites`, que hablan
// de la página y no afirman nada de otra norma.

import type { Punto } from "../lib/clases";

export interface Consulta {
  slug: string;
  // Título de la pestaña y de los resultados de búsqueda (≤ 60 caracteres).
  titulo: string;
  // Descripción de los resultados de búsqueda (≤ 160 caracteres).
  descripcion: string;
  etiqueta: string;
  // El color del punto que acompaña a la etiqueta en el índice.
  punto: Punto;
  // La fotografía de la consulta (src/assets/ilustraciones/) y su descripción.
  ilustracion: string;
  alt: string;
  pregunta: string;
  // La respuesta corta, sostenida por `cita`.
  respuesta: string;
  cita: string;
  puntos: { titulo: string; texto: string; cita: string }[];
  limites: string[];
  // Una forma de preguntárselo al asistente con kitlegal instalado.
  ejemplo: string;
}

// Lo que la disposición adicional primera de la Ley 39/2015 saca del
// procedimiento común: va en las dos consultas que citan esa ley.
const materiasAparte = {
  titulo: "No vale para todas las materias",
  cita: "lpacap-da1",
};

export const consultas: Consulta[] = [
  {
    slug: "plazo-recurso-de-alzada",
    titulo: "Plazo del recurso de alzada: un mes (art. 122 Ley 39/2015)",
    descripcion:
      "Un mes para recurrir en alzada una resolución expresa. Cómo se cuenta, qué pasa si no te contestaron y cuánto tarda la respuesta, con el texto vigente del BOE.",
    etiqueta: "Recursos administrativos",
    punto: "agente",
    ilustracion: "consulta-recurso-de-alzada",
    alt: "Unas manos abren un sobre con un abrecartas; al fondo, un reloj de mesa.",
    pregunta: "¿Qué plazo tengo para presentar un recurso de alzada?",
    respuesta:
      "Un mes, si la resolución que recurres es expresa. Si dejas pasar ese mes sin recurrir, la resolución queda firme.",
    cita: "lpacap-122",
    puntos: [
      {
        titulo: "Cómo se cuenta el mes",
        texto:
          "Empieza el día siguiente a la notificación y termina, en el mes siguiente, el mismo día en que te notificaron. Si ese último día es inhábil, el plazo pasa al primer día hábil siguiente.",
        cita: "lpacap-30",
      },
      {
        titulo: "Si lo que recurres es que no te contestaron",
        texto:
          "Entonces no hay un mes: puedes recurrir en alzada en cualquier momento a partir del día siguiente a aquel en que el silencio de la Administración produce efectos.",
        cita: "lpacap-122-presunto",
      },
      {
        titulo: "Qué se recurre en alzada y ante quién",
        texto:
          "Las resoluciones y los actos que no ponen fin a la vía administrativa, ante el órgano superior jerárquico del que los dictó.",
        cita: "lpacap-121",
      },
      {
        titulo: "Cuánto tarda la respuesta",
        texto:
          "La Administración tiene tres meses para resolver y notificar. Pasado ese plazo sin respuesta, puedes entender desestimado el recurso, con la excepción que señala el propio artículo.",
        cita: "lpacap-122-resolucion",
      },
      {
        ...materiasAparte,
        texto:
          "Los tributos, la Seguridad Social y el desempleo, las sanciones tributarias, del orden social y de tráfico, y la extranjería y el asilo se rigen por su normativa específica, y esta ley se les aplica solo de forma supletoria. Si tu resolución es de una de esas materias, mira primero su norma: el recurso o el plazo pueden ser otros.",
      },
    ],
    limites: [
      "Esta página no te dice si tu resolución pone fin a la vía administrativa ni cuál es el órgano superior jerárquico: depende de quién la dictó.",
      "Tampoco calcula el último día de tu plazo: para eso hacen falta la fecha de tu notificación y el calendario de días inhábiles.",
    ],
    ejemplo: "Me han notificado una resolución y quiero recurrirla en alzada: ¿qué plazo tengo y desde cuándo se cuenta?",
  },
  {
    slug: "silencio-administrativo",
    titulo: "La Administración no contesta: el silencio administrativo",
    descripcion:
      "Si la Administración no responde a tu solicitud en plazo, la regla general es que se entiende estimada, con excepciones. Art. 24 de la Ley 39/2015, vigente.",
    etiqueta: "Trámites con la Administración",
    punto: "agente",
    ilustracion: "consulta-silencio-administrativo",
    alt: "Una ventanilla de atención al ciudadano con la persiana a medio bajar y una silla vacía delante.",
    pregunta: "Presenté una solicitud y la Administración no contesta. ¿Qué pasa?",
    respuesta:
      "Si vence el plazo máximo sin que te notifiquen una resolución, la regla general es que puedes entender tu solicitud estimada por silencio administrativo, salvo que una ley o una norma europea o internacional diga lo contrario.",
    cita: "lpacap-24",
    puntos: [
      {
        titulo: "Cuál es el plazo máximo",
        texto:
          "El que fije la norma de tu procedimiento. Si no fija ninguno, tres meses, contados desde que tu solicitud entró en el registro electrónico de la Administración competente.",
        cita: "lpacap-21-computo",
      },
      {
        titulo: "Cuándo el silencio es negativo",
        texto:
          "El silencio desestima en el derecho de petición, en lo que daría al solicitante o a terceros facultades sobre el dominio público o el servicio público, en las actividades que puedan dañar el medio ambiente y en la responsabilidad patrimonial. También en los recursos y en la revisión de oficio.",
        cita: "lpacap-24-desestimatorio",
      },
      {
        titulo: "Para qué sirve cada silencio",
        texto:
          "El silencio que estima vale como un acto que termina el procedimiento. El que desestima sirve solo para una cosa: abrirte la puerta del recurso administrativo o del contencioso-administrativo.",
        cita: "lpacap-24-efectos",
      },
      {
        titulo: "Cómo se demuestra",
        texto:
          "El silencio se puede hacer valer ante la Administración y ante cualquier persona, y se acredita por cualquier medio de prueba, incluido un certificado que el órgano competente debe expedir de oficio en quince días. También puedes pedirlo tú en cualquier momento.",
        cita: "lpacap-24-certificado",
      },
      {
        ...materiasAparte,
        texto:
          "Los tributos, la Seguridad Social y el desempleo, las sanciones tributarias, del orden social y de tráfico, y la extranjería y el asilo se rigen por su normativa específica, y esta ley se les aplica solo de forma supletoria. Ahí el plazo y el sentido del silencio pueden ser otros.",
      },
    ],
    limites: [
      "Esta página no te dice cuál es el plazo máximo de tu procedimiento ni si una ley le da al silencio sentido negativo: depende de la norma que lo regule.",
    ],
    ejemplo: "Presenté una solicitud en mi ayuntamiento hace cuatro meses y no han contestado. ¿Qué dice la ley?",
  },
  {
    slug: "devolucion-fianza-alquiler",
    titulo: "Mi casero no me devuelve la fianza: qué dice la ley (LAU)",
    descripcion:
      "Pasado un mes desde que entregas las llaves, la fianza que te deban devolver genera el interés legal. Art. 36 de la Ley de Arrendamientos Urbanos, vigente.",
    etiqueta: "Alquiler y vivienda",
    punto: "vigente",
    ilustracion: "consulta-fianza-alquiler",
    alt: "Una mano deja unas llaves sobre una mesa, junto a una casita de madera y unas monedas.",
    pregunta: "Mi casero no me devuelve la fianza. ¿Qué dice la ley?",
    respuesta:
      "Que el saldo de la fianza que te tenga que devolver genera el interés legal cuando ha pasado un mes desde que entregaste las llaves sin que te lo haya devuelto.",
    cita: "lau-36",
    puntos: [
      {
        titulo: "Cuánto es la fianza",
        texto:
          "Una mensualidad de renta en el alquiler de vivienda y dos en el alquiler para uso distinto del de vivienda, en metálico.",
        cita: "lau-36-cuantia",
      },
      {
        titulo: "Al entregar las llaves: el documento de finalización",
        texto:
          "Al terminar el contrato, las dos partes debéis dejar por escrito el estado de la vivienda en un documento de finalización firmado por ambas. Si no se firma, o si no recoge ningún desperfecto, se presume, salvo prueba en contrario, que entregaste la vivienda en un estado adecuado de conservación.",
        cita: "lau-36-finalizacion",
      },
      {
        titulo: "Si te pidieron algo más que la fianza",
        texto:
          "El contrato puede incluir una garantía además de la fianza, pero no se te puede exigir un seguro de impago ni una cobertura parecida. En el alquiler de vivienda, en contratos de hasta cinco años (siete si el arrendador es una persona jurídica), esa garantía no puede pasar de dos mensualidades de renta; en los arrendamientos temporales, de una.",
        cita: "lau-36-garantia",
      },
      {
        titulo: "De cuándo es esta redacción",
        texto:
          "El apartado 5 se modificó y el 7 se añadió con el Real Decreto-ley 26/2026, de 29 de septiembre. Es una reforma muy reciente.",
        cita: "lau-36-nota",
      },
    ],
    limites: [
      "Esta página no te dice cómo afecta la reforma a un contrato firmado antes: lo regulan sus disposiciones transitorias, que esta página no cita.",
      "Tampoco te dice qué puede descontar tu casero de la fianza ni cómo reclamarla: el artículo 36 no lo regula.",
      "Tampoco cubre el depósito de la fianza ante la comunidad autónoma: solo cita el artículo 36 de la ley estatal.",
    ],
    ejemplo:
      "Entregué las llaves de mi piso hace dos meses y el casero no me ha devuelto la fianza. ¿Qué dice la Ley de Arrendamientos Urbanos?",
  },
  {
    slug: "preaviso-baja-voluntaria",
    titulo: "Preaviso de baja voluntaria: ¿son 15 días? Qué dice la ley",
    descripcion:
      "El Estatuto de los Trabajadores no fija los días de preaviso para dejar tu trabajo: remite a tu convenio o a la costumbre del lugar. Art. 49.1.d), vigente.",
    etiqueta: "Trabajo",
    punto: "sello",
    ilustracion: "consulta-preaviso-baja-voluntaria",
    alt: "Una caja de cartón con una planta sobre una mesa de oficina, junto a una silla vacía.",
    pregunta: "Quiero dejar mi trabajo. ¿Cuánto preaviso tengo que dar?",
    respuesta:
      "El Estatuto de los Trabajadores no fija un número de días: dice que debes dar el preaviso que señale tu convenio colectivo o la costumbre del lugar.",
    cita: "et-49",
    puntos: [
      {
        titulo: "De dónde salen los quince días",
        texto:
          "El mismo artículo habla de quince días, pero para otro caso: cuando un contrato de duración determinada de más de un año se acaba, la parte que lo comunica debe avisar a la otra con esa antelación mínima. No es la regla de la dimisión.",
        cita: "et-49-denuncia",
      },
      {
        titulo: "Al firmar el finiquito",
        texto:
          "Puedes pedir que esté presente un representante legal de los trabajadores, y en el recibo debe constar que firmaste en su presencia o que no usaste esa posibilidad.",
        cita: "et-49-finiquito",
      },
    ],
    limites: [
      "Esta página no te dice cuántos días fija tu convenio ni qué consecuencia tiene no dar el preaviso: la letra d) del artículo 49.1 no lo regula.",
      "kitlegal lee hoy la legislación consolidada del BOE; todavía no lee convenios colectivos.",
    ],
    ejemplo: "Quiero dejar mi trabajo. ¿Qué dice el Estatuto de los Trabajadores sobre el preaviso?",
  },
  {
    slug: "prescripcion-deudas-hacienda",
    titulo: "¿Cuándo prescribe una deuda con Hacienda? Cuatro años",
    descripcion:
      "Hacienda tiene cuatro años para liquidar una deuda tributaria y cuatro para cobrarla, pero el plazo se interrumpe y vuelve a empezar. Arts. 66 a 68 de la LGT.",
    etiqueta: "Hacienda",
    punto: "cian",
    ilustracion: "consulta-deudas-hacienda",
    alt: "Un reloj de arena sobre una mesa, junto a un taco de papel y unas monedas.",
    pregunta: "¿Cuánto tiempo tiene Hacienda para reclamarme una deuda?",
    respuesta:
      "Cuatro años. Es el plazo de prescripción del derecho de la Administración a determinar la deuda tributaria con una liquidación, y también del derecho a exigir el pago de las deudas ya liquidadas o autoliquidadas.",
    cita: "lgt-66",
    puntos: [
      {
        titulo: "Desde cuándo se cuentan",
        texto:
          "Para liquidar, desde el día siguiente a aquel en que termina el plazo para presentar la declaración o la autoliquidación. Para cobrar, desde el día siguiente al fin del plazo de pago en período voluntario.",
        cita: "lgt-67",
      },
      {
        titulo: "El plazo se interrumpe y vuelve a empezar",
        texto:
          "Cualquier actuación de la Administración tributaria dirigida de forma efectiva a recaudar la deuda, hecha con tu conocimiento formal, interrumpe la prescripción del cobro, y tras la interrupción los cuatro años empiezan a contar de nuevo. Por eso una deuda de hace más de cuatro años no está prescrita necesariamente.",
        cita: "lgt-68",
      },
      {
        titulo: "A ti te pasa lo mismo",
        texto:
          "También prescribe a los cuatro años tu derecho a solicitar una devolución, por ejemplo la de un ingreso indebido.",
        cita: "lgt-66-devoluciones",
      },
    ],
    limites: [
      "Esta página habla solo de deudas tributarias: no cubre las deudas con la Seguridad Social ni las multas de tráfico.",
      "Cita la Ley General Tributaria estatal, no las normas forales de Navarra y del País Vasco.",
      "El artículo 68 recoge más causas de interrupción que la que se muestra aquí.",
    ],
    ejemplo: "Hacienda me reclama ahora un impuesto de hace seis años. ¿Puede estar prescrito?",
  },
  {
    slug: "deduccion-maternidad-irpf",
    titulo: "Deducción por maternidad en el IRPF: hasta 1.200 € al año",
    descripcion:
      "Hasta 1.200 euros al año por cada hijo menor de tres años, y hasta 1.000 más por gastos de guardería. Requisitos del art. 81 de la Ley del IRPF, texto vigente.",
    etiqueta: "Impuestos y familia",
    punto: "cian",
    ilustracion: "consulta-deduccion-maternidad",
    alt: "Una cuna de madera con una manta de punto, un conejo de peluche y una hucha.",
    pregunta: "¿Qué deducción por maternidad recoge el IRPF?",
    respuesta:
      "Hasta 1.200 euros al año por cada hijo menor de tres años. Es para las mujeres con derecho al mínimo por descendientes que, al nacer el menor, cobren una prestación por desempleo o estén de alta en la Seguridad Social o en una mutualidad, o que se den de alta después y lleguen a 30 días cotizados.",
    cita: "lirpf-81-completo",
    puntos: [
      {
        titulo: "Hasta 1.000 euros más por guardería",
        texto:
          "La deducción puede aumentar hasta en 1.000 euros si has pagado gastos de custodia del menor de tres años en guarderías o centros de educación infantil autorizados.",
        cita: "lirpf-81-guarderia",
      },
      {
        titulo: "Se puede cobrar por adelantado",
        texto:
          "Puedes pedir a la Agencia Tributaria que te abone la deducción de forma anticipada; en ese caso ya no se resta de la cuota del impuesto.",
        cita: "lirpf-81-anticipo",
      },
    ],
    limites: [
      "El artículo 81 regula también la adopción y el acogimiento, y los casos en que la deducción pasa al padre o a un tutor: esta página no los resume.",
      "El procedimiento y las condiciones se regulan en un reglamento, que esta página no cita.",
      "Cita la ley estatal del IRPF, no las normas forales de Navarra y del País Vasco ni las deducciones de cada comunidad autónoma.",
    ],
    ejemplo: "Tengo un hijo de dos años y trabajo por cuenta ajena. ¿Qué dice la ley del IRPF sobre la deducción por maternidad?",
  },
];

export function consulta(slug: string): Consulta {
  const encontrada = consultas.find((c) => c.slug === slug);
  if (!encontrada) throw new Error(`consultas: no hay ninguna consulta con slug ${slug}`);
  return encontrada;
}

export function rutaDeConsulta(slug: string): string {
  return `/consultas/${consulta(slug).slug}/`;
}
