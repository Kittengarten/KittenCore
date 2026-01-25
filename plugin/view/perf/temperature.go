//go:build !windows

package perf

import (
	"cmp"
	"slices"
	"strconv"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"

	"github.com/shirou/gopsutil/v4/sensors"
)

// CPUTemperature 获取 CPU 温度（默认为所有传感器温度中最高的）
func CPUTemperature(_ fio.Path) string {
	t, err := sensors.SensorsTemperatures()
	if err != nil {
		return ErrNoData.Error()
	}
	if len(t) == 0 {
		return ErrNoData.Error()
	}
	return strconv.FormatFloat(slices.MaxFunc(t, func(i, j sensors.TemperatureStat) int {
		return cmp.Compare(i.Temperature, j.Temperature)
	}).Temperature, 'f', 2, 64)
}
