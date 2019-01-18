# Selpg
selpg 是从文本输入选择页范围的实用程序。该输入可以来自作为最后一个命令行参数指定的文件，在没有给出文件名参数时也可以来自标准输入。
[作业内容](https://pmlpml.github.io/ServiceComputingOnCloud/ex-cli-basic)
[参考资料](https://www.ibm.com/developerworks/cn/linux/shell/clutil/index.html)
## 使用
```
selpg -s start_page -e end_page [ -f | -l lines_per_page ][ -d dest ] [ in_filename ]

Usage of selpg:
  -e, --endPage int        the end page (default -1)
  -f, --flagPage           splits page using '/f', not compatible with '-l'
  -l, --pageLength int     line number in one page, not compatible with '-f' (default 72)
  -d, --printDest string   name of printer destination
  -s, --startPage int      the start page (default -1)

```
- 强制选项
1. -s 起始页
2. -e 结束页
- 可选选项
1. -l 每页的行数
2. -f 使用的分页方式, 
3. -d 将选定的页直接发送至打印机
4. 输入文件
## 设计
#### 1. 解析参数
用pflag包对参数进行解析，解析得到数据的指针。需要先绑定参数名，默认值，参数描述。用`flag.Parse()`解析。
```go
    startPage := flag.IntP("startPage", "s", -1, "the start page")
	endPage := flag.IntP("endPage", "e", -1, "the end page")
	pageLength := flag.IntP("pageLength", "l", 72, "line number in one page, not compatible with '-f'")
	pageType := flag.BoolP("flagPage", "f", false, "splits page using '/f', not compatible with '-l'")
	printDest := flag.StringP("printDest", "d", "", "name of printer destination")

	flag.Parse()
```
#### 2. 判断参数是否符合规范
按照[参考资料](https://www.ibm.com/developerworks/cn/linux/shell/clutil/index.html)中给定的规范
```go
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
```
#### 3.输入文件，并进行处理
一开始先用os.Stdin标准输入，如果有输入文件，再读取输入文件中的数据。
两种分页方式，一种是用‘\f’作为分页标识符，另一种是用固定行数进行分割。
```go
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
```
#### 4.输出文件
```go
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
```
## 测试
#### 生成测试数据
生成1-720共720行的等差数据到input_file文件中
```
$ seq 720 >input_file
```
#### 测试一
运行
```
selpg -s1 -e1 input_file
```
结果：1到72，共72行
#### 测试二
运行
```
selpg -s1 -e1 < input_file
```
结果：1到72，共72行
#### 测试三
运行
```
selpg -s1 -e2 input_file >output_file
```
结果： 得到文件output_file，包含1到144行数据

