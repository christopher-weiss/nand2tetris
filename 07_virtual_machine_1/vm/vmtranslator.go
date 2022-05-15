package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type CommandType int

const (
	C_ARITHMETIC CommandType = iota
	C_PUSH
	C_POP
	C_LABEL
	C_GOTO
	C_IF
	C_FUNCTION
	C_RETURN
	C_CALL
)

type Command struct {
	commandType CommandType
	command     string
	segment     string
	index       uint
}

var loopCount = 0

func main() {
	file := openFile()
	commands := parse(file)
	fmt.Println(commands)
	op := translateToAssembly(commands)
	writeFile(op)
}

func writeFile(op []string) {
	file, err := os.OpenFile("out.asm", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Println("Could not create file")
		os.Exit(1)
	}

	writer := bufio.NewWriter(file)

	for _, data := range op {
		_, _ = writer.WriteString(data + "\n")
	}

	writer.Flush()
	file.Close()
}

func translateToAssembly(stack []Command) []string {
	op := []string{}
	for _, command := range stack {
		switch command.commandType {
		case C_PUSH:
			op = append(op, push(command.index)...)
		case C_POP:
			op = append(op, pop()...)
		case C_ARITHMETIC:
			switch command.command {
			case "add":
				op = append(op, add()...)
			case "sub":
				op = append(op, sub()...)
			case "neg":
				op = append(op, neg()...)
			case "eq":
				op = append(op, eq()...)
			case "gt":
				op = append(op, gt()...)
			case "lt":
				op = append(op, lt()...)
			case "and":
				op = append(op, and()...)
			case "or":
				op = append(op, or()...)
			case "not":
				op = append(op, not()...)
			}
		}
	}
	return op
}

func not() []string {
	op := []string{}
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=!M")
	op = append(op, "M=D")
	op = append(op, "A=A+1")
	op = append(op, "M=0")
	op = append(op, "@SP")
	op = append(op, "AM=M+1")
	return op
}

func or() []string {
	op := []string{}
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=M")
	op = append(op, "A=A-1")
	op = append(op, "D=M|D")
	op = append(op, "M=D")
	op = append(op, "A=A+1")
	op = append(op, "M=0")
	return op
}

func and() []string {
	op := []string{}
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=M")
	op = append(op, "A=A-1")
	op = append(op, "D=M&D")
	op = append(op, "M=D")
	op = append(op, "A=A+1")
	op = append(op, "M=0")
	return op
}

func lt() []string {
	op := []string{}
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=M")
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=M-D")
	op = append(op, fmt.Sprintf("@LESSTHAN%d", loopCount))
	op = append(op, "D;JLT")
	op = append(op, "@SP")
	op = append(op, "A=M")
	op = append(op, "M=0")
	op = append(op, fmt.Sprintf("@GREATERTHAN%d", loopCount))
	op = append(op, "0;JMP")
	op = append(op, fmt.Sprintf("(LESSTHAN%d)", loopCount))
	op = append(op, "@SP")
	op = append(op, "A=M")
	op = append(op, "M=-1")
	op = append(op, fmt.Sprintf("(GREATERTHAN%d)", loopCount))
	op = append(op, "@SP")
	op = append(op, "AM=M+1")
	op = append(op, "@SP")
	op = append(op, "A=M")
	op = append(op, "M=0")
	loopCount++
	return op
}

func gt() []string {
	op := []string{}
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=M")
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=M-D")
	op = append(op, fmt.Sprintf("@GREATERTHAN%d", loopCount))
	op = append(op, "D;JGT")
	op = append(op, "@SP")
	op = append(op, "A=M")
	op = append(op, "M=0")
	op = append(op, fmt.Sprintf("@LESSTHAN%d", loopCount))
	op = append(op, "0;JMP")
	op = append(op, fmt.Sprintf("(GREATERTHAN%d)", loopCount))
	op = append(op, "@SP")
	op = append(op, "A=M")
	op = append(op, "M=-1")
	op = append(op, fmt.Sprintf("(LESSTHAN%d)", loopCount))
	op = append(op, "@SP")
	op = append(op, "AM=M+1")
	op = append(op, "@SP")
	op = append(op, "A=M")
	op = append(op, "M=0")
	loopCount++
	return op
}

func push(index uint) []string {
	op := []string{}
	op = append(op, fmt.Sprintf("@%d", index))
	op = append(op, "D=A")
	op = append(op, "@SP")
	op = append(op, "A=M")
	op = append(op, "M=D")
	op = append(op, "D=A+1")
	op = append(op, "@SP")
	op = append(op, "M=D")
	return op
}

