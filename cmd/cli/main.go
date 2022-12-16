package main

import (
	"fmt"
	"os"

	"github.com/ejuju/websizer/pkg/httputils"
)

func main() {
	if len(os.Args) <= 1 {
		panic("missing first argument (URL)")
	}

	pagesize, err := httputils.GetPageSize(os.Args[1])
	if err != nil {
		panic(err)
	}

	fmt.Println(pagesize)
}
