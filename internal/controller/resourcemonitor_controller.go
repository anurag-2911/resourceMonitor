package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
    log.Info("Reconcile function called", "ResourceMonitor", req.NamespacedName)

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
    log.Info("raise alert function called", "ResourceMonitor", req.NamespacedName)
    // Check the Pod threshold
    if podCount > instance.Spec.PodThreshold {
        log.Info(fmt.Sprintf("Pod count %d exceeds threshold %d", podCount, instance.Spec.PodThreshold))
        // Raise alert using Prometheus Alertmanager if available
        r.raiseAlert(ctx,instance)
    }else{
		log.Info(fmt.Sprintf("current pod count %d",podCount))
	}

    return reconcile.Result{}, nil
}

func (r *ResourceMonitorReconciler) raiseAlert(ctx context.Context, instance *monitorv1alpha1.ResourceMonitor) {
    // Discover the Alertmanager service based on configuration
    svc := &corev1.Service{}
    err := r.Client.Get(ctx, client.ObjectKey{Namespace: instance.Spec.Alertmanager.Namespace, Name: instance.Spec.Alertmanager.ServiceName}, svc)
    if err != nil {
        r.Log.Error(err, "unable to find Alertmanager service")
        return
    }

    // Construct the Alertmanager URL
    alertmanagerURL := fmt.Sprintf("http://%s.%s.svc:%d/api/v1/alerts", svc.Name, svc.Namespace, svc.Spec.Ports[0].Port)

    // Define the alert structure
    alert := []map[string]interface{}{
        {
            "labels": map[string]string{
                "alertname":  "PodCountThresholdExceeded",
                "severity":   "warning",
                "namespace":  instance.Namespace,
                "podMonitor": instance.Name,
            },
            "annotations": map[string]string{
                "summary":     "Pod count threshold exceeded",
                "description": fmt.Sprintf("The number of pods in the cluster has exceeded the threshold of %d. Current count is %d.", instance.Spec.PodThreshold, instance.Status.PodsCount),
            },
            "startsAt": time.Now().Format(time.RFC3339),
        },
    }

    // Convert alert to JSON
    alertBytes, err := json.Marshal(alert)
    if err != nil {
        r.Log.Error(err, "unable to marshal alert to JSON")
        return
    }

    // Send the alert to Alertmanager
    resp, err := http.Post(alertmanagerURL, "application/json", bytes.NewBuffer(alertBytes))
    if err != nil {
        r.Log.Error(err, "unable to send alert to Alertmanager")
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        r.Log.Error(fmt.Errorf("unexpected response code: %d", resp.StatusCode), "received non-OK response from Alertmanager")
    } else {
        r.Log.Info("alert successfully sent to Alertmanager")
    }
}


func (r *ResourceMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&monitorv1alpha1.ResourceMonitor{}).
        Complete(r)
}
