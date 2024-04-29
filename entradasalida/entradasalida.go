package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/utils/config"
)

func main() {

	// Establezco petición
	http.HandleFunc("GET /holamundo", entradaSalida)

	// Extrae info de config.json
	var configJson config.IO

	config.Iniciar("config.json", &configJson)

	// declaro puerto
	port := ":" + strconv.Itoa(configJson.Port)

	// Listen and serve con info del config.json
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al esuchar en el puerto " + port)
	}
}

func entradaSalida(w http.ResponseWriter, r *http.Request) {

	respuesta, err := json.Marshal("Hello world! Soy una consola de I/O")

	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)

	fmt.Println("Hello world! Soy una consola de I/O")
}
