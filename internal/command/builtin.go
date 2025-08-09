package command

import (
	"fmt"
	"strings"
	"time"

	"github.com/hongjun500/chat-go/internal/chat"
)

// RegisterBuiltins 注册内置命令
func RegisterBuiltins(r *Registry) (err error) {
	if err := r.Register(&Command{
		Name: "help",
		Help: "查看帮助",
		Handler: func(ctx *Context) error {
			list := r.List()
			lines := make([]string, 0, len(list))
			for _, c := range list {
				aliases := ""
				if len(c.Aliases) > 0 {
					aliases = " (别名: " + strings.Join(c.Aliases, ", ") + ")"
				}
				lines = append(lines, fmt.Sprintf("/%s - %s%s", c.Name, c.Help, aliases))
			}
			ctx.Client.Send(strings.Join(lines, "\n"))
			return nil
		},
		MinLevel: levelUser,
	}); err != nil {
		return err
	}

	if err := r.Register(&Command{
		Name: "quit",
		Help: "退出聊天室",
		Handler: func(ctx *Context) error {
			ctx.Client.Send("再见！")
			ctx.Hub.UnregisterClient(ctx.Client)
			return nil
		},
		MinLevel: levelUser,
	}); err != nil {
		return err
	}
	if err := r.Register(&Command{
		Name: "who",
		Help: "查看在线用户",
		Handler: func(ctx *Context) error {
			names := ctx.Hub.ListNames()
			ctx.Client.Send("在线用户：" + strings.Join(names, ","))
			return nil
		},
		MinLevel: levelUser,
	}); err != nil {
		return err
	}
	// 新增：系统通知
	if err := r.Register(&Command{
		Name: "notice",
		Help: "系统通知广播: /notice <level> <text>",
		Handler: func(ctx *Context) error {
			if len(ctx.Args) < 2 {
				return fmt.Errorf("用法: /notice <info|warn|error> <text>")
			}
			level := ctx.Args[0]
			text := strings.Join(ctx.Args[1:], " ")
			ctx.Hub.Emit(&chat.SystemNoticeEvent{When: time.Now(), Level: level, Content: text})
			return nil
		},
		MinLevel: levelUser,
	}); err != nil {
		return err
	}
	// 新增：心跳
	if err := r.Register(&Command{
		Name: "ping",
		Help: "发送心跳: /ping [detail]",
		Handler: func(ctx *Context) error {
			detail := strings.Join(ctx.Args, " ")
			ctx.Hub.Emit(&chat.HeartbeatEvent{When: time.Now(), FromID: ctx.Client.ID, Detail: detail})
			ctx.Client.Send("pong")
			return nil
		},
		MinLevel: levelUser,
	}); err != nil {
		return err
	}
	// 新增：文件传输事件（元数据）
	if err := r.Register(&Command{
		Name: "sendfile",
		Help: "发送文件: /sendfile <to|*> <name> <size> [mime]",
		Handler: func(ctx *Context) error {
			if len(ctx.Args) < 3 {
				return fmt.Errorf("用法: /sendfile <to|*> <name> <size> [mime]")
			}
			to := ctx.Args[0]
			name := ctx.Args[1]
			sizeStr := ctx.Args[2]
			var mime string
			if len(ctx.Args) >= 4 {
				mime = ctx.Args[3]
			}
			var size int64
			_, err := fmt.Sscan(sizeStr, &size)
			if err != nil {
				return fmt.Errorf("size 不是整数: %v", err)
			}
			ctx.Hub.Emit(&chat.FileTransferEvent{When: time.Now(), From: ctx.Client.Name, To: to, FileName: name, SizeBytes: size, MimeType: mime})
			ctx.Client.Send("文件事件已提交: " + name)
			return nil
		},
		MinLevel: levelUser,
	}); err != nil {
		return err
	}
	return nil

}
