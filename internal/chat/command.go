package chat

import (
	"fmt"
	"strings"
)

// processCommand 处理以 "/" 开头的命令，返回 true 表示命令已处理，不应广播
func processCommand(h *Hub, c *Client, line string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return true
	}
	cmd := fields[0]
	switch cmd {
	case "/help":
		help := "/who  查看在线用户\n/help 查看帮助\n/quit 退出聊天室\n"
		c.Send(help)
	case "/who":
		names := h.ListNames()
		c.Send("在线用户：" + strings.Join(names, ", "))
	case "/quit":
		c.Send("再见!")
		h.UnregisterClient(c)
		c.Conn.Close()
	default:
		c.Send(fmt.Sprintf("未知命令: %s（输入 /help 查看帮助）", cmd))
	}
	return true
}

// ProcessCommandWrapper 对外的命令处理入口，返回是否已处理（不需要广播）
func ProcessCommandWrapper(h *Hub, c *Client, line string) bool {
	return processCommand(h, c, line)
}
