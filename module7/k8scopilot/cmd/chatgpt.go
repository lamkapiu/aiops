/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"context"
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
	"github.com/lamkapiu/k8scopilot/utils"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/restmapper"
	"k8s.io/kubectl/pkg/scheme"
)

// chatgptCmd represents the chatgpt command
var chatgptCmd = &cobra.Command{
	Use:   "chatgpt",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		startChat()
	},
}

func startChat() {
	// bufio
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("我是 K8s Copilot，有什么可以帮助你：")

	// 1.接受标准输入，再去调调大模型;
	// 2.嗲用大模型生成yaml,前提需要先去封装util，就是openai的一个client的一个客户端的封装
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
			// fmt.Println("u say: ", input)
			response := processInput(input)
			fmt.Println(response)			
		}
	}
}

func processInput(input string) string {
	// 1. 先实现一个简单的回复
	//return fmt.Sprintf("你说的是: %s", input)
	client, err := utils.NewOpenAIClient()
	if err != nil {
		return err.Error()
	}
	// 2. 封装 utils/openai.go，调用 OpenAI API 得到回复
	// response, err := client.SendMessage("你现在是一个 K8s 资源生成器，根据用户输入生成 K8s YAML，注意除了 YAML 内容以外不要输出任何内容，此外不要把 YAML 放在 ``` 代码快里", input)

	// return response

	// 3. 调用 OpenAI Function calling
	response := functionCalling(input, client)
	return response
}

func functionCalling(input string, client *utils.OpenAI) string {
	// 用来生成 K8s YAML 并部署资源
	f1 := openai.FunctionDefinition{
		Name:        "generateAndDeployResource",
		Description: "生成 K8s YAML 并部署资源",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"user_input": {
					Type:        jsonschema.String,
					Description: "用户输出的文本内容，要求包含资源类型和镜像",
				},
			},
			Required: []string{"location"},
		},
	}
	t1 := openai.Tool{
		Type:     openai.ToolTypeFunction,  // openai.Tool 是一个工具类型，表示 OpenAI 系统可以调用的工具
		Function: &f1,
	}

	// 用来查询 K8s 资源
	f2 := openai.FunctionDefinition{
		Name:        "queryResource",
		Description: "查询 K8s 资源",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"namespace": {
					Type:        jsonschema.String,
					Description: "资源所在的命名空间",
				},
				"resource_type": {
					Type:        jsonschema.String,
					Description: "K8s 资源标准类型，例如 Pod、Deployment、Service 等",
				},
			},
		},
	}

	t2 := openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: &f2,
	}

	// 用来删除 K8s 资源
	f3 := openai.FunctionDefinition{
		Name:        "deleteResource",
		Description: "删除 K8s 资源",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"namespace": {
					Type:        jsonschema.String,
					Description: "资源所在的命名空间",
				},
				"resource_type": {
					Type:        jsonschema.String,
					Description: "K8s 资源标准类型，例如 Pod、Deployment、Service 等",
				},
				"resource_name": {
					Type:        jsonschema.String,
					Description: "资源名称",
				},
			},
		},
	}

	t3 := openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: &f3,
	}

	// ChatMessageRoleUser也就是user，组装message，dialogue的切片
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

	if err != nil || len(resp.Choices) != 1 {
		return fmt.Sprintf("Completion error: err:%v len(choices):%v\n", err,
			len(resp.Choices))

	}
	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) != 1 {
		return fmt.Sprintf("Completion error: len(toolcalls): %v\n", len(msg.ToolCalls))
	}

	// simulate calling the function & responding to OpenAI
	// 上下文对话历史
	dialogue = append(dialogue, msg)

	// 3. 到这里截止第三步，运行看输出效果
	// return fmt.Sprintf("OpenAI called us back wanting to invoke our function '%v' with params '%v'\n",
	// msg.ToolCalls[0].Function.Name, msg.ToolCalls[0].Function.Arguments)

	// 4. 解析 OpenAI 返回的消息，手动调用对应的函数
	result, err := callFunction(client, msg.ToolCalls[0].Function.Name, msg.ToolCalls[0].Function.Arguments)
	if err != nil {
		return fmt.Sprintf("Error calling function: %v\n", err)
	}
	return result
	
}

