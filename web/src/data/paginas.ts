// Ruta, título y descripción de las páginas que no son portada, y las órdenes de
// instalación. Los usan la propia página y /llms.txt, que no los repite.

export interface Pagina {
  ruta: string;
  // Título de la pestaña y de los resultados de búsqueda (≤ 60 caracteres).
  titulo: string;
  // Descripción de los resultados de búsqueda (≤ 160 caracteres).
  descripcion: string;
}

export const instalar = {
  ruta: "/instalar/",
  titulo: "Instalar kitlegal paso a paso en macOS, Linux y Windows",
  descripcion:
    "Instala kitlegal copiando y pegando dos líneas, sin saber programar, y úsalo con Claude Code, Codex o Antigravity. Con Homebrew, Scoop o paquetes .deb y .rpm.",
  macosLinux: [
    "curl -fsSL https://raw.githubusercontent.com/jmorenobl/kitlegal/main/scripts/install.sh | sh",
    "kitlegal skills install",
  ],
  windows: [
    "scoop bucket add jmorenobl https://github.com/jmorenobl/scoop-bucket",
    "scoop install kitlegal",
    "kitlegal skills install",
  ],
} satisfies Pagina & { macosLinux: string[]; windows: string[] };

export const consultas: Pagina = {
  ruta: "/consultas/",
  titulo: "Consultas legales habituales, con su cita del BOE",
  descripcion:
    "Plazos de recursos, silencio administrativo, fianza del alquiler, preaviso, Hacienda: preguntas habituales con el artículo vigente del BOE que las responde.",
};

export const rastreador: Pagina = {
  ruta: "/bot/",
  titulo: "El rastreador de kitlegal",
  descripcion:
    "Qué son las peticiones con el agente de usuario kitlegal, cómo se comportan con tu sitio y cómo limitarlas o bloquearlas con robots.txt.",
};
