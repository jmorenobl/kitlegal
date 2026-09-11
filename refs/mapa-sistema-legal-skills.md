# Mapa del sistema legal español y kit de skills agénticas

Objetivo: entender cómo está organizado el ordenamiento español, dónde vive cada tipo de información y qué skills (con sus herramientas) necesita un agente para obtenerla sin inventar nada.

Convención usada en todo el documento:
- 🟢 Fuente con API o datos abiertos estables → automatizable sin fricción.
- 🟡 Fuente web sin API formal (HTML/RSS/Atom) → scraping ligero, frágil, revisar TOS.
- 🔴 Requiere identidad (certificado, Cl@ve), CAPTCHA o está prohibida la automatización → **firma humana**, el agente solo prepara.

Dos reglas de lectura:
- **Las skills son el producto; las herramientas, su apoyo.** Los comandos `*.py` del catálogo son los nombres de la propuesta original: en `kitlegal` cada uno es un applet del binario Go (`kitlegal placsp …`, o `scripts/placsp …` desde la skill). Solo va a una herramienta lo que exige determinismo o verificabilidad; el razonamiento vive en la skill.
- **Genérico para cualquier municipio.** Donde un comando recibe un municipio, acepta cualquiera (nombre o código INE) y lo resuelve con `territorio`. Los ejemplos usan Leganés (INE 28074) porque la Comunidad de Madrid es el territorio de validación, que se implementa primero; fuera de él la salida declara su cobertura. Ninguna herramienta ni skill se particulariza para un municipio.

---

## 1. Mapa del sistema legal

### 1.1 Jerarquía normativa (art. 9.3 CE, art. 1 CC)

```
Constitución Española (1978)
 └─ Derecho de la UE (primacía; reglamentos directos, directivas transpuestas)
     └─ Tratados internacionales (art. 96 CE)
         └─ Leyes estatales
             ├─ Ley Orgánica (materias reservadas: derechos fundamentales, Estatutos, régimen electoral…)
             ├─ Ley ordinaria
             ├─ Real Decreto-ley (urgencia, art. 86) / Real Decreto Legislativo (texto refundido, art. 82-85)
             └─ Reglamentos: Real Decreto (Consejo de Ministros) > Orden Ministerial > Resolución/Instrucción
         └─ Leyes autonómicas (Estatuto de Autonomía + leyes del Parlamento autonómico)
             └─ Decretos y Órdenes de la Consejería
         └─ Normativa local (sin potestad legislativa)
             └─ Reglamentos orgánicos, Ordenanzas (fiscales, urbanísticas…), Bandos
 Fuentes subsidiarias: costumbre y principios generales del derecho (art. 1.1 CC)
 Jurisprudencia: complementa el ordenamiento (art. 1.6 CC) — no es fuente formal salvo doctrina del TC
```

Reglas de interpretación que el agente debe aplicar: competencia antes que jerarquía (Estado vs CCAA, arts. 148-149 CE), ley posterior deroga anterior, ley especial prevalece sobre general, reglamento nunca contra ley.

### 1.2 Reparto territorial (quién puede regular qué)

| Nivel | Órgano legislativo | Órgano ejecutivo | Boletín oficial | Ejemplos de competencia |
|---|---|---|---|---|
| UE | Parlamento + Consejo | Comisión | DOUE | Mercado interior, protección de datos, competencia |
| Estado | Cortes Generales (Congreso + Senado) | Gobierno | BOE | Civil, mercantil, penal, procesal, Hacienda general, Seguridad Social |
| CCAA (17 + Ceuta/Melilla) | Parlamento autonómico | Consejo de Gobierno | BOCM, DOGC, BOJA… | Sanidad, educación, urbanismo, vivienda, tributos cedidos |
| Provincia | Diputación (pleno) | Presidente | BOP | Asistencia a municipios, carreteras provinciales |
| Municipio | Pleno | Alcalde / Junta de Gobierno | BOP + tablón de edictos + sede electrónica | Urbanismo local, tributos locales (IBI, IAE, tasas), servicios |

### 1.3 Poder judicial y órdenes jurisdiccionales (LOPJ 6/1985)

