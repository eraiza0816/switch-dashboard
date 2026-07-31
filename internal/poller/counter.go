package poller

import (
	"math"
)

const (
	MaxUint32 = 4294967296

	Speed10G  int64 = 10_000_000_000
	Speed2_5G int64 = 2_500_000_000
	Speed2G   int64 = 2_000_000_000
	Speed1G   int64 = 1_000_000_000
	Speed100M int64 = 100_000_000
	Speed10M  int64 = 10_000_000
	DefaultBW int64 = 10_000_000_000

	BurstBuffer = 1.5
)

type CounterState struct {
	Tx    int64   `json:"tx"`
	Rx    int64   `json:"rx"`
	CumTX int64   `json:"cum_tx"`
	CumRX int64   `json:"cum_rx"`
	TS    float64 `json:"ts"`
}

type DeltaResult struct {
	DeltaTX int64
	DeltaRX int64
	CumTX   int64
	CumRX   int64
}

func ComputeDelta(prev *CounterState, curTX, curRX int64) DeltaResult {
	if prev == nil {
		return DeltaResult{CumTX: 0, CumRX: 0}
	}

	deltaTX := deltaValue(curTX, prev.Tx, prev.Tx > 2_000_000_000)
	deltaRX := deltaValue(curRX, prev.Rx, prev.Rx > 2_000_000_000)

	cumTX := prev.CumTX + deltaTX
	cumRX := prev.CumRX + deltaRX

	return DeltaResult{
		DeltaTX: deltaTX,
		DeltaRX: deltaRX,
		CumTX:   cumTX,
		CumRX:   cumRX,
	}
}

func deltaValue(cur, prev int64, wasHigh bool) int64 {
	if cur >= prev {
		return cur - prev
	}
	if wasHigh {
		return (MaxUint32 - prev) + cur
	}
	return cur
}

func ParseLinkSpeed(speed string) int64 {
	switch {
	case contains(speed, "10g"):
		return Speed10G
	case contains(speed, "2.5g") || contains(speed, "2500"):
		return Speed2_5G
	case contains(speed, "2g") || contains(speed, "2000"):
		return Speed2G
	case contains(speed, "1g") || contains(speed, "1000"):
		return Speed1G
	case contains(speed, "100m") || contains(speed, "100"):
		return Speed100M
	case contains(speed, "10m") || contains(speed, "10"):
		return Speed10M
	}
	return DefaultBW
}

func SanityCheckMaxBytes(linkSpeedBPS int64, dt float64) float64 {
	if dt <= 0 {
		dt = 15.0
	}
	return float64(linkSpeedBPS) * dt * BurstBuffer / 8.0
}

func ClampDelta(delta int64, maxBytes float64) int64 {
	if float64(delta) > maxBytes {
		return 0
	}
	return delta
}

func ComputeSpeed(deltaTX, deltaRX int64, dt float64) (txBPS, rxBPS int64) {
	if dt <= 0 {
		return 0, 0
	}
	txBPS = int64(float64(deltaTX) * 8 / dt)
	rxBPS = int64(float64(deltaRX) * 8 / dt)
	if txBPS < 0 {
		txBPS = 0
	}
	if rxBPS < 0 {
		rxBPS = 0
	}
	return txBPS, rxBPS
}

func UnifySpeed(speed string) string {
	if speed == "" {
		return ""
	}
	val, unit := parseSpeed(speed)
	if val == 0 {
		return speed
	}
	switch unit {
	case 'G':
		return formatG(val)
	case 'M', 0:
		if val >= 1000 {
			return formatG(val / 1000.0)
		}
		return formatM(val)
	default:
		return speed
	}
}

func parseSpeed(s string) (float64, byte) {
	var val float64
	var unit byte
	dot := false
	dec := 0.1
	for _, c := range s {
		if c >= '0' && c <= '9' {
			if dot {
				val += float64(c-'0') * dec
				dec *= 0.1
			} else {
				val = val*10 + float64(c-'0')
			}
		} else if c == '.' {
			dot = true
		} else if (c == 'M' || c == 'G' || c == 'm' || c == 'g') && unit == 0 {
			if c == 'm' || c == 'M' {
				unit = 'M'
			} else {
				unit = 'G'
			}
		} else {
			break
		}
	}
	return val, unit
}

func formatG(val float64) string {
	if val == math.Trunc(val) {
		return formatInt(int64(val)) + "G"
	}
	return formatFloat(val) + "G"
}

func formatM(val float64) string {
	if val == math.Trunc(val) {
		return formatInt(int64(val)) + "M"
	}
	return formatFloat(val) + "M"
}

func formatInt(v int64) string {
	return itoa(v)
}

func formatFloat(v float64) string {
	v = math.Round(v*10) / 10
	intPart := int64(v)
	decPart := int(math.Round((v - float64(intPart)) * 10))
	if decPart == 0 {
		return itoa(intPart)
	}
	return itoa(intPart) + "." + itoa(int64(decPart))
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := v < 0
	if neg {
		v = -v
	}
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	lower := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		lower[i] = c
	}
	for i := 0; i <= len(lower)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if lower[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

type PortSample struct {
	TS    float64
	CumTX int64
	CumRX int64
}

type SpeedSample struct {
	TS      float64
	SpeedTX int64
	SpeedRX int64
}

type History struct {
	Live   []PortSample
	Hourly []SpeedSample
	Daily  []SpeedSample
}

func (h *History) Record(now float64, cumTX, cumRX int64, speedTX, speedRX int64, refreshInterval int) {
	// Live: max 120 samples
	h.Live = append(h.Live, PortSample{TS: now, CumTX: cumTX, CumRX: cumRX})
	if len(h.Live) > 120 {
		h.Live = h.Live[1:]
	}

	// Hourly: recorded every cycle
	maxHourly := 3600 / refreshInterval
	if maxHourly < 1 {
		maxHourly = 1
	}
	h.Hourly = append(h.Hourly, SpeedSample{TS: now, SpeedTX: speedTX, SpeedRX: speedRX})
	if len(h.Hourly) > maxHourly {
		h.Hourly = h.Hourly[1:]
	}

	// Daily: 15-minute averages
	if len(h.Daily) == 0 || now-h.Daily[len(h.Daily)-1].TS >= 900 {
		h.Daily = append(h.Daily, SpeedSample{TS: now, SpeedTX: speedTX, SpeedRX: speedRX})
		if len(h.Daily) > 96 {
			h.Daily = h.Daily[1:]
		}
	}
}

func AvgSpeedOver(history []PortSample, duration float64, now float64) (txBPS, rxBPS int64) {
	if len(history) < 2 {
		return 0, 0
	}
	target := now - duration
	best := 0
	minDiff := math.Abs(history[0].TS - target)
	for i, s := range history {
		diff := math.Abs(s.TS - target)
		if diff < minDiff {
			minDiff = diff
			best = i
		}
	}
	start := history[best]
	end := history[len(history)-1]
	dt := end.TS - start.TS
	if dt <= 0 {
		return 0, 0
	}
	txDiff := end.CumTX - start.CumTX
	rxDiff := end.CumRX - start.CumRX
	if txDiff < 0 {
		txDiff = end.CumTX
	}
	if rxDiff < 0 {
		rxDiff = end.CumRX
	}
	return int64(float64(txDiff) * 8 / dt), int64(float64(rxDiff) * 8 / dt)
}
