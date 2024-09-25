# zuoye1：多阶段构建

### 设置环境变量

```bash
$ export TF_VAR_secret_id=
$ export TF_VAR_secret_key=


#window环境下
$env:TF_VAR_secret_id = ""
$env:TF_VAR_secret_key = ""
```
### 初始化并应用

```bash
cd modules3\zuoye1\
$ terraform init
$ terraform apply
```
#### 输出结果：

![图片](https://github.com/lamkapiu/aiops/blob/master/module3/images/0.png)
### 构建docker镜像

```bash
docker build -t your_image_name .
```
![图片](https://github.com/lamkapiu/aiops/blob/master/module3/images/1.png)
#### 输出结果：

![图片](https://github.com/lamkapiu/aiops/blob/master/module3/images/2.png)
# zuoye2：新增Pre-install Hooks

### helm安装helm chart应用

```bash
cd modules3\zuoye2\
$ helm dependency build   # 下载并构建缺失的依赖项
$ helm install vote . -n vote --create-namespace 
```
#### 运行结果：

![图片](https://github.com/lamkapiu/aiops/blob/master/module3/images/3.png)

![图片](https://github.com/lamkapiu/aiops/blob/master/module3/images/4.png)

### 销毁

```bash
# 先删除 k3s state，否则会出错
$ terraform state rm 'module.k3s'


# 再执行删除
$ terraform destroy -auto-approve
```

