package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type inputOutpConfig struct {
	Port               int    `json:"port"`
	Type               string `json:"type"`
	Unit_Work_Time     int    `json:"unit_work_time"`
	Ip_Kernel          string `json:"ip_kernel"`
	Port_Kernel        int    `json:"port_kernel"`
	Ip_Memory          string `json:"ip_memory"`
	Port_Memory        int    `json:"port_memory"`
	Dialfs_Path        string `json:"dialfs_path"`
	Dialfs_Block_Size  int    `json:"dialfs_block_size"`
	Dialfs_Block_Count int    `json:"dialfs_block_count"`
}

func main() {

	config := iniciarConfiguracion("config.json")

	http.HandleFunc("GET /holamundo", entradaSalida)

	port := ":" + strconv.Itoa(config.Port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al esuchar en el puerto " + port)
	}
	/*if err == nil {
		fmt.Println("Estoy escuchando en el puerto " + port)
	} else {
		fmt.Println("Error al esuchar en el puerto " + port)
	}
	PARA LOS LOGS*/
}

func iniciarConfiguracion(filePath string) *inputOutpConfig {
	//En el tp0 usan punteros y guardan la variable en un archivo "globals".
	// No estoy seguro del motivo, y por ahora no lo veo necesario
	var config *inputOutpConfig

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
