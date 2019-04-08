package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
)

// http://www.redis.cn/topics/protocol.html

type InputType byte

const (
	StringType InputType = '+'
	ErrorType  InputType = '-'
	IntType    InputType = ':'
	// 大体积，紧接着跟随一个数字表示数据大小,然后是实际数据
	BulkStringType InputType = '$'

	// 数组，紧接着一个是数组的长度，然后使用BulkString分别返回
	// *2
	// $3
	// bar
	// $5
	// hello
	ArrayType InputType = '*'
)

var (
	errPrefix      = errors.New("errors prefix")
	unexpectEnding = errors.New("unexpect ending")
)

type Input struct {
	Type  InputType
	Value []byte

	Array []*Input
}

func main() {
	fileInfo, _ := os.Stdin.Stat()
	if (fileInfo.Mode() & os.ModeNamedPipe) != os.ModeNamedPipe {
		fmt.Println("仅支持管道访问")
		os.Exit(1)
	}
	var chars []byte
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	})
	for scanner.Scan() {
		chars = scanner.Bytes()
	}

	if chars == nil {
		fmt.Println("输入为空")
		os.Exit(1)
	}

	parser := NewParser(chars)
	input, er := parser.Parse()
	if er != nil {
		fmt.Println(er)
		os.Exit(1)
	}
	fmt.Print("type: ", string(input.Type), "\n")
	if input.Type != ArrayType {
		fmt.Println(string(input.Value), "@from value")
	} else {
		for i := range input.Array {
			fmt.Println(string(input.Array[i].Value), "@from array value")
		}
	}
}

type Parser struct {
	bf *bytes.Buffer
}

func NewParser(chars []byte) *Parser {
	return &Parser{bf: bytes.NewBuffer(chars)}
}

func (p *Parser) Parse() (*Input, error) {
	firstByte, err := p.bf.ReadByte()
	if err != nil {
		return nil, err
	}

	res := &Input{}
	res.Type = InputType(firstByte)
	switch res.Type {
	// Input里面是Value是byte[]，所以这里intType放在一起处理
	case StringType, ErrorType, IntType:
		res.Value, err = p.parseStringType()
	case ArrayType:
		res.Array, err = p.parseArrayType()
	case BulkStringType:
		res.Value, err = p.parseBulkString()
	default:
		return nil, errPrefix
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *Parser) parseStringType() ([]byte, error) {
	line, err := p.bf.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	length := len(line) - 2 // 末尾必有一个CRLF
	if length < 0 || line[length] != '\r' {
		return nil, unexpectEnding
	}

	return line[:length], nil
}

func (p *Parser) getArrayOrBulkLen() (int64, error) {
	lenBytes, err := p.bf.ReadBytes('\n')
	if err != nil {
		return 0, err
	}
	length := len(lenBytes) - 2
	if length < 0 || lenBytes[length] != '\r' {
		return 0, unexpectEnding
	}

	return strconv.ParseInt(string(lenBytes[:length]), 10, 64)
}

func (p *Parser) parseBulkString() ([]byte, error) {
	bulkLen, err := p.getArrayOrBulkLen()

	if err != nil {
		return nil, err
	}

	switch {
	case bulkLen == -1:
		return nil, nil // 数据就是空的
	case bulkLen < -1:
		return nil, errors.New("bad bulk string") // TODO
	}
	bs := make([]byte, bulkLen+2)

	readLen, err := p.bf.Read(bs)
	if err != nil {
		return nil, err
	}

	if int64(readLen) != bulkLen+2 {
		return nil, errors.New("less bulk string") // TODO
	}

	if bs[bulkLen] != '\r' || bs[bulkLen+1] != '\n' {
		return nil, errors.New("bad bulk length")
	}

	return bs[:bulkLen], nil
}

func (p *Parser) parseArrayType() ([]*Input, error) {
	arrayLen, err := p.getArrayOrBulkLen()
	if err != nil {
		return nil, err
	}

	switch {
	case arrayLen < -1:
		return nil, errors.New("bad array length") // TODO
	case arrayLen == -1:
		return nil, nil // 说明数据就是空的
		// 有最大长度限制吗？
	}
	resultArray := make([]*Input, arrayLen)
	for i := range resultArray {
		r, er := p.Parse()
		if er != nil {
			return nil, er
		}
		resultArray[i] = r
	}
	return resultArray, nil
}
