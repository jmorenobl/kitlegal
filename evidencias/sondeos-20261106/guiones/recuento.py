"""Cuenta, por variante, modo y eval, las sesiones del sondeo de H41 que abren cada referencia, y cuántas
referencias abre cada sesión.

Uso: python3 -I recuento.py <tabla.md> [<tabla.md> …], las tablas que imprimió make laboratorio-tabla.

Solo lee el detalle de cada tabla (`## Detalle`, una fila por sesión), que ya cuenta las lecturas con el contrato de
H41 (specs/033-h41-las-skills-del/contracts/tabla.md §2, con la lectura «cd …/references && cat <f>»): no abre ninguna
transcripción y da lo mismo en cualquier equipo. Una celda abre su referencia si no es «no» ni está vacía. Los
intervalos son de Wilson al 95 %.
"""
import collections
import math
import sys


def wilson(k, n, z=1.959963984540054):
    if n == 0:
        return "—"
    p = k / n
    d = 1 + z * z / n
    c = (p + z * z / (2 * n)) / d
    h = z * math.sqrt(p * (1 - p) / n + z * z / (4 * n * n)) / d
    return f"[{100 * (c - h):.1f}–{100 * (c + h):.1f} %]".replace(".", ",")


def celdas(linea):
    return [c.strip() for c in linea.strip().strip("|").split("|")]


def detalle(ruta):
    with open(ruta, encoding="utf-8") as f:
        lineas = f.read().splitlines()
    i = lineas.index("## Detalle")
    cabecera = None
    filas = []
    for linea in lineas[i + 1:]:
        if not linea.startswith("|"):
            continue
        c = celdas(linea)
        if cabecera is None:
            cabecera = c
            continue
        if set(c[0]) <= set("-"):
            continue
        filas.append(dict(zip(cabecera, c)))
    referencias = [h for h in cabecera if h.endswith(".md")]
    return referencias, filas


def partes(sesion):
    """Variante, modo y sesión, aunque la ruta lleve delante el lanzamiento (a/, b/) que juntó la tabla."""
    p = sesion.split("/")
    i = next(i for i in range(len(p) - 1) if p[i + 1] in ("orden", "herramienta"))
    return p[i], p[i + 1], p[i + 2]


def abre(celda):
    return celda not in ("", "no")


def familia(ruta):
    referencias, filas = detalle(ruta)
    titulo = ruta.rsplit("/", 1)[-1]
    print(f"## {titulo}\n")
    print("Sesiones que abren cada referencia, de las que tiene cada eval (10), por variante y modo.\n")
    grupos = collections.OrderedDict()
    por_sesion = collections.OrderedDict()
    for fila in filas:
        variante, modo, sesion = partes(fila["Sesión"])
        evaluacion = sesion.rsplit("-", 1)[0]
        grupos.setdefault((variante, modo, evaluacion), []).append(fila)
        por_sesion.setdefault((variante, modo), collections.Counter())[
            sum(abre(fila[r]) for r in referencias)
        ] += 1
    print("| Variante | Modo | Eval | n | " + " | ".join(f"`{r}`" for r in referencias) + " |")
    print("|---|---|---|---|" + "---|" * len(referencias))
    for (variante, modo, evaluacion), fs in grupos.items():
        n = len(fs)
        cuentas = [sum(abre(f[r]) for f in fs) for r in referencias]
        print(f"| `{variante}` | {modo} | {evaluacion} | {n} | " + " | ".join(str(k) for k in cuentas) + " |")
    print()
    print("Sesiones por número de referencias que abren, por variante y modo.\n")
    maximo = len(referencias)
    print("| Variante | Modo | " + " | ".join(str(k) for k in range(maximo + 1)) + " |")
    print("|---|---|" + "---|" * (maximo + 1))
    for (variante, modo), c in por_sesion.items():
        print(f"| `{variante}` | {modo} | " + " | ".join(str(c.get(k, 0)) for k in range(maximo + 1)) + " |")
    print()
    print("Cada referencia sobre las 40 sesiones de la fila, con su intervalo.\n")
    print("| Variante | Modo | Referencia | Abren |")
    print("|---|---|---|---|")
    filas_por_variante = collections.OrderedDict()
    for fila in filas:
        variante, modo, _ = partes(fila["Sesión"])
        filas_por_variante.setdefault((variante, modo), []).append(fila)
    for (variante, modo), fs in filas_por_variante.items():
        n = len(fs)
        for r in referencias:
            k = sum(abre(f[r]) for f in fs)
            if k or any(f[r] != "" for f in fs):
                print(f"| `{variante}` | {modo} | `{r}` | {k} de {n} {wilson(k, n)} |")
    print()


def main(rutas):
    print("# Recuento por eval del sondeo de H41\n")
    print("Generado por `guiones/recuento.py` desde el detalle de cada tabla; no abre ninguna transcripción.\n")
    for ruta in rutas:
        familia(ruta)


if __name__ == "__main__":
    if len(sys.argv) < 2:
        sys.exit("uso: python3 -I recuento.py <tabla.md> [<tabla.md> …]")
    main(sys.argv[1:])
