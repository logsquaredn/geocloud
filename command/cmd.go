package command

import "context"

type Cmd interface {
	ExecuteContext(context.Context) error
}
