package controller

import (
    "context"
    "fmt"

    monitorv1alpha1 "github.com/anurag-2911/resourceMonitor/api/v1alpha1"
    "github.com/go-logr/logr"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    
    "sigs.k8s.io/controller-runtime/pkg/log"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ResourceMonitorReconciler struct {
    client.Client
    Log    logr.Logger
    Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=monitor.example.com,resources=resourcemonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitor.example.com,resources=resourcemonitors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

func (r *ResourceMonitorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
    log := log.FromContext(ctx)

    // Fetch the ResourceMonitor instance
    instance := &monitorv1alpha1.ResourceMonitor{}
    err := r.Get(ctx, req.NamespacedName, instance)
    if err != nil {
        log.Error(err, "unable to fetch ResourceMonitor")
        return reconcile.Result{}, client.IgnoreNotFound(err)
    }

    // List all Pods in the cluster
    podList := &corev1.PodList{}
    err = r.List(ctx, podList, &client.ListOptions{})
    if err != nil {
        log.Error(err, "unable to list pods")
        return reconcile.Result{}, err
    }

    // Count the number of Pods
    podCount := len(podList.Items)

    // Update the ResourceMonitor status
    instance.Status.PodsCount = podCount
    err = r.Status().Update(ctx, instance)
    if err != nil {
        log.Error(err, "unable to update ResourceMonitor status")
        return reconcile.Result{}, err
    }

    // Check the Pod threshold
    if podCount > instance.Spec.PodThreshold {
        log.Info(fmt.Sprintf("Pod count %d exceeds threshold %d", podCount, instance.Spec.PodThreshold))
        // Raise alert using Prometheus Alertmanager if available
        r.raiseAlert(instance)
    }

    return reconcile.Result{}, nil
}

func (r *ResourceMonitorReconciler) raiseAlert(instance *monitorv1alpha1.ResourceMonitor) {
    // Implement Prometheus Alertmanager integration
    // Ensure the operator does not crash if Prometheus is not available
}

func (r *ResourceMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&monitorv1alpha1.ResourceMonitor{}).
        Complete(r)
}
