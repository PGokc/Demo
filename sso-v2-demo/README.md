# ops_local_demo - 本地模拟 ops.kabeta4 登录

## 1. 目的与概述

本项目是一个 Go 实现的 Web 应用，旨在本地环境中模拟 `ops.kabeta4` 在字节跳动统一登录链路中的角色。它允许开发者在本地启动一个服务，该服务将使用**真实**的 `delivery.feishu.cn` (登录代理) 和 `sso.bytedance.com` (单点登录服务)，从而完整地在本地复现和调试从浏览器发起登录到最终建立会话的全过程。

这对于需要与该登录链路集成的应用，或者需要排查线上登录问题的开发者来说，是一个非常有用的工具。

**核心流程**:

1.  用户浏览器访问本地服务的 `/login`。
2.  本地服务将浏览器重定向到 `delivery.feishu.cn` 的登录代理。
3.  `delivery.feishu.cn` 负责与 `sso.bytedance.com` 完成认证。
4.  认证成功后，`sso` → `delivery` → 本地服务，完成两次回调。
5.  本地服务通过 `delivery` 返回的 code 换取用户信息，并建立本地会话（Cookie）。
6.  浏览器重定向到应用的最终目标页面，此时已是登录状态。

## 2. 先决条件

-   **Go**: 版本 `1.22` 或更高。

## 3. hosts 文件映射（关键步骤）

为了让 `delivery.feishu.cn` 在完成登录后能将回调请求发送到您的**本地服务**而不是线上服务，您必须修改操作系统的 `hosts` 文件，将线上域名 `ops.kabeta4.statusfeishu.cn` 指向本地环回地址 `127.0.0.1`。

**请注意**: 测试完成后，请务必将 `hosts` 文件改回来，否则您将无法访问真正的 `ops.kabeta4` 网站。

### macOS / Linux

1.  打开终端 (Terminal)。
2.  执行以下命令编辑 `hosts` 文件：

    ```bash
    sudo vim /etc/hosts
    ```

3.  在文件末尾添加以下一行，然后保存退出：

    ```
    127.0.0.1 ops.kabeta4.statusfeishu.cn
    ```

### Windows

1.  使用管理员权限打开记事本 (Notepad)。
2.  在记事本中，点击“文件” -> “打开”，然后浏览到 `C:\Windows\System32\drivers\etc`。
3.  在右下角文件类型处选择“所有文件 (*.*)”，然后打开 `hosts` 文件。
4.  在文件末尾添加以下一行，然后保存：

    ```
    127.0.0.1 ops.kabeta4.statusfeishu.cn
    ```

## 4. 环境变量配置

在启动服务前，您需要配置以下环境变量。建议创建一个 `.env` 文件或一个启动脚本来管理这些变量。

| 变量名                      | 必需 | 描述                                                                                                             | 示例值                               |
| --------------------------- | ---- | ---------------------------------------------------------------------------------------------------------------- | ------------------------------------ |
| `KA`                        | 是   | 在 `delivery` 注册的业务方标识 (Key Account)。                                                                     | `kagaussdb`                          |
| `SESSION_SECRET`            | 是   | 用于签名和加密会话 Cookie 的密钥，必须设置为一个不易被猜测的随机字符串。                                             | `a-very-strong-and-secret-key`       |
| `LISTEN_ADDR`               | 否   | 服务监听的地址和端口。默认为 `:8080`；当同时配置了 `TLS_CERT_FILE` 和 `TLS_KEY_FILE` 且此变量为空时，默认监听 `:443`。 | `:8080` / `:443`                     |
| `DELIVERY_BASE`             | 否   | `delivery` 服务的 Base URL。默认为 `https://delivery.feishu.cn`。                                                 | `https://delivery.feishu.cn`         |
| `OPS_HOST`                  | 否   | 期望的 Cookie Domain。用于在日志中记录，**但实际 Cookie domain 会自动从请求的 Host 头获取**。                       | `ops.kabeta4.statusfeishu.cn`        |
| `REDIRECT_PATH`             | 否   | 登录成功后最终重定向到应用内的路径。默认为 `/`。                                                                   | `/`                                  |
| `DELIVERY_PUBLIC_KEY_PATH`  | 否   | (可选但推荐) `delivery` 用于签名的公钥文件路径（PEM 格式）。如果提供，服务会验证 `delivery` 返回的用户信息签名。      | `./delivery_public_key.pem`          |
| `TLS_CERT_FILE`             | 否   | (HTTPS 模式) 服务端证书文件路径（PEM 格式）。配置后将优先以 HTTPS 启动。                                           | `./ops_kabeta4_local.crt`            |
| `TLS_KEY_FILE`              | 否   | (HTTPS 模式) 服务端私钥文件路径（PEM 格式）。必须与 `TLS_CERT_FILE` 搭配使用。                                      | `./ops_kabeta4_local.key`            |

### 关于签名验证

