package main

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////IMPORTS//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////STRUCTS//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type CpuConfig struct {
	Port               int    `json:"port"`
	Ip_Memory          string `json:"ip_memory"`
	Port_Memory        int    `json:"port_memory"`
	Number_Felling_tlb int    `json: "number_felling_tlb"`
	Algorithm_tlb      string `json: "algorithm_tlb"`
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////MAIN///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func main() {

	// Se establece el handler que se utilizará para las diversas situaciones recibidas por el server
	http.HandleFunc("PUT /plani", handler_iniciar_planificacion)
	http.HandleFunc("DELETE /plani", handler_detener_planificacion)

	// Extrae info de config.json
	config := iniciarConfiguracion("config.json")

	// declaro puerto
	port := ":" + strconv.Itoa(config.Port)

	// Listen and serve con info del config.json
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al esuchar en el puerto " + port)
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////FUNCIONES/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func iniciarConfiguracion(filePath string) *CpuConfig {
	//En el tp0 usan punteros y guardan la variable en un archivo "globals".
	// No estoy seguro del motivo, y por ahora no lo veo necesario
	var config *CpuConfig

	// Abre el archivo
	configFile, err := os.Open(filePath)
	if err != nil {
		// log.Fatal(err.Error())
		fmt.Println("Error: ", err)
	}
	// Cierra el archivo una vez que la función termina (ejecuta el return)
	defer configFile.Close()

	// Decodifica la info del json en la variable config
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	// Devuelve config
	return config
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////HANDLERS//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func handler_iniciar_planificacion(w http.ResponseWriter, r *http.Request) {

	// Respuesta vacía significa que manda una respuesta vacía, o que no hay respuesta?
	respuesta, err := json.Marshal("")

	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

func handler_detener_planificacion(w http.ResponseWriter, r *http.Request) {

	// Respuesta vacía significa que manda una respuesta vacía, o que no hay respuesta?
	respuesta, err := json.Marshal("")

	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}