```
Tribunal Constitucional (fuera del Poder Judicial: recurso/cuestión de inconstitucionalidad, amparo)
Tribunal Supremo — 5 Salas: Civil, Penal, Contencioso-Administrativo, Social, Militar
Audiencia Nacional — Penal (terrorismo, grandes delitos económicos), Contencioso (ministerios), Social
Tribunales Superiores de Justicia (1 por CCAA) — Civil/Penal, Contencioso, Social
Audiencias Provinciales — apelación civil y penal
Juzgados: Primera Instancia, Instrucción, Penal, Contencioso-Administrativo, Social, Mercantil, Violencia sobre la Mujer, Menores, Vigilancia Penitenciaria
+ Tribunales de Instancia (reorganización LO 1/2025, en despliegue)
Tribunal de Justicia de la UE / TEDH (Estrasburgo) — instancias supranacionales
```

Órdenes y sus leyes procesales: Civil (LEC 1/2000), Penal (LECrim 1882), Contencioso-Administrativo (LJCA 29/1998), Social (LRJS 36/2011).

### 1.4 Vía administrativa (la que más interesa para fiscalización)

```
Actuación administrativa (acto, resolución, contrato, subvención, silencio)
 ├─ Recurso de alzada (si el órgano no agota la vía administrativa, 1 mes) — LPAC 39/2015 arts. 121-122
 ├─ Recurso potestativo de reposición (1 mes) — arts. 123-124
 ├─ Recurso extraordinario de revisión — art. 125
 ├─ Reclamación económico-administrativa (tributos): TEAR → TEAC — LGT 58/2003
 ├─ Recurso especial en materia de contratación: TACRC / tribunales autonómicos — LCSP 9/2017 arts. 44-60
 ├─ Reclamación ante Consejo de Transparencia (Estado: CTBG; CCAA: consejos propios) — LTAIBG 19/2013 art. 24
 ├─ Queja al Defensor del Pueblo / defensores autonómicos (no vinculante)
 └─ Agotada la vía → recurso contencioso-administrativo (2 meses) — LJCA 29/1998
Paralelo: denuncia ante Fiscalía, Tribunal de Cuentas (responsabilidad contable), Oficinas antifraude (OAC, AVAF, Oficina de Conflictos de Intereses), AEPD, CNMC
```

Leyes vertebrales que el agente debe tener siempre a mano (con identificador BOE):

| Ley | BOE id |
|---|---|
| Constitución Española | BOE-A-1978-31229 |
| Código Civil | BOE-A-1889-4763 |
| LPAC 39/2015 (procedimiento administrativo) | BOE-A-2015-10565 |
| LRJSP 40/2015 (régimen jurídico sector público) | BOE-A-2015-10566 |
| LJCA 29/1998 | BOE-A-1998-16718 |
| LEC 1/2000 | BOE-A-2000-323 |
| LOPJ 6/1985 | BOE-A-1985-12666 |
| LRBRL 7/1985 (bases régimen local) | BOE-A-1985-5392 |
| TRLRHL RDL 2/2004 (haciendas locales) | BOE-A-2004-4214 |
| LCSP 9/2017 (contratos sector público) | BOE-A-2017-12902 |
| LGS 38/2003 (subvenciones) | BOE-A-2003-20977 |
| LTAIBG 19/2013 (transparencia) | BOE-A-2013-12887 |
| LGT 58/2003 | BOE-A-2003-23186 |
| LOPDGDD 3/2018 | BOE-A-2018-16673 |
| Ley 47/2003 General Presupuestaria | BOE-A-2003-21614 |

(Verificar los ids con `boe.py buscar` antes de fijarlos en un fichero de referencia; algunos están escritos de memoria.)

### 1.5 Ciclo de vida de una norma (para saber qué documentos existen en cada fase)

```
Iniciativa → Anteproyecto (consulta pública previa, MAIN, dictamen Consejo de Estado)
 → Proyecto/Proposición (Congreso: BOCG, Diario de Sesiones, enmiendas)
 → Senado → Sanción y promulgación → Publicación BOE (texto original)
 → Texto consolidado (BOE lo mantiene con cada modificación)
 → Desarrollo reglamentario → Interpretación (DGT, TEAC, Abogacía del Estado)
 → Impugnación (TC, TS) → Derogación / vigencia agotada
```

