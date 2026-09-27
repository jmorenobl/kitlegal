// Los esquemas publicados del repositorio (schemas/), en la dirección que
// declara su $id: https://kitlegal.es/schemas/<nombre>.json. Se sirven tal cual.
import { readdirSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import type { APIRoute, GetStaticPaths } from "astro";
import { RAIZ, SITIO } from "../../lib/proyecto";

const CARPETA = resolve(RAIZ, "schemas");

export const getStaticPaths: GetStaticPaths = () =>
  readdirSync(CARPETA)
    .filter((fichero) => fichero.endsWith(".json"))
    .map((fichero) => {
      const contenido = readFileSync(resolve(CARPETA, fichero), "utf8");
      const id = (JSON.parse(contenido) as { $id?: string }).$id;
      if (id !== `${SITIO}/schemas/${fichero}`) {
        throw new Error(`schemas: ${fichero} declara $id ${id}, no ${SITIO}/schemas/${fichero}`);
      }
      return { params: { nombre: fichero.replace(/\.json$/, "") }, props: { contenido } };
    });

export const GET: APIRoute = ({ props }) =>
  new Response(props.contenido as string, { headers: { "Content-Type": "application/schema+json" } });
