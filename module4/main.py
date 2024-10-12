from openai import OpenAI
import json
import time

client = OpenAI(
    api_key="sk-0gXNjSjCiwEvD2o6B4D24a99Ae9b43Cd8216C1BcE32a6d57",
    base_url="https://api.apiyi.com/v1",
)


def analyze_loki_log(query_str):
    print("\n函数调用的参数: ", query_str)
    return json.dumps({"log": "this is error log"})

def modify_config(service_name,key,value):
    print("\n函数调用的参数: ", service_name,key,value)
    return json.dumps({"修改gateway的配置": "vendor修改为alipay"})

def restart_service(service_name):
    print("\n函数调用的参数: ", service_name)
    return json.dumps({"restart_service": "重启gateway服务"})

def apply_manifest(resource_type,image):
    print("\n函数调用的参数: ", resource_type,image)
    return json.dumps({"部署一个deployment": "镜像是nginx"})

def run_conversation():
    """Query: 1.帮我修改gateway的配置,vendor修改为alipay; 2.帮我重启gateway服务; 3.帮我部署一个deployment, 镜像是nginx"""

    # 步骤一：把所有预定义的 function 传给 chatgpt
    query = input("输入查询指令：")
    messages = [
        {
            "role": "system",
            "content": "你是一个 service 应用助手，你可以帮助用户修改 service 配置或者是重启 service 应用，你也可以帮助用户部署 resources 应用，你可以调用多个函数来帮助用户完成任务。",
        },
        {
            "role": "user",
            "content": query,
        },
    ]
    tools = [
        {
            "type": "function",
            "function": {
                "name": "modify_config",
                "description": "修改配置",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "service_name": {
                            "type": "string",
                            "description": "修改service 的配置,例如: 帮我修改gateway的配置",
                        },
                        "key": {
                            "type": "string",
                            "description": "需要修改的key 值,例如: vendor",
                        },
                        "value": {
                            "type": "string",
                            "description": "需要修改成的value 值,例如: alipay",
                        },                       
                    },
                    "required": ["service_name","key","value"],
                },
            },
        },
        {
            "type": "function",
            "function": {
                "name": "restart_service",
                "description": "重启服务应用",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "service_name": {
                            "type": "string",
                            "description": "重启service 的服务,例如: 帮我重启gateway服务",
                        },                    
                    },
                    "required": ["service_name"],
                },
            },
        }, 
        {
            "type": "function",
            "function": {
                "name": "apply_manifest",
                "description": "部署应用定义的资源",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "resource_type": {
                            "type": "string",
                            "description": "部署应用定义 resources,例如: 部署一个deployment",
                        }, 
                        "image": {
                            "type": "string",
                            "description": "镜像参数,例如: 镜像是nginx",
                        },                  
                    },
                    "required": ["resource_type","image"],
                },
            },
        },              
    ]

    response = client.chat.completions.create(
        model="gpt-4o",
        messages=messages,
        tools=tools,
        tool_choice="auto",
    )
    response_message = response.choices[0].message
    tool_calls = response_message.tool_calls
    print("\nChatGPT want to call function: ", tool_calls)
    # 步骤二：检查 LLM 是否调用了 function
    if tool_calls is None:
        print("not tool_calls")
    if tool_calls:
        available_functions = {
            "modify_config": modify_config,
            "restart_service": restart_service,
            "apply_manifest": apply_manifest,

        }
        messages.append(response_message)
        # 步骤三：把每次 function 调用和返回的信息传给 model
        for tool_call in tool_calls:
            function_name = tool_call.function.name
            function_to_call = available_functions[function_name]
            function_args = json.loads(tool_call.function.arguments)
            function_response = function_to_call(**function_args)
            messages.append(
                {
                    "tool_call_id": tool_call.id,
                    "role": "tool",
                    "name": function_name,
                    "content": function_response,
                }
            )
        # 步骤四：把 function calling 的结果传给 model，进行对话
        response = client.chat.completions.create(
            model="gpt-4o",
            messages=messages,
        )
        return response.choices[0].message.content


print("LLM Res: ", run_conversation())