---

## 2. Mapa de fuentes de información

Cada fila = un dato que un agente legal necesita, y dónde está.

| Necesidad | Fuente primaria | Acceso | Semáforo |
|---|---|---|---|
| Texto vigente de una ley/reglamento estatal o autonómico | BOE Legislación Consolidada (incluye normas de CCAA) | API REST JSON/XML, sin auth | 🟢 |
| Qué se publicó hoy / un día concreto | BOE sumario diario | API JSON por fecha | 🟢 |
| Identificador permanente europeo de una norma | ELI (`boe.es/eli/es/...`) | URL determinista | 🟢 |
| Legislación UE, sentencias TJUE | EUR-Lex, CURIA | SPARQL (Cellar) + webservice con registro | 🟢/🟡 |
| Jurisprudencia TS/AN/TSJ/AP | CENDOJ (CGPJ) | Web con CAPTCHA, TOS prohíbe uso masivo/automatizado | 🔴 (solo metadatos/ECLI puntuales) |
| Jurisprudencia constitucional | TC — buscador HJ y publicación en BOE (Sección TC) | Web + BOE API | 🟡/🟢 |
| Doctrina tributaria vinculante | DGT — PETETE | Web (formularios), sin API oficial | 🟡 |
| Doctrina económico-administrativa | TEAC — DYCTEA | Web | 🟡 |
| Dictámenes Consejo de Estado | BOE — base de dictámenes | Web | 🟡 |
| Tramitación parlamentaria (enmiendas, votaciones) | Congreso Open Data, Senado | Datos abiertos (JSON/CSV) | 🟢 |
| Boletines autonómicos | BOCM, DOGC (tiene API), BOJA, BOPV… | Mixto: RSS/API/HTML | 🟡 |
| Boletines provinciales y edictos locales | BOP de cada Diputación; Tablón Edictal Único (TEU, BOE) | HTML/RSS; TEU tiene sección en BOE | 🟡 |
| Ordenanzas municipales | Sede electrónica / portal de transparencia municipal | HTML/PDF | 🟡 |
| Contratación pública (licitaciones, adjudicaciones) | PLACSP — sindicación Atom + datos abiertos; portales autonómicos | Atom feeds + XML CODICE | 🟢 |
| Subvenciones y ayudas | BDNS / infosubvenciones.es | API REST pública + datos abiertos | 🟢 |
| Presupuestos y ejecución | Ministerio de Hacienda (datos abiertos), portales autonómicos/municipales; Rendición de Cuentas | CSV/XLSX | 🟢/🟡 |
| Sociedades: constitución, administradores, cuentas depositadas | BORME (BOE API) + Registro Mercantil | BORME 🟢; RM 🔴 (pago + identidad) |
| Inmuebles: referencia catastral, titular | Catastro OVC (consulta libre sin titular), Registro de la Propiedad | OVC 🟢 (SOAP/JSON); nota simple 🔴 |
| Datos estadísticos | INE (API Tempus), datos.gob.es (CKAN) | JSON | 🟢 |
| Cargos públicos, agendas, retribuciones, bienes | Portal de Transparencia (AGE), portales autonómicos/locales | HTML/CSV | 🟡 |
| Resoluciones sobre acceso a información | CTBG y consejos autonómicos | Web, PDF | 🟡 |
| Informes de fiscalización | Tribunal de Cuentas y órgano de control externo de cada comunidad (Cámara de Cuentas de Madrid, Consejo de Cuentas de Castilla y León…) | PDF | 🟡 |
| Notificaciones administrativas propias | DEHú, Notifica, sede electrónica | Certificado/Cl@ve | 🔴 |
| Presentar solicitud, recurso, reclamación | Registro electrónico (REC), sedes | Certificado/Cl@ve + firma | 🔴 |
| Estado de un expediente propio | Carpeta Ciudadana | Certificado/Cl@ve | 🔴 |

---

## 3. Catálogo de skills agénticas

