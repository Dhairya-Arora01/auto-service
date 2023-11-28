package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// createService creates a service for the pod
func (r *PodReconciler) createService(ctx context.Context, pod *corev1.Pod) error {

	podPorts := getPodPorts(pod)
	if len(podPorts) == 0 {
		return nil
	}

	service := &corev1.Service{

		ObjectMeta: metav1.ObjectMeta{

			Namespace: pod.Namespace,
			Name:      pod.Name + "-auto-service",
			Labels: map[string]string{
				"auto-service": "true",
			},
		},
		Spec: corev1.ServiceSpec{

			Ports:    getPodPorts(pod),
			Selector: pod.Labels,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	if err := r.Create(ctx, service); err != nil {
		return err
	}

	return nil
}

// getPodPorts returns the ports assoicated with the given Pod's containers.
func getPodPorts(pod *corev1.Pod) []corev1.ServicePort {

	podPorts := []corev1.ServicePort{}

	for _, port := range pod.Spec.Containers[0].Ports {

		podPorts = append(podPorts, corev1.ServicePort{
			Port:       port.ContainerPort,
			TargetPort: intstr.IntOrString{IntVal: port.ContainerPort},
			Protocol:   port.Protocol,
		})
	}

	return podPorts

}

// isPodReady returns true if all the pod's containers are ready.
func isPodReady(pod *corev1.Pod) bool {

	for _, status := range pod.Status.ContainerStatuses {
		if !status.Ready {
			return false
		}
	}

	return true
}

func (r *PodReconciler) cleanAssociatedServices(ctx context.Context, podNamespacedName types.NamespacedName) error {

	serviceName := podNamespacedName.Name + "-auto-service"

	serviceNamespacedName := types.NamespacedName{
		Namespace: podNamespacedName.Namespace,
		Name:      serviceName,
	}

	service := &corev1.Service{}
	if err := r.Client.Get(ctx, serviceNamespacedName, service); err != nil {
		return client.IgnoreNotFound(err)
	}

	if err := r.Client.Delete(ctx, service); err != nil {
		return err
	}

	return nil
}
