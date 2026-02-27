package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	finopsv1 "github.com/migalsp/kubex-operator/api/v1"
)

func (s *Server) handleScalingGroups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	operatorNs := getOperatorNamespace()

	switch r.Method {
	case http.MethodGet:
		var list finopsv1.ScalingGroupList
		if err := s.Client.List(ctx, &list, client.InNamespace(operatorNs)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list.Items)

	case http.MethodPost:
		var group finopsv1.ScalingGroup
		if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		group.Namespace = operatorNs
		if err := s.Client.Create(ctx, &group); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(group)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleScalingGroupActions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	name := parts[4]
	operatorNs := getOperatorNamespace()

	group := &finopsv1.ScalingGroup{}
	if err := s.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: operatorNs}, group); err != nil {
		if errors.IsNotFound(err) {
			http.Error(w, "Group not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Sub-actions like /api/scaling/groups/{name}/manual
	if len(parts) > 5 && parts[5] == "manual" {
		s.handleScalingGroupManual(w, r, group)
		return
	}

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(group)

	case http.MethodPut:
		var updated finopsv1.ScalingGroup
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			current := &finopsv1.ScalingGroup{}
			if err := s.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: operatorNs}, current); err != nil {
				return err
			}
			current.Spec = updated.Spec
			return s.Client.Update(ctx, current)
		})
		
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(updated)

	case http.MethodDelete:
		if err := s.Client.Delete(ctx, group); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleScalingGroupManual(w http.ResponseWriter, r *http.Request, group *finopsv1.ScalingGroup) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Active *bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	group.Spec.Active = req.Active
	if err := s.Client.Update(r.Context(), group); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(group)
}

func (s *Server) handleScalingConfigs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	operatorNs := getOperatorNamespace()

	switch r.Method {
	case http.MethodGet:
		var list finopsv1.ScalingConfigList
		if err := s.Client.List(ctx, &list, client.InNamespace(operatorNs)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list.Items)

	case http.MethodPost:
		var config finopsv1.ScalingConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		config.Namespace = operatorNs
		if err := s.Client.Create(ctx, &config); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(config)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleScalingConfigActions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	name := parts[4]
	operatorNs := getOperatorNamespace()

	config := &finopsv1.ScalingConfig{}
	if err := s.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: operatorNs}, config); err != nil {
		if errors.IsNotFound(err) {
			http.Error(w, "Config not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if len(parts) > 5 && parts[5] == "manual" {
		s.handleScalingConfigManual(w, r, config)
		return
	}

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)

	case http.MethodPut:
		var updated finopsv1.ScalingConfig
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			current := &finopsv1.ScalingConfig{}
			if err := s.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: operatorNs}, current); err != nil {
				return err
			}
			current.Spec = updated.Spec
			return s.Client.Update(ctx, current)
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(updated)

	case http.MethodDelete:
		if err := s.Client.Delete(ctx, config); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleScalingConfigManual(w http.ResponseWriter, r *http.Request, config *finopsv1.ScalingConfig) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Active *bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	config.Spec.Active = req.Active
	if err := s.Client.Update(r.Context(), config); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(config)
}

func getOperatorNamespace() string {
	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		return "kubex"
	}
	return ns
}
