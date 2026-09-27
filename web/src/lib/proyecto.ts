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
