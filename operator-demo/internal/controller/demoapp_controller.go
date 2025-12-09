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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	demov1 "operator-demo/api/v1"
)

// DemoAppReconciler reconciles a DemoApp object
type DemoAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=demo.example.com,resources=demoapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=demo.example.com,resources=demoapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=demo.example.com,resources=demoapps/finalizers,verbs=update
// 新增Deployment的权限注解（根据实际需求调整verbs）
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// 若还需操作Pod、Service等其他内置资源，继续补充
// +kubebuilder:rbac:groups="",resources=pods;services;configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DemoApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *DemoAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// 从 context 中获取 logger
	log := logf.FromContext(ctx)

	// 1. 获取 DemoApp 实例
	// Reconcile 方法的输入参数 req 包含了触发调谐的对象的 Namespace 和 Name。
	// 我们使用 r.Get 从 apiserver 中获取最新的 DemoApp 对象状态。
	var demoApp demov1.DemoApp
	if err := r.Get(ctx, req.NamespacedName, &demoApp); err != nil {
		// 如果 DemoApp 资源未找到，很可能是因为它已经被删除了。
		// 在这种情况下，我们不需要做任何事情，直接返回即可。
		if errors.IsNotFound(err) {
			log.Info("DemoApp resource not found, skip reconciliation")
			return ctrl.Result{}, nil
		}
		// 如果获取 DemoApp 失败，记录错误并返回，controller-runtime 会自动重试。
		log.Error(err, "Failed to get DemoApp resource")
		return ctrl.Result{}, err
	}

	// 2. 调谐 Deployment: 检查 Deployment 是否存在，如果不存在则创建，如果存在则更新
	var deploy appsv1.Deployment
	deployName := demoApp.Name + "-nginx"
	// 尝试获取与 DemoApp 关联的 Deployment
	if err := r.Get(ctx, types.NamespacedName{Name: deployName, Namespace: demoApp.Namespace}, &deploy); err != nil {
		// Case 1: Deployment 不存在，需要创建
		if errors.IsNotFound(err) {
			log.Info("Creating a new Deployment", "Deployment.Namespace", demoApp.Namespace, "Deployment.Name", deployName)
			// 定义一个新的 Deployment 对象
			deploy = appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deployName,
					Namespace: demoApp.Namespace,
					// 设置 OwnerReference，将 Deployment 与 DemoApp 关联起来。
					// 这样当 DemoApp 被删除时，这个 Deployment 也会被 Kubernetes 自动垃圾回收。
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(&demoApp, demov1.GroupVersion.WithKind("DemoApp")),
					},
				},
				Spec: appsv1.DeploymentSpec{
					// 副本数从 DemoApp 的 Spec 中获取
					Replicas: &demoApp.Spec.Replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": deployName},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": deployName},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "nginx",
									Image: "nginx:1.25-alpine", // 使用的 Nginx 镜像
									Ports: []corev1.ContainerPort{
										{ContainerPort: 80},
									},
								},
							},
						},
					},
				},
			}
			// 调用 client-go 的 Create 方法在集群中创建 Deployment
			if err := r.Create(ctx, &deploy); err != nil {
				log.Error(err, "Failed to create new Deployment")
				return ctrl.Result{}, err
			}
			log.Info(fmt.Sprintf("Created Deployment %s", deployName))
			// 创建成功后，我们通常会返回并等待下一次调谐，以便在 Deployment 准备就绪后更新状态。
			return ctrl.Result{Requeue: true}, nil
		} else {
			// Case 2: 获取 Deployment 时发生其他错误
			log.Error(err, "Failed to get Deployment")
			return ctrl.Result{}, err
		}
	}

	// Case 3: Deployment 已经存在，确保其状态与 Spec 一致
	// 检查副本数是否与 DemoApp.Spec.Replicas 一致
	if *deploy.Spec.Replicas != demoApp.Spec.Replicas {
		// 如果不一致，则更新 Deployment 的副本数
		log.Info(fmt.Sprintf("Updating Deployment %s replicas to %d", deployName, demoApp.Spec.Replicas))
		deploy.Spec.Replicas = &demoApp.Spec.Replicas
		if err := r.Update(ctx, &deploy); err != nil {
			log.Error(err, "Failed to update Deployment replicas")
			return ctrl.Result{}, err
		}
		// 更新后，重新排队以等待 Deployment 完成更新
		return ctrl.Result{Requeue: true}, nil
	}

	// 3. 更新 DemoApp 的 Status
	// 将 Deployment 的就绪副本数 (ReadyReplicas) 同步到 DemoApp 的 Status 字段
	// 只有当状态发生变化时才进行更新，以避免不必要的 API 调用
	if demoApp.Status.ReadyReplicas != deploy.Status.ReadyReplicas {
		demoApp.Status.ReadyReplicas = deploy.Status.ReadyReplicas
		// 调用 Status().Update() 方法来更新 CR 的 status 子资源
		// 注意：更新 status 和更新 spec/metadata 是分开的
		if err := r.Status().Update(ctx, &demoApp); err != nil {
			log.Error(err, "Failed to update DemoApp status")
			return ctrl.Result{}, err
		}
	}

	// 调谐成功，返回空的 Result，告知 controller-runtime 不需要立即重新排队
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
// 这个函数用于向 Manager 注册 Controller。
func (r *DemoAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&demov1.DemoApp{}).     // 指定 Controller 关注的主要资源是 DemoApp
		Owns(&appsv1.Deployment{}). // 指定 Controller 拥有 (own) 的子资源是 Deployment
		Complete(r)                 // 构建并注册 Controller
}
