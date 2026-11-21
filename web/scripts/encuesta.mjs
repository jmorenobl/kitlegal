// Escribe en la salida estándar el guion de Google Apps Script que crea el
// formulario de Google de la encuesta de /plan/, con las preguntas de
// src/data/encuesta.ts. Se pega en https://script.google.com, se ejecuta la
// función crearEncuesta y el registro de ejecución devuelve el JSON que va en
// src/data/encuesta-formulario.json, y la hoja de cálculo de las respuestas.
//
// Uso: pnpm --dir web encuesta > encuesta.gs
// Node lee el .ts sin compilarlo (web/.node-version: Node 24).

import { bloques } from "../src/data/encuesta.ts";

const guion = `// Crea el formulario de la encuesta del plan de kitlegal. Generado por
// web/scripts/encuesta.mjs: no editar; se cambian las preguntas en
// web/src/data/encuesta.ts y se vuelve a generar.
function crearEncuesta() {
  const bloques = ${JSON.stringify(bloques, null, 2)};
  const form = FormApp.create("kitlegal · revisión del plan");
  form.setDescription("Encuesta anónima sobre el plan de kitlegal (https://kitlegal.es/plan/). Las respuestas llegan desde la web.");
  // Anónima: ni correo, ni inicio de sesión, ni límite por persona, que lo exigiría.
  form.setCollectEmail(false);
  try { form.setRequireLogin(false); } catch (e) {}
  form.setLimitOneResponsePerUser(false);
  form.setAllowResponseEdits(false);
  form.setShowLinkToRespondAgain(false);
  form.setConfirmationMessage("Gracias. Tus respuestas han llegado a kitlegal.");
  const hoja = SpreadsheetApp.create("kitlegal · respuestas a la encuesta del plan");
  form.setDestination(FormApp.DestinationType.SPREADSHEET, hoja.getId());
  const campos = {};
  for (const bloque of bloques) {
    form.addSectionHeaderItem().setTitle(bloque.titulo).setHelpText(bloque.texto);
    for (const p of bloque.preguntas) {
      let respuesta;
      if (p.tipo === "una") {
        respuesta = form.addMultipleChoiceItem().setTitle(p.titulo).setChoiceValues(p.opciones).createResponse(p.opciones[0]);
      } else if (p.tipo === "varias") {
        respuesta = form.addCheckboxItem().setTitle(p.titulo).setChoiceValues(p.opciones).createResponse([p.opciones[0]]);
      } else {
        const item = form.addParagraphTextItem().setTitle(p.titulo);
        if (p.ayuda) item.setHelpText(p.ayuda);
        respuesta = item.createResponse("x");
      }
      const url = form.createResponse().withItemResponse(respuesta).toPrefilledUrl();
      campos[p.clave] = "entry." + /entry\\.(\\d+)=/.exec(url)[1];
    }
  }
  const accion = form.getPublishedUrl().replace(/\\/viewform.*$/, "/formResponse");
  Logger.log("Copia esto en web/src/data/encuesta-formulario.json:\\n" + JSON.stringify({ accion: accion, campos: campos }, null, 2));
  Logger.log("Formulario: " + form.getEditUrl());
  Logger.log("Hoja de respuestas: " + hoja.getUrl());
}
`;

process.stdout.write(guion);
