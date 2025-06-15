package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type BrasilAPIResponse struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type ViaCEPResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

// struct para encapsular o resultado e a origem da API
type Result struct {
	API    string
	Data   interface{}
	Err    error
	Timing time.Duration
}

func fetchBrasilAPI(ctx context.Context, cep string, ch chan<- Result) {
	start := time.Now()
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- Result{API: "BrasilAPI", Err: fmt.Errorf("erro ao criar requisição BrasilAPI: %w", err)}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- Result{API: "BrasilAPI", Err: fmt.Errorf("erro ao fazer requisição BrasilAPI: %w", err)}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ch <- Result{API: "BrasilAPI", Err: fmt.Errorf("BrasilAPI retornou status: %d", resp.StatusCode)}
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- Result{API: "BrasilAPI", Err: fmt.Errorf("erro ao ler corpo da resposta BrasilAPI: %w", err)}
		return
	}

	var brasilAPIResponse BrasilAPIResponse
	err = json.Unmarshal(body, &brasilAPIResponse)
	if err != nil {
		ch <- Result{API: "BrasilAPI", Err: fmt.Errorf("erro ao decodificar JSON BrasilAPI: %w", err)}
		return
	}
	ch <- Result{API: "BrasilAPI", Data: brasilAPIResponse, Timing: time.Since(start)}
}

func fetchViaCEP(ctx context.Context, cep string, ch chan<- Result) {
	start := time.Now()
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- Result{API: "ViaCEP", Err: fmt.Errorf("erro ao criar requisição ViaCEP: %w", err)}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- Result{API: "ViaCEP", Err: fmt.Errorf("erro ao fazer requisição ViaCEP: %w", err)}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ch <- Result{API: "ViaCEP", Err: fmt.Errorf("ViaCEP retornou status: %d", resp.StatusCode)}
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- Result{API: "ViaCEP", Err: fmt.Errorf("erro ao ler corpo da resposta ViaCEP: %w", err)}
		return
	}

	var viaCEPResponse ViaCEPResponse
	err = json.Unmarshal(body, &viaCEPResponse)
	if err != nil {
		ch <- Result{API: "ViaCEP", Err: fmt.Errorf("erro ao decodificar JSON ViaCEP: %w", err)}
		return
	}
	ch <- Result{API: "ViaCEP", Data: viaCEPResponse, Timing: time.Since(start)}
}

func main() {
	cep := "29330000"

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel() // Garante que o contexto seja cancelado no final

	ch := make(chan Result, 2) // Canal com buffer para 2 resultados

	go fetchBrasilAPI(ctx, cep, ch)
	go fetchViaCEP(ctx, cep, ch)

	// Espera pelo primeiro resultado ou timeout
	select {
	case result := <-ch:
		if result.Err != nil {
			fmt.Printf("Erro da %s: %v\n", result.API, result.Err)
		} else {
			fmt.Printf("Resultado mais rápido da API: %s (Tempo: %s)\n", result.API, result.Timing)
			fmt.Println("Dados do Endereço:")
			switch v := result.Data.(type) {
			case BrasilAPIResponse:
				fmt.Printf("  CEP: %s\n", v.Cep)
				fmt.Printf("  Estado: %s\n", v.State)
				fmt.Printf("  Cidade: %s\n", v.City)
				fmt.Printf("  Bairro: %s\n", v.Neighborhood)
				fmt.Printf("  Rua: %s\n", v.Street)
			case ViaCEPResponse:
				fmt.Printf("  CEP: %s\n", v.Cep)
				fmt.Printf("  Estado: %s\n", v.Uf)
				fmt.Printf("  Cidade: %s\n", v.Localidade)
				fmt.Printf("  Bairro: %s\n", v.Bairro)
				fmt.Printf("  Rua: %s\n", v.Logradouro)
			}
		}
	case <-ctx.Done():
		fmt.Println("Erro: Timeout de 1 segundo excedido para obter o CEP.")
	}
}
