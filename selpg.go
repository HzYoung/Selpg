package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	flag "github.com/spf13/pflag"
)

const INT_MAX = int(^uint(0) >> 1)

func main() {
	/*==process_args()==*/
	startPage := flag.IntP("startPage", "s", -1, "the start page")
	endPage := flag.IntP("endPage", "e", -1, "the end page")
	pageLength := flag.IntP("pageLength", "l", 72, "line number in one page, not compatible with '-f'")
	pageType := flag.BoolP("flagPage", "f", false, "splits page using '/f', not compatible with '-l'")
	printDest := flag.StringP("printDest", "d", "", "name of printer destination")

	flag.Parse()

	//check the command-line arguments for validity
	if *startPage == -1 || *endPage == -1 {
		fmt.Println("selpg.go: not enough arguments")
		os.Exit(1)
	}
	//start page
	if *startPage < 1 || *startPage > (INT_MAX-1) {
		fmt.Println("selpg.go: invalid start page")
		os.Exit(1)
	}
	// end page
	if *endPage < 1 || *endPage > (INT_MAX-1) || *endPage < *startPage {
		fmt.Println("selpg.go: invalid end page")
		os.Exit(1)
	}
	// page type and page length are mutual exclusion
	if *pageType && *pageLength != 72 {
		fmt.Println("selpg.go: page type and page length are mutual exclusion")
		os.Exit(1)
	}
	// page length
	if *pageType == false && (*pageLength < 1 || *pageLength > (INT_MAX-1)) {
		fmt.Println("selpg.go: invalid page length")
		os.Exit(1)
	}

	/*==process_input()==*/
	reader := bufio.NewReader(os.Stdin)

	if flag.NArg() > 0 {
		file, err := os.Open(flag.Args()[0])
		if err != nil {
			panic(err)
			os.Exit(1)
		}
		defer file.Close()
		reader = bufio.NewReader(file)
	}
	result := ""
	pageCtr := 1
	lineCtr := 0

	if *pageType {
		for {
			str, err := reader.ReadString('\f')
			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
				os.Exit(1)
			}
			pageCtr++
			if pageCtr >= *startPage && pageCtr <= *endPage {
				result = strings.Join([]string{result, str}, "")
			}
		}
	} else {
		for {
			str, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
				os.Exit(1)
			}
			lineCtr++
			if lineCtr > *pageLength {
				pageCtr++
				lineCtr = 1
			}
			if pageCtr >= *startPage && pageCtr <= *endPage {
				result = strings.Join([]string{result, str}, "")
			}
		}
	}

	if pageCtr < *startPage {
		err := fmt.Sprintf("start page (%d) greater than total pages (%d)", *startPage, pageCtr)
		fmt.Println(err)
		os.Exit(1)
	} else if pageCtr < *endPage {
		err := fmt.Sprintf("end page (%d) greater than total pages: (%d)", *endPage, pageCtr)
		fmt.Println(err)
		os.Exit(1)
	}

	/*==process_output()==*/
	if *printDest != "" {
		cmd := exec.Command("lp", "-d"+*printDest)
		cmd.Stdin = strings.NewReader(result)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + " : " + stderr.String())
		}
	} else {
		fmt.Println(result)
	}

}
