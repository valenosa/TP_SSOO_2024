package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	/*
		read(GET, body) => que hay que hacer
		res
		switch res
	*/
	iniciar_proceso()
}

type BodyIniciarProceso struct {
	// Path del archivo que se utilizará como base para ejecutar un nuevo proceso
	Path string `json:"path"`
}

type ResponseIniciarProceso struct {
	Pid string `json:"pid"`
}

func iniciar_proceso() {

	// Establecer ip y puerto
	ip := "localhost"
	puerto := 8080

	// Codificar Body en un array de bytes (formato json)
	body, err := json.Marshal(BodyIniciarProceso{
		Path: "",
	})
	// Error Handler de la codificación
	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
	}

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto)
	url := fmt.Sprintf("http://%s:%d/paquetes", ip, puerto)

	// Se crea una request donde se "efectúa" un PUT hacia url, enviando el Body anteriormente mencionado
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))

	// Error Handler de la construcción de la request
	if err != nil {
		fmt.Printf("error creando request a ip:%s puerto:%d", ip, puerto)
	}

	// Se establecen los headers
	req.Header.Set("Content-Type", "application/json")

	// Se envía el request al servidor
	respuesta, err := cliente.Do(req)

	// Error handler de la request
	if err != nil {
		fmt.Printf("error enviando request a ip:%s puerto:%d", ip, puerto)
	}

	// Verificar el código de estado de la respuesta del servidor a nuestra request (de no ser OK)
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d", respuesta.StatusCode)
	}

	// Se declara una nueva variable que contendrá la respuesta del servidor
	var response ResponseIniciarProceso

	// Se decodifica la variable (codificada en formato json) en la estructura correspondiente
	err = json.NewDecoder(respuesta.Body).Decode(&response)

	// Error Handler para al decodificación
	if err != nil {
		fmt.Printf("Error decodificando")
	}

	// Imprime pid (parámetro de la estructura)
	fmt.Println(response.Pid)
}

func finalizar_proceso() {
	//implementar
}

func estado_proceso() {
	//implementar
}

func iniciar_planificacion() {
	//implementar
}

func detener_planificacion() {
	//implementar
}

func listar_proceso() {
	//implementar
}

func dispatch() {

}

func interrupt() {

}
