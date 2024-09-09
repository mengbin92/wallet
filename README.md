# wallet

## Usage

### 环境变量依赖

`wallet` 跟BTC网络使用RPC接口，需要设置环境变量来指定相应的RPC配置，如果没有指定则使用默认的配置。  

```go
// 默认环境变量配置
var (
    rpc_user     = "rpcusertest"
	rpc_password = "sjVj'rLmng;E>5)"
	rpc_endpoint = "127.0.0.1:8334"
	rpc_cert     = "./config/rpc.cert"
)
```  

配置环境变量：  

```bash
export RPC_USER=rpcusertest 
export RPC_PASSWORD="sjVj'rLmng;E>5)" 
export RPC_ENDPOINT=127.0.0.1:8334 
export RPC_CERT=./cmd/config/rpc.cert 
```  

### 编译

通过 `make` 命令可以编译 `wallet` 程序：  

```bash
# 下载项目  
$ git clone -b btc https://github.com/mengbin92/wallet.git
正克隆到 'wallet'...
remote: Enumerating objects: 213, done.
remote: Counting objects: 100% (213/213), done.
remote: Compressing objects: 100% (150/150), done.
remote: Total 213 (delta 101), reused 161 (delta 53), pack-reused 0 (from 0)
接收对象中: 100% (213/213), 55.89 KiB | 213.00 KiB/s, 完成.
处理 delta 中: 100% (101/101), 完成.  

# 编译  
$ cd wallet
$ make build

# 查看支持的命令
$ ./wallet --help
Command line BTC Wallet

Usage:
  wallet [command]

Available Commands:
  address     Manage btc address
  balance     Get the balance of the wallet
  block       Blockchain related commands
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  key         Manage key
  mnemonic    Manage mnemonic
  tx          Transaction operations
  version     Show version

Flags:
  -h, --help   help for wallet

Use "wallet [command] --help" for more information about a command.

# 查看版本信息
$ ./wallet version
wallet version: 0.0.1
```  

## 操作示例  

### 创建助记词  

```bash
./wallet mnemonic create ./key.key password
Create a new mnemonic
Your new mnemonic is:  help view super fabric media dad trust solid image behind flag acquire quantum clump ice nut cross outer dad model swear lab eye eternal
Your mnemonic is saved to file with password:  password
```  

### 创建新的地址  

```bash
# 创建私钥
# 当前版本生成的私钥加密后会与助记词存储在同一文件中，所以这里指定了文件路径为 `./key.key`
./wallet key create
key create
? Please input the file path of the key:  ./key.key
? Please input the password of the key:  ********
? 请选择网络类型： testnet
? Please input the account number:  0
? Please input the address index:  0
wif:  KzY7R7Uiqr5YufQS3qaKXvg75vQYbQJxrEJS9fJc34XHCGBuqvLN
key created successfully

# 创建新的地址
$ ./wallet address new
new btc address from wif key
? Please input wif key: KzY7R7Uiqr5YufQS3qaKXvg75vQYbQJxrEJS9fJc34XHCGBuqvLN
? 请选择网络类型： testnet
address:  tb1qe6wugez4ww4tz6fmjhda28r84l2kprd2f7ut4e

# 查看钱包中所有的地址
$ ./wallet address list
address list
? Please input key file path ./key.key
? Please input key file path ********
? 请选择网络类型： testnet
key:  cQu6t2UaGump56shSFPSuFBAi9hxFrQevGSuG5m7YBBHT1JTVSRV
address:  tb1qe6wugez4ww4tz6fmjhda28r84l2kprd2f7ut4e
```  

### 查询地址余额

```bash
$ ./wallet balance get
get balance
? 请选择网络类型： testnet
? Please input the address: tb1qndsh2mllf8g2hf29svazpxksa3ns4zga3n55mc
Address: tb1qndsh2mllf8g2hf29svazpxksa3ns4zga3n55mc Balance: 0.979885350 BTC
```  

### 查询区块链相关信息  

```bash
# 查询最新区块高度
$ ./wallet block getcount
block count: 2903999

# 查询区块链信息
$ ./wallet block chaininfo
chain info:
	 chain: testnet3
	 blocks: 2904022
	 headers: 2904022
	 best block hash: 0000000000000017e2bd22bf92453f82808203ebf3ce071e5a41566a4cc1c79b
	 difficulty: 139163970.894327

# 查询指定区块的hash
$ ./wallet block gethash 2903999
block hash: 00000000177d73a1001296640d8a9fbe3c5ca4fe76fbbb1b39cb5648dd769141 

# 查询指定区块的块头的信息
$ ./wallet block getheader 00000000177d73a1001296640d8a9fbe3c5ca4fe76fbbb1b39cb5648dd769141
block header:
	 version: 536870912
	 prev block: 000000008ec366eb6d3464dd14baca23cb386896d3d5cd28398a458a5a8f8eb6
	 merkle root: 620ae3c6a6bc1e2efc15de61d4bf0aeefa703b0d8d2da4c43e1d68e5a334b65a
	 timestamp:  2024-09-09 02:44:18 +0800 CST
	 bits: 486604799
	 nonce: 9530446

# 查询指定区块信息，目前仅展示指定区块中包含了多少笔交易
$ ./wallet block getblock 00000000177d73a1001296640d8a9fbe3c5ca4fe76fbbb1b39cb5648dd769141
block: 00000000177d73a1001296640d8a9fbe3c5ca4fe76fbbb1b39cb5648dd769141 has 3532 transactions
```

### 交易相关操作  

```bash
# 发送交易
$ ./wallet tx send
send btc
? Please input the file path of the key:  ./key.key
? Please input the password of the key:  ********
? 请选择网络类型： testnet
? Please input the sender address:  tb1qe6wugez4ww4tz6fmjhda28r84l2kprd2f7ut4e
? Please input the receiver address:  tb1qndsh2mllf8g2hf29svazpxksa3ns4zga3n55mc
? Please input the amount of bitcoins to send:  100000
txHash: 6fff869af50bc3b498451abec0d6bb29d1188542cc69cd952fa958107ff4c9b0
```  