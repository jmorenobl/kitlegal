// Las citas de la web (ADR 0024): cada fragmento de norma que se muestra está,
// literal, en el sobre que devolvió kitlegal. Aquí se comprueba al construir, y
// una cita que no lo cumple detiene la construcción.

import { readdirSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { parse } from "yaml";

interface Sobre<T> {
  ok: boolean;
  fuente: string;
  url: string;
  fecha_consulta: string;
  hash: string;
  data: T;
}

interface Aviso {
  codigo: string;
  texto: string;
}

interface Bloque {
  norma: string;
  bloque: string;
  texto: string;
  fecha_version?: string;
  avisos?: Aviso[];
}

interface Metadatos {
  norma: string;
  titulo: string;
  url_eli?: string;
  avisos?: Aviso[];
}

// Cómo se une un fragmento con el anterior: en la misma línea o en otra si en
// el texto van seguidos, con « […] » si entre los dos se omite algo.
export type Union = "espacio" | "linea" | "elision";

export interface Articulo {
  norma: string;
  bloque: string;
  // La versión del bloque (AAAAMMDD, la `fecha_version` del sobre) con la que
  // se escribió la cita y lo que la página dice de ella con otras palabras.
  version: string;
  referencia: string;
  fragmentos: string[];
  uniones: Union[];
  destacado?: string;
  // Los avisos de vigencia que el sobre trae con el artículo: la página los
  // muestra con la cita, como hace el agente antes de citar.
  avisos: Aviso[];
  enlace: string;
  fechaConsulta: Date;
  fechaVersion?: Date;
  hash: string;
}

export interface Norma {
  norma: string;
  referencia: string;
  titulo: string;
  avisos: Aviso[];
  enlace: string;
  fechaConsulta: Date;
  hash: string;
}

const DATOS = resolve(process.cwd(), "src/data");
const SOBRES = resolve(DATOS, "sobres");

// uniones busca cada fragmento, literal salvo en los espacios, en el texto
// del sobre, en orden y sin solaparse, y dice cómo se une cada uno con el
// anterior. Un fragmento que no está, o que está antes que el anterior, detiene
// la construcción: la cita no puede reordenar la norma.
function uniones(id: string, texto: string, fragmentos: string[]): Union[] {
  const resultado: Union[] = [];
  let fin = 0;
  fragmentos.forEach((fragmento, indice) => {
    const patron = fragmento
      .trim()
      .split(/\s+/)
      .map((palabra) => palabra.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"))
      .join("\\s+");
    const encontrado = new RegExp(patron, "g");
    encontrado.lastIndex = fin;
    const casa = encontrado.exec(texto);
    if (!casa) {
      const antes = new RegExp(patron).test(texto) ? ", o está antes que el fragmento anterior" : "";
      throw new Error(`citas: ${id}: «${fragmento}» no está literal en el texto del sobre${antes}`);
    }
    if (indice > 0) {
      const entre = texto.slice(fin, casa.index);
      resultado.push(entre.trim() !== "" ? "elision" : entre.includes("\n") ? "linea" : "espacio");
    }
    fin = casa.index + casa[0].length;
  });
  return resultado;
}

// fechaDelBoe convierte AAAAMMDD en una fecha.
function fechaDelBoe(valor: string | undefined): Date | undefined {
  const partes = valor?.match(/^(\d{4})(\d{2})(\d{2})$/);
  return partes ? new Date(Date.UTC(+partes[1], +partes[2] - 1, +partes[3])) : undefined;
}

function leerSobre<T>(fichero: string, usados: Set<string>): Sobre<T> {
  usados.add(fichero);
  let sobre: Sobre<T>;
  try {
    sobre = JSON.parse(readFileSync(resolve(SOBRES, fichero), "utf8"));
  } catch {
    throw new Error(`citas: falta el sobre ${fichero}; ejecuta make web-citas`);
  }
  if (!sobre.ok || !sobre.url || !sobre.fecha_consulta || !sobre.hash) {
    throw new Error(`citas: el sobre ${fichero} no es un sobre de éxito con fuente, url, fecha y huella`);
  }
  return sobre;
}

function enlaceAlBoe(norma: string, bloque?: string): string {
  return `https://www.boe.es/buscar/act.php?id=${norma}${bloque ? `#${bloque}` : ""}`;
}

function cargar() {
  const declaradas = parse(readFileSync(resolve(DATOS, "citas.yaml"), "utf8")) as {
    articulos: Record<
      string,
      Omit<Articulo, "uniones" | "avisos" | "enlace" | "fechaConsulta" | "fechaVersion" | "hash">
    >;
    normas: Record<string, { norma: string; referencia: string }>;
  };
  const usados = new Set<string>();

  const articulos: Record<string, Articulo> = {};
  for (const [id, cita] of Object.entries(declaradas.articulos)) {
    const sobre = leerSobre<Bloque>(`${cita.norma}-${cita.bloque}.json`, usados);
    if (sobre.data.norma !== cita.norma || sobre.data.bloque !== cita.bloque) {
      throw new Error(`citas: el sobre de ${id} es de ${sobre.data.norma} ${sobre.data.bloque}`);
    }
    // Si el BOE trae otra versión del bloque, los fragmentos pueden seguir
    // estando y lo que la web dice de ellos haber quedado corto: alguien tiene
    // que releerlo antes de publicar.
    if (String(cita.version) !== sobre.data.fecha_version) {
      throw new Error(
        `citas: ${id} se escribió con la versión ${cita.version} de ${cita.norma} ${cita.bloque} y el sobre trae la ${sobre.data.fecha_version}; relee lo que la web dice de ese bloque y actualiza version en citas.yaml`,
      );
    }
    const unidas = uniones(id, sobre.data.texto, cita.fragmentos);
    if (cita.destacado && !cita.fragmentos.some((f) => f.includes(cita.destacado!))) {
      throw new Error(`citas: ${id}: el destacado «${cita.destacado}» no está en ningún fragmento`);
    }
    articulos[id] = {
      ...cita,
      version: String(cita.version),
      uniones: unidas,
      avisos: sobre.data.avisos ?? [],
      enlace: enlaceAlBoe(cita.norma, cita.bloque),
      fechaConsulta: new Date(sobre.fecha_consulta),
      fechaVersion: fechaDelBoe(sobre.data.fecha_version),
      hash: sobre.hash,
    };
  }

  const normas: Record<string, Norma> = {};
  for (const [id, cita] of Object.entries(declaradas.normas)) {
    const sobre = leerSobre<Metadatos>(`${cita.norma}-metadatos.json`, usados);
    if (sobre.data.norma !== cita.norma) {
      throw new Error(`citas: el sobre de ${id} es de ${sobre.data.norma}`);
    }
    normas[id] = {
      ...cita,
      titulo: sobre.data.titulo,
      avisos: sobre.data.avisos ?? [],
      enlace: sobre.data.url_eli ?? enlaceAlBoe(cita.norma),
      fechaConsulta: new Date(sobre.fecha_consulta),
      hash: sobre.hash,
    };
  }

  const sobrantes = readdirSync(SOBRES).filter((f) => f.endsWith(".json") && !usados.has(f));
  if (sobrantes.length > 0) {
    throw new Error(`citas: sobres que no usa ninguna cita: ${sobrantes.join(", ")}; ejecuta make web-citas`);
  }

  return { articulos, normas };
}

const citas = cargar();

export function articulo(id: string): Articulo {
  const cita = citas.articulos[id];
  if (!cita) throw new Error(`citas: no hay ningún artículo declarado con id ${id}`);
  return cita;
}

// norma devuelve una norma declarada y exige que el BOE la marque con el aviso
// que la página va a mostrar.
export function norma(id: string, aviso: string): Norma {
  const cita = citas.normas[id];
  if (!cita) throw new Error(`citas: no hay ninguna norma declarada con id ${id}`);
  if (!cita.avisos.some((a) => a.codigo === aviso)) {
    throw new Error(`citas: ${id} no tiene el aviso ${aviso} en su sobre`);
  }
  return cita;
}

export function fechaCorta(fecha: Date): string {
  return new Intl.DateTimeFormat("es-ES", { dateStyle: "long", timeZone: "Europe/Madrid" }).format(fecha);
}

export function huellaCorta(hash: string): string {
  return hash.replace(/^(sha256:[0-9a-f]{12})[0-9a-f]+$/, "$1…");
}
