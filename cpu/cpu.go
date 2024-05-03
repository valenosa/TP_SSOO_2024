package main

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////IMPORTS//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/utils/APIs/kernel-cpu/planificacion"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////MAIN///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func main() {

	// Configura el logger
	config.Logger("CPU.log")

	log.Printf("Soy un logeano")

	// Se establece el handler que se utilizará para las diversas situaciones recibidas por el server
	http.HandleFunc("PUT /plani", planificacion.HandlerIniciar)
	http.HandleFunc("DELETE /plani", planificacion.HandlerDetener)

	// Extrae info de config.json
	var configJson config.Cpu

	config.Iniciar("config.json", &configJson)

	// declaro puerto
	port := ":" + strconv.Itoa(configJson.Port)

	// Listen and serve con info del config.json
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al esuchar en el puerto " + port)
	}
}
