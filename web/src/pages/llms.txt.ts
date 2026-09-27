// El índice de la web para los agentes de IA, en el formato de
// https://llmstxt.org/. No escribe a mano nada que la web o el repositorio ya
// digan (ADR 0024): los títulos y las descripciones son los de cada página; las
// skills, las de la versión publicada, que son las que instala
// `kitlegal skills install`; y los esquemas, los que sirve /schemas/.
import { readdirSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import type { APIRoute } from "astro";
import { parse } from "yaml";
import { ciudadania, despachos } from "../data/audiencias";
import { instalar, type Pagina, rastreador } from "../data/paginas";
import {
  carpetasPublicadas,
  formatoEntero,
  leerPublicado,
  numeroDeMunicipios,
  RAIZ,
  SITIO,
  urlPublicada,
  versionPublicada,
} from "../lib/proyecto";

// Una entrada de una lista de enlaces, en una sola línea: [nombre](url): notas.
function entrada(nombre: string, url: string, notas: string): string {
  if (/[[\]\n]/.test(nombre) || notas.includes("\n")) {
    throw new Error(`llms.txt: la entrada «${nombre}» rompería la lista de enlaces`);
  }
  return `- [${nombre}](${url}): ${notas}`;
}

function pagina({ ruta, titulo, descripcion }: Pagina): string {
  return entrada(titulo, new URL(ruta, SITIO).href, descripcion);
}

// Cada skill con el nombre y la descripción de su SKILL.md: la misma descripción
// con la que el agente decide cuándo usarla.
function skills(): string[] {
  return carpetasPublicadas("skills").map((carpeta) => {
    const ruta = `skills/${carpeta}/SKILL.md`;
    const cabecera = /^---\n([\s\S]*?)\n---\n/.exec(leerPublicado(ruta))?.[1];
    const { name, description } = (cabecera ? parse(cabecera) : {}) as { name?: unknown; description?: unknown };
    if (name !== carpeta || typeof description !== "string" || description === "") {
      throw new Error(`llms.txt: ${ruta} no declara name: ${carpeta} y una description`);
    }
    return entrada(name, urlPublicada(ruta), description);
  });
}

// De la descripción de cada esquema se toma la primera frase: las demás son
// avisos para quien mantiene el repositorio («no editar»).
function esquemas(): string[] {
  const carpeta = resolve(RAIZ, "schemas");
  return readdirSync(carpeta)
    .filter((fichero) => fichero.endsWith(".json"))
    .sort()
    .map((fichero) => {
      const { title, description } = JSON.parse(readFileSync(resolve(carpeta, fichero), "utf8")) as {
        title?: string;
        description?: string;
      };
      const notas = description ? description.split(/(?<=\.)\s+/)[0] : title && `Esquema de ${title}.`;
      if (!notas) {
        throw new Error(`llms.txt: schemas/${fichero} no tiene title ni description`);
      }
      return entrada(fichero, `${SITIO}/schemas/${fichero}`, notas);
    });
}

export const GET: APIRoute = () => {
  const texto = [
    "# kitlegal",
    "",
    "> kitlegal da a un agente de IA el texto vigente del BOE con su cita exacta —la norma, el bloque, la fecha de consulta y la huella del texto— para que responda preguntas legales sin citar de memoria. Son skills para Claude Code, Codex, Antigravity o cualquier agente que siga el estándar Agent Skills y trabaje en el equipo de quien lo usa, y un programa, `kitlegal`, que hace las consultas.",
    "",
    `El programa corre en ese equipo, sin cuentas ni servidores de kitlegal: lee la legislación consolidada del BOE, avisa si el BOE marca una norma como derogada o con la vigencia agotada y sitúa cualquiera de los ${formatoEntero(numeroDeMunicipios())} municipios del INE en su provincia, su comunidad y su régimen. Las skills enseñan al agente cuándo y cómo usarlo, y a no afirmar nada sobre una norma sin el texto que el programa acaba de leer. Es software libre (EUPL-1.2), y la última versión publicada es la ${versionPublicada()}.`,
    "",
    "Se instala con dos órdenes: la primera instala el programa; la segunda, las skills, en `.agents/skills/` del directorio en el que se ejecute (con `-g`, para todos los proyectos). Después basta con abrir el agente en ese directorio y preguntar. En macOS y Linux:",
    "",
    "```sh",
    ...instalar.macosLinux,
    "```",
    "",
    "En Windows, en PowerShell, con Scoop:",
    "",
    "```powershell",
    ...instalar.windows,
    "```",
    "",
    "## Instalar y usar",
    "",
    pagina(instalar),
    entrada("README", urlPublicada("README.md"), "Qué sabe hacer un agente con kitlegal, cómo instalarlo y qué garantiza."),
    "",
    "## Skills",
    "",
    ...skills(),
    "",
    "## Optional",
    "",
    pagina(despachos),
    pagina(ciudadania),
    pagina(rastreador),
    entrada("CHANGELOG", urlPublicada("CHANGELOG.md"), "Los cambios de comportamiento de cada versión publicada."),
    ...esquemas(),
    "",
  ].join("\n");
  return new Response(texto, { headers: { "Content-Type": "text/plain; charset=utf-8" } });
};
