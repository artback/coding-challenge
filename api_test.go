package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	SimpleContentRequest = httptest.NewRequest("GET", "/?offset=0&count=5", nil)
	LargeContentRequest  = httptest.NewRequest("GET", "/?offset=0&count=125", nil)
	OffsetContentRequest = httptest.NewRequest("GET", "/?offset=5&count=5", nil)
	wrongOffsetRequest   = httptest.NewRequest("GET", "/?offset=hello&count=5", nil)
	missingCountRequest  = httptest.NewRequest("GET", "/?offset=0", nil)
	wrongMethodRequest   = httptest.NewRequest("POST", "/?offset=0&count=5", nil)
)

func runRequest(srv http.Handler, r *http.Request, code int) ([]*ContentItem, error) {
	var content []*ContentItem
	response := httptest.NewRecorder()
	srv.ServeHTTP(response, r)

	if response.Code != code {
		return nil, fmt.Errorf("Response code is %d, want %d", response.Code, code)
	}

	err := json.NewDecoder(response.Body).Decode(&content)
	if err != nil {
		return nil, fmt.Errorf("couldn't decode Response json: %v", err)
	}

	return content, nil
}

func TestResponseCount(t *testing.T) {
	content, err := runRequest(app, SimpleContentRequest, 200)
	if err != nil {
		t.Fatal(err)
	}
	if len(content) != 5 {
		t.Fatalf("Got %d items back, want 5", len(content))
	}

}

func BenchmarkResponseLargeCount(b *testing.B) {
	response := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		app.ServeHTTP(response, LargeContentRequest)
		if response.Code != 200 {
			b.Fatalf("Response code is %d, want %d", response.Code, 200)
		}
	}
}

func BenchmarkResponseCount(b *testing.B) {
	response := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		app.ServeHTTP(response, SimpleContentRequest)
		if response.Code != 200 {
			b.Fatalf("Response code is %d, want %d", response.Code, 200)
		}
	}
}

func BenchmarkResponseFallback(b *testing.B) {
	app := App{
		ContentClients: map[Provider]Client{
			Provider1: FailedContentProvider{Source: Provider1},
			Provider2: SampleContentProvider{Source: Provider2},
			Provider3: SampleContentProvider{Source: Provider3},
		},
		Config: DefaultConfig,
	}
	response := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		app.ServeHTTP(response, SimpleContentRequest)
		if response.Code != 200 {
			b.Fatalf("Response code is %d, want %d", response.Code, 200)
		}
	}
}

func TestResponseLargerCount(t *testing.T) {
	content, err := runRequest(app, LargeContentRequest, 200)
	if err != nil {
		t.Fatal(err)
	}

	if len(content) != 125 {
		t.Fatalf("Got %d items back, want 125", len(content))
	}

}

func TestResponseCountSomeFailed(t *testing.T) {
	app := App{
		ContentClients: map[Provider]Client{
			Provider1: SampleContentProvider{Source: Provider1},
			Provider2: FailedContentProvider{Source: Provider2},
			Provider3: FailedContentProvider{Source: Provider3},
		},
		Config: DefaultConfig,
	}
	content, err := runRequest(app, SimpleContentRequest, 200)
	if err != nil {
		t.Fatal(err)
	}

	if len(content) != 2 {
		t.Fatalf("Got %d items back, want 2", len(content))
	}

}

func TestWrongMethodFailed(t *testing.T) {
	_, err := runRequest(app, wrongMethodRequest, http.StatusMethodNotAllowed)
	if err != nil {
		t.Fatal(err)
	}
}

func TestResponseWrongOffsetFailed(t *testing.T) {
	_, err := runRequest(app, wrongOffsetRequest, http.StatusBadRequest)
	if err != nil {
		t.Fatal(err)
	}
}
func TestResponseMissingCountFailed(t *testing.T) {
	_, err := runRequest(app, missingCountRequest, http.StatusBadRequest)
	if err != nil {
		t.Fatal(err)
	}
}

