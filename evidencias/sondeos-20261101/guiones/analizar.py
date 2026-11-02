# Por eval y modo: marcadas en las dos clases, cada frase marcada y Fisher unilateral de la eval 15 frente al control de evidencias/sondeos-20261025/ (16 y 17 de 80). Uso: analizar.py <temporal>; votos.json sale del temporal del sondeo.
import json, sys
from math import comb
from collections import Counter

def fisher(a, n1, b, n2):  # P(X <= a), unilateral, con a en el primer grupo
    k, n = a + b, n1 + n2
    return sum(comb(k, i) * comb(n - k, n1 - i) for i in range(a + 1)) / comb(n, n1)

votos = json.load(open(f'{sys.argv[1]}/votos.json'))
total, cuentas, marcas = Counter(), Counter(), []
for v in votos:
    s = v['sesion']
    grupo = (s[:2], 'herramienta' if '-herramienta-' in s else 'orden')
    total[grupo] += 1; total[(s[:2], 'ambos')] += 1
    for c in v['juicio']['Clases']:
        if c.get('marcada'):
            cuentas[grupo + (c['clase'],)] += 1; cuentas[(s[:2], 'ambos', c['clase'])] += 1
            marcas.append((c['clase'], s, [w.get('frase', '') for w in c['votos']]))
for g in sorted(total):
    print(g, total[g], {c: cuentas[g + (c,)] for c in ('afirma_lo_no_leido', 'cuenta_su_proceso')})
for clase, control in (('afirma_lo_no_leido', 16), ('cuenta_su_proceso', 17)):
    a = cuentas[('15', 'ambos', clase)]
    print(f'eval 15 {clase}: {a}/80 frente a {control}/80, p = {fisher(a, 80, control, 80):.5f}')
for m in marcas:
    print(m[0], '·', m[1])
    for f in m[2]: print('   ', f)
