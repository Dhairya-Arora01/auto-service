/*
Copyright 2023 Dhairya-Arora01.

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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	pod := &corev1.Pod{}

	if err := r.Client.Get(ctx, req.NamespacedName, pod); err != nil {

		if errors.IsNotFound(err) {
			l.Info("Pod no longer exists")

			if err := r.cleanAssociatedServices(ctx, req.NamespacedName); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil

		}

		return ctrl.Result{}, err
	}

	if pod.Status.Phase == corev1.PodRunning && pod.Labels["auto-service"] == "true" && isPodReady(pod) {

		if pod.Labels["service-active"] == "true" {
			return ctrl.Result{}, nil

		} else if pod.Labels["service-active"] == "false" {

			if err := r.createService(ctx, pod); err != nil {
				return ctrl.Result{}, client.IgnoreAlreadyExists(err)
			}

			pod.Labels["service-active"] = "true"
			if err := r.Update(ctx, pod); err != nil {
				return ctrl.Result{}, err
			}

		} else {

			pod.Labels["service-active"] = "false"

			if err := r.Update(ctx, pod); err != nil {
				return ctrl.Result{}, err
			}
		}

	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}
