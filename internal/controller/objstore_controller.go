/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mycninfv1apha1 "github.com/dradenvandewind/extending-kubernetes-cninf.git/api/v1apha1"
)

const configMapName = "%s-cm"

// ObjStoreReconciler reconciles a ObjStore object
type ObjStoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	S3svc  *s3.S3
}

// +kubebuilder:rbac:groups=mycninf.test.erwanleblond.com,resources=objstores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mycninf.test.erwanleblond.com,resources=objstores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mycninf.test.erwanleblond.com,resources=objstores/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ObjStore object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile
func (r *ObjStoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the ObjStore instance
	instance := &mycninfv1apha1.ObjStore{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "Failed to get ObjStore resource", "NamespacedName", req.NamespacedName)
		}
		// Return without requeueing if the resource is not found (deleted)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Initialize status if empty
	if instance.Status.State == "" {
		instance.Status.State = mycninfv1apha1.PendingState
		if err := r.Status().Update(ctx, instance); err != nil {
			log.Error(err, "Failed to update status to PENDING_STATE", "ObjStore", instance.Name)
			return ctrl.Result{}, err
		}
		log.Info("Initialized ObjStore status to PENDING_STATE", "ObjStore", instance.Name)
	}

	// TODO: Implement custom reconciliation logic
	log.Info("Reconcile logic is not yet implemented", "ObjStore", instance.Name)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ObjStoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mycninfv1apha1.ObjStore{}).
		Named("objstore").
		Complete(r)
}

func (r ObjStoreReconciler) createResources(ctx context.Context, objStore *mycninfv1apha1.ObjStore) error {
	// update first status
	objStore.Status.State = mycninfv1apha1.CreatingState
	err := r.Status().Update(ctx, objStore)
	if err != nil {
		return err
	}
	// create bucket
	b, err := r.S3svc.CreateBucket(&s3.CreateBucketInput{
		Bucket:                     aws.String(objStore.Spec.Name),
		ObjectLockEnabledForBucket: aws.Bool(objStore.Spec.Locked),
	})
	if err != nil {
		return err
	}
	// wait for it to be created
	err := r.S3svc.WaitUntilBucketExists(&s3.HeadBucketInput{Bucket: aws.String(objStore.Spec.Name)})
	if err != nil {
		return err

	}
	// create config map
	data := make(map[String]string, 0)
	data["bucketName"] = objStore.Spec.Name
	data["location"] = *b.Location
	configmap := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(configMapName, objStore.Spec.Name),
			Namespace: objStore.Namespace,
		},
		Data: data,
	}
	err = r.Create(ctx, configmap)
	if err != nil {
		return err
	}
	
}