Diseño coherente con lo que ya tienes: `boe-fiscal` es el patrón (skill = SKILL.md con protocolo de razonamiento + `scripts/` + `references/` con identificadores). Todas las herramientas que usan las skills devuelven **JSON** por stdout, con `fuente`, `url`, `fecha_consulta` en cada resultado, para que la cita sea verificable.

### Capa 0 — Núcleo transversal

**`legal-core`** (skill madre, se carga siempre)
- Contenido: jerarquía normativa, reparto competencial, mapa de recursos y plazos, tabla de leyes vertebrales, reglas de cita (`Art. X.Y de la Ley Z/AAAA, BOE-A-...`), reglas de no invención. Su protocolo empieza por identificar el territorio de la pregunta: de él dependen la normativa autonómica, los boletines, los festivos y los órganos competentes.
- Herramientas:
  - `territorio.py resolver "Leganés"` (o por código INE) → municipio, provincia, comunidad, DIR3 del ayuntamiento, régimen común o foral, boletines aplicables y cobertura. Datos de municipio desde registros nacionales (INE, DIR3); lo territorial, configurado por comunidad y por boletín.
  - `plazos.py calcular --tipo alzada --fecha-notificacion 2026-09-01 --municipio 28074` → fecha límite aplicando arts. 30-31 LPAC (días hábiles, agosto inhábil en judicial, festivos nacionales, autonómicos y locales; art. 30.6: inhábil en el municipio del interesado o en la sede del órgano). Si faltan los festivos locales del territorio, lo dice.
  - `competencia.py quien-regula "licencia de terraza"` → tabla heurística Estado/CCAA/Local con referencia al art. 148-149 CE y LRBRL.
- Datos: `references/leyes_vertebrales.md`, `references/plazos.md`; festivos en `data/festivos/` (nacionales y autonómicos, del BOE; locales, de la publicación de cada comunidad).

**`cita-verificada`**
- Recibe una cita (`Art. 21 LPAC`) y la resuelve a texto vigente + URL ELI + fecha de última modificación. Es la que impide alucinar.
- `cita.py resolver "art. 21 Ley 39/2015"` → llama a `boe.py articulo`.

### Capa 1 — Legislación (normas)

**`boe-legislacion`** — generalización de tu `boe-fiscal` a cualquier materia.
- Misma API de Legislación Consolidada. Reutilizar `boe.py` tal cual y añadir:
  - `boe.py sumario 20260909` → `GET https://www.boe.es/datosabiertos/api/boe/sumario/20260909` (Accept: application/json): todo lo publicado ese día, por sección y departamento.
  - `boe.py vigilar --materia "contratación" --desde 20260901` → usa `from`/`to` de la API para detectar normas actualizadas (capa de monitorización).
  - `boe.py eli "Ley 39/2015"` → construye `https://www.boe.es/eli/es/l/2015/10/01/39/con`.
- Cubre también normas autonómicas: el BOE consolida legislación de CCAA (filtrar por `departamento@codigo`).

**`boletines-autonomicos`**
- Un motor genérico configurado por YAML, un fichero por boletín: `boletin.py --boletin bocm sumario FECHA`, `boletin.py buscar --municipio 28074 "terrazas"` (URL, selectores CSS, RSS en `data/boletines/<boletin>.yaml`). Añadir un boletín es añadir un YAML; un adaptador propio solo si el motor no basta (p. ej. DOGC, que tiene API pública: verificar endpoint actual).
- Primero el BOCM, por ser el del territorio de validación; después el resto (BOCYL, DOGC, BOJA…), sin tocar la skill.
- Normalizar todo a un mismo esquema `{boletin, fecha, seccion, organo, titulo, url, pdf}`.

**`bop-y-edictos`**
- El boletín provincial del municipio y el Tablón Edictal Único (BOE Sección V-B, nacional). En las comunidades uniprovinciales (Madrid, Asturias, Cantabria, La Rioja, Murcia, Navarra, Baleares) no hay BOP y hace sus veces el boletín autonómico: en la Comunidad de Madrid, el BOCM. En las multiprovinciales hacen falta los dos (Tordesillas: BOCYL y BOP de Valladolid).
- `edictos.py buscar --municipio 28074 --desde 2026-01-01` (notificaciones por comparecencia, expropiaciones, licitaciones locales). El TEU funciona para cualquier municipio; los boletines, según la cobertura del territorio.

