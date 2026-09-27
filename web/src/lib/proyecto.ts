// Datos del proyecto que la web muestra y que no se escriben a mano (ADR 0024):
// salen del repositorio al construir, de modo que la web no puede decir otra
// cosa que lo que dicen data/ y la última versión publicada.

import { execFileSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { parse } from "yaml";

export const SITIO = "https://kitlegal.es";
export const REPOSITORIO = "https://github.com/jmorenobl/kitlegal";
export const LICENCIA = "https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12";

// La web se construye desde web/ (make web, el flujo de la web); la raíz del
// repositorio es su padre. Si no lo es, la construcción falla aquí y no más
// tarde con un dato vacío.
export const RAIZ = resolve(process.cwd(), "..");
if (!existsSync(resolve(RAIZ, "go.mod"))) {
  throw new Error(`web: ${RAIZ} no es la raíz del repositorio; construye la web desde web/ (make web)`);
}

function leerYaml(ruta: string): unknown {
  return parse(readFileSync(resolve(RAIZ, ruta), "utf8"));
}

let municipios: number | undefined;

// numeroDeMunicipios cuenta los municipios de data/territorio/municipios.yaml,
// la relación del INE que viaja en el binario.
export function numeroDeMunicipios(): number {
  if (municipios === undefined) {
    const datos = leerYaml("data/territorio/municipios.yaml") as { municipios?: Record<string, unknown> };
    municipios = Object.keys(datos.municipios ?? {}).length;
    if (municipios === 0) {
      throw new Error("web: data/territorio/municipios.yaml no tiene municipios");
    }
  }
  return municipios;
}

export function formatoEntero(n: number): string {
  return new Intl.NumberFormat("es-ES", { useGrouping: "always" }).format(n);
}

let version: string | undefined;

// versionPublicada es la última versión publicada. El flujo de la web la pasa
// en KITLEGAL_VERSION desde la API de releases; en local, la da la última
// etiqueta v* alcanzable.
export function versionPublicada(): string {
  if (version === undefined) {
    version =
      process.env.KITLEGAL_VERSION ||
      execFileSync("git", ["describe", "--tags", "--abbrev=0", "--match", "v[0-9]*"], {
        cwd: RAIZ,
        encoding: "utf8",
      }).trim();
    if (!/^v\d+\.\d+\.\d+$/.test(version)) {
      throw new Error(`web: la versión publicada «${version}» no tiene la forma vX.Y.Z`);
    }
  }
  return version;
}

// Lo que instala la versión publicada se lee del árbol de su etiqueta, no del de
// main, que puede ir por delante. El flujo de la web trae esa etiqueta; en local
// ya está.
function gitPublicado(argumentos: string[]): string {
  try {
    return execFileSync("git", argumentos, { cwd: RAIZ, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] });
  } catch (error) {
    throw new Error(`web: git ${argumentos.join(" ")} falla; ¿está la etiqueta ${versionPublicada()} en el clon?`, {
      cause: error,
    });
  }
}

// carpetasPublicadas lista las carpetas de <carpeta> en la versión publicada.
export function carpetasPublicadas(carpeta: string): string[] {
  return gitPublicado(["ls-tree", "-d", "--name-only", `${versionPublicada()}:${carpeta}`])
    .split("\n")
    .filter((nombre) => nombre !== "");
}

// leerPublicado devuelve un fichero del repositorio tal como está en la versión
// publicada.
export function leerPublicado(ruta: string): string {
  return gitPublicado(["show", `${versionPublicada()}:${ruta}`]);
}

// urlPublicada es la dirección de un fichero del repositorio en la versión
// publicada, en bruto: sin la página de GitHub alrededor, que es lo que un
// agente lee mejor.
export function urlPublicada(ruta: string): string {
  return `https://raw.githubusercontent.com/jmorenobl/kitlegal/${versionPublicada()}/${ruta}`;
}
