# go-task-platform

异步任务执行平台，基于 Go + PostgreSQL。通过 HTTP 提交任务，多个 Worker 并发抢占执行，失败自动重试，Worker 崩溃后任务自动恢复。

**语义：at-least-once（至少一次）。** 设计参考 [River](https://github.com/riverqueue/river)。

## Why this?

因为我想找工作（划掉）。

因为异步是计算机系统中一个很重要的思想。我一直对如何在复杂、不确定的执行条件下的任务执行感兴趣。我感觉这是信息技术的“技术”两个字的来源之一。

最初也考虑过 SAGA。它会更复杂太多。

## 为什么不是exactly-once？

稍微思考过后我发现我好像做不到exactly-once。并且rabbitMQ等消息平台涉及外部副作用的时候也是停留在at-least-once这个程度的，所以我就没有继续深究了。

我的思考大概是这样的：哪怕你把颗粒度切碎，再如何增加执行器的反馈频率 ---- 执行和反馈始终是两步。存在执行完成但还没来得及反馈的时间点。

有点类似于两将军问题。我猜这是各类MQ平台停留在at-least-once的原因。


## 技术要点

### 1. 并发安全的任务抢占：`FOR UPDATE SKIP LOCKED`

`Claim` 用一条 SQL 完成"选中任务并标记为 running"，并发的 Worker 会跳过已被锁住的行，而不是排队等待，因此同一个任务不会被重复领取。已有并发测试：10 个 goroutine 同时抢一个任务，只有一个成功。

### 2. 租约：用 `attempt` 校验执行权

任务超过 `JobTimeout` 仍是 running，会被恢复流程重新放回队列，由另一个 Worker 接手。这时第一个 Worker 可能只是慢，稍后才返回结果。

每次 Claim 会让 `attempt` 加 1，它同时充当"租约号"。所有状态更新（完成 / 重试 / 丢弃）都带条件：

```sql
WHERE id = ? AND state = 'running' AND attempt = ?
```

过期 Worker 的结果更新不到任何行，直接被忽略（记一条日志），不会覆盖当前执行者的状态。

River 的做法是只检查任务仍处于 running；我在此基础上额外加了 `attempt` 校验，避免过期 Worker 误伤已被重新领取的任务。

**取舍：**

- `completed`、`discarded` 是终态，不再回改（River、Oban 同样如此）。
- 代价是在极端情况下，数据库记录与现实可能不一致：例如旧 Worker 超时后其实成功了，但新 Worker 随后失败到次数耗尽，记录为 `discarded`。这需要"任务卡过超时线 + 重试耗尽"同时发生，条件比较苛刻，且属于至少一次语义的固有代价——数据库只能记录系统看到的结果。

### 3. 重试与超时恢复

- 失败后按 `attempt^4` 秒退避重试，达到 `MaxAttempts` 后进入 `discarded`。
- 后台定时扫描长时间停留在 running 的任务（如 Worker 崩溃），按剩余次数重新入队或丢弃。

## 任务状态

`available → running → completed`

失败：`running → available`（退避重试）或 `discarded`（次数用尽）。超时恢复同理。



## 测试

需要一个独立的测试库 `go_task_platform_test`


覆盖：并发抢占、超时恢复、超时且次数耗尽时丢弃、租约校验。

## 技术栈

Go · PostgreSQL · GORM · net/http
