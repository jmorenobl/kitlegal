// Las fotografías de la web, en src/assets/ilustraciones/: se nombran por su
// fichero sin extensión y Astro las convierte y las redimensiona al construir.
// Ninguna lleva texto legible, marcas ni logotipos de una institución: una web
// de «nada sin cita» no enseña un documento inventado.
import type { ImageMetadata } from "astro";

const ficheros = import.meta.glob<{ default: ImageMetadata }>("../assets/ilustraciones/*.jpg", { eager: true });

const porNombre = new Map(
  Object.entries(ficheros).map(([ruta, modulo]) => [ruta.replace(/^.*\/([^/]+)\.jpg$/, "$1"), modulo.default]),
);

export function ilustracion(nombre: string): ImageMetadata {
  const imagen = porNombre.get(nombre);
  if (!imagen) throw new Error(`ilustraciones: no hay ninguna imagen ${nombre}.jpg en src/assets/ilustraciones/`);
  return imagen;
}
