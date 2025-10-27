package main

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

// ----- mismas funciones / tipos que en Especulativo -----

func CalcularTrazaDeProductoDeMatrices(n int) int {
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

func EncontrarPrimos(ctx context.Context, max int) int {
	const checkEvery = 1024
	count := 0
	for i := 2; i < max; i++ {
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

func SimularProofOfWork(ctx context.Context, blockData string, dificultad int) ProofOfWorkResult {
	target := strings.Repeat("0", dificultad)
	nonce := 0
	for {
		select {
		case <-ctx.Done():
			return ProofOfWorkResult{}
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

func cmp(op string, lhs, rhs int) bool {
	switch op {
	case "<":
		return lhs < rhs
	case ">":
		return lhs > rhs
	default:
		return lhs > rhs
	}
}

// ----- main secuencial (solo ejecuta la rama elegida, pero con las mismas funciones) -----

func main() {
	var dimPtr = flag.Int("n", 500, "-Dimensión de las matrices (n x n)")
	var umbralPtr = flag.Int("umbral", 1000, "-Umbral para decidir rama")
	var nombreArchivoPtr = flag.String("nombre_archivo", "metricas.csv", "-Archivo CSV")
	var maxPrimosPtr = flag.Int("max_primos", 14000, "-Máximo hasta el cual buscar primos")
	var dificultadPtr = flag.Int("dificultad", 5, "-Dificultad PoW")
	var comparacionPtr = flag.String("comparacion", "<", "Comparación: < o >")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	fmt.Printf("Dimensión de matrices: %d\n", *dimPtr)
	fmt.Printf("Valor umbral: %d\n", *umbralPtr)
	fmt.Printf("Máximo número para buscar primos: %d\n", *maxPrimosPtr)
	fmt.Printf("Dificultad Proof of Work: %d\n", *dificultadPtr)
	fmt.Printf("Archivo de métricas: %s\n", *nombreArchivoPtr)
	fmt.Printf("Tipo de comparación: %s\n", *comparacionPtr)

	file, err := os.OpenFile(*nombreArchivoPtr, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		return
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		fmt.Println("Error al obtener el estado del archivo:", err)
		return
	}
	if fi.Size() == 0 {
		_, err = file.WriteString("Inicio,Fin,Estrategia,ValorUmbral,Comparacion,Resultado,TiempoTotal(ms)\n")
		if err != nil {
			fmt.Println("Error al escribir encabezado:", err)
			return
		}
	}

	n := *dimPtr
	startTime := time.Now()
	startIso := startTime.Format(time.RFC3339)
	// Predicado (misma función que en Especulativo)
	trace := CalcularTrazaDeProductoDeMatrices(n)

	var selectedBranch string
	var resultValue interface{}

	if cmp(*comparacionPtr, trace, *umbralPtr) {
		selectedBranch = "Rama A"
		// misma función con ctx
		cnt := EncontrarPrimos(context.Background(), *maxPrimosPtr)
		resultValue = cnt
		fmt.Printf("Traza=%d -> Rama A | Primos hasta %d: %d\n", trace, *maxPrimosPtr, cnt)
	} else {
		selectedBranch = "Rama B"
		res := SimularProofOfWork(context.Background(), "Bloque de datos de ejemplo", *dificultadPtr)
		resultValue = fmt.Sprintf("Hash=%s, Nonce=%d", res.hash, res.nonce)
		fmt.Printf("Traza=%d -> Rama B | PoW: Hash=%s, Nonce=%d\n", trace, res.hash, res.nonce)
	}

	endTime := time.Now()
	totalMs := endTime.Sub(startTime).Milliseconds()
	endIso := endTime.Format(time.RFC3339)

	fmt.Printf("Tiempo total de ejecución (rama): %s\n", endTime.Sub(startTime))

	_, err = file.WriteString(fmt.Sprintf("%s,%s,%s,%d,%s,%v,%d\n",
		startIso, endIso, selectedBranch, *umbralPtr, *comparacionPtr, resultValue, totalMs))
	if err != nil {
		fmt.Println("Error al escribir métricas:", err)
		return
	}
}
