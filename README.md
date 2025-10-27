# Ejecución Especulativa en Go (Control2-LP)

**Repositorio:** https://github.com/LucasRojasReyes/Control2-LP

## Descripción General

La **ejecución especulativa** es una técnica de optimización donde se aprovecha la concurrencia para ejecutar múltiples tareas en paralelo antes de saber cuál será necesaria. La idea es que mientras se evalúa una condición lógica (costosa en tiempos de ejecución), las ramas de ejecución que se derivan de la misma, se ejecutan en paralelo a la evaluación.

Una vez se determina el resultado de la condición lógica , se hace valida la rama correspondiente (se anulan las otras), lo que mejora el rendimiento del programa. 

El control consiste en implementar este patrón de ejecución especulativa en Go, utilizando tres funciones de computo intensivo para luego comparar dos estrategias de ejecución:

- **Especulativa:** ejecuta en paralelo la condición lógica costosa y las dos ramas (A y B), cancelando la que no sea necesaria.
- **Secuencial:** ejecuta la condición lógica y luego la rama correspondiente de forma secuencial.

Ambas versiones usan las mismas funciones de cómputo intensivo:

| Función | Descripción | Complejidad |
|----------|--------------|-------------|
| `CalcularTrazaDeProductoDeMatrices(n)` | Multiplica dos matrices NxN y devuelve la traza del resultado. | O(n³) |
| `EncontrarPrimos(max)` | Busca números primos hasta `max` usando prueba por división. | O(n^1.5) |
| `SimularProofOfWork(dificultad)` | Simula una búsqueda de *Proof-of-Work* con un número de ceros iniciales definido por la dificultad. | Exponencial |

---

## Estructura del Proyecto

```
Control2-LP/  
├── Especulativo
│   ├── Especulativo.go              
│   ├── go.mod                      
│   └── metricas.csv                   
│
├── Secuencial                     
│   ├── Secuencial.go             
│   ├── go.mod                     
│   └── metricas  
│
└── Readme.md
```

## Instrucciones de Compilación

Desde la carpeta del proyecto:

```bash
# Compilar versión especulativa
cd Especulativo
go build -o especulativo .

# Compilar versión secuencial
cd ../Secuencial
go build -o secuencial .
```
## Instrucciones de Ejecución

Ambas versiones aceptan los mismos parámetros:

| Parámetro         | Descripción                                 |
| ----------------- | ------------------------------------------- |
| `-n`              | Dimensión de las matrices NxN (predicado)   |
| `-umbral`         | Valor umbral usado en la comparación lógica |
| `-comparacion`    | Tipo de comparación: `<` o `>`              |
| `-max_primos`     | Límite superior para búsqueda de primos     |
| `-dificultad`     | Dificultad para el Proof-of-Work            |
| `-nombre_archivo` | Archivo CSV donde se guardan métricas       |

# Ejemplo de ejecución
```bash
# Rama ganadora = A (comparación "<")
./especulativo  go run especulativo.go -n 3000 -umbral 1000000000  -nombre_archivo "metricas.csv" -max_primos 3000000  -dificultad 5  -comparacion ">"

./secuencial  go run secuencial.go -n 3000 -umbral 1000000000  -nombre_archivo "metricas.csv" -max_primos 3000000  -dificultad 5  -comparacion ">"
  ```

# Ejecución Experimental (30 Corridas)

Para obtener resultados estadísticamente significativos, se ejecutaron 30 corridas para cada estrategia con los mismos parámetros.

En PoweShell (Windows)
```powershell
# 30 corridas Especulativo
1..30 | % {
  .\especulativo -n 500 -umbral 1000000 -comparacion "<" `
    -max_primos 500000 -dificultad 5 `
    -nombre_archivo "metricas.csv"
}

# 30 corridas Secuencial
1..30 | % {
  .\secuencial -n 500 -umbral 1000000 -comparacion "<" `
    -max_primos 500000 -dificultad 5 `
    -nombre_archivo "metricas.csv"
}
```
En Bash (Linux/macOS)
```bash
for i in $(seq 1 30); do
  ./especulativo -n 500 -umbral 1000000 -comparacion "<" \
    -max_primos 500000 -dificultad 5 \
    -nombre_archivo "metricas.csv"
done

for i in $(seq 1 30); do
  ./secuencial -n 500 -umbral 1000000 -comparacion "<" \
    -max_primos 500000 -dificultad 5 \
    -nombre_archivo "metricas.csv"
done
```

# Cálculo de Promedios y Speedup

Cada ejecución genera un archivo CSV con el formato:
```scss
Inicio,Fin,Estrategia,ValorUmbral,Comparacion,Resultado,TiempoTotal(ms)
```
Para calcular el promedio de tiempo y el speedup usamos la siguiente formula:
## $$Speedup=\frac{TEspeculativo}{​TSecuencial}​​$$

Para calcular el promedio de tiempo y el speedup en PowerShell:
```powershell
$avgEsp = (Import-Csv .\metricas_especulativo.csv | Measure-Object -Property "TiempoTotal(ms)" -Average).Average
$avgSec = (Import-Csv .\metricas_secuencial.csv | Measure-Object -Property "TiempoTotal(ms)" -Average).Average
$speedup = $avgSec / $avgEsp
"Especulativo promedio: $avgEsp ms"
"Secuencial  promedio: $avgSec ms"
"Speedup: $speedup x"
```
# Resultados Promedio (30 Corridas)
| Estrategia   | Promedio (ms) | Desv. Est. (ms) | Mediana (ms) | Mín (ms) | Máx (ms) |  n |
| ------------ | ------------: | --------------: | -----------: | -------: | -------: | -: |
| Especulativo |    **529.77** |           15.87 |        525.5 |      518 |      601 | 30 |
| Secuencial   |    **755.97** |            8.72 |        754.5 |      740 |      772 | 30 |

 $$Speedup=\frac{755.97}{​529.77} ≈ 1.43​​$$
  * Esto significa que la ejecución especulativa fue aproximadamente 43% más rápida que la versión secuencial bajo los mismos parámetros.



# Análisis de Rendimiento

* La versión especulativa logra ocultar la latencia de la evaluación del predicado al ejecutar simultáneamente las ramas A y B.

* Cuando la rama ganadora es costosa y la rama perdedora también lo es, la cancelación inmediata evita un desperdicio de tiempo, generando un speedup > 1.

* En cambio, si la rama perdedora es liviana o la decisión se toma rápido, el overhead de concurrencia puede hacer que el speedup sea cercano a 1 o incluso menor.

* En los experimentos realizados (30 repeticiones), la rama ganadora fue B (Proof-of-Work) y la ejecución especulativa obtuvo un speedup promedio de 1.43×, demostrando que este patrón concurrente aporta beneficios medibles en tareas computacionalmente intensivas.
