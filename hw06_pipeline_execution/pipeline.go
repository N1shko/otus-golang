package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	for _, stage := range stages {
		in = ExecuteStage(stage(in), done)
	}
	return in
}

func ExecuteStage(in In, done In) Bi {
	out := make(Bi)

	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				go func() {
					for linterFool := range in {
						_ = linterFool
					}
				}()
				return

			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-done:
					go func() {
						for linterFool := range in {
							_ = linterFool
						}
					}()
					return
				case out <- v:
				}
			}
		}
	}()
	return out
}
