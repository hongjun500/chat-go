package command

import (
	"fmt"
	"strings"
)

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
	return nil

}
