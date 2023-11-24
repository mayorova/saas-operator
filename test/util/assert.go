package util

import (
	"context"
	"fmt"
	"time"

	marin3rv1alpha1 "github.com/3scale-ops/marin3r/apis/marin3r/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExpectedWorkload struct {
	Namespace      string
	Name           string
	Replicas       int32
	ContainerName  string
	ContainerImage string
	ContainterCmd  []string
	ContainterArgs []string
	HPA            bool
	PDB            bool
	PodMonitor     bool
	EnvoyConfig    bool
	LastVersion    string
}

// checkWorkloadResources checks if all the k8s resource generated by a workload
// exists and matches the expectedWorkload specification
func (ew *ExpectedWorkload) Assert(c client.Client, dep *appsv1.Deployment, timeout, poll time.Duration) func() {
	return func() {

		By(fmt.Sprintf("%s workload Deployment", ew.Name),
			(&ExpectedResource{
				Name:        ew.Name,
				Namespace:   ew.Namespace,
				LastVersion: ew.LastVersion,
			}).Assert(c, dep, timeout, poll),
		)

		if ew.ContainerName != "" {
			Expect(dep.Spec.Template.Spec.Containers[0].Name).To(Equal(ew.ContainerName))
		}

		if ew.ContainerImage != "" {
			Expect(dep.Spec.Template.Spec.Containers[0].Image).To(Equal(ew.ContainerImage))
		}

		if ew.ContainterCmd != nil {
			Expect(dep.Spec.Template.Spec.Containers[0].Command).To(Equal(ew.ContainterCmd))
		}

		if ew.ContainterArgs != nil {
			Expect(dep.Spec.Template.Spec.Containers[0].Args).To(Equal(ew.ContainterArgs))
		}

		hpa := &autoscalingv2.HorizontalPodAutoscaler{}
		By(fmt.Sprintf("%s workload HPA", ew.Name),
			(&ExpectedResource{
				Name:      ew.Name,
				Namespace: ew.Namespace, Missing: !ew.HPA,
			}).Assert(c, hpa, timeout, poll),
		)
		if ew.HPA {
			Expect(hpa.Spec.ScaleTargetRef.Kind).Should(Equal("Deployment"))
			Expect(hpa.Spec.ScaleTargetRef.Name).Should(Equal(ew.Name))
			Expect(hpa.Spec.MinReplicas).Should(Equal(pointer.Int32(ew.Replicas)))
		} else {
			Expect(dep.Spec.Replicas).To(Equal(pointer.Int32(ew.Replicas)))
		}

		pdb := &policyv1.PodDisruptionBudget{}
		By(fmt.Sprintf("%s workload PDB", ew.Name),
			(&ExpectedResource{
				Name:      ew.Name,
				Namespace: ew.Namespace, Missing: !ew.PDB,
			}).Assert(c, pdb, timeout, poll),
		)
		if ew.PDB {
			Expect(pdb.Spec.Selector.MatchLabels["deployment"]).Should(Equal(ew.Name))
		}

		pm := &monitoringv1.PodMonitor{}
		By(fmt.Sprintf("%s workload PodMonitor", ew.Name),
			(&ExpectedResource{
				Name:      ew.Name,
				Namespace: ew.Namespace, Missing: !ew.PodMonitor,
			}).Assert(c, pm, timeout, poll),
		)
		if ew.PodMonitor {
			Expect(pm.Spec.Selector.MatchLabels["deployment"]).Should(Equal(ew.Name))
		}

		ec := &marin3rv1alpha1.EnvoyConfig{}
		By(fmt.Sprintf("%s workload EnvoyConfig", ew.Name),
			(&ExpectedResource{
				Name:      ew.Name,
				Namespace: ew.Namespace, Missing: !ew.EnvoyConfig,
			}).Assert(c, ec, timeout, poll),
		)
		if ew.EnvoyConfig {
			Expect(ec.Spec.NodeID).Should(Equal(ew.Name))
		}

	}
}

// getResourceVersion fetches the current resource version for an object,
// returns an empty string if the object doesn't exists
func GetResourceVersion(c client.Client, o client.Object, name, namespace string, timeout, poll time.Duration) string {
	By(fmt.Sprintf("%s feching resource version", name),
		(&ExpectedResource{
			Name:      name,
			Namespace: namespace,
		}).Assert(c, o, timeout, poll),
	)
	return o.GetResourceVersion()
}

type ExpectedResource struct {
	Namespace   string
	Name        string
	Missing     bool
	LastVersion string
}

// checkResource checks if a k8s resource exists, has been updated or is missing)
func (er *ExpectedResource) Assert(c client.Client, o client.Object, timeout, poll time.Duration) func() {

	if er.Missing {
		return func() {
			By(fmt.Sprintf("%s object does NOT exist", er.Name))
			Eventually(func() error {
				return c.Get(context.Background(),
					types.NamespacedName{Name: er.Name, Namespace: er.Namespace}, o,
				)
			}, timeout, poll).Should(HaveOccurred())
		}
	}

	if er.LastVersion != "" {
		return func() {
			By(fmt.Sprintf("%s object has been updated", er.Name))
			Eventually(func() bool {
				if err := c.Get(context.Background(),
					types.NamespacedName{Name: er.Name, Namespace: er.Namespace}, o,
				); err != nil {
					return false
				}
				return o.GetResourceVersion() != er.LastVersion
			}, timeout, poll).Should(BeTrue())
		}
	}

	return func() {
		By(fmt.Sprintf("%s object does exist", er.Name))
		Eventually(func() error {
			return c.Get(context.Background(),
				types.NamespacedName{Name: er.Name, Namespace: er.Namespace}, o,
			)
		}, timeout, poll).ShouldNot(HaveOccurred())
	}
}
