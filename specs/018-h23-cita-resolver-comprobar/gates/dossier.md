# Dossier: dependencia externa inaccesible

Tarea T010, paso grabar_datos (2026-10-06T23:44:15Z).

La grabación de `cendoj.jurisprudencia` falló dos veces. Final del log (specs/018-h23-cita-resolver-comprobar/gates/grabaciones.log):
```
== grabación de cendoj.jurisprudencia (internal/source/cendoj), intento 1
--- FAIL: TestGrabarConsultas (10.03s)
    grabacion_test.go:57: 
        	Error Trace:	/Users/jorge/Projects/kitlegal/internal/source/cendoj/grabacion_test.go:57
        	Error:      	Received unexpected error:
        	            	la consulta 1 (ecli ECLI:ES:TS:2023:3144) recibe una respuesta que no se reconoce, con el estado 302: no es una lista de resultados ni «No se ha encontrado ningún resultado», y no se copia nada a testdata/cendoj.jurisprudencia
        	Test:       	TestGrabarConsultas
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/source/cendoj	10.410s
FAIL
== grabación de cendoj.jurisprudencia (internal/source/cendoj), intento 2
--- FAIL: TestGrabarConsultas (10.03s)
    grabacion_test.go:57: 
        	Error Trace:	/Users/jorge/Projects/kitlegal/internal/source/cendoj/grabacion_test.go:57
        	Error:      	Received unexpected error:
        	            	la consulta 1 (ecli ECLI:ES:TS:2023:3144) recibe una respuesta que no se reconoce, con el estado 302: no es una lista de resultados ni «No se ha encontrado ningún resultado», y no se copia nada a testdata/cendoj.jurisprudencia
        	Test:       	TestGrabarConsultas
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/source/cendoj	10.263s
FAIL
```

Al resolverlo, reanudar con scripts/hito.sh --resume <run_id>: el bucle vuelve a esta tarea y la graba.
