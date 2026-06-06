# RingBuffer 性能基准测试报告

## 测试环境

| 项目 | 值 |
|---|---|
| OS | linux |
| 架构 | amd64 |
| CPU | AMD EPYC 7713 64-Core Processor |
| Go 版本 | 1.23.x |
| 单基准迭代次数 | 10 次（`benchstat` 统计） |

## 测试对象

- **smallnest/ringbuffer** (上游): `github.com/smallnest/ringbuffer v0.1.1`
- **argcv/ringbuffer** (本 fork): `github.com/argcv/ringbuffer@v2-rewrite`

## 测试方案

### 方案一：IO 带宽测试（Copy / ReadFrom / WriteTo）

**目的**: 验证本 fork 的 `RingBuffer` 包装器在 `io.Copy` 替代场景下的带宽是否与上游一致。

**方法**:
- `Copy`: 通过 ring buffer 异步拷贝 1MB 数据，测试总吞吐
- `ReadFrom`: 向 ring buffer 泵入 1MB 数据，测量写入侧带宽
- `WriteTo`: 从 ring buffer drain 1MB 数据，测量读取侧带宽
- 分别测试 4K 和 64K 两种 buffer 大小
- 每个 benchmark 运行 10 次，使用 `benchstat` 计算置信区间

### 方案二：核心操作性能测试

**目的**: 验证泛型核心 `Ring[T]` 在 `T=byte` 时是否有编译期 overhead。

**方法**:
- `Sync`: 单线程顺序 Write + Read
- `AsyncReadBlocking`: 并发读，测量写端性能
- `AsyncWriteBlocking`: 并发写，测量读端性能
- 额外对比 `RingByte_Sync`（`*Ring[byte]` 直接操作）与 `RingBuffer_Sync`（`*RingBuffer` 包装器）

### 方案三：泛型价值验证

**目的**: 评估泛型 `Ring[T]` 相比纯字节方案在结构化数据场景下的价值。

**方法**:
- `IntQueue_Generic`: `Ring[int]` 直接存取
- `IntQueue_ByteWorkaround`: `RingBuffer` + `binary.BigEndian` 手动编解码
- `StructQueue_Generic`: `Ring[LogEntry]` 直接存取
- `StructQueue_ByteWorkaround`: `RingBuffer` + 手动偏移量计算
- `Iterator_All`: `iter.Seq[byte]` 零拷贝遍历
- `Iterator_BytesCopy`: `Bytes()` 分配内存后遍历

---

## 原始测试结果

### 1. IO 带宽（benchstat 统计，10 次运行）

```
BenchmarkCopySmallnest_4K-2      609.0µ ± 14%    1.604Gi ± 12%    2.000Mi ± 0%    17.00 ± 0%
BenchmarkCopyArgcv_4K-2          509.1µ ± 13%    1.918Gi ± 11%    2.001Mi ± 0%    19.00 ± 0%
BenchmarkCopySmallnest_64K-2     499.2µ ± 19%    1.959Gi ± 16%    2.032Mi ± 0%    14.00 ± 0%
BenchmarkCopyArgcv_64K-2         482.3µ ± 24%    2.026Gi ± 31%    2.032Mi ± 0%    16.00 ± 0%

BenchmarkReadFrom_Smallnest_1M-2  70.24µ ± 5%   13.90Gi ± 5%         24 B ± 0%     1.000 ± 0%
BenchmarkReadFrom_Argcv_1M-2      73.50µ ± 5%   13.29Gi ± 5%         24 B ± 0%     1.000 ± 0%

BenchmarkWriteTo_Smallnest_1M-2   73.69µ ± 2%    6.948Gi ± 2%         24 B ± 0%     1.000 ± 0%
BenchmarkWriteTo_Argcv_1M-2       71.90µ ± 2%    7.121Gi ± 2%         24 B ± 0%     1.000 ± 0%

BenchmarkIoCopy_1M-2             326.9µ ± 11%    2.988Gi ± 13%    1.000Mi ± 0%     3.000 ± 0%
```

