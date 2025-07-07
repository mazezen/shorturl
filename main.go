package main

func main() {
	a := App{}
	env := getArgs()
	a.Initizlize(env)

	a.Run(":8001")
}
