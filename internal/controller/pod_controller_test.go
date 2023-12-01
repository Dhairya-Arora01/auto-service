package controller

import (
	"context"
	"reflect"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// unit test

func TestGetPodPorts(t *testing.T) {

	tests := []struct {
		name  string
		pod   *corev1.Pod
		ports []corev1.ServicePort
	}{
		{
			name: "Single-Port",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "Single",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
			ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.IntOrString{IntVal: 80},
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
		{
			name: "Multi-Port",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "Multi",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									ContainerPort: 443,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
			ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.IntOrString{IntVal: 80},
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Port:       443,
					TargetPort: intstr.IntOrString{IntVal: 443},
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
		{
			name: "No-Port",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "No",
							Ports: []corev1.ContainerPort{},
						},
					},
				},
			},
			ports: []corev1.ServicePort{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPodPorts(tt.pod); !reflect.DeepEqual(got, tt.ports) {
				t.Errorf("Test %s failed", tt.name)
			}
		})
	}
}

// Integration testing in BDD style using ginkgo

var _ = Describe("Pod Controller", func() {

	Context("When creating a pod with auto-service label", func() {
		It("should create a service in the same namespace", func() {
			By("Create demo pod with labels")
			ctx := context.Background()
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "demo-pod",
					Namespace: "default",
					Labels: map[string]string{
						"auto-service": "true",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "demo-container",
							Image: "nginx:alpine",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, pod)).Should(Succeed())

			By("An associated service should appear")
			service := &corev1.Service{}
			serviceNamespacedName := types.NamespacedName{
				Namespace: pod.Namespace,
				Name:      pod.Name + "-auto-service",
			}

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, serviceNamespacedName, service); err != nil {
					return false
				}
				return true
			}, time.Second*14, time.Second*2).Should(BeTrue())

			By("Deleting the pod")
			Expect(k8sClient.Delete(ctx, pod)).Should(Succeed())

			By("Associated service should be deleted")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, serviceNamespacedName, service); err != nil {
					return true
				}
				return false
			}, time.Second*12, time.Second*4).Should(BeTrue())

		})

	})

})