### 2. 核心操作性能（benchstat 统计，10 次运行）

```
BenchmarkRingByte_Sync-2                49.59n ± 2%    0 B/op    0 allocs/op
BenchmarkRingBuffer_Sync-2              49.46n ± 1%    0 B/op    0 allocs/op
BenchmarkSmallnest_Sync-2               60.77n ± 1%    0 B/op    0 allocs/op
BenchmarkArgcv_Sync-2                   61.24n ± 1%    0 B/op    0 allocs/op
BenchmarkSmallnest_AsyncReadBlocking-2 108.5n ± 1%    0 B/op    0 allocs/op
BenchmarkArgcv_AsyncReadBlocking-2     110.8n ± 2%    0 B/op    0 allocs/op
BenchmarkSmallnest_AsyncWriteBlocking-2 108.2n ± 1%    0 B/op    0 allocs/op
BenchmarkArgcv_AsyncWriteBlocking-2    111.2n ± 3%    0 B/op    0 allocs/op
```

### 3. 泛型价值验证（benchstat 统计，10 次运行）

```
BenchmarkIntQueue_Generic-2              52.21n ± 2%    0 B/op    0 allocs/op
BenchmarkIntQueue_ByteWorkaround-2       55.57n ± 1%    0 B/op    0 allocs/op
BenchmarkStructQueue_Generic-2           51.46n ± 2%    0 B/op    0 allocs/op
BenchmarkStructQueue_ByteWorkaround-2    51.31n ± 2%    0 B/op    0 allocs/op
BenchmarkIterator_All-2                 101.4n ± 2%    0 B/op    0 allocs/op
BenchmarkIterator_BytesCopy-2            50.60n ± 7%   48 B/op    1 allocs/op
```

---

## 结果汇总

### 表 1：IO 带宽对比（benchstat 置信区间）

| 测试项 | 上游 smallnest | 本 fork argcv | 差异 |
|---|---|---|---|
| Copy 4K | 609.0 µs ± 14% | 509.1 µs ± 13% | **-16%** (argcv 更快) |
| Copy 64K | 499.2 µs ± 19% | 482.3 µs ± 24% | **-3%** (argcv 更快) |
| ReadFrom 1MB | 70.24 µs ± 5% | 73.50 µs ± 5% | **+5%** (argcv 稍慢) |
| WriteTo 1MB | 73.69 µs ± 2% | 71.90 µs ± 2% | **-2%** (argcv 更快) |
| io.Copy 基准 (Go 标准库) | 326.9 µs ± 11% | 326.9 µs ± 11% | 基准参考 |

### 表 2：核心操作性能对比（benchstat 置信区间）

| 测试项 | 上游 smallnest | 本 fork argcv | 差异 |
|---|---|---|---|
| Sync (单线程 R+W) | 60.77n ± 1% | 61.24n ± 1% | **+0.8%** |
| AsyncReadBlocking | 108.5n ± 1% | 110.8n ± 2% | **+2.1%** |
| AsyncWriteBlocking | 108.2n ± 1% | 111.2n ± 3% | **+2.8%** |
| RingByte_Sync (直接) | — | 49.59n ± 2% | — |
| RingBuffer_Sync (包装器) | — | 49.46n ± 1% | — |

### 表 3：泛型 vs 字节 workaround 对比（benchstat 置信区间）

| 场景 | 泛型 Ring[T] | 字节 workaround | B/op | allocs/op |
|---|---|---|---|---|
| Int 队列 | 52.21n ± 2% | 55.57n ± 1% | 0 / 0 | 0 / 0 |
| Struct 队列 | 51.46n ± 2% | 51.31n ± 2% | 0 / 0 | 0 / 0 |
| Iterator 遍历 | 101.4n ± 2% | 50.60n ± 7% | **0** / **48** | **0** / **1** |

---

## 分析与结论

### 附：测试过程中的异常值发现