**`ordenanzas-locales`**
- Ordenanzas y reglamentos de cualquier ayuntamiento, obtenidos del boletín donde la ley obliga a publicarlos íntegros: el provincial o, en comunidades uniprovinciales, el autonómico (art. 70.2 LRBRL; ordenanzas fiscales, art. 17.4 TRLRHL). Así no hace falta un crawler ni una configuración por ayuntamiento.
- `ordenanzas.py listar --municipio 28074` → anuncios de aprobación definitiva y de modificación en el boletín del territorio, con fecha y URL; `ordenanzas.py ver <id>` → texto publicado. La skill no presenta como consolidado un texto que no lo es.
- La sede o el portal de transparencia del ayuntamiento, solo como complemento donde el boletín no baste.

**`eurlex`**
- `eurlex.py buscar "directiva servicios de pago"` → SPARQL contra Cellar (`https://publications.europa.eu/webapi/rdf/sparql`).
- `eurlex.py celex 32016R0679` → metadatos + transposición nacional (NIM) → cruzar con BOE.
- `curia.py ecli "ECLI:EU:C:2024:xxx"`.

**`tramitacion-parlamentaria`**
- Congreso Open Data (`https://www.congreso.es/opendata` — iniciativas, diarios, votaciones en JSON/CSV).
- `congreso.py iniciativa --expediente 121/000045` → estado, enmiendas, ponentes. Útil para "qué se está cocinando" antes de que llegue al BOE.

### Capa 2 — Jurisprudencia y doctrina

**`jurisprudencia`** — con las restricciones que ya conoces de Ventanilla Fiscal.
- CENDOJ: **no** ingesta masiva ni scraping del buscador. Estrategia viable:
  - `ecli.py resolver "ECLI:ES:TS:2025:1234"` → URL canónica + metadatos (fecha, sala, ponente) vía resolutor ECLI europeo (e-Justice).
  - Sentencias del TS que el BOE o notas de prensa del CGPJ referencian → capturar ECLI de ahí.
  - Fuentes secundarias legales para búsqueda de texto: bases de datos con licencia (el agente prepara la query, un humano la ejecuta) o el propio usuario pega la sentencia.
- TC: `tc.py buscar "amparo tutela judicial efectiva 2025"` contra el buscador HJ (HTML; revisar TOS) y `boe.py sumario` filtrando Sección TC para sentencias publicadas.
- Salida siempre con `ECLI` + `url` + `fecha`; nunca resumir una sentencia no descargada.

**`doctrina-administrativa`**
- DGT/PETETE: `dgt.py consulta V1234-24`, `dgt.py buscar "IRPF teletrabajo"` (HTML, respetar ritmo).
- TEAC/DYCTEA: `teac.py buscar "sanción 191 LGT"`.
- Consejo de Estado: `consejo-estado.py dictamen 123/2025` (base BOE).
- Abogacía del Estado, informes de la JCCA (contratación), Consultas de la DGSJFP (registros/notariado).

### Capa 3 — Actividad administrativa (la capa de fiscalización)

**`contratacion-publica`**
- PLACSP sindicación Atom (perfiles de contratante, licitaciones, adjudicaciones, contratos menores): `https://contrataciondelsectorpublico.gob.es/sindicacion/...` (feeds `licitacionesPerfilesContratanteCompleto3.atom`, `PlataformasAgregadasSinMenores.atom`, `contratosMenoresPerfilesContratantes.atom` — verificar nombres actuales).
- `placsp.py sync --desde 2026-01-01` → descarga incremental de Atom + parseo XML CODICE a SQLite.
- `placsp.py organo --municipio 28074` (el DIR3 del ayuntamiento lo da `territorio`), `placsp.py adjudicatario "XXXX SL"`, `placsp.py anomalias --organo ...` (fraccionamiento en menores, un solo licitador, modificados > 20%, plazos de publicación incumplidos). PLACSP es nacional: funciona para cualquier municipio.
- Plataformas autonómicas de contratación (la de la Comunidad de Madrid y las demás): PLACSP agrega parte de sus datos; lo que no llegue agregado (verificar, p. ej. los contratos menores) se declara en la cobertura y se añade como adaptador por plataforma cuando el territorio lo requiera.

