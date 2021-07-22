package router

type RouterOpt func(r *Router)

func WithGroup(group string) RouterOpt {
	return func(r *Router) {
		r.group = group
	}
}
