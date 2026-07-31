package poller

import (
	"math"
	"testing"
)

func TestComputeDelta_FirstPoll(t *testing.T) {
	r := ComputeDelta(nil, 1000, 2000)
	if r.DeltaTX != 0 || r.DeltaRX != 0 {
		t.Fatalf("expected zero delta for first poll, got tx=%d rx=%d", r.DeltaTX, r.DeltaRX)
	}
	if r.CumTX != 0 || r.CumRX != 0 {
		t.Fatalf("expected zero cum for first poll, got tx=%d rx=%d", r.CumTX, r.CumRX)
	}
}

func TestComputeDelta_NormalIncrement(t *testing.T) {
	prev := &CounterState{Tx: 1000, Rx: 2000, CumTX: 5000, CumRX: 6000, TS: 0}
	r := ComputeDelta(prev, 1500, 2500)
	if r.DeltaTX != 500 {
		t.Fatalf("expected deltaTX=500, got %d", r.DeltaTX)
	}
	if r.DeltaRX != 500 {
		t.Fatalf("expected deltaRX=500, got %d", r.DeltaRX)
	}
	if r.CumTX != 5500 {
		t.Fatalf("expected cumTX=5500, got %d", r.CumTX)
	}
	if r.CumRX != 6500 {
		t.Fatalf("expected cumRX=6500, got %d", r.CumRX)
	}
}

func TestComputeDelta_32BitWrap(t *testing.T) {
	prev := &CounterState{Tx: MaxUint32 - 100, Rx: MaxUint32 - 200, CumTX: 100000, CumRX: 200000, TS: 0}

	r := ComputeDelta(prev, 50, 80)

	expectedDeltaTX := (MaxUint32 - prev.Tx) + 50
	if r.DeltaTX != int64(expectedDeltaTX) {
		t.Fatalf("wrap deltaTX: expected %d, got %d", expectedDeltaTX, r.DeltaTX)
	}

	expectedDeltaRX := (MaxUint32 - prev.Rx) + 80
	if r.DeltaRX != int64(expectedDeltaRX) {
		t.Fatalf("wrap deltaRX: expected %d, got %d", expectedDeltaRX, r.DeltaRX)
	}
}

func TestComputeDelta_CounterReset(t *testing.T) {
	prev := &CounterState{Tx: 500000, Rx: 600000, CumTX: 10000, CumRX: 20000, TS: 0}

	r := ComputeDelta(prev, 100, 200)

	if r.DeltaTX != 100 {
		t.Fatalf("reset deltaTX: expected 100 (counter reset to 0), got %d", r.DeltaTX)
	}
	if r.DeltaRX != 200 {
		t.Fatalf("reset deltaRX: expected 200 (counter reset to 0), got %d", r.DeltaRX)
	}
	if r.CumTX != 10100 {
		t.Fatalf("reset cumTX: expected 10100, got %d", r.CumTX)
	}
}

func TestComputeDelta_NoChange(t *testing.T) {
	prev := &CounterState{Tx: 1000, Rx: 2000, CumTX: 5000, CumRX: 6000, TS: 0}
	r := ComputeDelta(prev, 1000, 2000)
	if r.DeltaTX != 0 || r.DeltaRX != 0 {
		t.Fatalf("expected zero delta for no change, got tx=%d rx=%d", r.DeltaTX, r.DeltaRX)
	}
}

