package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	finopsv1 "github.com/migalsp/kubex-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func buildMockServerWithK8s() *Server {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(finopsv1.AddToScheme(scheme))

	client := fakeclient.NewClientBuilder().WithScheme(scheme).Build()
	k8sClient := fake.NewSimpleClientset()

	k8sClient.Discovery().(*fakediscovery.FakeDiscovery).FakedServerVersion = &version.Info{
		GitVersion: "v1.35.0",
		Platform:   "linux/amd64",
	}

	return &Server{
		Client:    client,
		K8sClient: k8sClient,
	}
}

func TestHandleOperatorHealth(t *testing.T) {
	os.Setenv("HOSTNAME", "kubex-operator-1234")
	os.Setenv("POD_NAMESPACE", "kubex")
	defer os.Unsetenv("HOSTNAME")
	defer os.Unsetenv("POD_NAMESPACE")

	server := buildMockServerWithK8s()

	req, err := http.NewRequest("GET", "/api/operator/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleOperatorHealth)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	current, ok := response["current"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'current' object in response")
	}

	if current["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %v", current["status"])
	}
}

func TestHandleClusterInfo(t *testing.T) {
	server := buildMockServerWithK8s()

	req, err := http.NewRequest("GET", "/api/cluster-info", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleClusterInfo)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var info map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&info); err != nil {
		t.Fatal(err)
	}

	if info["version"] != "v1.35.0" {
		t.Errorf("expected version 'v1.35.0', got %v", info["version"])
	}
}

func TestHandleNamespaces(t *testing.T) {
	server := buildMockServerWithK8s()

	ns := &finopsv1.NamespaceFinOps{
		ObjectMeta: metav1.ObjectMeta{Name: "test-ns", Namespace: "kubex"},
	}
	server.Client.Create(context.Background(), ns)

	req, err := http.NewRequest("GET", "/api/namespaces", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleNamespaces)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var parsed []finopsv1.NamespaceFinOps
	if err := json.NewDecoder(rr.Body).Decode(&parsed); err != nil {
		t.Fatal(err)
	}

	if len(parsed) != 1 || parsed[0].Name != "test-ns" {
		t.Errorf("expected 1 namespace, got %v", parsed)
	}
}
