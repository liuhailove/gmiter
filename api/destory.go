package api

import "github.com/liuhailove/gmiter/spi"

func init() {
	// 默认销毁方法注册
	defaultDestroy()
}

// Destroy 当应用关闭是优先调用此方法，这样可以使容器在server上立刻下线，
// 这个方法主要是影响的动态路由，如果立刻移除应用的话，可以立刻修改此容器
// 对应的动态路由规则，以便最小量的使数据路由到不存在的IP上
func Destroy() {
	// 逐个销毁
	for _, destroyFunc := range spi.GetAllDestroyFunc() {
		destroyFunc.Destroy()
	}
}
