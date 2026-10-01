// Datos estructurados (schema.org) de kitlegal: el programa y su código.
import { LICENCIA, REPOSITORIO, SITIO, versionPublicada } from "./proyecto";

export function aplicacion(descripcion: string): Record<string, unknown>[] {
  const version = versionPublicada();
  return [
    {
      "@context": "https://schema.org",
      "@type": "SoftwareApplication",
      name: "kitlegal",
      description: descripcion,
      url: `${SITIO}/`,
      applicationCategory: "ReferenceApplication",
      operatingSystem: "macOS, Linux, Windows",
      softwareVersion: version.replace(/^v/, ""),
      downloadUrl: `${REPOSITORIO}/releases/tag/${version}`,
      installUrl: `${SITIO}/instalar/`,
      license: LICENCIA,
      inLanguage: "es-ES",
      isAccessibleForFree: true,
      offers: { "@type": "Offer", price: "0", priceCurrency: "EUR" },
    },
    {
      "@context": "https://schema.org",
      "@type": "SoftwareSourceCode",
      name: "kitlegal",
      codeRepository: REPOSITORIO,
      programmingLanguage: "Go",
      license: LICENCIA,
    },
  ];
}

// migas da el camino desde la portada hasta la página: cada argumento, un nivel.
export function migas(...paginas: { nombre: string; ruta: string }[]): Record<string, unknown>[] {
  return [
    {
      "@context": "https://schema.org",
      "@type": "BreadcrumbList",
      itemListElement: [
        { "@type": "ListItem", position: 1, name: "kitlegal", item: `${SITIO}/` },
        ...paginas.map(({ nombre, ruta }, indice) => ({
          "@type": "ListItem",
          position: indice + 2,
          name: nombre,
          item: `${SITIO}${ruta}`,
        })),
      ],
    },
  ];
}

// pregunta describe una página que responde a una pregunta.
export function pregunta(enunciado: string, respuesta: string): Record<string, unknown>[] {
  return [
    {
      "@context": "https://schema.org",
      "@type": "FAQPage",
      inLanguage: "es-ES",
      mainEntity: [
        {
          "@type": "Question",
          name: enunciado,
          acceptedAnswer: { "@type": "Answer", text: respuesta },
        },
      ],
    },
  ];
}
