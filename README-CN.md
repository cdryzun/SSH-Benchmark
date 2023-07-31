# SSH基准测试工具

> 这是一款支持离线和批处理执行的服务器基准测试工具。

---

## 说明

当前工具的实现流程描述：

1. 检查主机配置是否正确。
2. 检查主机是否可以连通（连通性测试）。
3. 启动Iperf3服务器端。
4. 将压力测试脚本上传到目标服务器并根据主机列表批量执行脚本（串行执行）。
5. 实际并显示结果。
6. 清理脚本数据（完成）。

---

## 工具打包和编译。

```bash
go mod tidy
go build -o ssh-benchmark main.go
```

如果您已安装了 [task](https://taskfile.dev/)，也可以使用它。

```bash
task build:binary
```

---

## 使用说明

1. 修改 `config.json` 中的配置

   > 1. 如果您没有Linux版Geekbench 5的许可证，可以通过电子邮件联系我（<yangzun@treesir.pub>）。我愿意与您分享。
   > 2. 如果您希望对多个节点同时进行测试，可以在 `Host` 字段中填入逗号分隔的列表。

2. 执行此项目预编译的 ssh-benchmark 文件

   ```bash
   ./ssh-benchmark
   ```

   ![2023-07-31 18.15.23 2](images/README/2023-07-31%2018.15.23%202.gif)
