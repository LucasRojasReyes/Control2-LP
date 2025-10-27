// Archivo Especulativo.go
// Uso desde línea de comandos (Para valores default, correr sin parámetros):
//
//	go run especulativo.go -n 500 -umbral 1000 -nombre_archivo "metricas.csv" -max_primos 1464563 -dificultad 5 -comparacion ">"
//
// Declaración del paquete al que pertenece este archivo
package main

// Imports de otros paquetes
import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

//--- Funcion de computo usada en la condicion logica costosa ---

func CalcularTrazaDeProductoDeMatrices(n int) int {
	// CalcularTrazaDeProductoDeMatrices multiplica dos matrices NxN y devuelve la traza
	// de la matriz resultante. La complejidad del cómputo es O(n^3).

	// Se crean dos matrices con valores aleatorios para la multiplicación.
	m1 := make([][]int, n)
	m2 := make([][]int, n)
	for i := 0; i < n; i++ {
		m1[i] = make([]int, n)
		m2[i] = make([]int, n)
		for j := 0; j < n; j++ {
			m1[i][j] = rand.Intn(10)
			m2[i][j] = rand.Intn(10)
		}
	}
	// Se realiza la multiplicación y se calcula la traza en el proceso.
	trace := 0
	for i := 0; i < n; i++ {
		sum := 0
		for k := 0; k < n; k++ {
			sum += m1[i][k] * m2[k][i]
		}
		trace += sum
	}
	return trace
}

//--- Funcion 1 de computo intensivo usaddas en las ramas especulativas ---

func EncontrarPrimos(ctx context.Context, max int) int {
	// EncontrarPrimos busca todos los números primos hasta un entero max.
	// Utiliza un enfoque de prueba por división, cuya complejidad es alta (aprox.O(n^1.5)).
	const checkEvery = 1024 // revisar cancelación cada 1024 números
	count := 0
	for i := 2; i < max; i++ {
		// cancelación cooperativa
		if (i % checkEvery) == 0 {
			select {
			case <-ctx.Done():
				return count
			default:
			}
		}
		isPrime := true
		for j := 2; j*j <= i; j++ {
			if i%j == 0 {
				isPrime = false
				break
			}
		}
		if isPrime {
			count++
		}
	}
	return count
}

type ProofOfWorkResult struct {
	hash  string
	nonce int
}

//--- Funcion 2 de computo intensivo usada en las ramas especulativas ---

func SimularProofOfWork(ctx context.Context, blockData string, dificultad int) ProofOfWorkResult {
	// SimularProofOfWork simula la búsqueda de una prueba de trabajo de blockchain.
	// La dificultad determina el número de ceros iniciales que debe tener el hash.
	// La complejidad crece exponencialmente con la dificultad.
	// Para un computador personal, dificultad 5-6 suele tardar unos segundos.
	// múltiples iteraciones chequean ctx.Done()
	target := strings.Repeat("0", dificultad)
	nonce := 0
	for {
		// cancelación cooperativa
		select {
		case <-ctx.Done():
			return ProofOfWorkResult{} // cancelado
		default:
		}

		data := fmt.Sprintf("%s%d", blockData, nonce)
		h := sha256.Sum256([]byte(data))
		hs := fmt.Sprintf("%x", h)
		if strings.HasPrefix(hs, target) {
			return ProofOfWorkResult{hash: hs, nonce: nonce}
		}
		nonce++
	}
}

// --- Funcion de computo usada en la condicion logica costosa ---
func cmp(ok string, lhs, rhs int) bool {
	switch ok {
	case "<":
		return lhs < rhs
	case ">":
		return lhs > rhs
	default:
		// por seguridad, trata cualquier otro valor como ">"
		return lhs > rhs
	}
}

