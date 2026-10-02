// Las clases de Tailwind que se repiten en varias páginas (DESIGN.md): los
// botones, los chips, las etiquetas de sección y el contenedor de la página.

// El ancho de la página y su margen lateral.
export const pagina = "mx-auto w-full max-w-[1280px] px-margen";

const baseBoton =
  "tocable inline-flex min-h-11 items-center justify-center gap-2 rounded-full text-center font-semibold";

export const boton = {
  primario: `${baseBoton} bg-agente px-[26px] py-[15px] text-base text-white`,
  secundario: `${baseBoton} border border-linea-2 bg-white px-[26px] py-[15px] text-base text-tinta`,
  oscuro: `${baseBoton} bg-tinta px-[26px] py-[15px] text-base text-white`,
  // Sobre el fondo azul del bloque final.
  claro: `${baseBoton} bg-white px-[26px] py-[15px] text-base text-tinta`,
  contorno: `${baseBoton} border border-white px-[26px] py-[15px] text-base text-white`,
  // El de la cabecera y el de las tarjetas.
  pequeno: `${baseBoton} bg-agente px-5 py-3 text-[15px] whitespace-nowrap text-white`,
};

// El chip de estado: vigente, en verde.
export const chipVigente =
  "inline-flex items-center rounded-full border border-vigente-borde bg-vigente-fondo px-3.5 py-1.5 text-sm font-semibold text-vigente-tinta";

// La etiqueta pequeña de una tarjeta, sin borde.
export const chipPequeno = "self-start rounded-full px-2.5 py-1 text-[13px] font-semibold";

// La etiqueta azul que precede al titular de una sección.
export const etiquetaSeccion = "text-sm font-semibold text-agente";

// La etiqueta en mayúsculas de una tarjeta.
export const etiquetaMayusculas = "text-[13px] font-semibold tracking-[0.06em] text-tenue uppercase";

export const titularSeccion = "font-titular text-h2 font-bold text-balance";

// Los puntos de color de las tarjetas y las categorías.
export type Punto = "agente" | "vigente" | "sello" | "cian";

export const punto: Record<Punto, string> = {
  agente: "bg-agente",
  vigente: "bg-vigente",
  sello: "bg-sello",
  cian: "bg-cian",
};
