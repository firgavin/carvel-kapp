// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resourcesmisc

import (
	"fmt"

	appv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	pkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	ctlres "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func init() {
	pkgv1alpha1.AddToScheme(scheme.Scheme)
}

type PackagingCarvelDevV1alpha1PackageInstall struct {
	resource ctlres.Resource
}

func NewPackagingCarvelDevV1alpha1PackageInstall(resource ctlres.Resource) *PackagingCarvelDevV1alpha1PackageInstall {
	matcher := ctlres.APIVersionKindMatcher{
		APIVersion: pkgv1alpha1.SchemeGroupVersion.String(),
		Kind:       "PackageInstall",
	}
	if matcher.Matches(resource) {
		return &PackagingCarvelDevV1alpha1PackageInstall{resource}
	}
	return nil
}

func (s PackagingCarvelDevV1alpha1PackageInstall) IsDoneApplying() DoneApplyState {
	pkgInstall := pkgv1alpha1.PackageInstall{}

	err := s.resource.AsTypedObj(&pkgInstall)
	if err != nil {
		return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf(
			"Error: Failed obj conversion: %s", err)}
	}

	if pkgInstall.Generation != pkgInstall.Status.ObservedGeneration {
		return DoneApplyState{Done: false, Message: fmt.Sprintf(
			"Waiting for generation %d to be observed", pkgInstall.Generation)}
	}

	for _, cond := range pkgInstall.Status.Conditions {
		switch {
		case cond.Type == appv1alpha1.Reconciling && cond.Status == corev1.ConditionTrue:
			return DoneApplyState{Done: false, Message: "Reconciling"}

		case cond.Type == appv1alpha1.ReconcileFailed && cond.Status == corev1.ConditionTrue:
			return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf(
				"Reconcile failed: %s (message: %s)", cond.Reason, cond.Message)}

		case cond.Type == appv1alpha1.DeleteFailed && cond.Status == corev1.ConditionTrue:
			return DoneApplyState{Done: true, Successful: false, Message: fmt.Sprintf(
				"Delete failed: %s (message: %s)", cond.Reason, cond.Message)}
		}
	}

	deletingRes := NewDeleting(s.resource)
	if deletingRes != nil {
		return deletingRes.IsDoneApplying()
	}

	return DoneApplyState{Done: true, Successful: true}
}