// Función principal
func main() {
	//------ Definir y parsear parámetros de entrada -------

	// Parámetro para la dimensión de las matrices
	var dimPtr = flag.Int("n", 500, "-Dimensión de las matrices que se multiplican (n x n)")

	// Umbral para decidir qué rama ejecutar
	var umbralPtr = flag.Int("umbral", 1000, "-Valor umbral para determinar la rama a ejecutar")

	// Nombre del archivo de salida
	var nombreArchivoPtr = flag.String("nombre_archivo", "metricas.csv", "-Nombre del archivo para guardar las métricas")

	// Parámetro para la cantidad de números primos a buscar
	var maxPrimosPtr = flag.Int("max_primos", 14000, "-Número máximo hasta el cual buscar números primos")

	// Parametro para la dificultad del Proof of Work
	var dificultadPtr = flag.Int("dificultad", 5, "-Dificultad para la simulación de Proof of Work")

	// Parámetro para especificar la comparación: "<" o ">"
	var comparacionPtr = flag.String("comparacion", "<", "Tipo de comparación para el umbral (opciones: <, >)")

	// Parsear los parámetros
	flag.Parse()

	// Mostrar los valores de los parámetros recibidos
	fmt.Printf("Dimensión de matrices: %d\n", *dimPtr)
	fmt.Printf("Valor umbral: %d\n", *umbralPtr)
	fmt.Printf("Máximo número para buscar primos: %d\n", *maxPrimosPtr)
	fmt.Printf("Dificultad Proof of Work: %d\n", *dificultadPtr)
	fmt.Printf("Archivo de métricas: %s\n", *nombreArchivoPtr)
	fmt.Printf("Tipo de comparación: %s\n", *comparacionPtr)

	// Crear o abrir el archivo de métricas
	file, err := os.OpenFile(*nombreArchivoPtr, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		return
	}
	defer file.Close()

	// Escribir encabezado en el archivo de métricas solo si está vacío
	fi, err := file.Stat()
	if err != nil {
		fmt.Println("Error al obtener el estado del archivo:", err)
		return
	}
	if fi.Size() == 0 {
		_, err = file.WriteString("Inicio,Fin,Estrategia,ValorUmbral,Resultado,TiempoTotal(ms)\n")
		if err != nil {
			fmt.Println("Error al escribir en el archivo:", err)
			return
		}
	}

	//------ Definir variables y canales para la ejecución especulativa ------

	// Definir el tamaño de la matriz para la condición lógica a evaluar
	var n = *dimPtr

	// Canal que espera el resultado de la condición lógica costosa
	var predChan = make(chan int, 1)

	// Canal que espera los resultados de EncontrarPrimos
	var primesChannel = make(chan int, 1)

	// Canal que espera los resultados de SimularProofOfWork
	var simWorkChannel = make(chan ProofOfWorkResult, 1)

	// Medir tiempo de inicio
	var startTime = time.Now()
	var startIso = startTime.Format(time.RFC3339)

	// Crear un contexto para manejar la cancelación de gorutinas
	ctxA, cancelA := context.WithCancel(context.Background())
	ctxB, cancelB := context.WithCancel(context.Background())

	//------ Lanzar gorutinas para ejecución especulativa ------

	go func() {
		// predicado costoso
		val := CalcularTrazaDeProductoDeMatrices(n)
		predChan <- val
	}()

	go func() {
		// Rama A: contar primos
		var cnt = EncontrarPrimos(ctxA, *maxPrimosPtr)
		// enviar solo si no está cancelado, no bloquear
		select {
		case <-ctxA.Done():
			return
		case primesChannel <- cnt:
		default:
			// si nadie lee (por cancel), salimos
		}
	}()

	go func() {
		// Rama B: simular proof of work
		res := SimularProofOfWork(ctxB, "Bloque de datos de ejemplo", *dificultadPtr)
		select {
		case <-ctxB.Done():
			return
		case simWorkChannel <- res:
		default:
		}
	}()

	//------ Evaluar la condición lógica y seleccionar la rama especulativa válida -------

	// Variables para almacenar la rama seleccionada y el resultado (traza, primos o proof of work)
	var trace = <-predChan
	var selectedBranch string
	var resultValue interface{}

	// Evaluar la condición lógica y seleccionar la rama especulativa válida
	if cmp(*comparacionPtr, trace, *umbralPtr) {
		// Condición verdadera => tu lógica original usa Rama A
		selectedBranch = "Rama A"
		// Cancelar B inmediatamente
		cancelB()
		// Esperar resultado de A
		cnt := <-primesChannel
		resultValue = cnt
	} else {
		selectedBranch = "Rama B"
		// Cancelar A inmediatamente
		cancelA()
		// Esperar resultado de B
		res := <-simWorkChannel
		resultValue = fmt.Sprintf("Hash=%s, Nonce=%d", res.hash, res.nonce)
	}

	// Medir tiempo de fin
	endTime := time.Now()
	totalMs := endTime.Sub(startTime).Milliseconds()
	endIso := endTime.Format(time.RFC3339)

	// Mostrar resultados por consola
	fmt.Printf("Traza=%d, Umbral=%d, Comp=%s -> Seleccionada: %s\n", trace, *umbralPtr, *comparacionPtr, selectedBranch)
	fmt.Printf("Resultado: %v\n", resultValue)
	fmt.Printf("Tiempo total de ejecución: %s\n", time.Since(startTime))

	//------ Guardar métricas en el archivo -------
	// Guardar métricas
	_, err = file.WriteString(fmt.Sprintf("%s,%s,%s,%d,%s,%v,%d\n",
		startIso, endIso, selectedBranch, *umbralPtr, *comparacionPtr, resultValue, totalMs))
	if err != nil {
		fmt.Println("Error al escribir métricas:", err)
		return
	}
}
