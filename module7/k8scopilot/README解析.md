# 解析代码

### 1.接收标准输入，再去调调大模型
```go
func startChat() {
	// bufio
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("我是 K8s Copilot，有什么可以帮助你：")
	for {
		fmt.Print("> ")
		if scanner.Scan() {
			input :=scanner.Text()
			if input == "exit" {
				fmt.Println("goodbye! ")
				break
			}
			if input == "" {
				continue
			}
			fmt.Println("u say: ", input)		
		}
	}
}
```
### 2.调用大模型生成yaml
前提需要先去封装util，就是openai的一个client的一个客户端的封装
### 2.1 创建utils目录，并创建openai.go文件
通过包go-openai去封装两模块方法。
```bash
"github.com/sashabaranov/go-openai"  // Go 语言的 OpenAI API 客户端库
```
NewOpenAIClient返回一个新的openai的客户端实例
SendMessage向 OpenAI 发送消息并返回响应
### 2.2 调用大模型
在chatgp.go文件中新增processInput。
在 processInput 函数中，您可以调用 SendMessage 来获取 OpenAI 的响应。通过 utils.NewOpenAIClient 创建客户端后，您就可以直接使用 client.SendMessage 发送用户的输入并获取返回的消息。

```go
func processInput(input string) string {
	// 1. 使用封装好的 OpenAI 客户端
	client, err := utils.NewOpenAIClient()
	if err != nil {
		return err.Error() // 如果客户端创建失败，返回错误信息
	}

	// 2. 发送消息给 OpenAI，获取回应
	response, err := client.SendMessage("你是一个K8s管理员，帮助用户执行K8s操作。", input)
	if err != nil {
		return err.Error() // 如果发送失败，返回错误信息
	}

	return response // 返回 OpenAI 的回应
}
```
然后修改startChat，直接调用新增的processInput函数输出结果。
### 3.简化大模型返回的输出结果，只保留yaml文件内容
取决于怎么写这个system prompt
```bash
response, err := client.SendMessage("你现在是一个 K8s 资源生成器，根据用户输入生成 K8s YAML，注意除了 YAML 内容以外不要输出任何内容，此外不要把 YAML 放在 ``` 代码快里", input)
```
### 4.定义openai的function calling
functionCalling 调用 OpenAI Function进行功能调用，下面是一个定义type为function的tool：
```go
f1 := openai.FunctionDefinition{
    Name:        "generateAndDeployResource", // 功能名称
    Description: "生成 K8s YAML 并部署资源",  // 功能描述
    Parameters: jsonschema.Definition{
        Type: jsonschema.Object, // 参数类型定义为对象
        Properties: map[string]jsonschema.Definition{
            "user_input": { // 用户输入的参数
                Type:        jsonschema.String,
                Description: "用户输出的文本内容，要求包含资源类型和镜像",
            },
        },
        Required: []string{"location"}, // 必须的参数
    },
}

```
定义好后，便可以调用openai api去获取对话响应
```go
	dialogue := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: input},
	}
    //调用 OpenAI API 获取对话响应
	resp, err := client.Client.CreateChatCompletion(context.TODO(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT4TurboPreview,
			Messages: dialogue,
			Tools:    []openai.Tool{t1, t2, t3},
		},
	)
```
输出对话响应内容
```go
msg := resp.Choices[0].Message.ToolCalls[0]
```
输出结果：
![alt text](<截屏2024-11-05 15.50.01.png>)
### 5.编写需要调用的函数
因为openai只会告知我们会调用哪个函数function，传哪些参数，但是没有真正去调用，调用函数是需要编写代码。
编写callfunction函数，用于分别调用之前生成的3个定义的generateAndDeployResource、queryResource、deleteResource
对应编写generateAndDeployResource、queryResource、deleteResource三个函数，并在函数内直接调用utils封装好的client-go去直接操作，比如生成、查询、删除。

因为是通过dynamic客户端是调用，所以需要生成unstructured 对象，然后再去取GVK。
```go
	// 将 YAML 转成 unstructured 对象
	unstructuredObj := &unstructured.Unstructured{}
	_, _, err = scheme.Codecs.UniversalDeserializer().Decode([]byte(yamlContent), nil, unstructuredObj)
	if err != nil {
		return "", err
	}

```
unstructured 对象可以使用GroupVersionKind自动获取GVK，然后再转成GVR便于clientGo.DynamicClient调用，转成的GVR通过RESTMapping便能获取到mapping.Resource。
```go
	// 从 unstructuredObj 中提取 GVK
	gvk := unstructuredObj.GroupVersionKind()


    // 用 GVK 转 GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return "", err
	}
```
使用dynamicClient.Resource()指定命名空间和资源选项, Create()方法创建deployment.
```go
_, err = clientGo.DynamicClient.Resource(mapping.Resource).Namespace(namespace).Create(context.TODO(), unstructuredObj, metav1.CreateOptions{})
```

#### 输出结果
部署
![alt text](<截屏2024-11-07 17.04.27.png>)
查询
![alt text](<截屏2024-11-07 17.05.00.png>)
删除
![alt text](<截屏2024-11-07 17.43.34.png>)



### 拓展
#### 匿名结构体
params := struct { ... }{}，它在 Go 中用于创建没有名字的结构体类型并立即初始化它。匿名结构体非常适合用作局部变量，尤其是在需要一个临时的数据结构时。

json:"user_input"是结构体标签（tag）。Go 允许你为结构体字段添加标签，json:"user_input" 告诉 Go 在进行 JSON 编解码时，这个字段在 JSON 中对应的键（key）是 user_input。
json.Unmarshal主要用来将 JSON 数据 解析（解码）为 Go 的数据结构（如结构体、切片、映射等）。
params := struct {
			UserInput string `json:"user_input"`
		}{}
则会解析成
params := struct {
    UserInput string `json:"user_input"`
}{
    UserInput: "deploy nginx",
}