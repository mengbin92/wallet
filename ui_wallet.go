package main

import (
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/mengbin92/wallet/internal/wallet"
)

func NewWalletTab(w fyne.Window) *container.TabItem {
	countEntry := widget.NewEntry()
	countEntry.SetPlaceHolder("数量")

	chainEntry := widget.NewEntry()
	chainEntry.SetText("bsc")

	outEntry := widget.NewEntry()
	outEntry.SetPlaceHolder("请选择存放keystore的目录...")
	browseBtn := NewFolderButton(outEntry, w)
	row := container.NewBorder(nil, nil, nil, browseBtn, outEntry)

	logBox := NewLogBox(10)
	lw := NewLogWriter(logBox)
	logger := log.New(lw, "[Wallet] ", log.LstdFlags)

	createBtn := widget.NewButton("生成钱包", func() {
		go func() {
			logger.Println("开始执行钱包生成...")
			count, err := strconv.Atoi(countEntry.Text)
			if err != nil {
				logger.Println("数量转换出错:", err.Error())
				return
			}
			if countEntry.Text == "" || chainEntry.Text == "" || outEntry.Text == "" {
				logger.Println("请填写完整参数")
				return
			}

			addresses, err := wallet.DeriveBatchEVM(chainEntry.Text, "", countEntry.Text, count)
			if err != nil {
				logger.Println("地址生成出错:", err.Error())
				return
			}
			logger.Println("地址生成完成")
			for _, address := range addresses {
				logger.Println("地址: " + address.Address)
			}

			err = wallet.SaveToKeystore(addresses, "", outEntry.Text)
			if err != nil {
				logger.Println("keystore存储出错:", err.Error())
				return
			}
			logger.Println("keystore存储完成")
		}()
	})

	content := container.NewVBox(
		widget.NewLabel("钱包创建"),
		widget.NewForm(
			widget.NewFormItem("数量", countEntry),
			widget.NewFormItem("底链", chainEntry),
			widget.NewFormItem("输出目录", row),
		),
		createBtn,
		logBox,
	)

	return container.NewTabItem("钱包创建", content)
}