func TestResponseOrder(t *testing.T) {
	content, err := runRequest(app, SimpleContentRequest, http.StatusOK)
	if err != nil {
		t.Fatal(err)
	}

	if len(content) != 5 {
		t.Fatalf("Got %d items back, want 5", len(content))
	}

	for i, item := range content {
		if Provider(item.Source) != DefaultConfig[i%len(DefaultConfig)].Type {
			t.Errorf(
				"Position %d: Got Provider %v instead of Provider %v",
				i, item.Source, DefaultConfig[i].Type,
			)
		}
	}
}

func TestOffsetResponseOrder(t *testing.T) {
	content, err := runRequest(app, OffsetContentRequest, http.StatusOK)
	if err != nil {
		t.Fatal(err)
	}

	if len(content) != 5 {
		t.Fatalf("Got %d items back, want 5", len(content))
	}

	for j, item := range content {
		i := j + 5
		if Provider(item.Source) != DefaultConfig[i%len(DefaultConfig)].Type {
			t.Errorf(
				"Position %d: Got Provider %v instead of Provider %v",
				i, item.Source, DefaultConfig[i].Type,
			)
		}
	}
}

func TestFallbackResponseOrder(t *testing.T) {
	app := App{
		ContentClients: map[Provider]Client{
			Provider1: FailedContentProvider{Source: Provider1},
			Provider2: SampleContentProvider{Source: Provider2},
			Provider3: SampleContentProvider{Source: Provider3},
		},
		Config: DefaultConfig,
	}
	content, err := runRequest(app, SimpleContentRequest, http.StatusOK)
	if err != nil {
		t.Fatal(err)
	}

	if len(content) != 4 {
		t.Fatalf("Got %d items back, want 4", len(content))
	}

	for j, item := range content {
		i := j
		c := DefaultConfig[i%len(DefaultConfig)]
		p := Provider(item.Source)
		if p != c.Type && p != *c.Fallback {
			t.Errorf(
				"Position %d: Got Provider %v instead of Provider %v",
				i, item.Source, *DefaultConfig[i].Fallback,
			)
		}
	}
}

func TestTotalFailResponse(t *testing.T) {
	app := App{
		ContentClients: map[Provider]Client{
			Provider1: FailedContentProvider{Source: Provider1},
			Provider2: FailedContentProvider{Source: Provider2},
			Provider3: FailedContentProvider{Source: Provider3},
		},
		Config: DefaultConfig,
	}
	content, err := runRequest(app, SimpleContentRequest, http.StatusOK)
	if err != nil {
		t.Fatal(err)
	}

	if len(content) != 0 {
		t.Fatalf("Got %d items back, want 0", len(content))
	}

	for j, item := range content {
		i := j + 5
		if Provider(item.Source) != DefaultConfig[i%len(DefaultConfig)].Type {
			t.Errorf(
				"Position %d: Got Provider %v instead of Provider %v",
				i, item.Source, DefaultConfig[i].Type,
			)
		}
	}
}

func TestNoConfigs(t *testing.T) {
	app := App{
		ContentClients: map[Provider]Client{
			Provider1: SampleContentProvider{Source: Provider1},
			Provider2: SampleContentProvider{Source: Provider2},
			Provider3: SampleContentProvider{Source: Provider3},
		},
		Config: []ContentConfig{},
	}
	content, err := runRequest(app, SimpleContentRequest, http.StatusOK)
	if err != nil {
		t.Fatal(err)
	}

	if len(content) != 0 {
		t.Fatalf("Got %d items back, want 0", len(content))
	}

	for j, item := range content {
		i := j
		if Provider(item.Source) != DefaultConfig[i%len(DefaultConfig)].Type {
			t.Errorf(
				"Position %d: Got Provider %v instead of Provider %v",
				i, item.Source, DefaultConfig[i].Type,
			)
		}
	}
}
