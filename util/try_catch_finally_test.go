package util

import (
	"errors"
	"fmt"
	"sort"
	"testing"
)

func TestTryCatchFinally(t *testing.T) {
	Try(func() {
		fmt.Println("hello world")
		panic(errors.New("panic zero"))
	}).CatchAll(func(err error) {
		fmt.Println("catch all error", err.Error())
	}).Finally(func() {
		fmt.Println("finally handler")
	})
}

func TestPxx(t *testing.T) {

	var numbers = []float64{1, 2, 3, 4, 5, 6}
	var p99, p95 = CalculateP99P95(numbers)
	fmt.Println(p99)
	fmt.Println(p95)
}

func CalculateP99P95(numbers []float64) (float64, float64) {
	// 对数值数组进行排序
	sort.Float64s(numbers)

	// 计算P99和P95的索引位置
	n := len(numbers)
	p99Index := int((99.0 / 100.0) * float64(n-1))
	p95Index := int((95.0 / 100.0) * float64(n-1))

	// 计算P99和P95的数值
	p99 := numbers[p99Index]
	p95 := numbers[p95Index]

	// 如果索引位置是小数，进行线性插值计算
	if p99Index != p99Index+1 {
		p99Fraction := (99.0/100.0)*float64(n-1) - float64(p99Index)
		p99 = numbers[p99Index] + p99Fraction*(numbers[p99Index+1]-numbers[p99Index])
	}

	if p95Index != p95Index+1 {
		p95Fraction := (95.0/100.0)*float64(n-1) - float64(p95Index)
		p95 = numbers[p95Index] + p95Fraction*(numbers[p95Index+1]-numbers[p95Index])
	}

	return p99, p95
}
