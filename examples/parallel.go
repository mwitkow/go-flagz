
package main
import (
	"flag"
	"fmt"
	"os"
	log "github.com/Sirupsen/logrus"
	"time"
	flagz_etcd "github.com/mwitkow/go-flagz/etcd"
)

var (
	myPackage = flag.NewFlagSet("custom_flagset", flag.ContinueOnError)

	myString = myPackage.String("somestring", "valueD", "someusage")
	myInt = myPackage.Int("someint", 1337, "someusage int")
	x = &flagz_etcd.Updater{}
)

func main() {
    fmt.Println("hello world")
	myPackage.Parse(os.Args[1:])


	go func() {
		for true {
			log.Infof("int: %v str: %v", *myInt, *myString)
			time.Sleep(50 * time.Millisecond)
		}
	}()

	for i := 0; i < 30; i++ {
		newInt := fmt.Sprintf("%v", 1337 + i)
		newStr := fmt.Sprintf("value%v", i)
		myPackage.Parse([]string{"-someint", newInt})
		myPackage.Parse([]string{"-somestring", newStr})
		log.Warningf("Updated")
		time.Sleep(5 * time.Second)
	}

}

