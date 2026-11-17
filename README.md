# qingdao_backend

基于 D:\simpleAdmin\go-zero 基础模板创建，使用 Go 1.26.7、go-zero v1.10.1 和 common-boot-starter v1.0.59。

## 目录

- main.go：嵌入配置并调用 boot.Run。
- etc：本地配置及 Nacos 示例。
- internal/bootstrap：初始化、路由注册和资源关闭。
- internal/config：配置定义及启动器接口。
- internal/svc：服务上下文，按需加入数据库、Redis 和客户端。
- internal/handler：HTTP 路由及请求处理。
- internal/logic：业务逻辑。
- internal/types：请求和响应类型。
- internal/client、internal/domain、internal/util：扩展位置。

## 本地运行

在项目目录执行 `go mod tidy`，然后执行 `go run .`。
默认监听 http://127.0.0.1:3010。
模板不包含 OAuth、短信、验证码等认证业务，默认不连接 Nacos 或 Redis。

common-boot-starter 是私有依赖，需要已有模块缓存或私有仓库访问权限。
本目录不含 .git；下载 Go 模块与关联本项目到 Git 仓库是两个独立操作。

## 配置和 Nacos

配置由 `//go:embed` 编入程序，修改 etc/config.yaml 后需要重新运行或构建。
默认 Service.Active 为空，跳过 Nacos。启用时将 config.nacos.example.yaml 的 Service 配置复制到 config.yaml，填写地址、命名空间 ID、Group 和两个 Data ID，并预先在 Nacos 创建对应配置。
示例仅使用本地地址，不含原 auth 项目的环境地址或密钥。

## 构建

Windows：`go build -trimpath -o dist/qingdao_backend.exe .`

Linux（PowerShell）：

```powershell
$env:CGO_ENABLED = '0'
$env:GOOS = 'linux'
$env:GOARCH = 'amd64'
go build -trimpath -o dist/qingdao_backend .
```

Linux 构建完成后可执行 `docker build -t qingdao_backend .`。
Makefile 用于 make 和 POSIX shell 环境，不依赖 Git 获取版本信息。

## 时间戳接口

原文件留在前端，前端计算哈希后提交摘要。后端不接收文件、不重复计算哈希。
三个接口均使用 POST 和 application/json：

| 路径 | 请求 | 用途 |
| --- | --- | --- |
| /api/timestamp/rsa | `{"rid":"64位十六进制SHA256摘要"}` | 申请 RSA 时间戳 |
| /api/timestamp/sm | `{"rid":"64位十六进制SM3摘要"}` | 申请国密时间戳 |
| /api/timestamp/verify | `{"rid":"64位十六进制摘要","algorithm":"SHA256","tsr":"Base64凭证"}` | 验证凭证并核对文件摘要；algorithm 也可为 SM3 |

申请结果结构为 `{"code":"OK","message":"成功","data":{"rid":"...","algorithm":"SHA256","sn":"...","ts":123,"tsr":"..."}}`。
ts 是上级签发时间的 Unix 毫秒数。tsr 的目标格式为原始时间戳响应的 Base64，需在上级适配时确认和转换。
验证结果 data 包含 valid、rid、algorithm，以及可选的 sn、ts、reason。
验证完成但凭证无效或摘要不匹配时，返回 HTTP 200、valid=false；上级不可用返回错误，不能解释为凭证无效。
验证接口要求前端同时提供原文件摘要和摘要算法，不能仅凭解析出凭证内容就返回 valid=true。

摘要必须是 64 位十六进制，大写会规范化为小写。凭证解码后最多 512 KiB，请求体最多 768 KiB。
参数错误返回 HTTP 400；请求体过大返回 413；非 JSON 请求返回 415。

## 上级接口预留

etc/config.yaml 的 Upstream 包含 BaseURL、TokenPath、RSAPath、SMPath、VerifyPath、AppKey、AppSecret，以及超时和令牌刷新等参数。
BaseURL 和凭据暂留空，四个路径默认分别为 /token、/rsa、/sm、/verify。
当前统一在该 YAML 的 Upstream 节点维护配置；各字段用途已在配置文件中注明。
AllowInsecureHTTP 默认 false，接入后应仅允许 HTTPS，仍需验证 HTTPS 证书。
TokenTimeoutMs=3000 表示 token 请求超时 3 秒；RequestTimeoutMs=2000 表示打戳、验证请求分别超时 2 秒，不含 token 申请时间。
RefreshBeforeSeconds=300 表示 token 到期前 5 分钟需要刷新，后续实现应适配上级有效期，避免短期 token 持续刷新。
MaxResponseBytes=4194304 表示上级响应体最多读取 4 MiB，与本地请求大小限制无关。
这些参数目前只完成配置定义；HTTP 调用、HTTPS 限制、超时、token 缓存刷新和响应大小限制将在上级客户端接入时实现。
internal/client/timestamp.go 集中预留 Token、RSA、SM、Verify 方法；Token 仅供后端内部使用，未提供公开令牌接口。

当前已实现本地路由、参数校验、逻辑分层和错误响应；上级 HTTP 调用尚未实现。
BaseURL 未配置时返回 HTTP 503、TIMESTAMP_NOT_CONFIGURED；填入后仍返回 HTTP 503、TIMESTAMP_INTEGRATION_PENDING。
不会调用真实上级，也不会生成模拟成功凭证。上级接入前，本地接口可用于验证请求协议和错误分支。

待取得上级文档后，需要补齐：

- token 请求字段、认证方式、响应格式、有效期及刷新规则。
- /rsa、/sm 的字段名、Content-Type、令牌传递方式和错误码。
- tsr 的编码/压缩方式、时间单位，以及 /verify 是否包含验签、证书信任和摘要核对结果。
- 超时、幂等和重试约定，避免重复签发或扣费。

真实凭据应通过部署配置提供，不要提交到版本库；本地配置由 go:embed 编入二进制，配置变更后需要重新构建。
当前未实现本项目调用方身份认证；对外部署前应接入项目约定的认证方式。

执行 `go test ./...` 验证摘要转发、无效请求拦截、上级未就绪响应以及验证结果的摘要匹配。
