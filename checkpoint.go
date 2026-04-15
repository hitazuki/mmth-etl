package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// loadCheckpoint 加载上次记录的断点
func (p *LogProcessor) loadCheckpoint() {
	if _, err := os.Stat(p.stateFilePath); os.IsNotExist(err) {
		// 文件不存在，使用空字符串表示从头开始
		p.checkpoint = ""
		return
	}

	data, err := os.ReadFile(p.stateFilePath)
	if err != nil {
		log.Fatalf("读取状态文件失败: %v", err)
	}

	var state struct {
		Checkpoint string `json:"checkpoint"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		log.Fatalf("解析状态文件失败: %v", err)
	}

	p.checkpoint = state.Checkpoint
	if p.checkpoint != "" {
		fmt.Printf("从 %s 开始处理日志\n", p.checkpoint)
	} else {
		fmt.Println("从头开始处理日志")
	}
}

// saveCheckpoint 保存当前断点
func (p *LogProcessor) saveCheckpoint() {
	state := struct {
		Checkpoint string `json:"checkpoint"`
	}{
		Checkpoint: p.checkpoint,
	}

	data, err := json.Marshal(state)
	if err != nil {
		log.Fatalf("序列化状态失败: %v", err)
	}

	if err := os.WriteFile(p.stateFilePath, data, 0644); err != nil {
		log.Fatalf("写入状态文件失败: %v", err)
	}
}
