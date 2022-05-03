package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"unicode"
)

var debugMode = false
var passes = 0
var address = 0
var variableAddress = 16

var predefSymbols = []string{"SP", "LCL", "ARG", "THIS", "THAT", "SCREEN", "KBD", "R0", "R1", "R2", "R3", "R4", "R5", "R6", "R7", "R8", "R9", "R10", "R11", "R12", "R13", "R14", "R15"}

var symbolTable map[string]int = map[string]int{
	"SP":     0,
	"LCL":    1,
	"ARG":    2,
	"THIS":   3,
	"THAT":   4,
	"R0":     0,
	"R1":     1,
	"R2":     2,
	"R3":     3,
	"R4":     4,
	"R5":     5,
	"R6":     6,
	"R7":     7,
	"R8":     8,
	"R9":     9,
	"R10":    10,
	"R11":    11,
	"R12":    12,
	"R13":    13,
	"R14":    14,
	"R15":    15,
	"SCREEN": 16384,
	"KBD":    24576,
}

type CommandType int

const (
	A_COMMAND CommandType = iota
	C_COMMAND
	L_COMMAND
)

type Command struct {
	commandType CommandType
	symbol      string
	dest        string
	comp        string
	jmp         string
	value       int
}

func main() {
	file := openFile()
	commands := parse(file)
	translateToMachineCode(commands)
}

/*
 * Parse .asm file in two passes.
 */
func parse(file *os.File) []Command {
	var commands = []Command{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var command = parseLine(scanner.Text())
		if command != (Command{}) {
			commands = append(commands, command)
		}
	}
	for index := range commands {
		command := commands[index]
		if command.commandType == A_COMMAND {
			value, exists := symbolTable[command.symbol]

			if isPredefSymbol(command.symbol) {
				commands[index].value = value
			} else {
				if exists {
					commands[index].value = value
				} else {
					// @<address> e.g. @123
					if address, err := strconv.Atoi(commands[index].symbol); err == nil {
						commands[index].value = address
					} else {
						//@<variable> e.g. @var
						commands[index].value = variableAddress
						symbolTable[command.symbol] = variableAddress
						variableAddress++
					}
				}
			}
		}
	}
	outputSymbolTable()

	return commands
}

/*
 * Output symbol table for debugging purposes
 */
func outputSymbolTable() {
	if debugMode {
		fmt.Println("=== Symbol Table ===")
		tw := new(tabwriter.Writer)
		tw.Init(os.Stdout, 8, 8, 0, '\t', 0)
		defer tw.Flush()

		fmt.Fprintf(tw, "\n %s\t%s\t", "Symbol", "Address")
		fmt.Fprintf(tw, "\n %s\t%s\t", "----", "----")

		for symbol, address := range symbolTable {
			fmt.Fprintf(tw, "\n %s\t%d\t", symbol, address)
		}
	}
}

/*
 * Translate parsed commands into Hack machine code, represented as a string of binary digits,
 * printed to STDOUT.
 */
func translateToMachineCode(commands []Command) {
	for _, command := range commands {
		if command.commandType == A_COMMAND {
			binary := int64(command.value)
			aCommand := fmt.Sprintf("0%s", strconv.FormatInt(binary, 2))

			fmt.Println(fillZeros(aCommand))
		}
		if command.commandType == C_COMMAND {
			var comp = "0000000"
			switch command.comp {
			case "0":
				comp = "0101010"
			case "1":
				comp = "0111111"
			case "-1":
				comp = "0111010"
			case "D":
				comp = "0001100"
			case "A":
				comp = "0110000"
			case "!D":
				comp = "0001101"
			case "!A":
				comp = "0110001"
			case "-D":
				comp = "0001111"
			case "-A":
				comp = "0110011"
			case "D+1":
				comp = "0011111"
			case "A+1":
				comp = "0110111"
			case "D-1":
				comp = "0001110"
			case "A-1":
				comp = "0110010"
			case "D+A":
				comp = "0000010"
			case "D-A":
				comp = "0010011"
			case "A-D":
				comp = "0000111"
			case "D&A":
				comp = "0000000"
			case "D|A":
				comp = "0010101"
			case "M":
				comp = "1110000"
			case "!M":
				comp = "1110001"
			case "-M":
				comp = "1110011"
			case "M+1":
				comp = "1110111"
			case "M-1":
				comp = "1110010"
			case "D+M":
				comp = "1000010"
			case "D-M":
				comp = "1010011"
			case "M-D":
				comp = "1000111"
			case "D&M":
				comp = "1000000"
			case "D|M":
				comp = "1010101"
			default:
				comp = "xxxxxxx"
			}
			var dest = "000"
			switch command.dest {
			case "":
				dest = "000"
			case "M":
				dest = "001"
			case "D":
				dest = "010"
			case "MD":
				dest = "011"
			case "A":
				dest = "100"
			case "AM":
				dest = "101"
			case "AD":
				dest = "110"
			case "AMD":
				dest = "111"
			}
			var jmp = "000"
			switch command.jmp {
			case "":
				jmp = "000"
			case "JGT":
				jmp = "001"
			case "JEQ":
				jmp = "010"
			case "JGE":
				jmp = "011"
			case "JLT":
				jmp = "100"
			case "JNE":
				jmp = "101"
			case "JLE":
				jmp = "110"
			case "JMP":
				jmp = "111"
			}
			cCommand := fmt.Sprintf("111%s%s%s", comp, dest, jmp)
			fmt.Println(cCommand)
		}
	}
}

