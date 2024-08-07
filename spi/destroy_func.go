package spi

// DestroyFunc 关闭Func
type DestroyFunc interface {
	// Destroy 销毁方法
	Destroy()
}

var (
	// 销毁Map
	destroyFuncMap = make(map[DestroyFunc]int)
)

func RegisterDestroy(destroyFunc DestroyFunc, order int) {
	destroyFuncMap[destroyFunc] = order
}

func GetAllDestroyFunc() []DestroyFunc {
	var funs = make([]DestroyFunc, 0, len(destroyFuncMap))
	for fun := range destroyFuncMap {
		funs = append(funs, fun)
	}
	return funs
}
