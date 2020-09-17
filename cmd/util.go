package cmd

func handleErrorWithPanic(err error) {
	if err != nil {
		panic(err)
	}
}
