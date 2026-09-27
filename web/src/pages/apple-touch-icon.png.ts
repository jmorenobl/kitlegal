import type { APIRoute } from "astro";
import { iconoCuadrado } from "../lib/imagenes";

export const GET: APIRoute = () => new Response(iconoCuadrado(180), { headers: { "Content-Type": "image/png" } });
