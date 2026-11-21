// El estado de cada hito, leído de docs/roadmap/ al construir (ADR 0044, punto
// 6): el directorio dice si está hecho, en curso, en la cola o en el backlog, y
// el prefijo NNN- de todo/ su puesto. La web no escribe ningún estado a mano.

import { readdirSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { parse } from "yaml";
import { RAIZ } from "./proyecto";

export type Estado = "hecho" | "en-curso" | "en-cola" | "pendiente";

export interface Hito {
  id: string;
  fase: string;
  dependeDe: string[];
  // El título del fichero, tal cual: el de la web, en lenguaje llano, está en
  // data/plan.ts.
  titulo: string;
  estado: Estado;
  // La posición en la cola de todo/, desde 1.
  puesto?: number;
  // La ruta del fichero en el repositorio.
  ruta: string;
}

const CARPETAS: [string, Estado][] = [
  ["done", "hecho"],
  ["in-progress", "en-curso"],
  ["todo", "en-cola"],
  ["backlog/transparencia-y-fiscalizacion", "pendiente"],
];

function leerCarpeta(carpeta: string, estado: Estado): Hito[] {
  const base = resolve(RAIZ, "docs/roadmap", carpeta);
  return readdirSync(base)
    .filter((f) => f.endsWith(".md") && f !== "README.md")
    .sort()
    .map((fichero, indice) => {
      const ruta = `docs/roadmap/${carpeta}/${fichero}`;
      const texto = readFileSync(resolve(base, fichero), "utf8");
      const cabecera = /^---\n([\s\S]*?)\n---\n/.exec(texto)?.[1];
      if (!cabecera) throw new Error(`roadmap: ${ruta} no tiene cabecera`);
      const datos = parse(cabecera) as { id?: unknown; fase?: unknown; depende_de?: unknown };
      const titulo = /^# (\S+) · (.+)$/m.exec(texto);
      if (typeof datos.id !== "string" || typeof datos.fase !== "string" || !titulo || titulo[1] !== datos.id) {
        throw new Error(`roadmap: ${ruta} no declara id, fase y un título «# ${String(datos.id)} · …»`);
      }
      return {
        id: datos.id,
        fase: datos.fase,
        dependeDe: Array.isArray(datos.depende_de) ? datos.depende_de.map(String) : [],
        titulo: titulo[2].replaceAll("`", ""),
        estado,
        ...(estado === "en-cola" ? { puesto: indice + 1 } : {}),
        ruta,
      };
    });
}

let cache: Hito[] | undefined;

// hitos devuelve todos los hitos con número: los hechos, el que está en curso,
// la cola en su orden y los de transparencia y fiscalización, que conservan el
// suyo en el backlog.
export function hitos(): Hito[] {
  cache ??= CARPETAS.flatMap(([carpeta, estado]) => leerCarpeta(carpeta, estado));
  return cache;
}

// cubierto devuelve lo que kitlegal declara cubierto hoy, de data/cobertura.yaml:
// una fila por capacidad, con su alcance en una frase (ADR 0058).
export function cubierto(): { capacidad: string; alcance: string; materia?: string }[] {
  const datos = parse(readFileSync(resolve(RAIZ, "data/cobertura.yaml"), "utf8")) as {
    cubierto?: { capacidad: string; alcance: string; materia?: string }[];
  };
  if (!datos.cubierto?.length) throw new Error("roadmap: data/cobertura.yaml no declara nada cubierto");
  return datos.cubierto;
}
