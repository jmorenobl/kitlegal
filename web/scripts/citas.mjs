// Regenera los sobres de las citas de la web (ADR 0024).
//
// Lee src/data/citas.yaml y, por cada artículo y cada norma que declara,
// ejecuta el binario —`kitlegal boe articulo … --json` o `kitlegal boe
// metadatos … --json`— y guarda el sobre tal cual en src/data/sobres/. La red
// la pone el binario, con su identificación, su ritmo y su caché; este guion no
// pide nada por su cuenta. Borra los sobres que ya no declara ninguna cita.
//
// El binario es el de KITLEGAL (por omisión, ../bin/kitlegal, el de make
// build). `make web-citas` lo construye antes.

import { execFileSync } from "node:child_process";
import { mkdirSync, readdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";
import { parse } from "yaml";

const raiz = resolve(import.meta.dirname, "..");
const kitlegal = process.env.KITLEGAL ?? resolve(raiz, "../bin/kitlegal");
const carpeta = resolve(raiz, "src/data/sobres");

const citas = parse(readFileSync(resolve(raiz, "src/data/citas.yaml"), "utf8"));

const consultas = new Map();
for (const { norma, bloque } of Object.values(citas.articulos ?? {})) {
  consultas.set(`${norma}-${bloque}.json`, ["boe", "articulo", norma, bloque, "--json"]);
}
for (const { norma } of Object.values(citas.normas ?? {})) {
  consultas.set(`${norma}-metadatos.json`, ["boe", "metadatos", norma, "--json"]);
}

mkdirSync(carpeta, { recursive: true });

for (const [fichero, argumentos] of consultas) {
  let salida;
  try {
    salida = execFileSync(kitlegal, argumentos, { encoding: "utf8", stdio: ["ignore", "pipe", "inherit"] });
  } catch (error) {
    console.error(`citas: \`kitlegal ${argumentos.join(" ")}\` salió con ${error.status}`);
    process.exit(1);
  }
  const sobre = JSON.parse(salida);
  writeFileSync(resolve(carpeta, fichero), `${JSON.stringify(sobre, null, 2)}\n`);
  console.log(`citas: ${fichero}`);
}

for (const fichero of readdirSync(carpeta)) {
  if (fichero.endsWith(".json") && !consultas.has(fichero)) {
    rmSync(resolve(carpeta, fichero));
    console.log(`citas: retirado ${fichero}`);
  }
}
