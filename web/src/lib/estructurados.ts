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
      applicationCategory: "DeveloperApplication",
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

export function migas(pagina: { nombre: string; ruta: string }): Record<string, unknown>[] {
  return [
    {
      "@context": "https://schema.org",
      "@type": "BreadcrumbList",
      itemListElement: [
        { "@type": "ListItem", position: 1, name: "kitlegal", item: `${SITIO}/` },
        { "@type": "ListItem", position: 2, name: pagina.nombre, item: `${SITIO}${pagina.ruta}` },
      ],
    },
  ];
}
