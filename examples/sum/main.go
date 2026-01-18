package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/atvirokodosprendimai/go-stdio"
)

func main() {
	// Apibrėžiame Handlerį
	invoiceHandler := func(req stdio.Request) {
		// 1. Deserializacija (iš gautų baitų į Blackbox struktūrą)
		var in ExternalType
		if err := json.Unmarshal(req.Data(), &in); err != nil {
			log.Printf("JSON klaida: %v", err)
			return // Arba galime rašyti klaidą į output
		}

		// 2. Kviečiame Blackbox logiką
		// Tarkime biblioteka turi funkciją ProcessInvoice(inv) -> Result
		result := ToUppercase(in)

		// 3. Serializacija (iš Blackbox rezultato į baitus)
		respBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Printf("Marshal klaida: %v", err)
			return
		}

		// 4. Atsakymas (Respond)
		req.Respond(respBytes)
	}

	// Konfigūracija ir Paleidimas
	err := stdio.RunService(stdio.Config{
		Name:    "InvoiceProcessor",
		Version: "1.0.0",
		Endpoint: &stdio.EndpointConfig{
			Subject: "file.process", // Čia tik dėl suderinamumo/informacijos
			Handler: stdio.HandlerFunc(invoiceHandler),
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}

type ExternalType struct {
	Name string `json:"name"`
}

func ToUppercase(in ExternalType) ExternalType {
	return ExternalType{
		Name: strings.ToUpper(in.Name),
	}
}
