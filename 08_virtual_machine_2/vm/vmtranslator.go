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
var callCount = 0

func main() {
	files := openFiles()
	commands := parse(files)
	fmt.Println(commands)
	op := translateToAssembly(commands)
	writeFile(op)
}

func writeFile(op []string) {
	filename := strings.Replace(os.Args[1], ".vm", ".asm", -1)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

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
	var op []string
	for _, command := range stack {
		switch command.commandType {
		case C_PUSH:
			op = append(op, push(command.segment, command.index)...)
		case C_POP:
			op = append(op, pop(command.segment, command.index)...)
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
		case C_LABEL:
			op = append(op, label(command.segment)...)
		case C_IF:
			op = append(op, gotoIf(command.segment)...)
		case C_RETURN:
			op = append(op, returnFromFunc()...)
		case C_FUNCTION:
			op = append(op, function(command.segment, command.index)...)
		case C_CALL:
			op = append(op, call(command.segment, command.index)...)
		}
	}
	return op
}

func returnFromFunc() []string {
	var op []string
	// store Memory[LCL] in R13
	op = append(op, "@LCL")
	op = append(op, "D=M")
	op = append(op, "@R13")
	op = append(op, "M=D")
	// put return address in R14
	op = append(op, "@LCL")
	op = append(op, "D=M")
	op = append(op, "@5")
	op = append(op, "D=D-A")
	op = append(op, "@R14")
	op = append(op, "M=D")
	// Reposition the return value for the caller
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=M")
	op = append(op, "@ARG")
	op = append(op, "A=M")
	op = append(op, "M=D")
	// Restore SP of the caller (SP=ARG+1)
	op = append(op, "@ARG")
	op = append(op, "D=M")
	op = append(op, "D=D+1")
	op = append(op, "@SP")
	op = append(op, "M=D")
	// Restore THAT of the caller (Memory[R13]-1)
	op = append(op, "@R13")
	op = append(op, "D=M-1")
	op = append(op, "A=D")
	op = append(op, "D=M")
	op = append(op, "@THAT")
	op = append(op, "M=D")
	// Restore THIS of the caller (Memory[R13]-2)
	op = append(op, "@R13")
	op = append(op, "D=M-1")
	op = append(op, "D=D-1")
	op = append(op, "A=D")
	op = append(op, "D=M")
	op = append(op, "@THIS")
	op = append(op, "M=D")
	// Restore ARG of the caller (Memory[R13]-3)
	op = append(op, "@R13")
	op = append(op, "D=M-1")
	op = append(op, "D=D-1")
	op = append(op, "D=D-1")
	op = append(op, "A=D")
	op = append(op, "D=M")
	op = append(op, "@ARG")
	op = append(op, "M=D")
	// Restore LCL of the caller (Memory[R13]-4)
	op = append(op, "@R13")
	op = append(op, "D=M-1")
	op = append(op, "D=D-1")
	op = append(op, "D=D-1")
	op = append(op, "D=D-1")
	op = append(op, "A=D")
	op = append(op, "D=M")
	op = append(op, "@LCL")
	op = append(op, "M=D")
	// Goto return address (Memory[R14])
	op = append(op, "@R14")
	op = append(op, "A=M")
	op = append(op, "0;JMP")
	return op
}

func bootstrap() []string {
	var op []string
	// initialize stack pointer to Mem[256]
	op = append(op, "@256")
	op = append(op, "D=A")
	op = append(op, "@SP")
	op = append(op, "M=D")
	return op
}

func function(fn string, argc uint) []string {
	var op []string

	if fn == "Sys.init" {
		op = append(op, bootstrap()...)
	} else {
		// create function label
		op = append(op, fmt.Sprintf("(%s)", fn))
		// push args
		for i := uint(0); i < argc; i++ {
			op = append(op, push("LCL", 0)...)
		}
	}
	return op
}

func call(fn string, argc uint) []string {
	returnAddr := fmt.Sprintf("RETURN%d", callCount)
	var op []string
	op = append(op, fmt.Sprintf("@%s", returnAddr))
	op = append(op, push(returnAddr, 0)...)
	op = append(op, push("LCL", 0)...)
	op = append(op, push("ARG", 0)...)
	op = append(op, push("THIS", 0)...)
	op = append(op, push("THAT", 0)...)
	op = append(op, push("THAT", 0)...)
	// reposition ARG
	op = append(op, fmt.Sprintf("@%d", argc+5))
	op = append(op, "D=D-A")
	op = append(op, "@ARG")
	op = append(op, "M=D")
	// reposition LCL
	op = append(op, "@SP")
	op = append(op, "D=M")
	op = append(op, "@LCL")
	op = append(op, "M=D")
	op = append(op, gotoLabel(fn)...)
	op = append(op, fmt.Sprintf("(%s)", returnAddr))
	return op
}

