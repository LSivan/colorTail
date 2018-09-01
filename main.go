/**
 * DateTime : 18-8-28 下午5:29
 * Author : liangxingwei
 */

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"io"
	"os"
	"strings"
)


type Reader struct {
	fileName string
	begin    int64 // 起点
}

func NewReader(fileName string) *Reader {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		panic(fmt.Sprintf("open file error : %v", err))
	}
	r := &Reader{}
	r.fileName = fileName
	r.begin = fileInfo.Size()
	return r
}

var follow bool
var fileName string
var lineNum int

func main() {

	flag.BoolVar(&follow, "f", false, "是否持续监听，缺省表示否")
	flag.IntVar(&lineNum, "n", 10, "非监听模式下生效，从后往前读取的行数")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "用法: %s [选项]... 文件 \n\n选项：\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	fileName = flag.Arg(0)

	if fileName == "" {
		fmt.Println("fileName connot be nil")
		return
	}
	r := NewReader(fileName)
	if follow { // 监听模式
		for {
			fileInfo, err := os.Stat(r.fileName)
			if err != nil {
				os.Exit(1)
			}

			if s := fileInfo.Size(); s > r.begin {
				f, err := os.Open(r.fileName)
				if err != nil {
					continue
				}
				f.Seek(r.begin, io.SeekStart)
				rd := bufio.NewReader(f)
				p, e := rd.ReadBytes('\n')

				for e == nil {
					colorLog(string(p))
					p, e = rd.ReadBytes('\n')
				}
				if e == io.EOF && len(p) > 0 {
					colorLog(string(p))
				}
				fileInfo, _ = os.Stat(r.fileName)
				r.begin = fileInfo.Size()
				//}
				f.Close()
			}
		}
	} else { // 读行模式
		lines, err := tail(r.fileName, lineNum)
		if err != nil {
			panic(err)
		}
		for i := len(lines) - 1; i >= 0; i-- {
			colorLog(lines[i] + "\n")
		}
	}

}

func colorLog(line string) {
	tmpLine := strings.ToLower(line)
	if strings.Contains(tmpLine, "level\":\"debug") {
		color.New(color.FgHiWhite).Print(line)
	} else if strings.Contains(tmpLine, "level\":\"info") {
		color.New(color.FgHiBlue).Print(line)
	} else if strings.Contains(tmpLine, "level\":\"warn") {
		color.New(color.FgHiYellow).Print(line)
	} else if strings.Contains(tmpLine, "level\":\"error") {
		color.New(color.FgHiRed).Print(line)
	} else if strings.Contains(tmpLine, "level\":\"fatal") {
		color.New(color.FgHiRed).Add(color.Bold).Print(line)
	} else if strings.Contains(tmpLine, "level\":\"panic") {
		color.New(color.FgHiRed).Add(color.Bold).Add(color.Underline).Print(line)
	} else {
		color.New(color.FgHiGreen).Print(line)
	}
}

const (
	defaultBufSize = 4096
)

func tail(filename string, n int) (lines []string, err error) {
	f, e := os.Stat(filename)
	if e == nil {
		size := f.Size()
		var fi *os.File
		fi, err = os.Open(filename)
		if err == nil {
			b := make([]byte, defaultBufSize)
			sz := int64(defaultBufSize)
			nn := n
			bTail := bytes.NewBuffer([]byte{})
			isStart := size
			readFlag := true
			for readFlag {
				if isStart < defaultBufSize {
					sz = isStart
					isStart = 0
					//readFlag = false
				} else {
					isStart -= sz
				}
				_, err = fi.Seek(isStart, io.SeekStart)
				if err == nil {
					mm, e := fi.Read(b)
					if e == nil && mm > 0 {
						j := mm
						for i := mm - 1; i >= 0; i-- {
							if b[i] == '\n' {
								bLine := bytes.NewBuffer([]byte{})
								bLine.Write(b[i+1 : j])
								j = i
								if bTail.Len() > 0 {
									bLine.Write(bTail.Bytes())
									bTail.Reset()
								}

								if (nn == n && bLine.Len() > 0) || nn < n { //skip last "\n"
									lines = append(lines, bLine.String())
									nn--
								}
								if nn == 0 {
									readFlag = false
									break
								}
							}
						}
						if readFlag && j > 0 {
							if isStart == 0 {
								bLine := bytes.NewBuffer([]byte{})
								bLine.Write(b[:j])
								if bTail.Len() > 0 {
									bLine.Write(bTail.Bytes())
									bTail.Reset()
								}
								lines = append(lines, bLine.String())
								readFlag = false
							} else {
								bb := make([]byte, bTail.Len())
								copy(bb, bTail.Bytes())
								bTail.Reset()
								bTail.Write(b[:j])
								bTail.Write(bb)
							}
						}
					}
				}
			}
		}
		defer fi.Close()
	}
	return
}
