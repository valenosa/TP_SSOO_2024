package main

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////IMPORTS//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/utils/config"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////STRUCTS//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type BodyIniciarProceso struct {
	// Path del archivo que se utilizará como base para ejecutar un nuevo proceso
	Path string `json:"path"`
}

type ResponseProceso struct {
	Pid    int    `json:"pid"`
	Estado string `json:"estado"`
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////MAIN///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func main() {

	// Extrae info de config.json
	var configJson config.Kernel

	config.Iniciar("config.json", &configJson)

	iniciar_proceso(configJson)
	finalizar_proceso(configJson)
	estado_proceso(configJson)
	detener_planificacion(configJson)
	iniciar_planificacion(configJson)
	listar_proceso(configJson)

	// Establezco petición
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

// Solamente esqueleto
func iniciar_proceso(configJson config.Kernel) {

	// Codificar Body en un array de bytes (formato json)
	body, err := json.Marshal(BodyIniciarProceso{
		Path: "string",
	})
	// Error Handler de la codificación
	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
		return
	}

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/process", configJson.Ip_Memory, configJson.Port_Memory)

	// Se crea una request donde se "efectúa" un PUT hacia url, enviando el Body anteriormente mencionado
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))

	// Error Handler de la construcción de la request
	if err != nil {
		fmt.Printf("error creando request a ip: %s puerto: %d\n", configJson.Ip_Memory, configJson.Port_Memory)
		return
	}

	// Se establecen los headers
	req.Header.Set("Content-Type", "application/json")

	// Se envía el request al servidor
	respuesta, err := cliente.Do(req)

	// Error handler de la request
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", configJson.Ip_Memory, configJson.Port_Memory)
		return
	}

	// Verificar el código de estado de la respuesta del servidor a nuestra request (de no ser OK)
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}

	// Se declara una nueva variable que contendrá la respuesta del servidor
	var response ResponseProceso

	// Se decodifica la variable (codificada en formato json) en la estructura correspondiente
	err = json.NewDecoder(respuesta.Body).Decode(&response)

	// Error Handler para al decodificación
	if err != nil {
		fmt.Printf("Error decodificando\n")
		return
	}

	// Imprime pid (parámetro de la estructura)
	fmt.Printf("pid: %d\n", response.Pid)
}

// Solamente esqueleto
func finalizar_proceso(configJson config.Kernel) {

	// Establecer pid (hardcodeado)
	pid := 0

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/process/%d", configJson.Ip_Memory, configJson.Port_Memory, pid)

	// Se crea una request donde se "efectúa" un GET hacia la url
	req, err := http.NewRequest("DELETE", url, nil)

	// Error Handler de la construcción de la request
	if err != nil {
		fmt.Printf("error creando request a ip: %s puerto: %d\n", configJson.Ip_Memory, configJson.Port_Memory)
		return
	}

	// Se establecen los headers
	req.Header.Set("Content-Type", "application/json")

	// Se envía el request al servidor
	respuesta, err := cliente.Do(req)

	// Error handler de la request
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", configJson.Ip_Memory, configJson.Port_Memory)
		return
	}

	// Verificar el código de estado de la respuesta del servidor a nuestra request (de no ser OK)
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}
}

func estado_proceso(configJson config.Kernel) {

	// Establecer pid (hardcodeado)
	pid := 0

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/process/%d", configJson.Ip_Memory, configJson.Port_Memory, pid)

	// Se crea una request donde se "efectúa" un GET hacia la url
	req, err := http.NewRequest("GET", url, nil)

	// Error Handler de la construcción de la request
	if err != nil {
		fmt.Printf("error creando request a ip: %s puerto: %d\n", configJson.Ip_Memory, configJson.Port_Memory)
		return
	}

	// Se establecen los headers
	req.Header.Set("Content-Type", "application/json")

	// Se envía el request al servidor
	respuesta, err := cliente.Do(req)

	// Error handler de la request
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", configJson.Ip_Memory, configJson.Port_Memory)
		return
	}

	// Verificar el código de estado de la respuesta del servidor a nuestra request (de no ser OK)
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}

	// Se declara una nueva variable que contendrá la respuesta del servidor
	var response ResponseProceso

	// Se decodifica la variable (codificada en formato json) en la estructura correspondiente
	err = json.NewDecoder(respuesta.Body).Decode(&response)

	// Error Handler para al decodificación
	if err != nil {
		fmt.Printf("Error decodificando\n")
		fmt.Println(err)
		return
	}

	// Imprime pid (parámetro de la estructura)
	fmt.Println(response)

}

func iniciar_planificacion(configJson config.Kernel) {

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/plani", configJson.Ip_CPU, configJson.Port_CPU)

	// Genera una petición HTTP.
	req, err := http.NewRequest("PUT", url, nil)

	// Check error generando una request.
	if err != nil {
		fmt.Printf("Error creando request: %s\n", err.Error())
		return
	}

	// Se envía la request al servidor.
	respuesta, err := cliente.Do(req)

	// Check request enviada.
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", configJson.Ip_CPU, configJson.Port_CPU)
		return
	}

	//Espera a que la respuesta se termine de utilizar para liberarla de memoria.
	defer respuesta.Body.Close()

	// Check response recibida.
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}

	// Todo salió bien, la planificación se inició correctamente.
	fmt.Println("Planificación iniciada exitosamente.")
}

func detener_planificacion(configJson config.Kernel) {

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/plani", configJson.Ip_CPU, configJson.Port_CPU)

	// Genera una petición HTTP.
	req, err := http.NewRequest("DELETE", url, nil)

	// Check error generando una request.
	if err != nil {
		fmt.Printf("Error creando request: %s\n", err.Error())
		return
	}

	// Se envía la request al servidor.
	respuesta, err := cliente.Do(req)

	// Check request enviada.
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", configJson.Ip_CPU, configJson.Port_CPU)
		return
	}

	//Espera a que la respuesta se termine de utilizar para liberarla de memoria.
	defer respuesta.Body.Close()

	// Check response recibida.
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}

	// Todo salió bien, la planificación se detuvo correctamente.
	fmt.Println("Planificación detenida exitosamente.")
}

/*
Se encargará de mostrar por consola y retornar por la api el listado de procesos
que se encuentran en el sistema con su respectivo estado dentro de cada uno de ellos.
*/
func listar_proceso(configJson config.Kernel) {

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/process", configJson.Ip_Memory, configJson.Port_Memory)

	// Genera una petición HTTP.
	req, err := http.NewRequest("GET", url, nil)

	// Check error generando una request.
	if err != nil {
		fmt.Printf("Error creando request: %s\n", err.Error())
		return
	}

	// Se envía la request al servidor.
	respuesta, err := cliente.Do(req)

	// Check request enviada.
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", configJson.Ip_Memory, configJson.Port_Memory)
		return
	}

	//Espera a que la respuesta se termine de utilizar para liberarla de memoria.
	defer respuesta.Body.Close()

	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return
	}

	fmt.Println(string(bodyBytes))
}

func dispatch() {

}

func interrupt() {

}
