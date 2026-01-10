# CLI 使用指南 - 并发归集功能

## 命令结构

```
wallet-tool collect [command] [flags]
```

### 可用命令

| 命令 | 说明 | 性能 |
|------|------|------|
| `run` | 串行归集 (原始方法) | 1x (基准) |
| `concurrent` | 并发归集 (推荐) | **5x-20x** ⚡ |
| `benchmark` | 性能对比测试 | - |
| `stats` | 查看历史统计 | 即将推出 |

---

## 1. 串行归集 (原始方法)

### 基本用法

```bash
wallet-tool collect run \
  --token 0x[Token合约地址] \
  --to 0x[目标地址] \
  --keystore keystore \
  --mainkey mainkey \
  --password [密码]
```

### 示例

```bash
wallet-tool collect run \
  --token 0x55d398326f99059ff775485246999027b3197955 \
  --to 0x6ee6Bb78166451E369CD6914190453C45A76ca6a \
  --keystore ./keystore \
  --mainkey ./mainkey \
  --password mypassword123
```

**输出**:
```
🔄 Starting serial token collection...
⚠️  For better performance with multiple accounts, use 'concurrent' command (5x-20x faster)
```

**适用场景**:
- 少量账户 (<5个)
- 调试和测试
- 需要逐个查看日志

---

## 2. 并发归集 (推荐) ⚡

### 基本用法

```bash
wallet-tool collect concurrent \
  --token 0x[Token合约地址] \
  --to 0x[目标地址] \
  --workers 10 \
  --config prod
```

### 配置选项

#### 2.1 使用默认配置

```bash
wallet-tool collect concurrent \
  --token 0x55d398326f99059ff775485246999027b3197955 \
  --to 0x6ee6Bb78166451E369CD6914190453C45A76ca6a
```

**默认配置**: 5个workers

#### 2.2 自定义Worker数量

```bash
# 10个workers (推荐用于中等规模)
wallet-tool collect concurrent \
  --token 0x... \
  --to 0x... \
  --workers 10

# 20个workers (大规模归集)
wallet-tool collect concurrent \
  --token 0x... \
  --to 0x... \
  --workers 20
```

#### 2.3 使用预设配置

```bash
# 生产配置 (10 workers, 5次重试)
wallet-tool collect concurrent \
  --token 0x... \
  --to 0x... \
  --config prod

# 测试配置 (2 workers, 快速确认)
wallet-tool collect concurrent \
  --token 0x... \
  --to 0x... \
  --config test
```

### 输出示例

```
╔════════════════════════════════════════════════════════════╗
║          Concurrent Token Collection Configuration          ║
╠════════════════════════════════════════════════════════════╣
║  RPC Endpoint:        https://bsc-dataseed.binance.org    ║
║  Max Workers:         10                                  ║
║  Max Retries:         5                                   ║
║  Token Confirmations: 6                                   ║
║  Max Wait Time:       1m0s                                ║
╚════════════════════════════════════════════════════════════╝

🚀 Starting concurrent token collection...
⚡ Expected speedup: 5x-10x with 10 workers

开始从 100 个账户归集代币 (并发数: 10)
发现余额 0xabc...: 100.5 tokens
发现余额 0xyz...: 50.2 tokens
...
找到 95 个有余额的账户

=== 归集统计 ===
总账户数: 95
成功: 93
失败: 2
总归集金额: 4752.5 tokens
成功率为: 97.89%
```

---

## 3. Worker数量选择指南

### 推荐配置

| 账户数 | 推荐Workers | 预期性能提升 | 命令示例 |
|--------|------------|-------------|----------|
| <10 | 5 | 5x | `--workers 5` 或 默认 |
| 10-50 | 10 | 10x ⭐ | `--workers 10` 或 `--config prod` |
| 50-100 | 10-20 | 10-20x | `--workers 15` |
| 100+ | 20 | 20x | `--workers 20 --config prod` |

### 性能对比 (100个账户)

| Workers | 预计耗时 | vs 串行 |
|---------|----------|---------|
| 1 (串行) | ~31分钟 | 1x |
| 5 | ~6.2分钟 | **5x** |
| 10 | ~3.1分钟 | **10x** ⭐ |
| 20 | ~1.6分钟 | **20x** ⚡ |

---

## 4. 配置预设详解

### Default (默认)

```bash
--config default
```

- **Workers**: 5
- **重试次数**: 3
- **确认数**: 6
- **最大等待**: 60秒

**适用**: 小规模归集，快速测试

---

### Production (生产)

```bash
--config prod
```

- **Workers**: 10
- **重试次数**: 5
- **确认数**: 6
- **最大等待**: 60秒

**适用**: 大规模生产环境归集

---

### Test (测试)

```bash
--config test
```

- **Workers**: 2
- **重试次数**: 3
- **确认数**: 6
- **最大等待**: 30秒

**适用**: 测试环境，快速验证

---

## 5. 实际使用场景

### 场景1: 日常归集 (50账户)

```bash
wallet-tool collect concurrent \
  --token 0x55d398326f99059ff775485246999027b3197955 \
  --to 0xYourWalletAddress \
  --keystore ./keystore \
  --mainkey ./mainkey \
  --password YourPassword \
  --config prod
```

**预计耗时**: ~1.5分钟 (串行需要15分钟)

