package base

import (
	"errors"
	"fmt"
	"github.com/liuhailove/gmiter/util"
	"strconv"
	"strings"
)

const metricPartSeparator = "|"

// MetricItem 表示每行的度量日志数据
type MetricItem struct {
	Resource       string `json:"resource"`
	Classification int32  `json:"classification"`
	Timestamp      uint64 `json:"timestamp"`

	PassQps         uint64 `json:"passQps"`
	BlockQps        uint64 `json:"blockQps"`
	CompleteQps     uint64 `json:"completeQps"`
	ErrorQps        uint64 `json:"errorQps"`
	AvgRt           uint64 `json:"avgRt"`
	OccupiedPassQps uint64 `json:"occupiedPassQps"`
	Concurrency     uint32 `json:"concurrency"`

	// 新增，以便Server收集数据，便于告警
	BlockedNumByFlow            uint64 `json:"blockedNumByFlow"`            // 被限流阻塞数
	BlockedNumByIsolation       uint64 `json:"blockedNumByIsolation"`       // 被资源隔离阻塞数
	BlockedNumByCircuitBreaking uint64 `json:"blockedNumByCircuitBreaking"` // 被熔断阻塞数
	BlockedNumBySystem          uint64 `json:"blockedNumBySystem"`          // 被系统限流阻塞数
	BlockedNumByHotspotParam    uint64 `json:"blockedNumByHotspotParam"`    // 被热点限流阻塞数
	BlockedNumByMock            uint64 `json:"blockedNumByMock"`            // 被Mock阻塞数
}

type MetricItemRetriever interface {
	MetricsOnCondition(predicate TimePredicate) []*MetricItem
}

func (m *MetricItem) ToFatString() (string, error) {
	b := strings.Builder{}
	timeStr := util.FormatTimeMillis(m.Timestamp)
	// 所有“|”资源名称中的内容将替换为“_”
	finalName := strings.ReplaceAll(m.Resource, "|", "-")
	_, err := fmt.Fprintf(&b, "%d|%s|%s|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d",
		m.Timestamp, timeStr, finalName, m.PassQps,
		m.BlockQps, m.CompleteQps, m.ErrorQps, m.AvgRt,
		m.OccupiedPassQps, m.Concurrency, m.Classification,
		m.BlockedNumByFlow, m.BlockedNumByIsolation, m.BlockedNumByCircuitBreaking,
		m.BlockedNumBySystem, m.BlockedNumByHotspotParam, m.BlockedNumByMock)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (m *MetricItem) ToThinString() (string, error) {
	b := strings.Builder{}
	finalName := strings.ReplaceAll(m.Resource, "|", "-")
	_, err := fmt.Fprintf(&b, "%d|%s|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d",
		m.Timestamp, finalName, m.PassQps,
		m.BlockQps, m.CompleteQps, m.ErrorQps, m.AvgRt,
		m.OccupiedPassQps, m.Concurrency, m.Classification,
		m.BlockedNumByFlow, m.BlockedNumByIsolation, m.BlockedNumByCircuitBreaking,
		m.BlockedNumBySystem, m.BlockedNumByHotspotParam, m.BlockedNumByMock)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func MetricItemFromFatString(line string) (*MetricItem, error) {
	if len(line) == 0 {
		return nil, errors.New("invalid metric line: empty string")
	}
	item := &MetricItem{}
	arr := strings.Split(line, metricPartSeparator)
	if len(arr) < 8 {
		return nil, errors.New("invalid metric line: invalid format")
	}
	ts, err := strconv.ParseUint(arr[0], 10, 64)
	if err != nil {
		return nil, err
	}
	item.Timestamp = ts
	item.Resource = arr[2]
	p, err := strconv.ParseUint(arr[3], 10, 64)
	if err != nil {
		return nil, err
	}
	item.PassQps = p
	b, err := strconv.ParseUint(arr[4], 10, 64)
	if err != nil {
		return nil, err
	}
	item.BlockQps = b
	c, err := strconv.ParseUint(arr[5], 10, 64)
	if err != nil {
		return nil, err
	}
	item.CompleteQps = c
	e, err := strconv.ParseUint(arr[6], 10, 64)
	if err != nil {
		return nil, err
	}
	item.ErrorQps = e
	rt, err := strconv.ParseUint(arr[7], 10, 64)
	if err != nil {
		return nil, err
	}
	item.AvgRt = rt

	if len(arr) >= 9 {
		oc, err := strconv.ParseUint(arr[8], 10, 64)
		if err != nil {
			return nil, err
		}
		item.OccupiedPassQps = oc
	}
	if len(arr) >= 10 {
		concurrency, err := strconv.ParseUint(arr[9], 10, 32)
		if err != nil {
			return nil, err
		}
		item.Concurrency = uint32(concurrency)
	}
	if len(arr) >= 11 {
		cl, err := strconv.ParseInt(arr[10], 10, 32)
		if err != nil {
			return nil, err
		}
		item.Classification = int32(cl)
	}

	// 新增，以便Server收集数据，便于告警
	if len(arr) >= 12 {
		blockedNumByFlow, err := strconv.ParseUint(arr[11], 10, 64)
		if err != nil {
			return nil, err
		}
		item.BlockedNumByFlow = blockedNumByFlow
	}
	if len(arr) >= 13 {
		blockedNumByIsolation, err := strconv.ParseUint(arr[12], 10, 64)
		if err != nil {
			return nil, err
		}
		item.BlockedNumByIsolation = blockedNumByIsolation
	}
	if len(arr) >= 14 {
		blockedNumByCircuitBreaking, err := strconv.ParseUint(arr[13], 10, 64)
		if err != nil {
			return nil, err
		}
		item.BlockedNumByCircuitBreaking = blockedNumByCircuitBreaking
	}
	if len(arr) >= 15 {
		blockedNumBySystem, err := strconv.ParseUint(arr[14], 10, 64)
		if err != nil {
			return nil, err
		}
		item.BlockedNumBySystem = blockedNumBySystem
	}
	if len(arr) >= 16 {
		blockedNumByHotspotParam, err := strconv.ParseUint(arr[15], 10, 64)
		if err != nil {
			return nil, err
		}
		item.BlockedNumByHotspotParam = blockedNumByHotspotParam
	}
	if len(arr) >= 17 {
		blockedNumByMock, err := strconv.ParseUint(arr[16], 10, 64)
		if err != nil {
			return nil, err
		}
		item.BlockedNumByMock = blockedNumByMock
	}
	return item, nil
}
