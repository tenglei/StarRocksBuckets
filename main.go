package main

import (
	_ "setbuckets/init"
	"setbuckets/service"
	"setbuckets/util"
)

func main()  {
	util.Parm()
	service.Run()
}
