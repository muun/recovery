package main

import (
	"fmt"

	"github.com/muun/recovery/electrum"
	"github.com/muun/recovery/scanner"
)

var failedToConnect []string
var withBatching []string
var withoutBatching []string

func main() {
	client := electrum.NewClient()

	for _, server := range scanner.PublicElectrumServers {
		surveyServer(client, server)
	}

	fmt.Println("// With batch support:")
	for _, server := range withBatching {
		fmt.Printf("\"%s\"\n", server)
	}

	fmt.Println("// Without batch support:")
	for _, server := range withoutBatching {
		fmt.Printf("\"%s\"\n", server)
	}

	fmt.Println("// Unclassified:")
	for _, server := range failedToConnect {
		fmt.Printf("\"%s\"\n", server)
	}
}

func surveyServer(client *electrum.Client, server string) {
	fmt.Println("Surveyng", server)
	err := client.Connect(server)

	if err != nil {
		failedToConnect = append(failedToConnect, server)
		return
	}

	if client.SupportsBatching() {
		withBatching = append(withBatching, server)
	} else {
		withoutBatching = append(withoutBatching, server)
	}
}
