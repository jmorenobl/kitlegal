# Sondeo de la skill de laboratorio

| Variante | Modo | Modelo | Claude Code | Sin terminar | n | Activan antes | Entrada antes | Las dos antes | Abren la referencia | Turno de la primera apertura | Cumplen la regla |
|---|---|---|---|---|---|---|---|---|---|---|---|
| control | orden | claude-sonnet-5-5 | 2.1.317 | 0 | 4 | 2 de 4 · 50,0 % [15,0–85,0 %] |  |  |  |  | 2 de 4 · 50,0 % [15,0–85,0 %] |
| control | herramienta | claude-sonnet-5-5 | 2.1.300 | 0 | 1 | 1 de 1 · 100,0 % [20,7–100,0 %] |  |  |  |  | 1 de 1 · 100,0 % [20,7–100,0 %] |
| control | herramienta | claude-sonnet-5-5 | 2.1.317 | 1 | 3 | 2 de 3 · 66,7 % [20,8–93,9 %] |  |  |  |  | 2 de 3 · 66,7 % [20,8–93,9 %] |
| entrada | herramienta | claude-sonnet-5-5 | 2.1.317 | 0 | 3 | 2 de 3 · 66,7 % [20,8–93,9 %] | 2 de 3 · 66,7 % [20,8–93,9 %] | 1 de 3 · 33,3 % [6,1–79,2 %] |  |  | 1 de 3 · 33,3 % [6,1–79,2 %] |
| referencia-secundaria | herramienta | claude-sonnet-5-5 | 2.1.317 | 0 | 2 | 1 de 2 · 50,0 % [9,5–90,5 %] |  |  | 0 de 2 · 0,0 % [0,0–65,8 %] |  | 1 de 2 · 50,0 % [9,5–90,5 %] |
| parametro | herramienta | claude-sonnet-5-5 | 2.1.300 | 2 | 0 |  |  |  |  |  |  |
| referencia-paso | orden | claude-sonnet-5-5 | 2.1.317 | 0 | 3 | 3 de 3 · 100,0 % [43,8–100,0 %] |  |  | 2 de 3 · 66,7 % [20,8–93,9 %] | 2,5 (2–3) | 2 de 3 · 66,7 % [20,8–93,9 %] |
| descripcion-sin-herramientas | orden | claude-sonnet-5-5 | 2.1.300 | 1 | 1 | 0 de 1 · 0,0 % [0,0–79,3 %] |  |  |  |  | 0 de 1 · 0,0 % [0,0–79,3 %] |

## Detalle

| Sesión | Terminada | Activa antes | Entrada antes | Abre la referencia | Cumple |
|---|---|---|---|---|---|
| h39-a/control/orden/01-constitucion-articulo-140-01 | sí | sí |  |  | sí |
| h39-a/control/orden/02-lgt-prescripcion-01 | sí | sí |  |  | no |
| h39-a/control/orden/03-lrjsp-principio-de-legalidad-01 | sí | no |  |  | sí |
| h39-a/control/orden/04-et-vacaciones-01 | sí | no |  |  | no |
| h39-b/control/herramienta/01-constitucion-articulo-140-02 | sí | sí |  |  | sí |
| h39-a/control/herramienta/01-constitucion-articulo-140-01 | sí | no |  |  | sí |
| h39-a/control/herramienta/02-lgt-prescripcion-01 | sí | sí |  |  | no |
| h39-a/control/herramienta/03-lrjsp-principio-de-legalidad-01 | sí | sí |  |  | sí |
| h39-a/control/herramienta/04-et-vacaciones-01 | no: tope de 240 s agotado (código 124) |  |  |  |  |
| h39-a/entrada/herramienta/01-constitucion-articulo-140-01 | sí | sí | sí |  | sí |
| h39-a/entrada/herramienta/02-lgt-prescripcion-01 | sí | no | sí |  | no |
| h39-a/entrada/herramienta/03-lrjsp-principio-de-legalidad-01 | sí | sí | no |  | no |
| h39-a/referencia-secundaria/herramienta/01-constitucion-articulo-140-01 | sí | sí |  | no | sí |
| h39-a/referencia-secundaria/herramienta/02-lgt-prescripcion-01 | sí | no |  | no | no |
| h39-b/parametro/herramienta/01-constitucion-articulo-140-01 | no: código 1: result con is_error: API Error: 529 Overloaded |  |  |  |  |
| h39-b/parametro/herramienta/04-et-vacaciones-01 | no: sin mensaje result |  |  |  |  |
| h39-a/referencia-paso/orden/01-constitucion-articulo-140-01 | sí | sí |  | turno 2 | sí |
| h39-a/referencia-paso/orden/02-lgt-prescripcion-01 | sí | sí |  | turno 3 | no |
| h39-a/referencia-paso/orden/03-lrjsp-principio-de-legalidad-01 | sí | sí |  | no | sí |
| h39-b/descripcion-sin-herramientas/orden/02-lgt-prescripcion-01 | no: sin respuesta |  |  |  |  |
| h39-b/descripcion-sin-herramientas/orden/03-lrjsp-principio-de-legalidad-01 | sí | no |  |  | no |
