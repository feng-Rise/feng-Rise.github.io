package session

import (
	"context"
	"net/http"
)

// 传播者
type Propagator interface {
	Inject(id string, writer http.ResponseWriter) error
	Extract(req *http.Request) (string, error)
	// Remove 将 session id 从 http.ResponseWriter 中删除
	// 例如删除对应的 cookie
	Remove(writer http.ResponseWriter) error
}

// 这个接口使用seesion中的key获取value
// 可以用不同的对象实现，redis存储，内存存储等
type Session interface {
	GET(ctx context.Context, key string) (string, error)
	SET(ctx context.Context, key, value string) error
	ID() string
}

//这个接口是用来管理seesion

type Store interface {
	Generate(ctx context.Context, id string) (Session, error)
	// Refresh 这种设计是一直用同一个 id 的
	// 如果想支持 Refresh 换 ID，那么可以重新生成一个，并移除原有的
	// 又或者 Refresh(ctx context.Context, id string) (Session, error)
	// 其中返回的是一个新的 Session
	Refresh(ctx context.Context, id string) error
	Remove(ctx context.Context, id string) error
	GET(ctx context.Context, id string) (Session, error)
}
