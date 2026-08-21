# Bug Reproduction

## 包的性质

当前 test_model_fix 保存的是被测模型修复后的结果源码，不是初始含 Bug 源码。要复现原始缺陷，必须检出下面固定的 parent SHA；不要在当前修复结果源码上期待重新出现修复前失败。生成系统使用的可信验证补丁和完整验证日志仅在本地留存，不提交到结果分支。

## 问题现象

航班动态推送给两个值班席位后，第一个席位为了本地高亮，修改收到的优先级、经停节点或载荷；第二个席位随后读到的同一条消息也跟着改变，稍后断线重连拿到的最后事件同样被污染。即使发布方在 Publish 返回后才改原始对象，重放内容也会被带偏。请修复事件交付的隔离，确保每个订阅者和保留副本都有独立数据，同时维持租户过滤与正常重放。

## 含 Bug 版本

- 仓库：VanceMichael/go-label-11
- 仓库地址：https://github.com/VanceMichael/go-label-11.git
- parent SHA：6d0885785579d31b45b5ba31a378d65bcf51d9e3

## 复现步骤

```bash
git clone -- https://github.com/VanceMichael/go-label-11.git bug-repro
cd bug-repro
git checkout --detach 6d0885785579d31b45b5ba31a378d65bcf51d9e3
go test ./internal/stream -run ^TestSubscriberMutationDoesNotPolluteFanoutOrReplay$ -count=1
```

## 双架构完整错误信息

### linux/amd64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/stream -run ^TestSubscriberMutationDoesNotPolluteFanoutOrReplay$ -count=1
--- FAIL: TestSubscriberMutationDoesNotPolluteFanoutOrReplay (0.00s)
    replay_test.go:54: second console priority = "locally-highlighted", want normal
FAIL
FAIL	github.com/VanceMichael/go-base-airbridge/internal/stream	0.032s
FAIL

```

stderr：

```text
(empty)
```

### linux/arm64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/stream -run ^TestSubscriberMutationDoesNotPolluteFanoutOrReplay$ -count=1
--- FAIL: TestSubscriberMutationDoesNotPolluteFanoutOrReplay (0.00s)
    replay_test.go:54: second console priority = "locally-highlighted", want normal
FAIL
FAIL	github.com/VanceMichael/go-base-airbridge/internal/stream	0.002s
FAIL

```

stderr：

```text
(empty)
```

## 通过条件

同一航班事件送达多个值班席位后，任一席位对优先级头、经停节点或载荷所做的本地修改都不能改变其他席位已收到的数据，也不能污染断线重连读取的最近事件；发布方在 Publish 返回后的修改同样应与总线内状态隔离，既有租户筛选和合法重放保持可用。TestSubscriberMutationDoesNotPolluteFanoutOrReplay 必须从失败转为通过，stream 与 event 相关测试、全仓 go test ./... 和 go build ./... 均无回归，不得删改断言或跳过目标订阅路径。