func add() []string {
	op := []string{}
	op = append(op, "@SP")
	op = append(op, "D=D-1")
	op = append(op, "M=D")
	op = append(op, "A=D")
	op = append(op, "D=M")
	op = append(op, "A=A-1")
	op = append(op, "D=M+D")
	op = append(op, "M=D")
	op = append(op, "A=A+1")
	op = append(op, "M=0")
	return op
}

func sub() []string {
	op := []string{}
	op = append(op, "@SP")
	op = append(op, "D=D-1")
	op = append(op, "M=D")
	op = append(op, "A=D")
	op = append(op, "D=M")
	op = append(op, "A=A-1")
	op = append(op, "D=M-D")
	op = append(op, "M=D")
	op = append(op, "A=A+1")
	op = append(op, "M=0")
	return op
}

func eq() []string {
	op := []string{}
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=M")
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=M-D")
	op = append(op, fmt.Sprintf("@ISEQUAL%d", loopCount))
	op = append(op, "D;JEQ")
	op = append(op, "@SP")
	op = append(op, "A=M")
	op = append(op, "M=0")
	op = append(op, fmt.Sprintf("@ISNOTEQUAL%d", loopCount))
	op = append(op, "0;JMP")
	op = append(op, fmt.Sprintf("(ISEQUAL%d)", loopCount))
	op = append(op, "@SP")
	op = append(op, "A=M")
	op = append(op, "M=-1")
	op = append(op, fmt.Sprintf("(ISNOTEQUAL%d)", loopCount))
	op = append(op, "@SP")
	op = append(op, "AM=M+1")
	op = append(op, "@SP")
	op = append(op, "A=M")
	op = append(op, "M=0")
	loopCount++
	return op
}

func neg() []string {
	op := []string{}
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=M")
	op = append(op, "D=-D")
	op = append(op, "M=D")
	op = append(op, "@SP")
	op = append(op, "AM=M+1")
	return op
}

func pop() []string {
	op := []string{}
	op = append(op, "D=M")
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	return op
}

func openFile() *os.File {
	if len(os.Args) <= 1 {
		fmt.Println("No path to file provided: vmtranslator <filepath>")
		os.Exit(1)
	}

	file, error := os.Open(os.Args[1])

	if error != nil {
		fmt.Println("Could not open file")
		os.Exit(1)
	}

	return file
}

func parse(file *os.File) []Command {
	var commands = []Command{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var command, err = parseLine(scanner.Text())
		if err == nil {
			commands = append(commands, command)
		}
	}
	return commands
}

func parseLine(line string) (Command, error) {
	// remove comments && trim whitespace
	commentsRemoved := stripComment(line)
	trimmedLine := strings.TrimSpace(commentsRemoved)

	// ignore empty lines
	if len(trimmedLine) == 0 {
		return Command{}, errors.New("comment")
	}

	command := Command{}

	tokens := strings.Split(trimmedLine, " ")

	switch tokens[0] {
	case "add":
		command.commandType = C_ARITHMETIC
		command.command = "add"
		command.segment = ""
		command.index = 0
	case "sub":
		command.commandType = C_ARITHMETIC
		command.command = "sub"
	case "neg":
		command.commandType = C_ARITHMETIC
		command.command = "neg"
	case "push":
		command.commandType = C_PUSH
		command.segment = tokens[1]
		value, err := strconv.Atoi(tokens[2])
		if err == nil {
			command.index = uint(value)
		} else {
			fmt.Println("Error: Expecting uint value for index")
			os.Exit(1)
		}
	case "eq":
		command.commandType = C_ARITHMETIC
		command.command = "eq"
		command.segment = ""
		command.index = 0
	case "lt":
		command.commandType = C_ARITHMETIC
		command.command = "lt"
	case "gt":
		command.commandType = C_ARITHMETIC
		command.command = "gt"
	case "and":
		command.commandType = C_ARITHMETIC
		command.command = "and"
	case "or":
		command.commandType = C_ARITHMETIC
		command.command = "or"
	case "not":
		command.commandType = C_ARITHMETIC
		command.command = "not"
	}

	return command, nil
}

func stripComment(source string) string {
	if comment := strings.IndexAny(source, "///"); comment >= 0 {
		return strings.TrimRightFunc(source[:comment], unicode.IsSpace)
	}
	return source
}
