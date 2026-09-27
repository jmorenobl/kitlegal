// Imágenes que se generan al construir: la de cada página al compartirla
// (1200 × 630) y el icono de Apple. Sin ficheros binarios en el repositorio: el
// texto sale de las propias páginas y el icono, de public/favicon.svg.
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { Resvg } from "@resvg/resvg-js";
import satori from "satori";

const MODULOS = resolve(process.cwd(), "node_modules");

function fuente(ruta: string): Buffer {
  return readFileSync(resolve(MODULOS, ruta));
}

export const ICONO = readFileSync(resolve(process.cwd(), "public/favicon.svg"), "utf8");

function png(svg: string, ancho: number): Uint8Array<ArrayBuffer> {
  return new Uint8Array(new Resvg(svg, { fitTo: { mode: "width", value: ancho } }).render().asPng());
}

// iconoCuadrado es el icono sin esquinas redondeadas: iOS aplica su propia
// máscara y rellena de negro lo transparente.
export function iconoCuadrado(lado: number): Uint8Array<ArrayBuffer> {
  return png(ICONO.replace(/ rx="\d+"/, ""), lado);
}

type Nodo = { type: string; props: Record<string, unknown> };

function caja(style: Record<string, unknown>, ...children: (Nodo | string)[]): Nodo {
  return { type: "div", props: { style: { display: "flex", ...style }, children } };
}

export async function imagenParaCompartir(titulo: string, subtitulo: string): Promise<Uint8Array<ArrayBuffer>> {
  const icono = `data:image/svg+xml;base64,${Buffer.from(ICONO).toString("base64")}`;

  const arbol = caja(
    {
      width: "100%",
      height: "100%",
      flexDirection: "column",
      justifyContent: "space-between",
      padding: "64px 72px",
      backgroundColor: "#f8f9ff",
      borderTop: "12px solid #d97706",
    },
    caja(
      { alignItems: "center", gap: "20px" },
      { type: "img", props: { src: icono, width: 72, height: 72 } },
      caja(
        { fontFamily: "Newsreader", fontWeight: 700, fontSize: "52px" },
        caja({ color: "#0b1325" }, "kit"),
        caja({ color: "#b45309" }, "legal"),
      ),
    ),
    caja(
      { flexDirection: "column", gap: "20px" },
      caja(
        { fontFamily: "Newsreader", fontWeight: 600, fontSize: "60px", lineHeight: 1.15, color: "#0b1c30" },
        titulo,
      ),
      caja({ fontFamily: "Inter", fontWeight: 500, fontSize: "28px", color: "#45464c" }, subtitulo),
    ),
    caja(
      { fontFamily: "Inter", fontWeight: 500, fontSize: "24px", color: "#45464c", borderTop: "1px solid #e2e8f0", paddingTop: "24px" },
      "kitlegal.es · Texto vigente del BOE, con la cita exacta",
    ),
  );

  const svg = await satori(arbol as unknown as Parameters<typeof satori>[0], {
    width: 1200,
    height: 630,
    fonts: [
      { name: "Newsreader", weight: 600, style: "normal", data: fuente("@fontsource/newsreader/files/newsreader-latin-600-normal.woff") },
      { name: "Newsreader", weight: 700, style: "normal", data: fuente("@fontsource/newsreader/files/newsreader-latin-700-normal.woff") },
      { name: "Inter", weight: 500, style: "normal", data: fuente("@fontsource/inter/files/inter-latin-500-normal.woff") },
    ],
  });

  return png(svg, 1200);
}
