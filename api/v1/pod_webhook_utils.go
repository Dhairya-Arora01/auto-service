package v1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func checkIfPortsPresent(pod *corev1.Pod) bool {

	podPorts := []corev1.ServicePort{}

	for _, port := range pod.Spec.Containers[0].Ports {

		podPorts = append(podPorts, corev1.ServicePort{
			Port:       port.ContainerPort,
			TargetPort: intstr.IntOrString{IntVal: port.ContainerPort},
			Protocol:   port.Protocol,
		})
	}

	return len(podPorts) > 0
}
