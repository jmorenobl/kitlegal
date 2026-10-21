---
id: H37
fase: "1"
---

# H37 · Destinatarios, territorio y asunto genéricos

- **Objetivo**: que el modelo de dominio no dé por hecho que el destinatario es una administración ni que el territorio es solo el municipio.
- **Entrega**: una interfaz de destinatario con varios esquemas de identificador (DIR3, NIF…), con el ayuntamiento de `territorio` como primera implementación y la cobertura declarada para lo demás; `territorio` con sus ejes declarados —municipio y comunidad hoy; partido judicial, vecindad civil y ámbito de un convenio, declarados no cubiertos—; y el diseño del asunto genérico en el grafo —vía, contraparte, hechos con fecha y `Documento` con su huella—, que H10 implementa, con su ADR.
- **Riesgo**: interfaces sin segundo uso; solo interfaces y cobertura, sin datos nuevos.
