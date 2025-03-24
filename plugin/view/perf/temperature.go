//go:build !windows

package perf

import (
	"cmp"
	"slices"
	"strconv"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"

	"github.com/shirou/gopsutil/v4/sensors"
)

// 获取 CPU 温度（默认为所有传感器温度中最高的）
func cpuTemperature(_ fio.Path) string {
	t, err := sensors.SensorsTemperatures()
	if err != nil {
		kitten.Warn(err)
		return defaultTemperature
	}
	return strconv.FormatFloat(slices.MaxFunc(t, func(i, j sensors.TemperatureStat) int {
		return cmp.Compare(i.Temperature, j.Temperature)
	}).Temperature, 'f', 2, 64)
}
