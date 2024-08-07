package micro

//
//import (
//	"context"
//	"fmt"
//	sea "github.com/liuhailove/gmiter/api"
//	"github.com/liuhailove/gmiter/core/base"
//	"github.com/liuhailove/gmiter/core/flow"
//	"github.com/liuhailove/gmiter/core/stat"
//	proto "github.com/liuhailove/gmiter/pkg/adapters/micro/test"
//	"github.com/liuhailove/gmiter/util"
//	"github.com/micro/go-micro"
//	micro_error "github.com/micro/go-micro/errors"
//	"github.com/micro/go-micro/server"
//	"github.com/pkg/errors"
//	"github.com/stretchr/testify/assert"
//	"log"
//	"regexp"
//	"strings"
//	"testing"
//	"time"
//)
//
//const FakeErrorMsg = "fake error for testing"
//
//type TestHandler struct {
//}
//
//func (h *TestHandler) Ping(ctx context.Context, req *proto.Request, rsp *proto.Response) error {
//	rsp.Result = "Pong"
//	return nil
//}
//
//func TestServerLimiter(t *testing.T) {
//	srv := micro.NewService(
//		micro.Name("sea.test.server"),
//		micro.Address("localhost:56436"),
//		micro.Version("latest"),
//		micro.WrapHandler(NewHandlerWrapper(
//			// add custom fallback function to return a fake error for assertion
//			WithServerBlockFallback(
//				func(ctx context.Context, request server.Request, blockError *base.BlockError) error {
//					return errors.New(FakeErrorMsg)
//				}),
//		)))
//	srv.Init()
//	_ = proto.RegisterTestHandler(srv.Server(), &TestHandler{})
//	go srv.Run()
//
//	time.Sleep(time.Second)
//
//	c := srv.Client()
//	req := c.NewRequest("sea.test.server", "Test.Ping", &proto.Request{})
//
//	err := sea.InitDefault()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	_, err = flow.LoadRules([]*flow.Rule{
//		{
//			Resource:               req.Method(),
//			Threshold:              1.0,
//			TokenCalculateStrategy: flow.Direct,
//			ControlBehavior:        flow.Reject,
//		},
//	})
//
//	assert.Nil(t, err)
//
//	var rsp = &proto.Response{}
//
//	t.Run("success", func(t *testing.T) {
//		var _, err = flow.LoadRules([]*flow.Rule{
//			{
//				Resource:               req.Method(),
//				Threshold:              1.0,
//				TokenCalculateStrategy: flow.Direct,
//				ControlBehavior:        flow.Reject,
//			},
//		})
//		assert.Nil(t, err)
//		err = c.Call(context.TODO(), req, rsp)
//		assert.Nil(t, err)
//		assert.EqualValues(t, "Pong", rsp.Result)
//		assert.True(t, util.Float64Equals(1.0, stat.GetResourceNode(req.Method()).GetQPS(base.MetricEventPass)))
//
//		t.Run("second fail", func(t *testing.T) {
//			err := c.Call(context.TODO(), req, rsp)
//			assert.Error(t, err)
//			assert.True(t, util.Float64Equals(1.0, stat.GetResourceNode(req.Method()).GetQPS(base.MetricEventPass)))
//		})
//	})
//
//}
//
//func TestName2(t *testing.T) {
//	var err error = nil
//	if microErr, ok := err.(*micro_error.Error); ok {
//		fmt.Println(microErr)
//	}
//	fmt.Println("hello")
//}
//
//func TestToLower(t *testing.T) {
//	str := "HelloWorld"
//	//lowerTitle := strings.Title(strings.ToLower(str))
//	//fmt.Println(lowerTitle) // Output: hello world
//
//	re := regexp.MustCompile(`\b\w`)
//	lowerTitle := re.ReplaceAllStringFunc(str, strings.ToLower)
//	fmt.Println(lowerTitle) // Output: hello world
//
//}
