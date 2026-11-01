# Clima por CEP

API em Go que recebe um CEP, descobre a cidade pelo [ViaCEP](https://viacep.com.br) e retorna a temperatura atual (via [WeatherAPI](https://www.weatherapi.com)) em Celsius, Fahrenheit e Kelvin.

## URL em produção (Google Cloud Run)

https://clima-por-cep-640606615516.southamerica-east1.run.app

Exemplo de consulta:
https://clima-por-cep-640606615516.southamerica-east1.run.app/weather/01001000

## Contrato

`GET /weather/{cep}` (também aceita `GET /weather?cep={cep}`)

| Cenário | Status | Corpo |
|---|---|---|
| Sucesso | `200` | `{"temp_C": 28.5, "temp_F": 83.3, "temp_K": 301.5}` |
| CEP sem 8 dígitos ou com caracteres inválidos | `422` | `invalid zipcode` |
| CEP válido, mas inexistente | `404` | `can not find zipcode` |
| ViaCEP ou WeatherAPI indisponível | `502` | `upstream error` |

Conversões: `F = C × 1.8 + 32` e `K = C + 273`, arredondadas para 2 casas.

```bash
curl https://clima-por-cep-640606615516.southamerica-east1.run.app/weather/01001000
```

## Rodando localmente com Docker

Pré-requisitos: Docker e uma chave gratuita do [WeatherAPI](https://www.weatherapi.com/signup.aspx).

```bash
cp .env.example .env        # edite e coloque sua WEATHER_API_KEY
docker compose up --build
```

A aplicação sobe em `http://localhost:8080`:

```bash
curl http://localhost:8080/weather/01001000
```

## Rodando os testes

Via Docker (não precisa de Go instalado nem de chave):

```bash
docker compose --profile test run --rm test
```

Ou localmente, com Go 1.27+:

```bash
go test -v -cover ./...
```

Os testes não acessam a internet: ViaCEP e WeatherAPI são substituídos por servidores `httptest`.

## Deploy no Cloud Run

```bash
gcloud auth login
gcloud config set project SEU_PROJETO

# guarda a chave no Secret Manager
printf '%s' "SUA_CHAVE" | gcloud secrets create weather-api-key --data-file=-

gcloud run deploy clima-por-cep \
  --source . \
  --region southamerica-east1 \
  --allow-unauthenticated \
  --set-secrets WEATHER_API_KEY=weather-api-key:latest
```

O comando imprime a URL do serviço; copie-a para a seção acima.

## Estrutura

| Arquivo | Responsabilidade |
|---|---|
| `main.go` | Configuração (`PORT`, `WEATHER_API_KEY`) e start do servidor |
| `handler.go` | Rota, validação do CEP e mapeamento de erros para status HTTP |
| `viacep.go` / `weather.go` | Clientes das APIs externas |
| `convert.go` | Conversões de temperatura |