func TestParseLinkSpeed(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"10G", Speed10G},
		{"10g", Speed10G},
		{"2.5G", Speed2_5G},
		{"2500M", Speed2_5G},
		{"2500", Speed2_5G},
		{"1G", Speed1G},
		{"1000M", Speed1G},
		{"1000", Speed1G},
		{"100M", Speed100M},
		{"100", Speed100M},
		{"10M", Speed10M},
		{"10", Speed10M},
		{"Auto", DefaultBW},
		{"", DefaultBW},
		{"2G", Speed2G},
		{"2000M", Speed2G},
	}
	for _, tc := range tests {
		got := ParseLinkSpeed(tc.input)
		if got != tc.want {
			t.Errorf("ParseLinkSpeed(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestSanityCheckMaxBytes(t *testing.T) {
	dt := 10.0
	linkSpeed := Speed1G
	maxBytes := SanityCheckMaxBytes(linkSpeed, dt)

	expectedMax := float64(Speed1G) * dt * BurstBuffer / 8.0
	if maxBytes != expectedMax {
		t.Fatalf("SanityCheckMaxBytes(%d, %f) = %f, want %f", linkSpeed, dt, maxBytes, expectedMax)
	}
}

func TestSanityCheckMaxBytes_ZeroDT(t *testing.T) {
	maxBytes := SanityCheckMaxBytes(Speed1G, 0)
	if maxBytes == 0 {
		t.Fatal("SanityCheckMaxBytes should not return 0 for dt=0 (should use default 15s)")
	}
}

func TestClampDelta_UnderLimit(t *testing.T) {
	result := ClampDelta(1000, 10000)
	if result != 1000 {
		t.Fatalf("expected 1000, got %d", result)
	}
}

func TestClampDelta_OverLimit(t *testing.T) {
	result := ClampDelta(10000, 1000)
	if result != 0 {
		t.Fatalf("expected 0 (clamped), got %d", result)
	}
}

func TestClampDelta_AtLimit(t *testing.T) {
	result := ClampDelta(1000, 1000)
	if result != 1000 {
		t.Fatalf("expected 1000, got %d", result)
	}
}

func TestComputeSpeed(t *testing.T) {
	txBPS, rxBPS := ComputeSpeed(1000, 2000, 10.0)
	if txBPS != 800 {
		t.Fatalf("expected txBPS=800 (1000*8/10), got %d", txBPS)
	}
	if rxBPS != 1600 {
		t.Fatalf("expected rxBPS=1600 (2000*8/10), got %d", rxBPS)
	}
}

func TestComputeSpeed_ZeroDT(t *testing.T) {
	txBPS, rxBPS := ComputeSpeed(1000, 2000, 0)
	if txBPS != 0 || rxBPS != 0 {
		t.Fatalf("expected 0,0 for dt=0, got %d,%d", txBPS, rxBPS)
	}
}

func TestComputeSpeed_NegativeDelta(t *testing.T) {
	txBPS, _ := ComputeSpeed(-100, 2000, 10.0)
	if txBPS != 0 {
		t.Fatalf("expected txBPS=0 for negative delta, got %d", txBPS)
	}
}

func TestHistory_Record(t *testing.T) {
	h := &History{}
	h.Record(10.0, 1000, 2000, 800, 1600, 30)
	h.Record(20.0, 2000, 4000, 800, 1600, 30)

	if len(h.Live) != 2 {
		t.Fatalf("expected 2 live samples, got %d", len(h.Live))
	}
	if len(h.Hourly) != 2 {
		t.Fatalf("expected 2 hourly samples, got %d", len(h.Hourly))
	}
}

func TestHistory_LiveMaxLength(t *testing.T) {
	h := &History{}
	for i := 0; i < 150; i++ {
		h.Record(float64(i), int64(i*100), int64(i*200), 0, 0, 30)
	}
	if len(h.Live) != 120 {
		t.Fatalf("expected 120 live samples (max), got %d", len(h.Live))
	}
	if h.Live[0].CumTX != int64(30*100) {
		t.Fatalf("live[0] should be sample 30, got cumTX=%d", h.Live[0].CumTX)
	}
}

func TestHistory_DailyInterval(t *testing.T) {
	h := &History{}
	// 100 second intervals for 3000 seconds = 30 samples
	for i := 0; i < 30; i++ {
		h.Record(float64(i*100), 0, 0, 0, 0, 30)
	}
	// 900 seconds between daily samples, so 30*100/900 = 3 daily samples + first
	expectedDaily := 1 + (30*100)/900
	if len(h.Daily) != expectedDaily {
		t.Fatalf("expected %d daily samples, got %d", expectedDaily, len(h.Daily))
	}
}

func TestHistory_DailyMaxLength(t *testing.T) {
	h := &History{}
	for i := 0; i < 200; i++ {
		h.Record(float64(i*900), 0, 0, 0, 0, 30)
	}
	if len(h.Daily) > 96 {
		t.Fatalf("daily samples exceeded 96, got %d", len(h.Daily))
	}
	if len(h.Daily) != 96 {
		t.Fatalf("expected 96 daily samples, got %d", len(h.Daily))
	}
}

func TestAvgSpeedOver(t *testing.T) {
	h := []PortSample{
		{TS: 0, CumTX: 0, CumRX: 0},
		{TS: 10, CumTX: 1000, CumRX: 2000},
	}
	txBPS, rxBPS := AvgSpeedOver(h, 5, 10)
	if txBPS != 800 {
		t.Fatalf("expected avg txBPS=800 (1000*8/10), got %d", txBPS)
	}
	if rxBPS != 1600 {
		t.Fatalf("expected avg rxBPS=1600 (2000*8/10), got %d", rxBPS)
	}
}

func TestAvgSpeedOver_Wrap(t *testing.T) {
	h := []PortSample{
		{TS: 0, CumTX: math.MaxInt64 - 1000, CumRX: math.MaxInt64 - 2000},
		{TS: 10, CumTX: 500, CumRX: 1000},
	}
	txBPS, rxBPS := AvgSpeedOver(h, 5, 10)
	if txBPS != 400 {
		t.Fatalf("expected avg txBPS=400 (500*8/10=400), got %d", txBPS)
	}
	if rxBPS != 800 {
		t.Fatalf("expected avg rxBPS=800 (1000*8/10=800), got %d", rxBPS)
	}
}

func TestAvgSpeedOver_InsufficientSamples(t *testing.T) {
	txBPS, rxBPS := AvgSpeedOver([]PortSample{{TS: 0, CumTX: 0, CumRX: 0}}, 5, 10)
	if txBPS != 0 || rxBPS != 0 {
		t.Fatalf("expected 0,0 for <2 samples, got %d,%d", txBPS, rxBPS)
	}
}

func TestUnifySpeed(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1000M", "1G"},
		{"1000", "1G"},
		{"2500M", "2.5G"},
		{"2500", "2.5G"},
		{"100M", "100M"},
		{"100", "100M"},
		{"10G", "10G"},
		{"1G", "1G"},
		{"Auto", "Auto"},
		{"", ""},
	}
	for _, tc := range tests {
		got := UnifySpeed(tc.input)
		if got != tc.want {
			t.Errorf("UnifySpeed(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestCounterIntegration_RealisticScenario(t *testing.T) {
	prev := &CounterState{Tx: 0, Rx: 0, CumTX: 0, CumRX: 0}
	var results []DeltaResult
	for i := 0; i < 5; i++ {
		curTX := int64((i + 1) * 1000000)
		curRX := int64((i + 1) * 2000000)
		r := ComputeDelta(prev, curTX, curRX)
		results = append(results, r)
		prev = &CounterState{Tx: curTX, Rx: curRX, CumTX: r.CumTX, CumRX: r.CumRX, TS: float64(i * 10)}
	}
	if results[4].CumTX != 5000000 {
		t.Fatalf("cumulative TX after 5 polls: expected 5000000, got %d", results[4].CumTX)
	}
	if results[4].CumRX != 10000000 {
		t.Fatalf("cumulative RX after 5 polls: expected 10000000, got %d", results[4].CumRX)
	}
}