如果配置了 `DELIVERY_PUBLIC_KEY_PATH`，应用会执行验签。这能确保从 `delivery` 收到的用户信息是真实且未被篡改的。在生产环境中，**必须进行验签**。在本地测试时，如果无法获取公钥，可以暂时不设置此变量，应用会跳过验签并打印警告日志。

## 5. 构建与运行

1.  **设置环境变量**:

    ```bash
    export KA="kagaussdb"
    export SESSION_SECRET="your-secret-goes-here"
    # 如果需要验签，请同时设置
    # export DELIVERY_PUBLIC_KEY_PATH="/path/to/your/public_key.pem"
    ```

2.  **构建**:

    在项目根目录 (`ops_local_demo`) 下执行：

    ```bash
    go build -o ops_local_demo_server .
    ```

    这会生成一个名为 `ops_local_demo_server` 的可执行文件。

3.  **运行**:

    ```bash
    ./ops_local_demo_server
    ```

    如果一切正常，您会看到类似以下的日志输出：
    ```
    2026/01/04 12:00:00.000000 step=startup delivery_base=https://delivery.feishu.cn ka=kagaussdb redirect_path=/ ops_host= public_key_path= session_secret_set=true
    2026/01/04 12:00:00.000000 step=load_public_key result=skipped reason=DELIVERY_PUBLIC_KEY_PATH_not_set
    2026/01/04 12:00:00.000000 step=listen addr=:8080
    ```

## 6. 逐步测试流程

1.  **准备浏览器**: 打开 Chrome 或 Edge 浏览器，并开启**隐身模式** (Incognito Mode) 以避免现有 Cookie 干扰。

2.  **打开开发者工具**: 按下 `F12` (或 `Cmd+Opt+I`) 打开开发者工具，切换到 **Network (网络)** 面板。

3.  **保留日志**: 勾选 **Preserve log** 选项，确保在页面跳转时网络请求记录不会被清空。

4.  **发起登录**:
    -   在地址栏输入 `http://ops.kabeta4.statusfeishu.cn:8080` 并回车。
        *注意: 域名是我们通过 `hosts` 文件映射的，端口是服务监听的端口。*
    -   页面会显示 "当前未登录"。点击 **“使用 Delivery 字节登录”** 链接。

5.  **观察网络请求链**:

    在 Network 面板中，您应该能按顺序观察到以下关键请求：

    a.  **`login`**: 对 `http://ops.kabeta4.statusfeishu.cn:8080/login` 的请求。这是第一步。

    b.  **`proxy` (302)**: 上一步的响应是一个 `302 Found` 重定向，Location 指向 `delivery.feishu.cn/api/sign/proxy?...`。浏览器自动跟随此跳转。

    c.  **`authorize` (302)**: `delivery` 服务会再次返回一个 `302`，将浏览器重定向到 `sso.bytedance.com/oauth2/authorize?...`。

    d.  **SSO 登录**: 浏览器加载字节 SSO 登录页面。您可能需要输入用户名和密码。如果已在其他地方登录，此步骤可能会自动跳过。

    e.  **`ka/callback` (302)**: SSO 认证成功后，会携带一个 `code` 重定向回 `delivery.feishu.cn/api/sign/ka/callback?...`。

    f.  **`byted/callback` (302)**: `delivery` 处理完上一步的 `code` 后，会生成一个新的 `code`，并最终重定向回我们的本地服务：`http://ops.kabeta4.statusfeishu.cn:8080/login/sign/byted/callback?...`。

    g.  **`/` (根路径)**: 本地服务在 `byted/callback` 接口中成功处理了回调、换取了用户信息、并设置了会话 Cookie。最后，它会重定向到 `redirect_uri` 指定的路径 (默认为 `/`)。

6.  **验证结果**:
    -   浏览器最终停留在 `http://ops.kabeta4.statusfeishu.cn:8080/`。
    -   页面显示 "已登录" 和从 `delivery` 获取的用户信息 (JSON 格式)。
    -   在开发者工具的 **Application (应用)** -> **Cookies** 面板下，可以看到为 `ops.kabeta4.statusfeishu.cn` 域设置了一个名为 `ops_local_session` 的 Cookie。

## 7. 故障排查

-   **错误: "KA 未配置"**
    -   原因: `KA` 环境变量未设置或为空。
    -   解决: 确保在启动服务前已正确 `export KA=...`。

-   **点击登录后，页面停在 Delivery 或 SSO 报错**
    -   原因: 可能是 `KA` 值不正确，或者 `delivery` 服务侧的配置 (如回调地址白名单) 有问题。
    -   解决: 确认 `KA` 值是否正确。检查本地服务的日志，看是否有 `delivery` 返回的错误信息。

-   **回调到本地服务后，出现 "验签失败"**
    -   原因: `DELIVERY_PUBLIC_KEY_PATH` 指向的公钥文件不正确或已过期。
    -   解决: 确认公钥是否与 `delivery` 环境匹配。如果只是临时测试，可以暂时不设置该环境变量以跳过验签。

