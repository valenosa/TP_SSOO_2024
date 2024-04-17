package main

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////IMPORTS//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
	/*
		read(GET, body) => que hay que hacer
		res
		switch res
	*/
	//finalizar_proceso()
	detener_planificacion()
	iniciar_planificacion()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////API's//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Solamente esqueleto
func iniciar_proceso() {

	// Establecer ip_memory y puerto (hardcodeado)
	ip_memory := "localhost"
	port_memory := 8002

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
	url := fmt.Sprintf("http://%s:%d/process", ip_memory, port_memory)

	// Se crea una request donde se "efectúa" un PUT hacia url, enviando el Body anteriormente mencionado
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))

	// Error Handler de la construcción de la request
	if err != nil {
		fmt.Printf("error creando request a ip: %s puerto: %d\n", ip_memory, port_memory)
		return
	}

	// Se establecen los headers
	req.Header.Set("Content-Type", "application/json")

	// Se envía el request al servidor
	respuesta, err := cliente.Do(req)

	// Error handler de la request
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", ip_memory, port_memory)
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
func finalizar_proceso() {

	// Establecer ip, puerto y pid (hardcodeado)
	ip_memory := "localhost"
	port_memory := 8002
	pid := 0

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/process/%d", ip_memory, port_memory, pid)

	// Se crea una request donde se "efectúa" un GET hacia la url
	req, err := http.NewRequest("DELETE", url, nil)

	// Error Handler de la construcción de la request
	if err != nil {
		fmt.Printf("error creando request a ip: %s puerto: %d\n", ip_memory, port_memory)
		return
	}

	// Se establecen los headers
	req.Header.Set("Content-Type", "application/json")

	// Se envía el request al servidor
	respuesta, err := cliente.Do(req)

	// Error handler de la request
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", ip_memory, port_memory)
		return
	}

	// Verificar el código de estado de la respuesta del servidor a nuestra request (de no ser OK)
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}
}

func estado_proceso() {
	//implementar
}

func iniciar_planificacion() {

	// Establecer ip_cpu y puerto
	ip_cpu := "localhost"
	port_cpu := 8006

	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/plani", ip_cpu, port_cpu)

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
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", ip_cpu, port_cpu)
		return
	}

	//Espera a que la respuesta se termine de utilizar para liberarla de memoria.
	//defer respuesta.Body.Close()

	// Check response recibida.
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}

	// Todo salió bien, la planificación se inició correctamente.
	fmt.Println("Planificación iniciada exitosamente.")
}

func detener_planificacion() {
	// Establecer ip_cpu y puerto
	ip_cpu := "localhost"
	port_cpu := 8006

	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/plani", ip_cpu, port_cpu)

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
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", ip_cpu, port_cpu)
		return
	}

	//Espera a que la respuesta se termine de utilizar para liberarla de memoria.
	// defer respuesta.Body.Close()

	// Check response recibida.
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}

	// Todo salió bien, la planificación se detuvo correctamente.
	fmt.Println("Planificación detenida exitosamente.")
}

func listar_proceso() {
	//implementar
}

func dispatch() {

}

func interrupt() {

}
