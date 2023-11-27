package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *PodReconciler) createService(ctx context.Context, pod *corev1.Pod) error {

	podPorts := getPodPorts(pod)
	if len(podPorts) == 0 {
		return nil
	}

	service := &corev1.Service{

		ObjectMeta: metav1.ObjectMeta{

			Namespace: pod.Namespace,
			Name:      pod.Name + "auto-service",
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
