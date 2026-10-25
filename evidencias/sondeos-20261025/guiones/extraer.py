# Extrae informe.json del log de un trabajo de evals (gh run view --log, columnas ya cortadas).
import json, re, sys
pref = re.compile(r'^﻿?\d{4}-\d\d-\d\dT[\d:.]+Z ?')
dentro, lineas = False, []
for l in open(sys.argv[1], encoding='utf-8', errors='replace'):
    l = pref.sub('', l.rstrip('\n'))
    if '--- inicio de informe.json ---' in l: dentro, lineas = True, []; continue
    if '--- fin de informe.json ---' in l: dentro = False; continue
    if dentro: lineas.append(l)
json.dump(json.loads('\n'.join(lineas)), open(sys.argv[2], 'w'), ensure_ascii=False)
