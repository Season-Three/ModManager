# 整合包模组管理

TownStory 整合包的模组功能开关工具。按 SHA256 哈希识别 `mods/` 目录中的模组文件，
把预设的可选功能（动画拓展、视觉优化、性能优化等）打包成开关面板，通过重命名
`.jar` / `.jar.disabled` 来启用或禁用，不依赖模组加载器本身的配置。

## 使用方法

将编译出的 `modmanager.exe` 放在整合包根目录（与 `mods` 文件夹同级），双击运行即可。

### 扫描模组哈希

新增可选功能前，需要知道目标模组文件的 SHA256。  
命令行运行：

```
modmanager.exe --scan
```

会列出 `mods/` 目录下所有 `.jar` / `.jar.disabled` 文件的文件名与哈希，复制哈希
填入 `app.go` 的 `features` 配置即可。

## 修改功能分组

功能分组定义在 [`app.go`](app.go) 的 `features` 变量里，每组包含名称、描述和该
功能对应的模组哈希列表：

```go
var features = []Feature{
    {
        Name:        "动画拓展",
        Description: "...",
        Hashes:      []string{"<sha256>", "..."},
    },
    ...
}
```

同一个哈希只要模组文件内容不变，跨版本更新目录结构也不受影响。

## 开发

需要 
- [Go](https://go.dev/dl/)
- [Node.js](https://nodejs.org/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

```bash
wails dev      # 开发模式，前端热重载
wails build    # 构建生产版单文件 exe，输出到 build/bin/
```
