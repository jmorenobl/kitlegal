// La imagen de cada página al compartirla: /og/<imagen>.png, con el titular de
// la propia página.
import type { APIRoute, GetStaticPaths } from "astro";
import { audiencias } from "../../data/audiencias";
import { consultas } from "../../data/consultas";
import { imagenDeConsulta, imagenParaCompartir } from "../../lib/imagenes";

const imagenes: Record<string, { titulo: string; subtitulo: string }> = {
  ...Object.fromEntries(audiencias.map(({ imagen, h1, nombre }) => [imagen, { titulo: h1, subtitulo: nombre }])),
  instalar: { titulo: "Instalar kitlegal", subtitulo: "Paso a paso, en macOS, Linux y Windows" },
  consultas: { titulo: "Consultas habituales, con su cita", subtitulo: "La pregunta y el artículo vigente del BOE que la responde" },
  ...Object.fromEntries(
    consultas.map(({ slug, pregunta, etiqueta }) => [
      imagenDeConsulta(slug),
      { titulo: pregunta, subtitulo: `${etiqueta} · con la cita del BOE` },
    ]),
  ),
  bot: { titulo: "El rastreador de kitlegal", subtitulo: "Qué es, cómo se comporta con tu sitio y cómo limitarlo" },
};

export const getStaticPaths: GetStaticPaths = () =>
  Object.entries(imagenes).map(([imagen, textos]) => ({ params: { imagen }, props: textos }));

export const GET: APIRoute = async ({ props }) => {
  const { titulo, subtitulo } = props as { titulo: string; subtitulo: string };
  return new Response(await imagenParaCompartir(titulo, subtitulo), { headers: { "Content-Type": "image/png" } });
};
