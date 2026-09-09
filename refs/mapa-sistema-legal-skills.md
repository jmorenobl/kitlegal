# Mapa del sistema legal español y kit de skills agénticas

Objetivo: entender cómo está organizado el ordenamiento español, dónde vive cada tipo de información y qué skills (con sus herramientas) necesita un agente para obtenerla sin inventar nada.

Convención usada en todo el documento:
- 🟢 Fuente con API o datos abiertos estables → automatizable sin fricción.
- 🟡 Fuente web sin API formal (HTML/RSS/Atom) → scraping ligero, frágil, revisar TOS.
- 🔴 Requiere identidad (certificado, Cl@ve), CAPTCHA o está prohibida la automatización → **firma humana**, el agente solo prepara.

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
| Informes de fiscalización | Tribunal de Cuentas, Cámara de Cuentas de Madrid, etc. | PDF | 🟡 |
| Notificaciones administrativas propias | DEHú, Notifica, sede electrónica | Certificado/Cl@ve | 🔴 |
| Presentar solicitud, recurso, reclamación | Registro electrónico (REC), sedes | Certificado/Cl@ve + firma | 🔴 |
| Estado de un expediente propio | Carpeta Ciudadana | Certificado/Cl@ve | 🔴 |

---

## 3. Catálogo de skills agénticas

Diseño coherente con lo que ya tienes: `boe-fiscal` es el patrón (skill = SKILL.md con protocolo de razonamiento + `scripts/` + `references/` con identificadores). Todas las skills devuelven **JSON** por stdout, con `fuente`, `url`, `fecha_consulta` en cada resultado, para que la cita sea verificable.

### Capa 0 — Núcleo transversal

**`legal-core`** (skill madre, se carga siempre)
- Contenido: jerarquía normativa, reparto competencial, mapa de recursos y plazos, tabla de leyes vertebrales, reglas de cita (`Art. X.Y de la Ley Z/AAAA, BOE-A-...`), reglas de no invención.
- Herramientas:
  - `plazos.py calcular --tipo alzada --fecha-notificacion 2026-09-01` → fecha límite aplicando arts. 30-31 LPAC (días hábiles, agosto inhábil en judicial, festivos locales).
  - `competencia.py quien-regula "licencia de terraza"` → tabla heurística Estado/CCAA/Local con referencia al art. 148-149 CE y LRBRL.
- Datos: `references/leyes_vertebrales.md`, `references/plazos.md`, `references/calendario_festivos.json` (descargado del BOE/Ayuntamiento).

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
- Un adaptador por boletín; empezar por BOCM (tu ámbito).
  - DOGC tiene API pública (`https://dogc.gencat.cat` → verificar endpoint actual).
  - BOCM: RSS + buscador HTML → `bocm.py sumario 2026-09-09`, `bocm.py buscar "Leganés"`.
  - Patrón genérico: `boletin.py --config bocm.yaml sumario FECHA` con configuración YAML (URL, selectores CSS, RSS).
- Normalizar todo a un mismo esquema `{boletin, fecha, seccion, organo, titulo, url, pdf}`.

**`bop-y-edictos`**
- BOP de la Comunidad de Madrid (lo publica el BOCM en sección III/IV) y Tablón Edictal Único (BOE Sección V-B).
- `edictos.py buscar --municipio Leganés --desde 2026-01-01` (notificaciones por comparecencia, expropiaciones, licitaciones locales).

**`ordenanzas-locales`**
- Descarga e indexa ordenanzas y reglamentos de un ayuntamiento desde su sede/portal de transparencia.
- `ordenanzas.py listar --municipio leganes` → crawler con config por ayuntamiento; guarda PDF + texto extraído (`pdftotext`) + fecha BOCM de aprobación definitiva.

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
- `placsp.py organo "Ayuntamiento de Leganés"`, `placsp.py adjudicatario "XXXX SL"`, `placsp.py anomalias --organo ...` (fraccionamiento en menores, un solo licitador, modificados > 20%, plazos de publicación incumplidos).
- Portal de Contratación de la Comunidad de Madrid: adaptador aparte.

**`subvenciones`**
- BDNS API pública (`https://www.infosubvenciones.es/bdnstrans/api/...` — convocatorias, concesiones, beneficiarios; verificar swagger actual en `bdnstrans/api/swagger`).
- `bdns.py convocatorias --organo "Leganés" --anio 2026`, `bdns.py beneficiario "NIF|nombre"`, `bdns.py concesiones --convocatoria 123456`.
- Cruce con PLACSP y BORME: mismo beneficiario ↔ adjudicatario ↔ administrador (señal de conflicto de interés).

**`presupuestos-y-cuentas`**
- Presupuestos Generales del Estado (datos abiertos Hacienda), liquidaciones de EELL (Ministerio: `datos.gob.es` datasets "liquidaciones entidades locales"), Rendición de Cuentas (`rendiciondecuentas.es`).
- `presupuesto.py descargar --entidad 28074 --ejercicio 2025` (código INE del municipio), `presupuesto.py comparar --aprobado --liquidado` → desviaciones por capítulo y programa.

