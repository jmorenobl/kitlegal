# Por informe: árbol de la skill, recuento medido de afirma_lo_no_leido y respuestas marcadas.
import json, glob, os, subprocess, sys
D = sys.argv[1]
filas = []
for f in sorted(glob.glob(D + '/inf/*.json')):
    d = json.load(open(f))
    c = d['commit']
    r = subprocess.run(['git', 'rev-parse', '--short', c + ':skills/boe-legislacion'], capture_output=True, text=True)
    arbol = r.stdout.strip() or '?'
    conj = next((u for u in d['umbrales'] if u['nombre'] == 'afirma_lo_no_leido:claude-sonnet-5-5'), None)
    marcadas = []
    for resp in d['juez']['respuestas']:
        for cl in resp['clases']:
            if cl['clase'] == 'afirma_lo_no_leido' and cl['marcada']:
                frases = sorted({v.get('frase', '') for v in cl['votos']})
                prec = sorted({v.get('precepto', '') for v in cl['votos']})
                marcadas.append({'sesion': resp['sesion'], 'frases': frases, 'precepto': prec})
    filas.append({'run': os.path.basename(f)[:-5], 'commit': c[:8], 'arbol': arbol,
                  'medida': conj and conj['medida'], 'total': conj and conj['total'], 'umbral': conj and conj['umbral'],
                  'marcadas': marcadas})
json.dump(filas, open(D + '/tabla.json', 'w'), ensure_ascii=False, indent=1)
for x in filas:
    print(x['run'], x['commit'], x['arbol'], x['medida'], x['total'], len(x['marcadas']))
