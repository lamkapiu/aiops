package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/workqueue"
)

// 这两个先定义
// control结构体
type Controller struct {
	indexer  cache.Indexer  // 存储和检索对象的索引器
	queue    workqueue.TypedRateLimitingInterface[string] // 速率限制的工作队列
	informer cache.Controller  // 事件通知的控制器
}

// 初始化control结构体
func NewController(queue workqueue.TypedRateLimitingInterface[string], indexer cache.Indexer, informer cache.Controller) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
	}
}

// 然后定义 main

// 处理下一个
func (c *Controller) processNextItem() bool {
	// c 是接收者的名称，你可以用它在方法内部引用 Controller 实例的字段和方法
	// 这里的 c 是指向 Controller 实例的指针
	key, quit := c.queue.Get() // 从队列中获取下一个项
	if quit {
		return false
	}
	// defer 语句用于在当前函数（方法）执行结束时，自动执行一个指定的操作
	defer c.queue.Done(key)

	err := c.syncToStdout(key)
	c.handleErr(err, key)
	return true
}

// 输出日志
func (c *Controller) syncToStdout(key string) error {
	// 通过 key 从 indexer 中获取完整的对象
	obj, exists, err := c.indexer.GetByKey(key) // GetByKey用于从索引器中根据提供的 key 获取相应的对象
	if err != nil {
		fmt.Printf("Fetching object with key %s from store failed with %v\n", key, err)
		return err
	}

	if !exists {
		fmt.Printf("Pod %s does not exist anymore\n", key)
	} else {
		pod := obj.(*corev1.Pod) // 类型断言，获取 Deployment 对象，获取一个接口类型的值并将其转换为具体的 *v1.Deployment 类型
		fmt.Printf("Sync/Add/Update for Pod %s,\n", pod.Name)
		if pod.Name == "test-pod" {
			time.Sleep(2 * time.Second)
			return fmt.Errorf("simulated error for pod %s", pod.Name)
		}
	}
	return nil
}

// 错误处理
func (c *Controller) handleErr(err error, key string) {
	if err == nil {
		c.queue.Forget(key) // 如果有错误，忘记该项
		return
	}

	if c.queue.NumRequeues(key) < 5 {
		fmt.Printf("Retry %d for key %s\n", c.queue.NumRequeues(key), key)
		// 重新加入队列，并且进行速率限制，这会让他过一段时间才会被处理，避免过度重试
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key) // 超过重试次数，丢弃该项
	fmt.Printf("Dropping pod %q out of the queue: %v\n", key, err)
}

func main() {
	var err error
	var config *rest.Config

	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "[可选] kubeconfig 绝对路径")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "kubeconfig 绝对路径")
	}

	// 初始化 rest.Config 对象
	if config, err = rest.InClusterConfig(); err != nil {
		if config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
			panic(err.Error())
		}
	}

	// 创建 Clientset 对象
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 初始化 informer factory
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Hour*12)

	// 创建速率限制队列
	queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[string]())

	// 对 Deployment 监听
	podInformer := informerFactory.Core().V1().Pods()
	informer := podInformer.Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { onAddPod(obj, queue) },
		UpdateFunc: func(old, new interface{}) { onUpdatePod(new, queue) },
		DeleteFunc: func(obj interface{}) { onDeletePod(obj, queue) },
	})

	controller := NewController(queue, podInformer.Informer().GetIndexer(), informer)

	stopper := make(chan struct{})
	defer close(stopper)

	// 启动 informer，List & Watch
	informerFactory.Start(stopper)
	informerFactory.WaitForCacheSync(stopper)

	// 处理队列中的事件
	// go 关键字用于启动一个新的 goroutine。这意味着代码将在一个并发执行的环境中运行，不会阻塞主程序的执行。
	go func() {
		for {
			// 调用控制器的 processNextItem 方法
			if !controller.processNextItem() {
				break
			}
		}
	}()

	<-stopper
}

func onAddPod(obj interface{}, queue workqueue.TypedRateLimitingInterface[string]) {
	// 生成 key
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err == nil {
		queue.Add(key)
	}
}

func onUpdatePod(new interface{}, queue workqueue.TypedRateLimitingInterface[string]) {
	key, err := cache.MetaNamespaceKeyFunc(new)
	if err == nil {
		queue.Add(key)
	}
}

func onDeletePod(obj interface{}, queue workqueue.TypedRateLimitingInterface[string]) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err == nil {
		queue.Add(key)
	}
}
