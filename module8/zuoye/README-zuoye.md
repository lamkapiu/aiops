# zuoye：增强实战一，增加configmap字段，实现一并生成comfigMap

### 1.进入对应目录并出初始化
```bash
cd module8/zuoye && go mod init github.com/lamkapiu/zuoye
```
### 2.kubebuilder
```bash
kubebuilder init --domain=aiops.com

kubebuilder create api --group application --version v1 --kind Application
```
### 3.完善api/v1/application_type.go
这边仅显示出zuoye中提到的configmap
```go
type ApplicationSpec struct {
	
    ...
    // 添加 ConfigMap 字段
	ConfigMap  map[string]string        `json:"configMap"`
}
```
### 4.生成crd
```bash
make manifests
```
### 5.编写reconcile业务逻辑部分
这边仅显示出zuoye中提到的configmap
#### 5.1.定义reconcileConfigMap函数
```go
func (r *ApplicationReconciler) reconcileConfigMap(ctx context.Context, app *v1.Application) error {
	// Create or update configMap
	configMapName := app.Name + "-configmap"
	// 定义 ConfigMap 资源
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: app.Namespace,
		},
	}
	// 使用 controllerutil 创建或更新 ConfigMap
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, configMap, func() error {
		configMap.Data = app.Spec.ConfigMap
		return controllerutil.SetControllerReference(app, configMap, r.Scheme)
	})

	if err != nil {
		log.Log.Error(err, "unable to create or update configMap")
		return err
	}

	log.Log.Info("configMap created or updated", "Name", configMap.Name)
	return nil
}
```
#### 5.2 在reconcile函数中调用
```go
	// configMap
	if err := r.reconcileConfigMap(ctx, &app); err != nil {
		logger.Error(err, "Failed to reconcile ConfigMap")
		return ctrl.Result{}, err
	}
```
### 6.将crd安装到集群中
```bash
make install
```
### 7.编写资源定义
编写config/simple下的文件的spec
```yaml
apiVersion: application.aiops.com/v1
kind: Application
metadata:
  labels:
    app.kubernetes.io/name: zuoye
    app.kubernetes.io/managed-by: kustomize
  name: application-sample
spec:
  # TODO(user): Add fields here
  deployment:
    image: nginx
    replicas: 1
    port: 80
  service:
    ports:
      - port: 80
        targetPort: 80
  ingress:
    ingressClassName: nginx
    rules:
      - host: example.foo.com
        http:
          paths:
            - path: /
              pathType: Prefix
              backend:
                service:
                  name: application-sample
                  port:
                    number: 80
  configMap:
    lesson: "8"
    from: aiops
```
### 8.运行operator并部署
```bash
make run
kubectl apply -f config/samples/application_v1_application.yaml
```
### 9.输出结果
当apply应用定义部署到集群中后，可以看到operator便开始工作。
![alt text](<截屏2024-11-22 11.41.44.png>)

从日志中可以看到spec中定义的四个资源类型都成功部署了。

![alt text](<截屏2024-11-22 11.42.55.png>)

由于设置了 OwnerReference，确保 ConfigMap 会随 Application 一起被删除。
所以在执行delete操作后，资源也随之删除。