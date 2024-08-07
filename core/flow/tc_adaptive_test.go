package flow

import (
	"encoding/json"
	"fmt"
	"git.garena.com/honggang.liu/seamiter-go/core/system_metric"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemoryAdaptiveTrafficShapingCalculator_CalculateAllowedTokens(t *testing.T) {
	tc1 := &MemoryAdaptiveTrafficShapingCalculator{
		owner:                 nil,
		lowMemUsageThreshold:  1000,
		highMemUsageThreshold: 100,
		memLowWaterMark:       1024,
		memHighWaterMark:      2048,
	}

	system_metric.SetSystemMemoryUsage(100)
	assert.True(t, util.Float64Equals(tc1.CalculateAllowedTokens(0, 0), tc1.lowMemUsageThreshold))

	system_metric.SetSystemMemoryUsage(1024)
	assert.True(t, util.Float64Equals(tc1.CalculateAllowedTokens(0, 0), tc1.lowMemUsageThreshold))

	system_metric.SetSystemMemoryUsage(1536)
	assert.True(t, util.Float64Equals(tc1.CalculateAllowedTokens(0, 0), 550))

	system_metric.SetSystemMemoryUsage(2048)
	assert.True(t, util.Float64Equals(tc1.CalculateAllowedTokens(0, 0), 100))

	system_metric.SetSystemMemoryUsage(3072)
	assert.True(t, util.Float64Equals(tc1.CalculateAllowedTokens(0, 0), 100))
}

func TestInt2Float(t *testing.T) {
	var mjson = `{"lowMemUsageThreshold":100.55}`
	var mat = &Rule{}
	err := json.Unmarshal([]byte(mjson), mat)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(mat.LowMemUsageThreshold)
}
