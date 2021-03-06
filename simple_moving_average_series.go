package chart

const (
	// DefaultMovingAverageWindowSize is the default number of values to average.
	DefaultMovingAverageWindowSize = 5
)

// SimpleMovingAverageSeries is a computed series.
type SimpleMovingAverageSeries struct {
	Name  string
	Style Style
	YAxis YAxisType

	WindowSize  int
	InnerSeries ValueProvider

	valueBuffer *RingBuffer
}

// GetName returns the name of the time series.
func (mas SimpleMovingAverageSeries) GetName() string {
	return mas.Name
}

// GetStyle returns the line style.
func (mas SimpleMovingAverageSeries) GetStyle() Style {
	return mas.Style
}

// GetYAxis returns which YAxis the series draws on.
func (mas SimpleMovingAverageSeries) GetYAxis() YAxisType {
	return mas.YAxis
}

// Len returns the number of elements in the series.
func (mas *SimpleMovingAverageSeries) Len() int {
	return mas.InnerSeries.Len()
}

// GetValue gets a value at a given index.
func (mas *SimpleMovingAverageSeries) GetValue(index int) (x float64, y float64) {
	if mas.InnerSeries == nil {
		return
	}
	if mas.valueBuffer == nil || index == 0 {
		mas.valueBuffer = NewRingBufferWithCapacity(mas.GetWindowSize())
	}
	if mas.valueBuffer.Len() >= mas.GetWindowSize() {
		mas.valueBuffer.Dequeue()
	}
	px, py := mas.InnerSeries.GetValue(index)
	mas.valueBuffer.Enqueue(py)
	x = px
	y = mas.getAverage(mas.valueBuffer)
	return
}

// GetLastValue computes the last moving average value but walking back window size samples,
// and recomputing the last moving average chunk.
func (mas SimpleMovingAverageSeries) GetLastValue() (x float64, y float64) {
	if mas.InnerSeries == nil {
		return
	}
	windowSize := mas.GetWindowSize()
	seriesLength := mas.InnerSeries.Len()
	startAt := seriesLength - windowSize
	if startAt < 0 {
		startAt = 0
	}
	vb := NewRingBufferWithCapacity(windowSize)
	for index := startAt; index < seriesLength; index++ {
		xn, yn := mas.InnerSeries.GetValue(index)
		vb.Enqueue(yn)
		x = xn
	}
	y = mas.getAverage(vb)
	return
}

// GetWindowSize returns the window size.
func (mas SimpleMovingAverageSeries) GetWindowSize(defaults ...int) int {
	if mas.WindowSize == 0 {
		if len(defaults) > 0 {
			return defaults[0]
		}
		return DefaultMovingAverageWindowSize
	}
	return mas.WindowSize
}

func (mas SimpleMovingAverageSeries) getAverage(valueBuffer *RingBuffer) float64 {
	var accum float64
	valueBuffer.Each(func(v interface{}) {
		if typed, isTyped := v.(float64); isTyped {
			accum += typed
		}
	})
	return accum / float64(valueBuffer.Len())
}

// Render renders the series.
func (mas *SimpleMovingAverageSeries) Render(r Renderer, canvasBox Box, xrange, yrange Range, defaults Style) {
	style := mas.Style.WithDefaultsFrom(defaults)
	DrawLineSeries(r, canvasBox, xrange, yrange, style, mas)
}
