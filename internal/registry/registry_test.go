package registry

import (
	"testing"
	"time"
)

func TestRegister_Upsert_NoDuplicate(t *testing.T) {
	r := NewRegistry()

	req := RegisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-1",
		Address:     "localhost:8080",
	}

	// First register
	_ = r.Register(req)

	// Second register (same instance)
	_ = r.Register(req)

	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.services["auth"]) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(r.services["auth"]))
	}
}

func TestRegister_Upsert_UpdatesLastSeen(t *testing.T) {
	r := NewRegistry()

	req := RegisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-1",
		Address:     "localhost:8080",
	}

	_ = r.Register(req)

	r.mu.RLock()
	first := r.services["auth"]["auth-1"].LastSeen
	r.mu.RUnlock()

	time.Sleep(10 * time.Millisecond)

	_ = r.Register(req)

	r.mu.RLock()
	second := r.services["auth"]["auth-1"].LastSeen
	r.mu.RUnlock()

	if !second.After(first) {
		t.Fatalf("expected LastSeen to update")
	}
}

func TestDeregister_MultiInstance(t *testing.T) {
	r := NewRegistry()
	req1 := RegisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-1",
		Address:     "localhost:8080",
	}
	req2 := RegisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-2",
		Address:     "localhost:8081",
	}
	_ = r.Register(req1)
	_ = r.Register(req2)
	deregReq1 := DeregisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-1",
	}
	err := r.Deregister(deregReq1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, ok := r.services["auth"]["auth-1"]; ok {
		t.Fatalf("expected instance auth-1 to be deregistered")
	}
	if len(r.services["auth"]) != 1 {
		t.Fatalf("expected 1 instance remaining, got %d", len(r.services["auth"]))
	}
}

func TestDeregister_Idempotent(t *testing.T) {
	r := NewRegistry()
	req := RegisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-1",
		Address:     "localhost:8080",
	}
	_ = r.Register(req)
	deregReq := DeregisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-1",
	}
	// First deregister
	_ = r.Deregister(deregReq)

	// Second deregister (should be idempotent)
	err := r.Deregister(deregReq)
	if err != nil {
		t.Fatalf("expected no error on second deregistration, got %v", err)
	}
}

func TestDeregister_CleansUpService(t *testing.T) {
	r := NewRegistry()
	req := RegisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-1",
		Address:     "localhost:8080",
	}
	_ = r.Register(req)
	deregReq := DeregisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-1",
	}
	_ = r.Deregister(deregReq)
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, ok := r.services["auth"]; ok {
		t.Fatalf("expected service to be cleaned up")
	}
}

func TestHeartbeat(t *testing.T) {
	r := NewRegistry()
	req := RegisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-1",
		Address:     "localhost:8080",
	}
	_ = r.Register(req)

	r.mu.RLock()
	lastSeen := r.services["auth"]["auth-1"].LastSeen
	r.mu.RUnlock()

	time.Sleep(10 * time.Millisecond)

	heartbeatReq := HeartbeatRequest{
		ServiceName: "auth",
		InstanceID:  "auth-1",
	}
	err := r.ReceiveHeartbeat(heartbeatReq)
	if err != nil {
		t.Fatalf("expected no error on heartbeat, got %v", err)
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if !r.services["auth"]["auth-1"].LastSeen.After(lastSeen) {
		t.Fatalf("expected LastSeen to update on heartbeat")
	}
}

func TestGetInstances_SingleInstance(t *testing.T) {
	r := NewRegistry()
	req := RegisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-1",
		Address:     "localhost:8080",
	}
	_ = r.Register(req)
	getInstanceReq := GetInstanceRequest{
		ServiceName: "auth",
	}
	instances, err := r.GetInstances(getInstanceReq)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(instances))
	}
	if instances[0].ID != "auth-1" || instances[0].Address != "localhost:8080" {
		t.Fatalf("instance data mismatch")
	}
}

func TestGetInstances_MultiInstance(t *testing.T) {
	r := NewRegistry()
	req1 := RegisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-1",
		Address:     "localhost:8080",
	}
	req2 := RegisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-2",
		Address:     "localhost:8081",
	}
	_ = r.Register(req1)
	_ = r.Register(req2)

	getInstanceReq := GetInstanceRequest{
		ServiceName: "auth",
	}
	instances, err := r.GetInstances(getInstanceReq)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
}

func TestGetInstance_NoService(t *testing.T) {
	r := NewRegistry()

	getInstanceReq := GetInstanceRequest{
		ServiceName: "nonexistent",
	}

	_, err := r.GetInstances(getInstanceReq)
	if err == nil {
		t.Fatalf("expected error for nonexistent service")
	}
}

func TestGetInstance_EmptyService(t *testing.T) {
	r := NewRegistry()

	getInstanceReq := GetInstanceRequest{
		ServiceName: "",
	}

	instances, err := r.GetInstances(getInstanceReq)
	if err == nil {
		t.Fatalf("expected error for empty service name")
	}
	if len(instances) != 0 {
		t.Fatalf("expected 0 instances for empty service name, got %d", len(instances))
	}
}

func TestGetInstance_MultiService(t *testing.T) {
	r := NewRegistry()
	req1 := RegisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-1",
		Address:     "localhost:8080",
	}
	req2 := RegisterRequest{
		ServiceName: "payment",
		InstanceID:  "payment-1",
		Address:     "localhost:8081",
	}
	_ = r.Register(req1)
	_ = r.Register(req2)

	getInstanceReq := GetInstanceRequest{
		ServiceName: "auth",
	}
	instances, err := r.GetInstances(getInstanceReq)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(instances))
	}
	if instances[0].ID != "auth-1" || instances[0].Address != "localhost:8080" {
		t.Fatalf("instance data mismatch")
	}
}

func TestGetInstance_AfterDereg(t *testing.T) {
	r := NewRegistry()
	req1 := RegisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-1",
		Address:     "localhost:8080",
	}
	req2 := RegisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-2",
		Address:     "localhost:8081",
	}
	req3 := RegisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-3",
		Address:     "localhost:8082",
	}
	_ = r.Register(req1)
	_ = r.Register(req2)
	_ = r.Register(req3)

	deregReq := DeregisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-1",
	}
	_ = r.Deregister(deregReq)

	getInstanceReq := GetInstanceRequest{
		ServiceName: "auth",
	}
	instances, err := r.GetInstances(getInstanceReq)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
	for _, inst := range instances {
		if inst.ID == "auth-1" {
			t.Fatalf("expected auth-1 to be deregistered")
		}
	}

	deregReq2 := DeregisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-2",
	}
	_ = r.Deregister(deregReq2)

	deregReq3 := DeregisterRequest{
		ServiceName: "auth",
		InstanceID:  "auth-3",
	}
	_ = r.Deregister(deregReq3)

	instances, err = r.GetInstances(getInstanceReq)
	if err == nil {
		t.Fatalf("expected error for service with no instances")
	}

}
