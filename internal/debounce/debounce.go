package debounce

import "time"

type Debouncer struct {
	input  <-chan string
	output chan struct{}
}

func NewDebouncer(input <-chan string) *Debouncer {
	d := &Debouncer{
		input:  input,
		output: make(chan struct{}, 1),
	}
	go d.run()
	return d
}

func (d *Debouncer) Output() <-chan struct{} {
	return d.output
}

func (d *Debouncer) run() {
	const delay = 500 * time.Millisecond

	var timer *time.Timer
	var timerC <-chan time.Time

	for {
		select {
		case _, ok := <-d.input:
			if !ok {
				if timer != nil {
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
				}
				close(d.output)
				return
			}

			if timer == nil {
				timer = time.NewTimer(delay)
				timerC = timer.C
				continue
			}

			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(delay)
			timerC = timer.C

		case <-timerC:
			select {
			case d.output <- struct{}{}:
			default:
			}
			timerC = nil
		}
	}
}
