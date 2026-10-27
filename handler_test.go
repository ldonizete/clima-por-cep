package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestServer wires the handler to fake ViaCEP and WeatherAPI backends.
func newTestServer(t *testing.T, viacep, weather http.HandlerFunc) *server {
	t.Helper()

	via := httptest.NewServer(viacep)
	t.Cleanup(via.Close)
	wx := httptest.NewServer(weather)
	t.Cleanup(wx.Close)

	return &server{
		viacep:  &viaCEPClient{baseURL: via.URL, http: via.Client()},
		weather: &weatherClient{baseURL: wx.URL, apiKey: "test-key", http: wx.Client()},
		timeout: time.Second,
	}
}

func get(s *server, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	s.routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func viaCEPOK(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"cep":"01001-000","localidade":"São Paulo","uf":"SP"}`))
}

func weatherOK(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"current":{"temp_c":28.5}}`))
}

func unexpectedCall(t *testing.T, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("%s must not be called", name)
	}
}

func TestWeatherSuccess(t *testing.T) {
	var gotQuery string
	s := newTestServer(t, viaCEPOK, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("q")
		if r.URL.Query().Get("key") != "test-key" {
			t.Errorf("key = %q", r.URL.Query().Get("key"))
		}
		weatherOK(w, r)
	})

	for _, target := range []string{"/weather/01001000", "/weather?cep=01001000"} {
		rec := get(s, target)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, want 200", target, rec.Code)
		}
		var got weatherResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if want := (weatherResponse{TempC: 28.5, TempF: 83.3, TempK: 301.5}); got != want {
			t.Errorf("%s: body = %+v, want %+v", target, got, want)
		}
	}
	if gotQuery != "São Paulo, SP, Brazil" {
		t.Errorf("weather q = %q", gotQuery)
	}
}

func TestWeatherJSONKeys(t *testing.T) {
	s := newTestServer(t, viaCEPOK, weatherOK)

	body := get(s, "/weather/01001000").Body.String()
	for _, key := range []string{`"temp_C"`, `"temp_F"`, `"temp_K"`} {
		if !strings.Contains(body, key) {
			t.Errorf("body %s lacks %s", body, key)
		}
	}
}

func TestWeatherInvalidZipcode(t *testing.T) {
	s := newTestServer(t, unexpectedCall(t, "viacep"), unexpectedCall(t, "weather"))

	for _, target := range []string{
		"/weather/1234567",   // 7 digits
		"/weather/123456789", // 9 digits
		"/weather/0100100a",  // letter
		"/weather/01001-000", // hyphen
		"/weather/%20%20%20%20%20%20%20%20",
		"/weather", // missing
		"/weather?cep=abc",
	} {
		rec := get(s, target)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status = %d, want 422", target, rec.Code)
		}
		if got := strings.TrimSpace(rec.Body.String()); got != "invalid zipcode" {
			t.Errorf("%s: body = %q", target, got)
		}
	}
}

func TestWeatherZipcodeNotFound(t *testing.T) {
	for name, body := range map[string]string{
		"bool flag":   `{"erro": true}`,
		"string flag": `{"erro": "true"}`,
	} {
		t.Run(name, func(t *testing.T) {
			s := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(body))
			}, unexpectedCall(t, "weather"))

			rec := get(s, "/weather/99999999")

			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", rec.Code)
			}
			if got := strings.TrimSpace(rec.Body.String()); got != "can not find zipcode" {
				t.Errorf("body = %q", got)
			}
		})
	}
}

func TestWeatherCityUnknownToWeatherAPI(t *testing.T) {
	s := newTestServer(t, viaCEPOK, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"code":1006,"message":"No matching location found."}}`))
	})

	rec := get(s, "/weather/01001000")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestWeatherUpstreamFailures(t *testing.T) {
	fail := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}

	t.Run("viacep down", func(t *testing.T) {
		s := newTestServer(t, fail, unexpectedCall(t, "weather"))
		if rec := get(s, "/weather/01001000"); rec.Code != http.StatusBadGateway {
			t.Errorf("status = %d, want 502", rec.Code)
		}
	})

	t.Run("weatherapi down", func(t *testing.T) {
		s := newTestServer(t, viaCEPOK, fail)
		if rec := get(s, "/weather/01001000"); rec.Code != http.StatusBadGateway {
			t.Errorf("status = %d, want 502", rec.Code)
		}
	})

	t.Run("weatherapi bad key", func(t *testing.T) {
		s := newTestServer(t, viaCEPOK, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":{"code":2008,"message":"API key has been disabled."}}`))
		})
		if rec := get(s, "/weather/01001000"); rec.Code != http.StatusBadGateway {
			t.Errorf("status = %d, want 502", rec.Code)
		}
	})
}
