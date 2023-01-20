package resources

import (
	"context"
	"fmt"

	marin3r "github.com/3scale-ops/marin3r/apis/marin3r/v1alpha1"
	basereconciler "github.com/3scale/saas-operator/pkg/reconcilers/basereconciler/v2"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var _ basereconciler.Resource = EnvoyConfigTemplate{}

// EnvoyConfigTemplate has methods to generate and reconcile a EnvoyConfig
type EnvoyConfigTemplate struct {
	Template  func() (*marin3r.EnvoyConfig, error)
	IsEnabled bool
}

// Build returns a EnvoyConfig resource
func (ect EnvoyConfigTemplate) Build(ctx context.Context, cl client.Client) (client.Object, error) {

	ec, err := ect.Template()
	if err != nil {
		return nil, err
	}
	return ec.DeepCopy(), nil
}

// Enabled indicates if the resource should be present or not
func (ect EnvoyConfigTemplate) Enabled() bool {
	return ect.IsEnabled
}

// ResourceReconciler implements a generic reconciler for EnvoyConfig resources
func (ect EnvoyConfigTemplate) ResourceReconciler(ctx context.Context, cl client.Client, obj client.Object) error {
	logger := log.FromContext(ctx, "kind", "EnvoyConfig", "resource", obj.GetName())

	needsUpdate := false
	desired := obj.(*marin3r.EnvoyConfig)

	instance := &marin3r.EnvoyConfig{}
	err := cl.Get(ctx, types.NamespacedName{Name: desired.GetName(), Namespace: desired.GetNamespace()}, instance)
	if err != nil {
		if errors.IsNotFound(err) {

			if ect.Enabled() {
				err = cl.Create(ctx, desired)
				if err != nil {
					return fmt.Errorf("unable to create object: " + err.Error())
				}
				logger.Info("resource created")
				return nil

			} else {
				return nil
			}
		}

		return err
	}

	/* Delete and return if not enabled */
	if !ect.Enabled() {
		err := cl.Delete(ctx, instance)
		if err != nil {
			return fmt.Errorf("unable to delete object: " + err.Error())
		}
		logger.Info("resource deleted")
		return nil
	}

	/* Reconcile metadata */
	if !equality.Semantic.DeepEqual(instance.GetAnnotations(), desired.GetAnnotations()) {
		instance.ObjectMeta.Annotations = desired.GetAnnotations()
		needsUpdate = true
	}
	if !equality.Semantic.DeepEqual(instance.GetLabels(), desired.GetLabels()) {
		instance.ObjectMeta.Labels = desired.GetLabels()
		needsUpdate = true
	}

	/* Reconcile spec */
	if !equality.Semantic.DeepEqual(instance.Spec, desired.Spec) {
		instance.Spec = desired.Spec
		needsUpdate = true
	}

	if needsUpdate {
		err := cl.Update(ctx, instance)
		if err != nil {
			return err
		}
		logger.Info("resource updated")
	}

	return nil
}