/*
 * Takes a string representation of a binary number and fixes the length to 16,
 * pre-appending 0s.
 */
func fillZeros(binaryStr string) string {
	verb := fmt.Sprintf("%%%d.%ds", 16, 16)
	resultWithSpaces := fmt.Sprintf(verb, binaryStr)
	return strings.Replace(resultWithSpaces, " ", "0", 16)
}

func parseLine(line string) Command {
	// remove comments && trim whitespace
	commentsRemoved := stripComment(line)
	trimmedLine := strings.TrimSpace(commentsRemoved)

	// ignore empty lines
	if len(trimmedLine) == 0 {
		return Command{}
	}

	var commandType CommandType
	var value = 0
	if trimmedLine[0] == '@' {
		commandType = A_COMMAND
	} else if trimmedLine[0] == '(' {
		commandType = L_COMMAND
	} else {
		commandType = C_COMMAND
	}

	var symbol = ""
	if commandType == A_COMMAND {
		symbol = trimmedLine[1:]
	}
	if commandType == L_COMMAND {
		symbol = trimmedLine[1 : len(trimmedLine)-1]
		symbolTable[symbol] = address
	}

	var dest = ""
	var comp = ""
	var jmp = ""

	if commandType == C_COMMAND {
		dest, comp, jmp = parseCCommand(trimmedLine)
	}

	command := Command{commandType: commandType, symbol: symbol, dest: dest, comp: comp, jmp: jmp, value: value}

	if command.commandType != L_COMMAND {
		address++
	}

	passes++

	return command
}

func isPredefSymbol(symbol string) bool {
	for _, predefSymbol := range predefSymbols {
		if predefSymbol == symbol {
			return true
		}
	}
	return false
}

/*
 * Parse components of C-Command (dest=comp;jmp)
 */
func parseCCommand(cCommand string) (string, string, string) {
	var dest = ""
	var comp = ""
	var jmp = ""

	if strings.Contains(cCommand, ";") {
		var cCommandParts = strings.Split(cCommand, ";")
		jmp = cCommandParts[1]
		if strings.Contains(cCommandParts[0], "=") {
			var leftSide = strings.Split(cCommandParts[0], "=")
			dest = leftSide[0]
			comp = leftSide[1]
		} else {
			comp = strings.Split(cCommand, ";")[0]
		}
	} else {
		var leftSide = strings.Split(cCommand, "=")
		dest = leftSide[0]
		comp = leftSide[1]
	}
	return dest, comp, jmp
}

func stripComment(source string) string {
	if comment := strings.IndexAny(source, "///"); comment >= 0 {
		return strings.TrimRightFunc(source[:comment], unicode.IsSpace)
	}
	return source
}

func openFile() *os.File {
	if len(os.Args) <= 1 {
		fmt.Println("No path to file provided: hackasm <filepath>")
		os.Exit(1)
	}

	file, error := os.Open(os.Args[1])

	if error != nil {
		fmt.Println("Could not open file")
		os.Exit(1)
	}

	return file
}
