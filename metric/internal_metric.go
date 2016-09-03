package metric

// InternalMetrics holds the key:value pairs for counters/gauges
type InternalMetrics struct {
	Counters map[string]float64
	Gauges   map[string]float64
}

// NewInternalMetrics initializes the internal components of InternalMetrics
func NewInternalMetrics() *InternalMetrics {
	inst := new(InternalMetrics)
	inst.Counters = make(map[string]float64)
	inst.Gauges = make(map[string]float64)
	return inst
}

// CollectorEmission counts collector emissions
type CollectorEmission struct {
	Name          string
	EmissionCount uint64
}
