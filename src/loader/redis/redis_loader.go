package main

import (
    "fmt"
    "github.com/simonz05/godis"    
)

func main() {
    r := godis.New("", 10, "")
    r.Set("test", 12224)
    fmt.Println("Done!")
}
