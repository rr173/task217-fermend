基于 Go 实现的发酵罐代谢终点漂移诊断服务，一款纯后端生物工艺分析服务，处理多通道时序校正、代谢拐点识别与偏差诊断快照发布。

# task217-fermend 评测说明

## 项目定位

发酵罐代谢终点漂移诊断服务：接收多通道时序（溶氧/pH/补料/代谢物）与工艺阶段标记，校正传感器滞后、识别代谢拐点、比较预测终点与离线采样，发布可追溯的诊断快照。

## 启动与命令

```bash
# 端到端自检（退出码 0 为通过，Docker 判据）
go run ./cmd/task217-fermend --smoke-test

# 启动 HTTP 服务
go run ./cmd/task217-fermend --addr :8080 --db fermend.db
```

## --smoke-test 契约

1. 登记批次并推进到 running；
2. 登记溶氧/pH/补料三通道（单位校验）；
3. 标记 lag/log/stationary 三阶段（拒绝逆序）；
4. 写入三通道采样曲线（溶氧谷 t=500、pH 拐点 t=530、补料拐点 t=500）；
5. 首次诊断，pH 拐点记录为 530；
6. 标记 pH 通道滞后段（lag=30s）并应用校正；
7. 重算，pH 拐点前移到 500；
8. 以离线采样终点 500 比对综合预测，判定 confirmed（偏差 0）；
9. 发布诊断快照（草稿→发布），批次推进到 confirmed；
10. 关闭并重开同一数据库，验证批次状态与已发布快照恢复。

全部通过输出 `SMOKE TEST PASSED` 并返回 0；任一断言失败返回非 0。

## API 概览（前缀 /api）

- 批次：`POST /api/batches`、`GET /api/batches`、`GET /api/batches/{id}`、`POST /api/batches/{id}/transition`、`POST /api/batches/{id}/seal`
- 通道：`POST /api/batches/{id}/channels`、`GET /api/batches/{id}/channels`、`POST /api/channels/{id}/exclude`
- 阶段：`POST /api/batches/{id}/stages`、`GET /api/batches/{id}/stages`
- 采样：`POST /api/batches/{id}/channels/{channelID}/samples`、`GET /api/channels/{id}/samples`
- 校正：`POST /api/channels/{id}/segments`、`GET /api/channels/{id}/segments`、`POST /api/channels/{id}/estimate-lag`、`POST /api/segments/{id}/apply`、`POST /api/segments/{id}/exclude`
- 诊断：`POST /api/batches/{id}/diagnose`、`GET /api/batches/{id}/endpoints`、`POST /api/batches/{id}/resolve`、`POST /api/endpoints/{id}/confirm`、`POST /api/endpoints/{id}/veto`
- 快照：`POST /api/batches/{id}/snapshots`、`GET /api/batches/{id}/snapshots`、`POST /api/snapshots/{id}/publish`、`POST /api/snapshots/{id}/supersede`
- 自检：`GET /api/health`、`GET /api/stats`

## Docker 双架构

```bash
# 单架构构建
bash build_benzhi_docker.sh my-project linux/amd64
bash build_benzhi_docker.sh my-project linux/arm64

# 运行自检
docker run --rm my-project --smoke-test
```

Dockerfile 与 benzhi.Dockerfile 同源，`ENTRYPOINT ["/app/fermend"]` + `CMD ["--smoke-test"]`，镜像默认执行自检。

## 持久化

SQLite（`modernc.org/sqlite`，WAL），7 张表：batches / channels / stages / samples / segments / endpoints / snapshots。重启重开同一数据库全量恢复。
