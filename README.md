# go-stdio

Šis package yra skirtas apdoroti duomenų srautus (pvz., failus arba stdin/stdout) naudojant standartizuotą struktūrą. Jis leidžia sujungti dvi skirtingas sistemas: standartinę įvestį/išvestį ir išorinę biblioteką (blackbox). 

Pagrindinė jo funkcija yra:
1. Nuskaityti įvesties duomenis JSON formatu ir juos deserializuoti į struktūrą, kurią supranta išorinė biblioteka.
2. Panaudoti išorinę biblioteką tam tikram skaičiavimui ar apdorojimui atlikti.
3. Gautą rezultatą serializuoti atgal į JSON formatą ir išvesti į srautą (stdout arba failą).


```go
package main

import (
	"encoding/json"
	"log"
	"os"

	// Importuojame mūsų paketą
	"github.com/atvirokodosprendimai/go-stdio"

	// Tavo "Blackbox" biblioteka, kurios negali keisti
	external "github.com/some/blackbox-lib"
)

func main() {
	// Apibrėžiame Handlerį 
	invoiceHandler := func(req stdio.Request) {
		// 1. Deserializacija (iš gautų baitų į Blackbox struktūrą)
		var inv external.Invoice
		if err := json.Unmarshal(req.Data(), &inv); err != nil {
			log.Printf("JSON klaida: %v", err)
			return // Arba galime rašyti klaidą į output
		}

		// 2. Kviečiame Blackbox logiką
		// Tarkime biblioteka turi funkciją ProcessInvoice(inv) -> Result
		result := external.ProcessInvoice(inv)

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
```
