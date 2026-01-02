package kubernetes

import (
	"context"
	"deploy-platform/internal/models"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// CreateOrUpdateDeployment creates or updates a Kubernetes deployment (Vercel-style: updates existing)
func (c *Client) CreateOrUpdateDeployment(ctx context.Context, deployment *models.Deployment, hostname string, envVars map[string]string) error {
	return c.CreateDeployment(ctx, deployment, hostname, envVars)
}

func (c *Client) CreateDeployment(ctx context.Context, deployment *models.Deployment, hostname string, envVars map[string]string) error {
	namespace := "default" // Or create per-project namespace
	// Use project-based name (Vercel-style: one deployment per project that updates)
	deploymentName := fmt.Sprintf("project-%d", deployment.ProjectID)

	// Create Deployment
	k8sDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: deployment.ImageTag,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							Env: convertEnvVars(envVars),
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := c.clientset.AppsV1().Deployments(namespace).Create(ctx, k8sDeployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// Create Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": deploymentName,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}

	// Try to create service, if exists, update it
	_, err = c.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			_, updateErr := c.clientset.CoreV1().Services(namespace).Update(ctx, service, metav1.UpdateOptions{})
			if updateErr != nil {
				return fmt.Errorf("failed to update service: %v", updateErr)
			}
		} else {
			return fmt.Errorf("failed to create service: %v", err)
		}
	}

	// Create Ingress
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: hostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: func() *networkingv1.PathType { p := networkingv1.PathTypePrefix; return &p }(),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: deploymentName,
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Try to create ingress, if exists, update it
	_, err = c.clientset.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			_, updateErr := c.clientset.NetworkingV1().Ingresses(namespace).Update(ctx, ingress, metav1.UpdateOptions{})
			if updateErr != nil {
				return fmt.Errorf("failed to update ingress: %v", updateErr)
			}
		} else {
			return fmt.Errorf("failed to create ingress: %v", err)
		}
	}
	return nil
}

func convertEnvVars(envVars map[string]string) []corev1.EnvVar {
	var env []corev1.EnvVar
	for k, v := range envVars {
		env = append(env, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return env
}

func int32Ptr(i int32) *int32 { return &i }
