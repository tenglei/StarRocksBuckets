# StarRocks连接重构测试指南

## 测试环境准备

### 1. 使用Docker部署StarRocks

```bash
# 拉取StarRocks官方镜像
docker pull starrocks/allin1-ubuntu:latest

# 启动容器（端口9030用于MySQL协议连接）
docker run -d --name starrocks-test -p 9030:9030 -p 8030:8030 starrocks/allin1-ubuntu:latest

# 等待容器启动（约30秒）
docker logs -f starrocks-test

# 验证StarRocks是否启动成功
docker exec -it starrocks-test mysql -h 127.0.0.1 -P 9030 -u root
```

### 2. 编译项目

```bash
# Windows环境
cd C:\Users\t00469468\.qoder\worktree\StarRocksBuckets\qoder\starrocks-connection-refactor-1766404366
go build -o setbuckets.exe

# Linux/Mac环境
go build -o setbuckets
```

## 测试用例

### 测试用例1：基本连接测试

**目标：** 验证使用默认配置成功连接StarRocks

**步骤：**
1. 运行程序：`setbuckets.exe -t test.demo_table`
2. 按提示输入连接参数：
   - FE地址：直接回车（使用默认localhost）
   - FE端口：直接回车（使用默认9030）
   - 用户名：直接回车（使用默认root）
   - 密码：直接回车（空密码）
3. 输入`y`确认配置

**预期结果：**
- 显示"✓ 连接成功！"
- 程序继续执行后续业务逻辑

### 测试用例2：密码隐藏验证

**目标：** 验证密码输入时不显示明文

**步骤：**
1. 运行程序
2. 到达密码输入提示时
3. 输入任意字符（例如：test123）
4. 观察终端显示

**预期结果：**
- 输入时终端不显示任意字符
- 回车后显示"密码: ******"

### 测试用例3：参数验证测试

**目标：** 验证非法参数的拦截

#### 3.1 端口号超出范围
- 输入端口：99999
- 预期：提示"端口号必须在1-65535范围内"，要求重新输入

#### 3.2 端口号非数字
- 输入端口：abc
- 预期：提示"端口号必须为有效数字"，要求重新输入

#### 3.3 主机地址为空
- FE地址留空且没有默认值
- 预期：提示"主机地址不能为空"，要求重新输入

### 测试用例4：连接失败处理

**目标：** 验证连接失败时的错误处理

#### 4.1 错误的主机地址
- 输入：192.168.999.999
- 预期：显示网络连接失败提示和检查建议

#### 4.2 错误的端口
- 输入：3306
- 预期：显示无法连接提示

#### 4.3 错误的密码（如果StarRocks配置了密码）
- 输入错误密码
- 预期：显示"用户名或密码错误"提示

### 测试用例5：重新输入测试

**目标：** 验证配置确认后可以重新输入

**步骤：**
1. 运行程序并输入所有连接参数
2. 在确认提示时输入`n`
3. 观察是否重新开始输入流程

**预期结果：**
- 显示"重新输入配置..."
- 重新显示连接配置输入界面

## 性能测试

### 连接时间测试

**目标：** 验证连接建立时间在可接受范围内

**方法：** 记录从确认配置到显示"✓ 连接成功！"的时间

**预期：** 小于3秒（本地环境）

## 兼容性测试

### 平台兼容性

- [ ] Windows 10/11
- [ ] Linux (Ubuntu 20.04+)
- [ ] macOS

### 终端兼容性

- [ ] PowerShell
- [ ] Windows Terminal
- [ ] CMD
- [ ] Bash
- [ ] Zsh

## 回归测试清单

确保原有功能不受影响：

- [ ] 表分桶查看功能正常
- [ ] 排序键查看功能正常
- [ ] 分桶数修改功能正常
- [ ] 空分区清理功能正常
- [ ] 分区级别分桶调整功能正常
- [ ] 日志输出格式正常
- [ ] 错误处理机制正常

## 测试报告模板

```
测试日期：____年__月__日
测试环境：[Windows/Linux/macOS]
测试人员：______

| 测试用例 | 测试结果 | 备注 |
|---------|---------|------|
| 基本连接测试 | ✓ PASS / ✗ FAIL |  |
| 密码隐藏验证 | ✓ PASS / ✗ FAIL |  |
| 参数验证测试 | ✓ PASS / ✗ FAIL |  |
| 连接失败处理 | ✓ PASS / ✗ FAIL |  |
| 重新输入测试 | ✓ PASS / ✗ FAIL |  |
| 性能测试 | ✓ PASS / ✗ FAIL |  |

发现的问题：
1. 
2. 

改进建议：
1. 
2. 
```

## 快速验证命令

```bash
# Windows环境快速测试
echo localhost | echo 9030 | echo root | echo | echo y | setbuckets.exe -t test.table

# 注意：上述命令在PowerShell中可能不工作，建议手动交互测试
```

## 故障排除

### 问题1：无法连接StarRocks

**解决方案：**
1. 检查Docker容器是否运行：`docker ps`
2. 检查端口映射是否正确：`docker port starrocks-test`
3. 使用MySQL客户端测试连接：`mysql -h localhost -P 9030 -u root`

### 问题2：密码不隐藏

**解决方案：**
1. 检查是否为TTY终端环境
2. 查看终端日志，应有"警告: 非TTY终端环境"提示
3. 在标准终端（非IDE集成终端）中测试

### 问题3：编译失败

**解决方案：**
1. 确保Go版本 >= 1.23
2. 运行 `go mod tidy` 更新依赖
3. 检查golang.org/x/term是否正确安装

## Docker测试环境清理

```bash
# 停止并删除测试容器
docker stop starrocks-test
docker rm starrocks-test

# 删除镜像（可选）
docker rmi starrocks/allin1-ubuntu:latest
```
