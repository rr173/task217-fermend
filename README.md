# task217-fermend 发酵罐代谢终点漂移诊断服务

面向生物工艺工程师的纯后端诊断服务：接收发酵批次的多通道时序（溶氧、pH、补料、代谢物等）与工艺阶段标记，校正传感器滞后、识别代谢拐点、比较预测终点与离线采样，最终发布可追溯的偏差诊断快照。

## 业务闭环

1. 登记发酵批次，推进状态机 `preparing → running → pending_diagnosis → confirmed → sealed`。
2. 登记传感器通道（按物理量类型校验单位），写入多通道采样曲线（按设备序号幂等）。
3. 标记工艺阶段时间窗（拒绝阶段逆序）。
4. 标记传感器滞后/缺口段、估计滞后并应用校正（滞后段的时间轴平移）。
5. 诊断代谢终点：识别溶氧回升点、pH 拐点、补料拐点，多通道取中位数综合。
6. 用离线采样终点比对综合预测，判定 confirmed / conflict。
7. 冻结诊断快照（draft → published → superseded），封存快照不被覆盖。

## 核心状态机

- 发酵批次：准备 / 运行中 / 待诊断 / 已确认 / 封存
- 传感器段：待校准 / 有效 / 滞后 / 缺口 / 剔除
- 终点候选：预测 / 冲突 / 确认 / 否决
- 诊断快照：草稿 / 发布 / 替代

## 持久化与重启恢复

SQLite（`modernc.org/sqlite`，WAL + busy_timeout + 单写连接）。批次、通道、阶段、采样点、传感器段、终点候选、诊断快照共 7 张表全量持久化。`--smoke-test` 会关闭并重开同一数据库验证重启恢复（批次状态、已发布快照全量还原）。

## 标准命令

```bash
# 构建 / 静态检查 / 测试
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...

# 端到端自检（Docker 判据）
go run ./cmd/task217-fermend --smoke-test

# 启动 HTTP 服务
go run ./cmd/task217-fermend --addr :8080 --db fermend.db
```

## API 入口（前缀 /api）

| 能力 | 入口 |
| --- | --- |
| 批次登记/列表/详情/推进/封存 | POST `/api/batches`、GET `/api/batches`、GET `/api/batches/{id}`、POST `/api/batches/{id}/transition`、POST `/api/batches/{id}/seal` |
| 通道登记/列表/剔除 | POST `/api/batches/{id}/channels`、GET `/api/batches/{id}/channels`、POST `/api/channels/{id}/exclude` |
| 阶段标记/列表 | POST `/api/batches/{id}/stages`、GET `/api/batches/{id}/stages` |
| 曲线写入/查询 | POST `/api/batches/{id}/channels/{channelID}/samples`、GET `/api/channels/{id}/samples` |
| 段标记/列表/估计滞后/应用/剔除 | POST `/api/channels/{id}/segments`、GET `/api/channels/{id}/segments`、POST `/api/channels/{id}/estimate-lag`、POST `/api/segments/{id}/apply`、POST `/api/segments/{id}/exclude` |
| 终点诊断/列表/采样比对/确认/否决 | POST `/api/batches/{id}/diagnose`、GET `/api/batches/{id}/endpoints`、POST `/api/batches/{id}/resolve`、POST `/api/endpoints/{id}/confirm`、POST `/api/endpoints/{id}/veto` |
| 快照创建/列表/发布/替代 | POST `/api/batches/{id}/snapshots`、GET `/api/batches/{id}/snapshots`、POST `/api/snapshots/{id}/publish`、POST `/api/snapshots/{id}/supersede` |
| 健康/统计 | GET `/api/health`、GET `/api/stats` |

## 模块责任

| 模块 | 责任 |
| --- | --- |
| `internal/sampling` | 采样接收：通道登记、曲线写入、单位校验、幂等去重 |
| `internal/stage` | 阶段时间窗维护，拒绝逆序 |
| `internal/calibration` | 传感器滞后估计与校正，段标记、通道剔除 |
| `internal/diagnosis` | 拐点识别（溶氧谷/pH 拐点/补料拐点）、终点推断与采样比对 |
| `internal/snapshot` | 诊断快照冻结、发布与替代 |
| `internal/service` | 跨模块用例编排 |
| `internal/httpapi` | JSON HTTP 层，路由前缀 /api |
