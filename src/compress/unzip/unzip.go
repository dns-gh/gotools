package main

import (
	"compress/compress"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	flag.Usage = func() {
		// http://patorjk.com/software/taag/#p=display&f=Big
		fmt.Fprintf(os.Stderr, ""+
			`unzip [OPTIONS]

----------------------------
  _    _           _       
 | |  | |         (_)      
 | |  | |_ __  _____ _ __  
 | |  | | '_ \|_  / | '_ \ 
 | |__| | | | |/ /| | |_) |
  \____/|_| |_/___|_| .__/ 
                    | |    
                    |_|    
 ---------------------------

Usage:

  unzip -f test.zip

starts "unzip.exe"
 - on the zip file named test.zip
 to create test folder in the current folder

Options:
`)
		flag.PrintDefaults()
	}
	file := flag.String("f", "", "file to unzip")
	flag.Parse()
	if len(*file) <= 0 {
		log.Fatalf("you must specify a file to unzip")
	}
	log.Println("file (-f)", *file)
	dst, err := compress.Unzip(*file)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println("unzipped to", dst)
}
