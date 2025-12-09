// idl/mgr.thrift

namespace go thrift.kitex.api

// 简化版 Mgr 请求，保证示例可独立运行
struct MgrReq {
    1: string Action,
    2: string JobID,
    3: bool Async,
}

// 业务私有请求，可嵌套在 MgrReq 中
struct PrivateReq {
    1: string SomeParameter,
}

// 通用响应结构
struct Response {
    1: i32 Code,
    2: string Msg,
    3: MgrResp Data,
}

// 简化版 Mgr 响应
struct MgrResp {
    1: string JobID,
    2: string JobStatus,
    3: string JobStage,
}

// 定义服务与 RPC 接口
service AppService {
    Response Action(1: MgrReq req)
}