**`transparencia-portal`**
- Portal de Transparencia AGE (datos abiertos), portal de la CAM, portal municipal.
- `transparencia.py altos-cargos --organo`, `transparencia.py agenda`, `transparencia.py retribuciones`.
- `ctbg.py resoluciones "Leganés"` → resoluciones del Consejo de Transparencia (precedentes útiles para redactar reclamaciones).

**`entidades-y-registros`**
- BORME vía BOE API: `GET https://www.boe.es/datosabiertos/api/borme/sumario/AAAAMMDD` → `borme.py empresa "XXXX SL"`, `borme.py administrador "Nombre"` (histórico de nombramientos/ceses, constituciones, disoluciones).
- Catastro OVC: `catastro.py rc 1234567AB1234C` / `catastro.py direccion "calle X 1, Leganés"` (sin titular; el titular es 🔴).
- Registro de Fundaciones, Registro de Asociaciones (Ministerio del Interior), Registro de Entidades Locales.
- `entidad.py perfil "nombre|NIF"` → agrega BORME + BDNS + PLACSP + Catastro en un único perfil.

**`datos-abiertos`**
- INE Tempus: `ine.py tabla 2852` (`https://servicios.ine.es/wstempus/js/ES/DATOS_TABLA/{id}`).
- datos.gob.es CKAN: `datosgob.py buscar "Leganés"` (`https://datos.gob.es/apidata/catalog/dataset?...`).
- Portal de datos abiertos de Leganés / CAM.

### Capa 4 — Acción (todo 🔴, el agente redacta, el humano firma)

**`redaccion-escritos`**
- Plantillas: solicitud de acceso a información (LTAIBG art. 17), recurso de alzada/reposición, reclamación ante CTBG, queja al Defensor, denuncia a Tribunal de Cuentas, escrito de alegaciones, recurso especial en contratación.
- `escrito.py generar --tipo acceso-informacion --hechos hechos.md --organo "..."` → DOCX/PDF con fundamentación jurídica resuelta por `cita-verificada` y plazo calculado por `legal-core`.
- Nunca presenta: entrega el fichero y la URL de la sede donde presentarlo.

**`seguimiento-expedientes`**
- Registro propio (SQLite/Markdown) de solicitudes presentadas, fechas, plazos de silencio (1 mes LTAIBG, 3 meses LPAC por defecto), próximos pasos.
- `expedientes.py vencimientos` → alerta de silencios administrativos que habilitan reclamación.

---

## 4. Arquitectura del kit

```
kitlegal (binario Go multicall)
 ├─ legal-core ───────────── siempre cargada
 ├─ cita-verificada ───────── usada por todas las demás al citar
 ├─ normas/    boe-legislacion · boletines-autonomicos · bop-y-edictos · ordenanzas-locales · eurlex · tramitacion-parlamentaria
 ├─ doctrina/  jurisprudencia · doctrina-administrativa
 ├─ actividad/ contratacion-publica · subvenciones · presupuestos-y-cuentas · transparencia-portal · entidades-y-registros · datos-abiertos
 └─ accion/    redaccion-escritos · seguimiento-expedientes
```

Principios de implementación:
1. **Una fuente, un script, un esquema JSON.** Cada script imprime JSON con `fuente`, `url`, `fecha_consulta`, `hash_contenido`. Sin eso no hay cita.
2. **Caché local con TTL** (SQLite en `~/.cache/legal-kit/`): las normas cambian poco, los feeds de contratación cada hora.
3. **Config por entidad en YAML** (`entidades/leganes.yaml`: código INE, URLs de sede, portal de transparencia, perfil de contratante, boletín). Añadir un ayuntamiento = añadir un YAML.
4. **Rate limiting y `User-Agent` identificable** en todo scraping 🟡; respetar `robots.txt`; nada de CENDOJ masivo.
5. **Monitorización = diff sobre fuentes 🟢**: `kit vigilar` ejecuta `boe.py vigilar`, `placsp.py sync`, `bdns.py convocatorias`, `edictos.py` y emite solo novedades.
6. **Frontera humana explícita**: cualquier skill de Capa 4 termina en "fichero listo para firmar", nunca en un POST a una sede.

---

## 5. Orden de construcción sugerido

1. `legal-core` + `cita-verificada` + generalizar `boe-fiscal` → `boe-legislacion` (reutilizas el 90% del código existente).
2. `contratacion-publica` (PLACSP) y `subvenciones` (BDNS): las dos fuentes 🟢 con más valor fiscalizador y las que dan resultados en Leganés desde el primer día.
3. `entidades-y-registros` (BORME + Catastro) para cruzar beneficiarios/adjudicatarios.
4. `bop-y-edictos` + `boletines-autonomicos` (BOCM) + `ordenanzas-locales`.
5. `presupuestos-y-cuentas` + `transparencia-portal`.
6. `redaccion-escritos` + `seguimiento-expedientes`.
7. `jurisprudencia` y `doctrina-administrativa` al final: son las de mayor fricción legal/técnica y las menos necesarias para fiscalizar.

Riesgos a verificar antes de codificar: endpoints exactos de BDNS y PLACSP (cambian), TOS de PETETE/DYCTEA/HJ-TC, y el estado de despliegue de los Tribunales de Instancia (afecta a cómo se citan órganos desde 2025).
