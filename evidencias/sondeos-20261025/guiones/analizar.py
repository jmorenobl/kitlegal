# Por brazo: marcadas por modo, referencias leídas por las sesiones y si la respuesta nombra el art. 7. Uso: analizar.py <carpeta> <brazos separados por comas>; los votos salen de brazo-<b>/votos.json, y las referencias leídas, de las transcripciones de brazo-<b>/sesiones/<sesión>/sesion.jsonl, que el temporal del sondeo conserva y esta carpeta no versiona.
import json, sys, os, re, glob, collections
S = sys.argv[1]
for b in sys.argv[2].split(','):
    votos = {v['sesion']: v for v in json.load(open(f'{S}/brazo-{b}/votos.json'))}
    filas = []
    for d in sorted(glob.glob(f'{S}/brazo-{b}/sesiones/*')):
        n = os.path.basename(d)
        leidas, skill = set(), False
        for l in open(f'{d}/sesion.jsonl'):
            try: m = json.loads(l)
            except Exception: continue
            for c in (m.get('message') or {}).get('content') or []:
                if isinstance(c, dict) and c.get('type') == 'tool_use':
                    i = c.get('input') or {}
                    if c['name'] == 'Skill': skill = True
                    ruta = i.get('file_path') or i.get('path') or i.get('command') or ''
                    for r in re.findall(r'references/(\w+)\.md', ruta): leidas.add(r)
        v = votos.get(n, {})
        afirma = next((c for c in v.get('juicio', {}).get('Clases', []) if c.get('clase') == 'afirma_lo_no_leido'), {})
        marcada = afirma.get('marcada', False)
        resp = v.get('respuesta', '')
        nombra7 = bool(re.search(r'art(ículo|\.)\s*7\b', resp))
        modo = 'herramienta' if '-herramienta-' in n else 'orden'
        filas.append((modo, marcada, leidas, nombra7))
    print(f'== {b}')
    for modo in ('orden', 'herramienta'):
        f = [x for x in filas if x[0] == modo]
        print(f'  {modo}: marcadas {sum(x[1] for x in f)}/{len(f)}, nombran el art. 7 {sum(x[3] for x in f)}/{len(f)}')
    c = collections.Counter(r for x in filas for r in x[2])
    print('  referencias leídas (sesiones):', dict(c.most_common()))
    if b == 'D':
        con = [x for x in filas if 'responder' in x[2]]; sin = [x for x in filas if 'responder' not in x[2]]
        print(f'  D con responder.md: {sum(x[1] for x in con)}/{len(con)} marcadas; sin leerla: {sum(x[1] for x in sin)}/{len(sin)}')
    print('  marcadas que nombran el art. 7:', sum(1 for x in filas if x[1] and x[3]), '· sin marcar que lo nombran:', sum(1 for x in filas if not x[1] and x[3]))
