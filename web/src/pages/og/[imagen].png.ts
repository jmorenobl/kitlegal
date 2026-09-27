// La imagen de cada página al compartirla: /og/<imagen>.png, con el titular de
// la propia página.
import type { APIRoute, GetStaticPaths } from "astro";
import { ciudadania, despachos } from "../../data/audiencias";
import { imagenParaCompartir } from "../../lib/imagenes";

const imagenes = {
  inicio: { titulo: despachos.h1, subtitulo: "Para despachos y abogados" },
  ciudadania: { titulo: ciudadania.h1, subtitulo: "Para ciudadanos y trámites" },
  instalar: { titulo: "Instalar kitlegal", subtitulo: "Dos órdenes: el programa y sus skills, en macOS, Linux y Windows" },
  bot: { titulo: "El rastreador de kitlegal", subtitulo: "Qué es, cómo se comporta con tu sitio y cómo limitarlo" },
};

export const getStaticPaths: GetStaticPaths = () =>
  Object.entries(imagenes).map(([imagen, textos]) => ({ params: { imagen }, props: textos }));

export const GET: APIRoute = async ({ props }) => {
  const { titulo, subtitulo } = props as { titulo: string; subtitulo: string };
  return new Response(await imagenParaCompartir(titulo, subtitulo), { headers: { "Content-Type": "image/png" } });
};
