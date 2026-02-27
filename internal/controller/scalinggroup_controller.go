/*
Copyright 2026 migalsp.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	finopsv1 "github.com/migalsp/kubex-operator/api/v1"
	"github.com/migalsp/kubex-operator/internal/scaling"
)

// ScalingGroupReconciler reconciles a ScalingGroup object
type ScalingGroupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Engine *scaling.Engine
}

// +kubebuilder:rbac:groups=finops.kubex.io,resources=scalinggroups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=finops.kubex.io,resources=scalinggroups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=finops.kubex.io,resources=scalinggroups/finalizers,verbs=update
// +kubebuilder:rbac:groups=finops.kubex.io,resources=scalingpolicies,verbs=get;list;watch

func (r *ScalingGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := logf.FromContext(ctx)

	// 1. Fetch the ScalingGroup
	group := &finopsv1.ScalingGroup{}
	if err := r.Get(ctx, req.NamespacedName, group); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 2. Determine desired state
	targetActive := r.Engine.IsActive(group.Spec.Schedules, group.Spec.Active)
	l.Info("Reconciling ScalingGroup", "category", group.Spec.Category, "namespaces", group.Spec.Namespaces, "targetActive", targetActive)

	// Initialize status maps if nil
	if group.Status.OriginalReplicas == nil {
		group.Status.OriginalReplicas = make(map[string]int32)
	}

	// 3. Define stages from group.Spec.Sequence
	// Default: all namespaces in one stage if no sequence defined
	managedNamespaces := group.Spec.Namespaces
	var stages [][]string

	if len(group.Spec.Sequence) > 0 {
		for _, s := range group.Spec.Sequence {
			nsInStage := strings.Fields(s)
			stages = append(stages, nsInStage)
		}
		// Add namespaces not mentioned in sequence as the last stage
		var missing []string
		for _, ns := range managedNamespaces {
			found := false
			for _, stage := range stages {
				for _, sn := range stage {
					if sn == ns {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				missing = append(missing, ns)
			}
		}
		if len(missing) > 0 {
			stages = append(stages, missing)
		}
	} else {
		stages = append(stages, managedNamespaces)
	}

	// Reverse stages for Scaling Up if needed?
	// Usually sequence is defined for "Shutdown" order.
	// User said: "first 1, then 3,4,5, then 2".
	// This usually means Up order. Let's assume sequence defines UP order, and reverse for DOWN.
	if !targetActive {
		for i, j := 0, len(stages)-1; i < j; i, j = i+1, j-1 {
			stages[i], stages[j] = stages[j], stages[i]
		}
	}

	allReady := true
	managedCount := 0

	// 4. Iterate over stages
	for i, stage := range stages {
		l.Info("Processing scaling stage", "stageIndex", i, "namespaces", stage)

		stageReady := true
		for _, ns := range stage {
			managedCount++

			// a. Fetch individual ScalingConfig for exclusions and sequence inheritance
			var exclusions []string
			var nsSequence []string

			// Try to find a ScalingConfig that manages this target namespace
			configList := &finopsv1.ScalingConfigList{}
			if err := r.List(ctx, configList, client.InNamespace(group.Namespace)); err == nil {
				for _, cfg := range configList.Items {
					if cfg.Spec.TargetNamespace == ns {
						exclusions = cfg.Spec.Exclusions
						nsSequence = cfg.Spec.Sequence
						l.Info("Found ScalingConfig for inheritance", "namespace", ns, "config", cfg.Name)
						break
					}
				}
			}

			// b. Scale Target
			nsKeyPrefix := ns + "/"
			nsReplicas := make(map[string]int32)
			for k, v := range group.Status.OriginalReplicas {
				if strings.HasPrefix(k, nsKeyPrefix) {
					nsReplicas[strings.TrimPrefix(k, nsKeyPrefix)] = v
					delete(group.Status.OriginalReplicas, k)
				}
			}

			updatedOriginals, nsReady, err := r.Engine.ScaleTarget(ctx, ns, targetActive, nsSequence, exclusions, nsReplicas)
			if err != nil {
				l.Error(err, "failed to scale namespace", "namespace", ns)
				allReady = false
				stageReady = false
				continue
			}

			if !nsReady {
				stageReady = false
				allReady = false
			}

			// Merge back
			for k, v := range updatedOriginals {
				group.Status.OriginalReplicas[nsKeyPrefix+k] = v
			}

			// c. Check if namespace reached target phase
			phase := r.Engine.ComputePhase(ctx, ns, targetActive)
			if (targetActive && phase != "ScaledUp") || (!targetActive && phase != "ScaledDown") {
				stageReady = false
				allReady = false
			}
		}

		if !stageReady {
			l.Info("Stage not ready, waiting before next stage", "stageIndex", i)
			break // Stop at this stage, wait for next reconcile
		}
	}

	// 5. Update Status
	group.Status.ManagedCount = managedCount
	group.Status.LastAction = metav1.Now()

	if allReady {
		if targetActive {
			group.Status.Phase = "ScaledUp"
		} else {
			group.Status.Phase = "ScaledDown"
		}
	} else {
		if targetActive {
			group.Status.Phase = "ScalingUp"
		} else {
			group.Status.Phase = "ScalingDown"
		}
	}

	if err := r.Status().Update(ctx, group); err != nil {
		return ctrl.Result{}, err
	}

	// Requeue faster if scaling is in progress
	if !allReady {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalingGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Engine == nil {
		r.Engine = &scaling.Engine{Client: r.Client}
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&finopsv1.ScalingGroup{}).
		Named("scalinggroup").
		Complete(r)
}
