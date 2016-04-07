package file

import "time"

type Timer struct {
	t0   int64
	Rate float64
}

func MakeTimer() Timer {
	timer := Timer{}
	timer.t0 = time.Now().UnixNano()
	return timer
}

func (timer *Timer) Tick() float64 {
	t := time.Now().UnixNano()
	dt := t - timer.t0
	if dt <= 0 {
		return 0
	}
	rate := 1.0e9 / float64(dt)
	timer.t0 = t
	timer.Rate = (0.0*timer.Rate + 1.0*rate)
	return timer.Rate
}
