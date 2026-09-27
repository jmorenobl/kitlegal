import type { APIRoute } from "astro";
import { SITIO } from "../lib/proyecto";

// Todo el sitio es público, también para los rastreadores de los buscadores con
// IA: kitlegal es para asistentes de IA, y que lo encuentren es el propósito.
export const GET: APIRoute = () =>
  new Response(`User-agent: *\nAllow: /\n\nSitemap: ${SITIO}/sitemap-index.xml\n`, {
    headers: { "Content-Type": "text/plain; charset=utf-8" },
  });