func not() []string {
	var op []string
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
	var op []string
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
	var op []string
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
	var op []string
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
	var op []string
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

func push(segment string, index uint) []string {
	var op []string
	// segment is empty if a constant is being pushed
	if segment == "" || segment == "constant" {
		op = append(op, fmt.Sprintf("@%d", index))
		op = append(op, "D=A")
		op = append(op, "@SP")
		op = append(op, "A=M")
		op = append(op, "M=D")
		op = append(op, "D=A+1")
		op = append(op, "@SP")
		op = append(op, "M=D")
	} else {
		if segment == "LCL" {
			op = append(op, fmt.Sprintf("@%d", index))
			op = append(op, "D=A")
			op = append(op, "@SP")
			op = append(op, "A=M")
			op = append(op, "M=D")
			op = append(op, "@SP")
			op = append(op, "M=M+1")
		}
		if segment == "ARG" {
			op = append(op, "@ARG")
			op = append(op, "D=M")
			op = append(op, fmt.Sprintf("@%d", index))
			op = append(op, "D=D+A") // target argument address
			op = append(op, "A=D")   // load target arg address to A
			op = append(op, "D=M")   // load target arg value to D
			op = append(op, "@SP")
			op = append(op, "A=M")
			op = append(op, "M=D")
			op = append(op, "@SP")
			op = append(op, "AM=M+1") // increase SP
		}
	}

	return op
}

func pop(segment string, index uint) []string {
	var op []string
	if segment == "" {
		op = append(op, "@SP")
		op = append(op, "AM=M-1")
		op = append(op, fmt.Sprintf("@%d", index))
		op = append(op, "D=A")
		op = append(op, "@SP")
		op = append(op, "A=M")
		op = append(op, "M=0")
	}
	if segment == "pointer" {
		if index == 0 {
			op = append(op, "@SP")
			op = append(op, "AM=M-1")
			op = append(op, "D=M")
			op = append(op, "@THIS")
			op = append(op, "M=D")
			op = append(op, "@SP")
			op = append(op, "AM=M+1")
		} else if index == 1 {
			op = append(op, "@SP")
			op = append(op, "AM=M-1")
			op = append(op, "D=M")
			op = append(op, "@THAT")
			op = append(op, "M=D")
			op = append(op, "@SP")
			op = append(op, "AM=M+1")
		} else {
			panic("Index out of range for pointer (only 0 and 1 are valid)")
		}
	}
	if segment == "temp" {

	}
	return op
}

func add() []string {
	var op []string
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=M")
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=D+M")
	op = append(op, "M=D")
	op = append(op, "@SP")
	op = append(op, "AM=M+1")
	op = append(op, "M=0")
	return op
}

func sub() []string {
	var op []string
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=M")
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=M-D")
	op = append(op, "M=D")
	op = append(op, "@SP")
	op = append(op, "AM=M+1")
	op = append(op, "M=0")
	return op
}

func eq() []string {
	var op []string
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
	var op []string
	op = append(op, "@SP")
	op = append(op, "AM=M-1")
	op = append(op, "D=M")
	op = append(op, "D=-D")
	op = append(op, "M=D")
	op = append(op, "@SP")
	op = append(op, "AM=M+1")
	return op
}

func label(label string) []string {
	var op []string
	op = append(op, fmt.Sprintf("(%s)", label))
	return op
}

func gotoLabel(label string) []string {
	var op []string
	op = append(op, fmt.Sprintf("@%s", label))
	op = append(op, "0;JMP")
	return op
}

func gotoIf(label string) []string {
	var op []string
	op = append(op, pop("local", 0)...)
	op = append(op, fmt.Sprintf("@%s", label))
	op = append(op, "JNE;JMP")
	return op
}

func openFiles() []*os.File {
	if len(os.Args) <= 1 {
		fmt.Println("No path to file provided: vmtranslator <filepath>")
		os.Exit(1)
	}

	files := []*os.File{}
	for i,arg := range os.Args {
		if i > 0 {
			file, error := os.Open(arg)
			files = append(files, file)
			if error != nil {
					fmt.Println("Could not open file")
					os.Exit(1)
			}
		}
	}

	return files
}

func parse(files []*os.File) []Command {
	var commands []Command
	for _,file := range(files) {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var command, err = parseLine(scanner.Text())
			if err == nil {
				commands = append(commands, command)
			}
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
	case "function":
		command.commandType = C_FUNCTION
		command.command = "function"
		command.segment = tokens[1]
		index, _ := strconv.Atoi(tokens[2])
		command.index = uint(index)
	case "call":
		command.commandType = C_CALL
		command.command = "call"
		command.segment = tokens[1]
		index, _ := strconv.Atoi(tokens[2])
		command.index = uint(index)
	case "return":
		command.commandType = C_RETURN
		command.command = tokens[0]
		command.segment = ""
		command.index = 0
	case "label":
		command.commandType = C_LABEL
		command.command = tokens[0]
		command.segment = tokens[1]
		command.index = 0
	case "goto":
		command.commandType = C_GOTO
		command.command = tokens[0]
		command.segment = tokens[1]
		command.index = 0
	case "if-goto":
		command.commandType = C_IF
		command.command = tokens[0]
		command.segment = tokens[1]
		command.index = 0
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
		switch tokens[1] {
		case "constant":
			command.segment = "constant"
		case "local":
			command.segment = "LCL"
		case "argument":
			command.segment = "ARG"
		case "pointer":
			if tokens[2] == "0" {
				command.segment = "THIS"
			}
			if tokens[2] == "1" {
				command.segment = "THAT"
			}
		case "temp":

		}

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
