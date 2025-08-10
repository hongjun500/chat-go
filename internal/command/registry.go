package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/internal/observe"
)

type Level int

const (
	levelUser Level = iota
	levelAdmin
)

type Context struct {
	Hub    *chat.Hub
	Client *chat.Client
	Args   []string
	Raw    string
}

type HandlerFunc func(ctx *Context) error

type Command struct {
	Name     string
	Aliases  []string
	Help     string
	MinLevel Level
	Handler  HandlerFunc
}

type Registry struct {
	mu     sync.RWMutex
	byName map[string]*Command
	list   []*Command
}

func NewRegistry() *Registry {
	return &Registry{
		byName: make(map[string]*Command),
		list:   make([]*Command, 0),
	}
}

func (r *Registry) Register(cmd *Command) (err error) {
	if cmd == nil {
		return errors.New("command is nil")
	}
	name := strings.ToLower(strings.TrimSpace(cmd.Name))
	if name == "" {
		return errors.New("command name is empty")
	}
	if strings.Contains(name, "/") {
		return fmt.Errorf("command name must not contain '/':%s", name)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byName[name]; exists {
		return fmt.Errorf("command %s already registered", name)
	}
	r.byName[name] = cmd
	for _, item := range cmd.Aliases {
		alias := strings.ToLower(strings.TrimSpace(item))
		if alias == "" {
			continue
		}
		if _, exists := r.byName[alias]; exists {
			return fmt.Errorf("command alias %s already registered", alias)
		}
		r.byName[alias] = cmd
	}
	r.list = append(r.list, cmd)
	return nil
}

func (r *Registry) Get(name string) (*Command, bool) {
	k := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(name), "/"))
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmd, ok := r.byName[k]
	return cmd, ok
}

func (r *Registry) List() []*Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Command, len(r.list))
	copy(out, r.list)
	return out
}

func (r *Registry) Execute(raw string, ctx *Context) (handled bool, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasPrefix(raw, "/") {
		return false, nil
	}
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return true, nil
	}
	cmdName := strings.TrimPrefix(parts[0], "/")
	cmd, ok := r.Get(cmdName)
	if !ok {
		observe.IncCommandError("not_found")
		return true, fmt.Errorf("command %s not found", cmdName)
	}

	if !r.checkPermission(ctx.Client, cmd.MinLevel) {
		observe.IncCommandError("permission")
		return true, errors.New("permission denied")
	}
	ctx.Args = parts[1:]
	observe.IncCommand(cmd.Name)
	if err := cmd.Handler(ctx); err != nil {
		observe.IncCommandError("handler")
		return true, err
	}
	return true, nil

}

func (r *Registry) checkPermission(c *chat.Client, need Level) bool {
	if c == nil {
		return true
	}
	// 用户级命令默认允许
	if need <= levelUser {
		return true
	}
	if c.Meta == nil {
		return false
	}
	if s, ok := c.Meta["level"]; ok && s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			return Level(v) >= need
		}
	}
	return false
}
