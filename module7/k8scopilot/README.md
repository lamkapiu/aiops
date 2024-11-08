### 实现deleteResource方法
#### 1.在chatgpt.go文件中deleteResource方法处编写
#### 2.构建并运行
```bash
go bulid -o k8scopilot
```
```bash
./k8scopilot ask chatgpt
```
#### 3.实现结果
输入**删除default ns的nginx-deployment deploy**
![alt text](<截屏2024-11-08 14.31.46.png>)
输入**删除default ns的test-deployment deploy**
![alt text](<截屏2024-11-08 14.32.44.png>)
说明能够正常调用到deleteResource方法，并根据用户输入二次确认后进行删除资源。