**`subvenciones`**
- BDNS API pública (`https://www.infosubvenciones.es/bdnstrans/api/...` — convocatorias, concesiones, beneficiarios; verificar swagger actual en `bdnstrans/api/swagger`). Nacional: funciona para cualquier municipio.
- `bdns.py convocatorias --municipio 28074 --anio 2026`, `bdns.py beneficiario "NIF|nombre"`, `bdns.py concesiones --convocatoria 123456`.
- Cruce con PLACSP y BORME: mismo beneficiario ↔ adjudicatario ↔ administrador (señal de conflicto de interés).

**`presupuestos-y-cuentas`**
- Presupuestos Generales del Estado (datos abiertos Hacienda), liquidaciones de EELL (Ministerio: `datos.gob.es` datasets "liquidaciones entidades locales"), Rendición de Cuentas (`rendiciondecuentas.es`).
- `presupuesto.py descargar --entidad 28074 --ejercicio 2025` (código INE del municipio; sirve para cualquiera), `presupuesto.py comparar --aprobado --liquidado` → desviaciones por capítulo y programa.

**`transparencia-portal`**
- Portal de Transparencia AGE (datos abiertos), portal autonómico y portal municipal del territorio.
- `transparencia.py altos-cargos --organo`, `transparencia.py agenda`, `transparencia.py retribuciones`.
- `transparencia.py resoluciones --municipio 28074` → resoluciones del consejo de transparencia competente según el territorio (CTBG o el autonómico; algunas comunidades han convenido con el CTBG): precedentes útiles para redactar reclamaciones.

**`entidades-y-registros`**
- BORME vía BOE API: `GET https://www.boe.es/datosabiertos/api/borme/sumario/AAAAMMDD` → `borme.py empresa "XXXX SL"`, `borme.py administrador "Nombre"` (histórico de nombramientos/ceses, constituciones, disoluciones).
- Catastro OVC: `catastro.py rc 1234567AB1234C` / `catastro.py direccion "calle X 1, Leganés"` (sin titular; el titular es 🔴). No cubre País Vasco ni Navarra, que tienen catastros forales: ahí la salida declara la cobertura.
- Registro de Fundaciones, Registro de Asociaciones (Ministerio del Interior), Registro de Entidades Locales.
- `entidad.py perfil "nombre|NIF"` → agrega BORME + BDNS + PLACSP + Catastro en un único perfil.

**`datos-abiertos`**
- INE Tempus: `ine.py tabla 2852` (`https://servicios.ine.es/wstempus/js/ES/DATOS_TABLA/{id}`).
- datos.gob.es CKAN: `datosgob.py buscar --municipio 28074` (`https://datos.gob.es/apidata/catalog/dataset?...`), que federa buena parte de los portales autonómicos y municipales.
- Portales de datos abiertos autonómicos y municipales, solo cuando no estén federados en datos.gob.es.

### Capa 4 — Acción (todo 🔴, el agente redacta, el humano firma)

**`redaccion-escritos`**
- Plantillas: solicitud de acceso a información (LTAIBG art. 17), recurso de alzada/reposición, reclamación ante CTBG, queja al Defensor, denuncia a Tribunal de Cuentas, escrito de alegaciones, recurso especial en contratación.
- `escrito.py generar --tipo acceso-informacion --hechos hechos.md --municipio 28074` → documento con fundamentación jurídica resuelta por `cita-verificada`, órgano destinatario resuelto por `territorio` y plazo calculado por `legal-core`.
- Nunca presenta: entrega el fichero y dónde presentarlo. El Registro Electrónico General de la AGE admite escritos dirigidos a cualquier administración (art. 16.4 LPAC), así que siempre hay una vía genérica; la sede del ayuntamiento, cuando conste.

