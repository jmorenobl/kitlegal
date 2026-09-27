// Web de kitlegal (ADR 0024): HTML estático en https://kitlegal.es, publicado
// en GitHub Pages por .github/workflows/web.yml.
import sitemap from "@astrojs/sitemap";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "astro/config";

export default defineConfig({
  site: "https://kitlegal.es",
  // GitHub Pages sirve /ruta/ desde ruta/index.html y redirige /ruta a
  // /ruta/: la URL canónica lleva siempre la barra final.
  trailingSlash: "always",
  build: {
    format: "directory",
    // Ninguna hoja de estilos en línea: la política de seguridad de contenido
    // de la página solo admite estilos propios servidos como fichero.
    inlineStylesheets: "never",
  },
  devToolbar: { enabled: false },
  integrations: [sitemap({ filter: (pagina) => !pagina.endsWith("/404/") })],
  vite: { plugins: [tailwindcss()] },
});