// 4 所需要的函数方法
// 根据 OpenAI 返回的消息，调用对应的函数
func callFunction(client *utils.OpenAI, name, arguments string) (string, error) {
	if name == "generateAndDeployResource" {
		// 匿名结构体 params
		params := struct {
			UserInput string `json:"user_input"`
		}{}
		// 函数通过 json.Unmarshal 将 arguments 解析为一个匿名结构体 params，该结构体包含一个 UserInput 字段。
		if err := json.Unmarshal([]byte(arguments), &params); err != nil {
			return "", fmt.Errorf("failed to parse function call name=%s arguments=%s", name, arguments)
		}
		return generateAndDeployResource(client, params.UserInput)
	}
	if name == "queryResource" {
		params := struct {
			Namespace    string `json:"namespace"`
			ResourceType string `json:"resource_type"`
		}{}
		if err := json.Unmarshal([]byte(arguments), &params); err != nil {
			return "", fmt.Errorf("failed to parse function call name=%s arguments=%s", name, arguments)
		}
		return queryResource(params.Namespace, params.ResourceType)
	}
	if name == "deleteResource" {
		params := struct {
			Namespace    string `json:"namespace"`
			ResourceType string `json:"resource_type"`
			ResourceName string `json:"resource_name"`
		}{}
		if err := json.Unmarshal([]byte(arguments), &params); err != nil {
			return "", fmt.Errorf("failed to parse function call name=%s arguments=%s", name, arguments)
		}
		return deleteResource(params.Namespace, params.ResourceType, params.ResourceName)

	}
	return "", fmt.Errorf("unknown function: %s", name)
}

// 4. 生成 K8s YAML 并部署资源
func generateAndDeployResource(client *utils.OpenAI, userInput string) (string, error) {
	yamlContent, err := client.SendMessage("你现在是一个 K8s 资源生成器，根据用户输入生成 K8s YAML，注意除了 YAML 内容以外不要输出任何内容，此外不要把 YAML 放在 ``` 代码快里", userInput)
	if err != nil {
		return "", fmt.Errorf("ChatGPT error: %v", err)
	}
	// 这里可以看一下调用结果
	// return yamlContent, nil
	// 调用 dynamic client 部署资源，封装到 utils/clien_go.go 中
	clientGo, err := utils.NewClientGo(kubeconfig) // kubeconfig 是一个全局 Flag
	if err != nil {
		return "", fmt.Errorf("Error creating Kubernetes clients: %v", err)
	}
	resources, err := restmapper.GetAPIGroupResources(clientGo.DiscoveryClient)
	if err != nil {
		return "", err
	}

	// 将 YAML 转成 unstructured 对象
	unstructuredObj := &unstructured.Unstructured{}
	_, _, err = scheme.Codecs.UniversalDeserializer().Decode([]byte(yamlContent), nil, unstructuredObj)
	if err != nil {
		return "", err
	}
	// 创建 mapper
	mapper := restmapper.NewDiscoveryRESTMapper(resources)
	// 从 unstructuredObj 中提取 GVK
	gvk := unstructuredObj.GroupVersionKind()
	// 用 GVK 转 GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return "", err
	}

	namespace := unstructuredObj.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	_, err = clientGo.DynamicClient.Resource(mapping.Resource).Namespace(namespace).Create(context.TODO(), unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("YAML content:\n%s\n\nDeployment successful.", yamlContent), nil
}

// 5. 查询 K8s 资源
func queryResource(namespace, resourceType string) (string, error) {
	clientGo, err := utils.NewClientGo(kubeconfig)
	resourceType = strings.ToLower(resourceType)
	var gvr schema.GroupVersionResource
	switch resourceType {
	case "deployment":
		gvr = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	case "service":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	case "pod":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	default:
		return "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	// Query the resources using the dynamic client
	resourceList, err := clientGo.DynamicClient.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list resources: %w", err)
	}

	// Iterate over the resources and print their names (or handle them as needed)
	result := ""
	for _, item := range resourceList.Items {
		result += fmt.Sprintf("Found %s: %s\n", resourceType, item.GetName())
	}

	return result, nil
}

// 删除 K8s 资源，课后作业
func deleteResource(namespace, resourceType, resourceName string) (string, error) {
	fmt.Printf("Are you sure you want to delete the %s resource '%s' in namespace '%s'? (yes/no): ", resourceType, resourceName, namespace)

	// 获取用户输入
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %v", err)
	}

	// 去掉输入的换行符并将输入转换为小写
	input = strings.TrimSpace(input)
	if input != "yes" {
		return "Operation cancelled by user.", nil
	}
	
	clientGo, err := utils.NewClientGo(kubeconfig)
	resourceType = strings.ToLower(resourceType)
	var gvr schema.GroupVersionResource
	switch resourceType {
	case "deployment":
		gvr = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	case "service":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	case "pod":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	default:
		return "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	err = clientGo.DynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
	
	if err != nil {
		return "", fmt.Errorf("failed to delete resource %s of type %s in namespace %s: %v", resourceName, resourceType, namespace, err)
	}

	return fmt.Sprintf("successed to delete resource %s of type %s in namespace %s: %v", resourceName, resourceType, namespace), nil

}

func init() {
	askCmd.AddCommand(chatgptCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// chatgptCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// chatgptCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}