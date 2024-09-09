package command

import (
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// blockchainCmd 创建一个与区块链相关的命令
func (c *WalletCommand) blockchainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block",
		Short: "Blockchain related commands",
		Long:  "Blockchain related commands",
	}
	cmd.AddCommand(
		c.chainInfoCmd(),
		c.blockCountCmd(),
		c.blockHashCmd(),
		c.blockHeaderCmd(),
		c.blockCmd(),
	)
	return cmd
}

// blockCountCmd 创建一个获取当前区块数量的命令
func (c *WalletCommand) blockCountCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "getcount",
		Short: "Get the current block count",
		Long:  "Get the current block count, example: ./wallet block getcount",
		RunE: func(cmd *cobra.Command, args []string) error {
			bc, err := client.GetBlockCount()
			if err != nil {
				return errors.Wrap(err, "failed to get block count")
			}
			fmt.Printf("block count: %d\n", bc)
			return nil
		},
	}
}

// blockHashCmd 创建一个通过区块编号获取区块哈希的命令
func (c *WalletCommand) blockHashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gethash",
		Short: "Get the block hash by block number",
		Long:  "Get the block hash by block number, example: ./wallet block gethash 100",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var bc uint64
			if len(args) != 1 {
				bc, err = askOneNumber("Please enter the block number: ")
				if err != nil {
					return errors.Wrap(err, "failed to get block number")
				}
			} else {
				bc, err = strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return errors.Wrap(err, "failed to parse block number")
				}
			}

			hash, err := client.GetBlockHash(int64(bc))
			if err != nil {
				return errors.Wrap(err, "failed to get block hash")
			}
			fmt.Printf("block hash: %s\n", hash.String())
			return nil
		},
	}
}

// blockHeaderCmd 创建一个通过区块哈希获取区块头的命令
func (c *WalletCommand) blockHeaderCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "getheader",
		Short: "Get the block header by block hash",
		Long:  "Get the block header by block hash, example: ./wallet block getheader hash",
		RunE: func(cmd *cobra.Command, args []string) error {
			var hashStr string
			var err error
			if len(args) != 1 {
				hashStr, err = askOneString("Please enter the block hash: ")
				if err != nil {
					return errors.Wrap(err, "failed to get block hash")
				}
			} else {
				hashStr = args[0]
			}
			hash, err := chainhash.NewHashFromStr(hashStr)
			if err != nil {
				return errors.Wrap(err, "failed to parse block hash")
			}
			header, err := client.GetBlockHeader(hash)
			if err != nil {
				return errors.Wrap(err, "failed to get block header")
			}
			fmt.Println("block header:")
			fmt.Printf("\t version: %d\n", header.Version)
			fmt.Printf("\t prev block: %s\n", header.PrevBlock.String())
			fmt.Printf("\t merkle root: %s\n", header.MerkleRoot.String())
			fmt.Println("\t timestamp: ", header.Timestamp)
			fmt.Printf("\t bits: %d\n", header.Bits)
			fmt.Printf("\t nonce: %d\n", header.Nonce)
			return nil
		},
	}
}

// blockCmd 创建一个通过区块哈希获取区块的命令
func (c *WalletCommand) blockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "getblock",
		Short: "Get the block by block hash",
		Long:  "Get the block by block hash, example: ./wallet block getblock hash",
		RunE: func(cmd *cobra.Command, args []string) error {
			var hashStr string
			var err error
			if len(args) != 1 {
				hashStr, err = askOneString("Please enter the block hash: ")
				if err != nil {
					return errors.Wrap(err, "failed to get block hash")
				}
			} else {
				hashStr = args[0]
			}
			hash, err := chainhash.NewHashFromStr(hashStr)
			if err != nil {
				return errors.Wrap(err, "failed to parse block hash")
			}
			block, err := client.GetBlock(hash)
			if err != nil {
				return errors.Wrap(err, "failed to get block")
			}
			fmt.Printf("block: %s has %d transactions\n", args[0], len(block.Transactions))
			return nil
		},
	}
}

// chainInfoCmd 创建一个获取链信息的命令
func (c *WalletCommand) chainInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "chaininfo",
		Short: "Get the chain information",
		Long:  "Get the chain information, example: ./wallet block chaininfo",
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := client.GetBlockChainInfo()
			if err != nil {
				return errors.Wrap(err, "failed to get chain info")
			}
			fmt.Println("chain info:")
			fmt.Printf("\t chain: %s\n", info.Chain)
			fmt.Printf("\t blocks: %d\n", info.Blocks)
			fmt.Printf("\t headers: %d\n", info.Headers)
			fmt.Printf("\t best block hash: %s\n", info.BestBlockHash)
			fmt.Printf("\t difficulty: %f\n", info.Difficulty)
			return nil
		},
	}
}
