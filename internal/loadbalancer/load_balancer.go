package loadbalancer

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"service-mesh-sim/internal/registry"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type LoadBalancer struct {
	registryURL     string // Final, safe for concurrent reads
	getInstancesURL string // Final, safe for concurrent reads
	httpClient      *http.Client
	algorithm       BalancingAlgorithm
	rrCounter       map[string]*atomic.Uint64 // Round-robin counters per service
	rrCounterMu     sync.Mutex                // Mutex to protect access to rrCounter map
}

func NewLoadBalancer(registryURL string, algorithm BalancingAlgorithm) *LoadBalancer {
	lb := &LoadBalancer{
		registryURL:     registryURL,
		getInstancesURL: registryURL + "/get-instances",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{ // Connection pooling settings, tune as needed
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		algorithm: algorithm,
	}
	if algorithm == RoundRobin {
		lb.rrCounter = make(map[string]*atomic.Uint64)
	}
	return lb
}

// Orchestrates the load balancing process for incoming requests
func (lb *LoadBalancer) HandleRequest(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	// Extract service name from request path, e.g., /serviceA/endpoint -> serviceA
	serviceName, remainingPath, err := parseServicePath(req.URL.Path)
	if err != nil {
		http.Error(w, "Invalid request path", http.StatusBadRequest)
		return
	}

	// Fetch instances for the requested service from the registry
	instances, err := lb.getInstances(serviceName)
	if err != nil {
		http.Error(w, "Failed to fetch service instances", http.StatusInternalServerError)
		return
	}

	// Select an instance based on the configured load balancing algorithm
	instance, err := lb.selectInstance(instances)
	if err != nil {
		http.Error(w, "Failed to select service instance", http.StatusInternalServerError)
		return
	}

	// Forward the incoming request to the selected instance and relay the response back to the client
	lb.forwardRequest(w, req, instance, remainingPath)

}

func parseServicePath(path string) (string, string, error) {
	// Simple extraction logic: assumes path format is /serviceName/remainingPath
	// parts[0] is empty due to leading slash
	// parts[1] is the service name
	// parts[2] is the remaining path to forward to the service instance
	parts := strings.SplitN(path, "/", 3)

	// Validate that we have at least a service name for routing
	if len(parts) < 2 || parts[1] == "" {
		return "", "", ErrMissingServiceName
	}

	// Construct the path to forward to the service instance
	instancePath := "/"
	if len(parts) == 3 {
		instancePath += parts[2]
	}

	return parts[1], instancePath, nil
}

// Fetches list of instances for the requested service from the registry
func (lb *LoadBalancer) getInstances(serviceName string) ([]registry.InstanceSnapshot, error) {

	// 1. Construct request body
	reqBody := registry.GetInstancesRequest{
		ServiceName: serviceName,
	}
	bodyBytes, err := json.Marshal(reqBody) // Encode the request body as JSON
	if err != nil {
		return nil, err
	}

	// 2. Construct http request to registry's /get-instances endpoint
	req, err := http.NewRequest(http.MethodPost, lb.getInstancesURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// 3. Send request to registry
	resp, err := lb.httpClient.Do(req)
	if err != nil {
		return nil, err

	}
	defer resp.Body.Close()

	// 4. Read and decode response into GetInstancesResponse struct
	if resp.StatusCode != http.StatusOK { // Handle non-200 responses from registry
		return nil, ErrRegistryError
	}
	var getInstancesResp registry.GetInstancesResponse
	dec := json.NewDecoder(resp.Body)
	dec.DisallowUnknownFields() // Strict decoding to prevent unexpected fields
	if err := dec.Decode(&getInstancesResp); err != nil {
		return nil, err
	}

	return getInstancesResp.Instances, nil

}

// Selects an instance based on the configured load balancing algorithm
func (lb *LoadBalancer) selectInstance(instances []registry.InstanceSnapshot) (registry.InstanceSnapshot, error)

// Forwards the incoming request to the selected instance and relays the response back to the client
func (lb *LoadBalancer) forwardRequest(w http.ResponseWriter, req *http.Request, instance registry.InstanceSnapshot, remainingPath string) {
	// Construct the URL to forward the request to, e.g., http://instanceURL/remainingPath
	targetURL := instance.URL + remainingPath

	// Create a new request to forward to the service instance
	forwardReq, err := http.NewRequest(req.Method, targetURL, req.Body)
	if err != nil {
		http.Error(w, "Failed to create request to service instance", http.StatusInternalServerError)
		return
	}

	// Copy headers from the original request to the forwarded request
	forwardReq.Header = req.Header.Clone()

	// Forward to client and wait for response
	resp, err := lb.httpClient.Do(forwardReq)
	if err != nil {
		http.Error(w, "Failed to forward request to service instance", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	// Relay response back to client
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy the response body to the client
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		// Log the error but we can't do much at this point since we've already sent the headers
		// In a real implementation, consider using a logger to log this error
		return
	}

}
