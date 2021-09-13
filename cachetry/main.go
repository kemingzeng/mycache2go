package main

import (
	"fmt"
	"time"

	"try.com/mycache2go"
)

func main() {
	ct := mycache2go.NewCacheTable("KemCache")
	ct.AddItem("weak", "tyy", 6*time.Millisecond)
	ct.AddItem("strong", "kem", 5*time.Millisecond)
	item1, _ := ct.Data("weak")
	item2, _ := ct.Data("strong")
	time.Sleep(20 * time.Millisecond)
	_, err := ct.Data("weak")
	fmt.Printf("%s represents %s\n", item1.Key(), item1.Value())
	fmt.Printf("%s represents %s\n", item2.Key(), item2.Value())
	if err != nil {
		fmt.Println(err)
	}
}