-   **登录成功了，但刷新页面后又回到未登录状态**
    -   原因: Cookie 未被正确设置或发送。
    -   解决:
        1.  确认 `SESSION_SECRET` 环境变量已设置。
        2.  检查浏览器开发者工具，看 `ops_local_session` Cookie 是否已成功写入。
        3.  确认 Cookie 的 `Domain` 属性是否为 `ops.kabeta4.statusfeishu.cn`。如果 `Domain` 不对，可能是因为您访问的地址不是 `hosts` 文件中配置的域名。

-   **修改 `hosts` 文件后，域名仍然指向线上 IP**
    -   原因: 操作系统缓存了 DNS 解析结果。
    -   解决: 尝试刷新系统的 DNS 缓存。
        -   macOS: `sudo dscacheutil -flushcache; sudo killall -HUP mDNSResponder`
        -   Windows: `ipconfig /flushdns`


---

## 附录 A: HTTPS 本地直连模式

除了在 `ops_local_demo` 前面再架设一层 Nginx 等反向代理来实现 HTTPS 外，也可以直接让本应用支持 HTTPS。这在某些场景下更简单，因为它减少了外部依赖。

### A.1. 为何需要 HTTPS

`delivery.feishu.cn` 在完成认证并回调时，会严格按照其后台配置的回调地址发起请求。对于 `ops.kabeta4`，这个地址是 `https://ops.kabeta4.statusfeishu.cn/login/sign/byted/callback`。请注意：

1.  **协议是 `https://`**: 这意味着浏览器会向 443 端口发起 TLS 连接。
2.  **域名是 `ops.kabeta4.statusfeishu.cn`**: 这就是为什么我们必须修改 `hosts` 文件。
3.  **不带端口号**: 这隐含地表示目标端口是 HTTPS 的标准端口 443。

因此，我们的本地服务**必须在 443 端口上监听 HTTPS 流量**，才能成功接收回调。

### A.2. 步骤

#### 1. 生成自签名证书

为了使应用能处理 HTTPS 请求，你需要一对 TLS 证书和私钥。对于本地开发，可以使用 `openssl` 快速生成一个自签名证书。

**重要**: 证书的通用名称 (Common Name, CN) **必须**是 `ops.kabeta4.statusfeishu.cn`，以匹配 `hosts` 文件中的域名。

在终端执行以下命令，它会在当前目录下生成 `ops_kabeta4_local.crt` (证书) 和 `ops_kabeta4_local.key` (私钥) 两个文件：

```bash
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout ops_kabeta4_local.key \
  -out ops_kabeta4_local.crt \
  -days 3650 \
  -subj "/CN=ops.kabeta4.statusfeishu.cn"
```

#### 2. 配置并以特权模式运行

现在，设置 `TLS_CERT_FILE` 和 `TLS_KEY_FILE` 环境变量指向刚刚生成的文件。

由于需要监听 443 端口（小于 1024 的端口是特权端口），你**必须**使用 `sudo` 来运行程序，或者通过 `setcap` 命令为可执行文件授予绑定低位端口的权限。

**方法一：使用 sudo (最简单)**

```bash
# 确保 KA 和 SESSION_SECRET 已设置
export KA="kagaussdb"
export SESSION_SECRET="a-very-strong-and-secret-key"

# 设置证书路径 (假设证书在当前目录)
export TLS_CERT_FILE="./ops_kabeta4_local.crt"
export TLS_KEY_FILE="./ops_kabeta4_local.key"
# LISTEN_ADDR 会自动默认为 :443，也可以显式指定

# 使用 sudo 启动
sudo ./ops_local_demo_server
```

**方法二：使用 setcap (免 root 运行，仅限 Linux)**

如果你不想每次都用 `sudo`，可以在 Linux 上给可执行文件一次性授权：

```bash
# 先构建
go build -o ops_local_demo_server .

# 授权 (只需执行一次)
sudo setcap 'cap_net_bind_service=+ep' ./ops_local_demo_server

# 之后就可以普通用户身份启动了
export ... # (设置其他环境变量)
./ops_local_demo_server
```

#### 3. 浏览器体验

-   **首次访问**: 当你用浏览器打开 `https://ops.kabeta4.statusfeishu.cn` 时，浏览器会显示一个安全警告（例如 “您的连接不是私密连接”），因为证书是自签名的，不受信任。
-   **如何处理**:
    -   点击 “高级” 或 “显示详细信息”。
    -   找到并点击 “继续前往 ops.kabeta4.statusfeishu.cn (不安全)” 的链接。
    -   这样，浏览器就会临时信任这个证书，后续的登录和回调流程就可以正常进行了。

完成以上步骤后，整个登录链路就能在纯 HTTPS 环境下顺畅运行，并且设置的会话 Cookie 会自动带上 `Secure` 标记。
