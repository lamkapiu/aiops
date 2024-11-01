# zuoye1: 使用informer + RateLimitingQueue监听Pod事件

### 初始化go
```bash
go mod init github.com/lamkapiu/module6-zuoye1
```

### 加载依赖包
```bash
go mod tity
```

### 执行main.go
```bash
go run main.go
```

输出结果：
![alt text](<截屏2024-11-01 14.31.14.png>)

开始监听中，手动创建pod触发
```bash
kubectl run test-pod --image=nginx
```
再查看监听状况：
![alt text](<截屏2024-11-01 14.34.50.png>)

监听结果符合预期，说明监听正常运行且成功。

# zuoye2: 创建自定义的CRD并使用dynamicClient获取

### 初始化go
```bash
go mod init github.com/lamkapiu/module6-zuoye
```

### 加载依赖包
```bash
go mod tity
```

### 应用定义资源
```bash
kubectl apply -f crd.yaml
kubectl apply -f aiops.yaml
```

### 执行go文件
```bash
go run main6-zuoye2.go get aiops
```

输出结果：
![alt text](<截屏2024-11-01 14.41.08.png>)

手动获取，通过kubetcl查看：
![alt text](<截屏2024-11-01 15.29.51.png>)

结果一致，说明程序运行正常且成功。