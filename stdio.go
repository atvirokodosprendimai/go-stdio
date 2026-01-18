package stdio

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

// Request interfeisas imituoja micro.Request
type Request interface {
	// Data grąžina input duomenis (byte slice)
	Data() []byte
	// Respond išsiunčia atsakymą į output (failą arba stdout)
	Respond(data []byte) error
}

// HandlerFunc yra funkcija, kuri apdoroja užklausą
type HandlerFunc func(req Request)

// Config struktūra, identiška micro.Config stiliui
type Config struct {
	Name        string
	Version     string
	Description string
	Endpoint    *EndpointConfig
}

// EndpointConfig aprašo, kas apdoros duomenis
type EndpointConfig struct {
	Subject string      // CLI atveju tai tik dekoracija arba logams
	Handler HandlerFunc // Tavo funkcija
}

// -- Vidinė implementacija --

// stdRequest realizuoja Request interfeisą
type stdRequest struct {
	data   []byte
	output io.Writer
}

func (r *stdRequest) Data() []byte {
	return r.data
}

func (r *stdRequest) Respond(data []byte) error {
	_, err := r.output.Write(data)
	return err
}

// RunService yra "entry point", kuris pakeičia micro.AddService
func RunService(cfg Config) error {
	// 1. Setup - Flagai ir Loggeriai
	// Naudojame atskirą FlagSet, kad netrukdytų, jei programa turi kitų init
	fs := flag.NewFlagSet(cfg.Name, flag.ExitOnError)
	useStd := fs.Bool("std", false, "Naudoti stdin/stdout pipe režimą")
	showInfo := fs.Bool("info", false, "Rodyti serviso informaciją")
	showVer := fs.Bool("version", false, "Rodyti versiją")

	// Parsiname argumentus
	fs.Parse(os.Args[1:])
	args := fs.Args()

	if *showVer {
		fmt.Printf("%s v%s\n", cfg.Name, cfg.Version)
		return nil
	}

	if *showInfo {
		printInfo(cfg)
		return nil
	}

	if !*useStd && len(fs.Args()) < 2 {
		printInfo(cfg) // Jei vartotojas nieko nepadavė, parodome info kaip pagalbą
		fmt.Fprintf(os.Stderr, "\nNaudojimas: %s <input> <output> ARBA --std\n", os.Args[0])
		return fmt.Errorf("trūksta argumentų")
	}

	logger := log.New(os.Stderr, fmt.Sprintf("[%s %s] ", cfg.Name, cfg.Version), log.LstdFlags)

	var input io.Reader = os.Stdin
	var output io.Writer = os.Stdout
	var closers []io.Closer

	// 2. IO Konfigūracija
	if !*useStd {
		if len(args) < 2 {
			// Jei nėra argumentų ir nėra --std, rodome pagalbą
			fmt.Fprintf(os.Stderr, "Usage: %s <input> <output> OR %s --std\n", cfg.Name, cfg.Name)
			return fmt.Errorf("neteisingi argumentai")
		}

		fIn, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("klaida atidarant input: %w", err)
		}
		input = fIn
		closers = append(closers, fIn)

		fOut, err := os.Create(args[1])
		if err != nil {
			return fmt.Errorf("klaida kuriant output: %w", err)
		}
		output = fOut
		closers = append(closers, fOut)
	} else {
		logger.Println("Startuojama PIPE režimu...")
	}

	defer func() {
		for _, c := range closers {
			c.Close()
		}
	}()

	// 3. Nuskaitome duomenis ("Request")
	// Pastaba: Dideliems failams reiktų srautinio apdorojimo,
	// bet Request.Data() pattern'as implikuoja, kad turime viską atmintyje.
	inputData, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("klaida skaitant duomenis: %w", err)
	}

	req := &stdRequest{
		data:   inputData,
		output: output,
	}

	// 4. Kviečiame Handlerį
	logger.Printf("Apdorojama užklausa (%d bytes)...", len(inputData))

	// Čia kviečiame vartotojo funkciją
	cfg.Endpoint.Handler(req)

	logger.Println("Baigta.")
	return nil
}

func printInfo(cfg Config) {
	json.NewEncoder(os.Stdout).Encode(cfg)
}
