package main

//
//import (
//	"fmt"
//	"github.com/liuhailove/gmiter/pkg/adapters/go_scheduler/middleware_test/task"
//	"log"
//	"os"
//
//	gs "git.garena.com/shopee/loan-service/credit_backend/fast-escrow/go-scheduler-executor-go"
//	sea "github.com/liuhailove/gmiter/api"
//	"github.com/liuhailove/gmiter/pkg/adapters/go_scheduler"
//	"github.com/liuhailove/gretry/logging"
//)
//
//func main() {
//	wd, _ := os.Getwd()
//	err := sea.InitWithConfigFile(wd + "/middleware_test/resources/config/sea.yml")
//	if err != nil {
//		logging.Warn("Upexpected error:", "err", err)
//	}
//	exec := gs.NewExecutor(
//		gs.ServerAddr("http://localhost:8082/xxl-job-admin"),
//		gs.RegistryKey("my-golang-jobs-2"), //执行器名称
//		gs.SetLogger(&logger{}),            //自定义日志
//		gs.SetSync(true),
//		gs.WithTaskWrapper(go_scheduler.SeaMiddleware()),
//	)
//	exec.Init()
//	// 注册任务handler
//	exec.RegTask("task.test", task.Test)
//	exec.Run()
//}
//
//// xxl.Logger接口实现
//type logger struct{}
//
//func (l *logger) Info(format string, a ...interface{}) {
//	fmt.Println(fmt.Sprintf("自定义日志 [Info]- "+format, a...))
//}
//
//func (l *logger) Error(format string, a ...interface{}) {
//	log.Println(fmt.Sprintf("自定义日志 [Error]- "+format, a...))
//}
//
//func (l *logger) Debug(format string, a ...interface{}) {
//	log.Println(fmt.Sprintf("自定义日志 [Debug]- "+format, a...))
//}