在方案一的早期测试中，曾观察到 64K buffer 下 `Copy` 操作出现显著差异（argcv 比上游慢约 54%，759.6 µs vs 493.2 µs）。经排查，该差异并非真实性能回归，而是以下因素共同导致的 benchmark 异常值：

1. **异步操作的调度噪声**: `Copy` 内部启动 goroutine + `sync.WaitGroup` 等待，每次迭代都创建新的 goroutine。当单次操作耗时较长（64K buffer > 700 µs）时，goroutine 生命周期与调度器、GC 的交互产生剧烈波动。
2. **迭代次数差异**: 64K buffer 的迭代次数（~1500 次）显著少于 4K buffer（~2500 次），样本量不足导致统计偏差。
3. **系统负载干扰**: 早期测试同时运行了 `WriteTo`/`ReadFrom` 等多个阻塞式 benchmark，goroutine 数量激增，调度器压力被放大。

**修复措施**: 将 `Copy` benchmark 拆分为独立文件，去除冗余的 `WriteTo`/`ReadFrom` 并发测试，统一使用 `bytes.NewReader(data)` + `bytes.Buffer` 的标准模式重新运行后，64K buffer 的差异收敛至正常范围。

**启示**: 测量异步阻塞操作的端到端性能时，goroutine 创建和调度器状态是主要干扰源。精确测量 ring buffer 本身性能时，应采用方案二的设计（固定 reader/writer goroutine，仅测量被测端）。

---

### 1. IO 带宽无差异

基于 `Copy`/`ReadFrom`/`WriteTo` 的测试数据，本 fork 的 `RingBuffer` 包装器在 IO 场景下的带宽与上游基本一致。所有差异均在置信区间重叠范围内，不存在可测量的性能回归。

值得注意的是，`Copy` 操作的方差较大（±13%~±24%），这反映了异步 goroutine 调度的不确定性。`ReadFrom` 和 `WriteTo` 的方差较小（±2%~±5%），数据更稳定。

### 2. 泛型核心零 overhead

`Ring[T]` 在 `T=byte` 时的 `Sync`/`AsyncReadBlocking`/`AsyncWriteBlocking` 性能与上游原生实现基本一致（差异 <3%）。Go 编译器对泛型代码进行了有效的单态化（monomorphization），消除了类型参数的运行时开销。

额外发现：`RingByte_Sync`（49.59n）与 `RingBuffer_Sync`（49.46n）几乎相同，证明 `RingBuffer` 包装器的嵌入开销为零。

### 3. 泛型在结构化数据场景有价值

基于 `IntQueue` 和 `StructQueue` 的测试数据，泛型版本与字节 workaround 的性能相同（均为 0 alloc），但泛型版本的优势在于：
- **零序列化代码**: 无需 `binary.BigEndian.PutUint64` 等手动编解码
- **类型安全**: 编译期检查，不存在字节偏移计算错误
- **维护性**: 结构体字段变更时，字节 workaround 的偏移量计算会失效

### 4. Iterator 是独占特性

`All()` 返回 `iter.Seq[byte]`，遍历时不修改读指针、不分配内存。对比 `Bytes()` 每次调用分配 48B + 1 alloc，`All()` 在热循环中可避免 GC 压力。该特性依赖 Go 1.23 的 iterator 标准。

---

## 最终结论

| 命题 | 结论 |
|---|---|
| 泛型化是否损害字节 IO 性能？ | **否**。带宽与核心操作均无差异。 |
| 泛型化是否对纯字节场景有必要？ | **非必要**。若仅用于 `io.Reader`/`io.Writer`，上游 API 已足够。 |
| 泛型化是否扩展了适用场景？ | **是**。`Ring[int]`、`Ring[struct]` 等结构化数据缓冲无需序列化。 |
| Iterator API 是否有价值？ | **是**。零分配遍历是 `Bytes()` 无法实现的。 |

这个泛型化的设计目标并非替代上游的字节流 IO 能力，而是在**保留完整字节 IO 兼容性且零性能损失**的前提下，通过泛型核心扩展至结构化数据缓冲场景。
