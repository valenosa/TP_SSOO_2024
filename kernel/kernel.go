package main

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////IMPORTS//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/utils/APIs/kernel-cpu/planificacion"
	"github.com/sisoputnfrba/tp-golang/utils/APIs/kernel-memoria/proceso"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////STRUCTS//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////MAIN///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func main() {

	// Extrae info de config.json
	var configJson config.Kernel

	config.Iniciar("config.json", &configJson)

	proceso.Iniciar(configJson)
	proceso.Finalizar(configJson)
	proceso.Estado(configJson)
	proceso.Listar(configJson)
	planificacion.Detener(configJson)
	planificacion.Iniciar(configJson)

	//Establezco petición
	http.HandleFunc("GET /holamundo", kernel)

	// declaro puerto
	port := ":" + strconv.Itoa(configJson.Port)

	// Listen and serve con info del config.json
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al esuchar en el puerto " + port)
	}

}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////FUNCIONES/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func kernel(w http.ResponseWriter, r *http.Request) {

	respuesta, err := json.Marshal("Hello world! Soy una consola del kernel")

	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)

	fmt.Println("Hello world! Soy una consola del kernel")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////API's//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func dispatch() {

}

func interrupt() {

}
