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

// imagenDeConsulta es el nombre de la imagen para compartir de una página de
// /consultas/.
export function imagenDeConsulta(slug: string): string {
  return `consulta-${slug}`;
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
      backgroundColor: "#f7f8fb",
      borderTop: "12px solid #2563eb",
    },
    caja(
      { alignItems: "center", gap: "18px" },
      { type: "img", props: { src: icono, width: 64, height: 64 } },
      caja({ fontFamily: "Bricolage Grotesque", fontWeight: 700, fontSize: "48px", color: "#10182b" }, "kitlegal"),
    ),
    caja(
      { flexDirection: "column", gap: "20px" },
      caja(
        {
          fontFamily: "Bricolage Grotesque",
          fontWeight: 700,
          fontSize: "62px",
          lineHeight: 1.05,
          letterSpacing: "-0.03em",
          color: "#10182b",
        },
        titulo,
      ),
      caja({ fontFamily: "Figtree", fontWeight: 500, fontSize: "28px", color: "#45464c" }, subtitulo),
    ),
    caja(
      { fontFamily: "Figtree", fontWeight: 500, fontSize: "24px", color: "#45464c", borderTop: "1px solid #e3e7ef", paddingTop: "24px" },
      "kitlegal.es · Texto vigente del BOE, con la cita exacta",
    ),
  );

  const svg = await satori(arbol as unknown as Parameters<typeof satori>[0], {
    width: 1200,
    height: 630,
    fonts: [
      {
        name: "Bricolage Grotesque",
        weight: 700,
        style: "normal",
        data: fuente("@fontsource/bricolage-grotesque/files/bricolage-grotesque-latin-700-normal.woff"),
      },
      { name: "Figtree", weight: 500, style: "normal", data: fuente("@fontsource/figtree/files/figtree-latin-500-normal.woff") },
    ],
  });

  return png(svg, 1200);
}
