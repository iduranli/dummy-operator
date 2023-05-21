/*
Copyright 2023.

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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	interviewcomv1alpha1 "github.com/iduranli/dummy-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DummyReconciler reconciles a Dummy object
type DummyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=interview.com,resources=dummies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=interview.com,resources=dummies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=interview.com,resources=dummies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Dummy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *DummyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Initialize logging
	// log := ctrllog.FromContext(ctx)
	// log := ctrllog.Log.WithValues("DummyApp", req.NamespacedName)
	log := ctrllog.Log.WithValues()

	log.Info("Reconcile for Dummy Controller")

	// get the operator object
	dummy := &interviewcomv1alpha1.Dummy{}
	err := r.Get(ctx, req.NamespacedName, dummy)

	if err != nil {
		if errors.IsNotFound(err) {
			// STEP 4
			// Dummy object not found, could have been deleted.
			log.Info("Not found! Dummy Object must be deleted")

			// Find and delete the associated Pod when Dummy is deleted
			pod := r.getPodForDummy(dummy)
			podFound := &corev1.Pod{}
			err = r.Get(ctx, client.ObjectKey{Namespace: pod.ObjectMeta.Namespace, Name: pod.ObjectMeta.Name}, podFound)
			if err != nil {
				// This pod is already deleted
				log.Info("Failed to get Pod for deletion, must be already deleted")
				return ctrl.Result{}, nil
			}
			err = r.Delete(ctx, pod)
			if err != nil {
				log.Error(err, "Failed to delete Pod")
				return ctrl.Result{}, nil
			}
			log.Info("Deleted the associated Pod to Dummy")
			return ctrl.Result{}, nil
		}
		// Error reading the Dummy object.
		log.Error(err, "Failed to get Dummy instance")
		return ctrl.Result{}, err
	} else {
		// STEP 2
		// Log object name, object namespace and spec.message value
		log.Info("Retrieved values", "object name", dummy.Name, "object namespace", dummy.Namespace, "value of spec.message", dummy.Spec.Message)

		// STEP 3
		// Echo spec.message into status.specEcho
		if dummy.Status.SpecEcho != dummy.Spec.Message {
			dummy.Status.SpecEcho = dummy.Spec.Message
			err := r.Status().Update(ctx, dummy)
			if err != nil {
				log.Error(err, "Failed to update SpecEcho field")
				return ctrl.Result{}, err
			}
			log.Info("Updated the SpecEcho field as: " + dummy.Spec.Message)
		} else {
			log.Info("No change in SpecEcho status... " + dummy.Spec.Message)
		}
	}

	// STEP 4
	// Associate a Pod to each Dummy API object
	pod := r.getPodForDummy(dummy)
	podFound := &corev1.Pod{}
	err = r.Get(ctx, client.ObjectKey{Namespace: pod.ObjectMeta.Namespace, Name: pod.ObjectMeta.Name}, podFound)
	if err != nil {
		if errors.IsNotFound(err) {
			// The pod did not exist, create a new pod
			err = r.Create(ctx, pod)
			if err != nil {
				log.Error(err, "Failed to create Pod", "Namespace", pod.Namespace, "Name", pod.Name)
				return ctrl.Result{}, err
			}
			// pod successfully created
			log.Info("Created a new Pod", "Namespace", pod.Namespace, "Name", pod.Name)
		} else if err != nil {
			log.Error(err, "Failed to get Pod")
			return ctrl.Result{}, err
		}
	}

	// STEP 4
	// Keep track of the Pod status in the  PodStatus field
	if dummy.Status.PodStatus != string(podFound.Status.Phase) {
		dummy.Status.PodStatus = string(podFound.Status.Phase)
		err = r.Status().Update(ctx, dummy)
		if err != nil {
			log.Error(err, "Failed to update PodStatus field")
			return ctrl.Result{}, err
		}
		// status successfully updated
		log.Info("Updated the PodStatus field as: " + string(podFound.Status.Phase) + " for Pod: " + podFound.Name)
	} else {
		log.Info("No change in PodStatus... " + string(podFound.Status.Phase))
	}

	return ctrl.Result{}, nil
}

// Define and return a new Pod object
func (r *DummyReconciler) getPodForDummy(m *interviewcomv1alpha1.Dummy) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dummy-nginx",
			Namespace: "default",
			Labels: map[string]string{
				"app": "dummy-nginx",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "dummy-nginx",
					Image: "nginx:latest",
					Env: []corev1.EnvVar{
						{
							Name:  "CLIENT_ID",
							Value: "DUMMY-POD",
						},
					},
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 8080,
							Protocol:      corev1.ProtocolTCP,
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
	// Ensure that this pod is observed for state changes
	ctrl.SetControllerReference(m, pod, r.Scheme)
	return pod
}

// SetupWithManager sets up the controller with the Manager.
func (r *DummyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&interviewcomv1alpha1.Dummy{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