---

### 场景2: 大规模归集 (500账户)

```bash
wallet-tool collect concurrent \
  --token 0x55d398326f99059ff775485246999027b3197955 \
  --to 0xYourWalletAddress \
  --keystore ./keystore \
  --mainkey ./mainkey \
  --password YourPassword \
  --workers 20 \
  --config prod
```

**预计耗时**: ~7.5分钟 (串行需要2.5小时)

---

### 场景3: 快速测试 (10账户)

```bash
wallet-tool collect concurrent \
  --token 0x55d398326f99059ff775485246999027b3197955 \
  --to 0xYourWalletAddress \
  --keystore ./test_keystore \
  --mainkey ./test_mainkey \
  --password TestPassword \
  --config test
```

**预计耗时**: ~30秒

---

## 6. 高级选项

### 自定义RPC节点

```bash
wallet-tool collect concurrent \
  --rpc https://your-rpc-node.com \
  --token 0x... \
  --to 0x... \
  --workers 10
```

### 启用详细统计

```bash
wallet-tool collect concurrent \
  --token 0x... \
  --to 0x... \
  --workers 10 \
  --stats
```

### 组合多个选项

```bash
wallet-tool collect concurrent \
  --rpc https://bsc-dataseed1.binance.org \
  --token 0x55d398326f99059ff775485246999027b3197955 \
  --to 0xYourAddress \
  --keystore ./keystore \
  --mainkey ./mainkey \
  --password YourPassword \
  --workers 15 \
  --config prod \
  --stats
```

---

## 7. 性能基准测试

### 运行基准测试

```bash
wallet-tool collect benchmark \
  --token 0x... \
  --to 0x... \
  --keystore ./keystore \
  --password YourPassword
```

**输出示例**:
```
🔬 Running performance benchmark...
This will run both serial and concurrent collection
WARNING: This will take extra time but provide useful performance data

=== Benchmark Results ===

Serial Collection:
  Time: 15 minutes 32 seconds
  Accounts: 50
  Success: 48/50 (96%)

Concurrent Collection (10 workers):
  Time: 1 minute 34 seconds
  Accounts: 50
  Success: 49/50 (98%)

Speedup: 9.9x ⚡
Time Saved: 13 minutes 58 seconds
```

---

## 8. 故障排查

### 问题1: RPC连接失败

**错误**: `failed to connect rpc`

**解决**:
```bash
# 尝试不同的RPC节点
--rpc https://bsc-dataseed1.binance.org
--rpc https://bsc-dataseed2.binance.org
--rpc https://bsc-dataseed3.binance.org
```

---

### 问题2: 并发过高被限流

**错误**: `too many requests` 或 `rate limit exceeded`

**解决**:
```bash
# 减少worker数量
--workers 5  # 从10降到5

# 或使用默认配置
--config default
```

---

### 问题3: 内存不足

**错误**: `out of memory`

**解决**:
```bash
# 减少并发数
--workers 5

# 分批处理
# 先处理一半，再处理另一半
```

---

## 9. 最佳实践

### ✅ DO (推荐)

1. **使用并发模式进行大规模归集**
   ```bash
   wallet-tool collect concurrent --workers 10 ...
   ```

2. **根据账户数选择合适的worker数**
   - <10账户: 5 workers
   - 10-50账户: 10 workers
   - 50+账户: 15-20 workers

3. **使用生产配置进行重要归集**
   ```bash
   --config prod
   ```

4. **先测试再大规模使用**
   ```bash
   --config test  # 先用测试配置验证
   ```

5. **保存归集日志**
   ```bash
   wallet-tool collect concurrent ... 2>&1 | tee collection.log
   ```

---

### ❌ DON'T (避免)

1. **不要在公共RPC上使用过多workers**
   - 公共RPC通常限制10并发
   - 超过会被限流

2. **不要在重要操作上首次使用大规模并发**
   - 先用小规模测试
   - 确认无误后再扩大规模

3. **不要忽略失败日志**
   - 失败的账户需要手动处理
   - 可能需要重试或调查原因

---

## 10. 快速参考

### 常用命令模板

```bash
# 快速归集 (默认配置)
wallet-tool collect concurrent --token 0x... --to 0x...

# 生产归集 (推荐)
wallet-tool collect concurrent --token 0x... --to 0x... --config prod

# 大规模归集
wallet-tool collect concurrent --token 0x... --to 0x... --workers 20 --config prod

# 测试验证
wallet-tool collect concurrent --token 0x... --to 0x... --config test

# 查看帮助
wallet-tool collect concurrent --help
```

---

## 11. 性能预期

| 场景 | 账户数 | 串行耗时 | 并发耗时(10w) | 提升 |
|------|--------|----------|--------------|------|
| 小规模 | 10 | 1.5分钟 | 9秒 | 10x |
| 中规模 | 50 | 15分钟 | 1.5分钟 | 10x |
| 大规模 | 100 | 31分钟 | 3.1分钟 | 10x |
| 超大 | 500 | 2.5小时 | 7.5分钟 | 20x |

---

**更多信息**:
- 性能测试报告: `PERFORMANCE_TEST_REPORT.md`
- 优化总结: `OPTIMIZATION_SUMMARY.md`
- 基准测试结果: `benchmark_results.txt`
