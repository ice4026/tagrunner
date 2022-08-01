package tagrunner

type Opt func(runner *Runner)

func WithDive(dive bool) Opt {
	return func(runner *Runner) {
		runner.dive = dive
	}
}
