package notificacao

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func EnviarWebhookAlerta(cnpj string) {
	payload := map[string]string{
		"mensagem": "Alerta: nova negociação iniciada com CNPJ já existente",
		"cnpj":     cnpj,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post("https://seu-webhook-url.com/alerta", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Erro ao enviar webhook: %v", err)
		return
	}
	defer resp.Body.Close()
}