**`seguimiento-expedientes`**
- Registro propio (SQLite/Markdown) de solicitudes presentadas, fechas, plazos de silencio (1 mes LTAIBG, 3 meses LPAC por defecto; la normativa autonómica puede fijar otros, y se aplica cuando el territorio la configura), próximos pasos.
- `expedientes.py vencimientos` → alerta de silencios administrativos que habilitan reclamación.

---

## 4. Arquitectura del kit

```
skills (el producto)
 ├─ legal-core ───────────── siempre cargada; identifica el territorio
 ├─ cita-verificada ───────── usada por todas las demás al citar
 ├─ normas/    boe-legislacion · boletines-autonomicos · bop-y-edictos · ordenanzas-locales · eurlex · tramitacion-parlamentaria
 ├─ doctrina/  jurisprudencia · doctrina-administrativa
 ├─ actividad/ contratacion-publica · subvenciones · presupuestos-y-cuentas · transparencia-portal · entidades-y-registros · datos-abiertos
 └─ accion/    redaccion-escritos · seguimiento-expedientes
        │ scripts/<applet> …
        ▼
kitlegal (binario Go multicall: las herramientas deterministas de todas las skills)
```

Principios de implementación:
1. **Una fuente, un script, un esquema JSON.** Cada script imprime JSON con `fuente`, `url`, `fecha_consulta`, `hash_contenido`. Sin eso no hay cita.
2. **Caché local con TTL** (SQLite en `~/.cache/legal-kit/`): las normas cambian poco, los feeds de contratación cada hora.
3. **Territorio por datos, nunca por municipio.** Los datos de cada municipio (código INE, provincia, comunidad, DIR3) salen de registros nacionales; lo que varía por territorio (boletines, festivos locales, órganos de control y de transparencia, plataformas de contratación, régimen foral) se configura por comunidad o por boletín (`data/territorio/`, `data/boletines/`). No hay YAML por ayuntamiento: el municipio de cada persona va en su `.kitlegal/config.yaml`. Se configura primero la Comunidad de Madrid (validación en Leganés); fuera de ella cada herramienta declara su cobertura.
4. **Rate limiting y `User-Agent` identificable** en todo scraping 🟡; respetar `robots.txt`; nada de CENDOJ masivo.
5. **Monitorización = diff sobre fuentes 🟢**: `kit vigilar` ejecuta `boe.py vigilar`, `placsp.py sync`, `bdns.py convocatorias`, `edictos.py` y emite solo novedades.
6. **Frontera humana explícita**: cualquier skill de Capa 4 termina en "fichero listo para firmar", nunca en un POST a una sede.

---

## 5. Orden de construcción sugerido

Prioridad: lo que sirve para actuar en el propio municipio, genérico para cualquiera y validado primero en la Comunidad de Madrid. El detalle por hitos está en `docs/ROADMAP.md`.

1. `legal-core` (con `territorio` y `plazos`) + `cita-verificada`, sobre el `boe` portado de `boe-fiscal`.
2. `contratacion-publica` (PLACSP) y `subvenciones` (BDNS): fuentes 🟢 nacionales, con valor fiscalizador para cualquier municipio desde el primer día.
3. `bop-y-edictos` (BOCM + TEU) + `ordenanzas-locales` (vía boletín).
4. `redaccion-escritos` (acceso a información, reposición) + `seguimiento-expedientes`.
5. `entidades-y-registros` (BORME) + `presupuestos-y-cuentas`, y los cruces del pack fiscalizador.
6. `boe-legislacion` (sumario, vigilancia) y el resto de territorios: boletines autonómicos y provinciales, plataformas autonómicas, régimen foral.
7. `transparencia-portal`, `datos-abiertos`, `eurlex`, `tramitacion-parlamentaria`, `jurisprudencia` y `doctrina-administrativa` al final: son las de mayor fricción legal/técnica o las menos necesarias para actuar en el municipio.

Riesgos a verificar antes de codificar: endpoints exactos de BDNS y PLACSP (cambian), TOS de PETETE/DYCTEA/HJ-TC, y el estado de despliegue de los Tribunales de Instancia (afecta a cómo se citan órganos desde 2025).
