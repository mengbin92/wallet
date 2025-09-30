package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/mengbin92/wallet/internal/collector"
)

func NewCollectTab(w fyne.Window) *container.TabItem {
	rpcEntry := widget.NewEntry()
	rpcEntry.SetText("http://127.0.0.1:8545")

	mainKeyEntry := widget.NewEntry()
	mainKeyEntry.SetPlaceHolder("请输入主账户的目录...")

	ksEntry := widget.NewEntry()
	ksEntry.SetPlaceHolder("请选择存放keystore的目录...")
	browseBtn := NewFolderButton(ksEntry, w)
	row := container.NewBorder(nil, nil, nil, browseBtn, ksEntry)

	tokenEntry := widget.NewEntry()
	toEntry := widget.NewEntry()

	scroll, rt := NewLogBox(150)
	lw := NewLogWriter(rt)
	logger := log.New(lw, "[Collect] ", log.LstdFlags)

	collectBtn := widget.NewButton("开始归集", func() {
		go func() {
			logger.Println("开始执行归集...")
			if rpcEntry.Text == "" || mainKeyEntry.Text == "" || ksEntry.Text == "" || tokenEntry.Text == "" || toEntry.Text == "" {
				logger.Println("请填写完整参数")
				return
			}
			err := collector.CollectTokens(rpcEntry.Text, mainKeyEntry.Text, ksEntry.Text, "", tokenEntry.Text, toEntry.Text, logger)
			if err != nil {
				logger.Println("代币归集出错:", err.Error())
				return
			}
			logger.Println("代币归集完成")
		}()
	})

	content := container.NewVBox(
		widget.NewLabel("代币归集"),
		widget.NewForm(
			widget.NewFormItem("RPC", rpcEntry),
			widget.NewFormItem("主私钥", mainKeyEntry),
			widget.NewFormItem("Keystore", row),
			widget.NewFormItem("Token合约地址", tokenEntry),
			widget.NewFormItem("接收地址", toEntry),
		),
		collectBtn,
		scroll,
	)

	return container.NewTabItem("代币归集", content)
}
