package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	MemorySize = 30000
)

var debugMode bool

type TokenType int

const (
	INC TokenType = iota
	DEC
	PRINT
	READ
	PTR_RIGHT
	PTR_LEFT
	LOOP_START
	LOOP_END
	UNKNOWN
)

type Token struct {
	Type  TokenType
	Value string
}

const (
	runeBeh = '\u0628'
	runeAin = '\u0639'
)

func lexAndParse(source string) ([]Token, error) {
	var tokens []Token
	runes := []rune(source)
	i := 0
	loopBalance := 0

	if debugMode {
		fmt.Fprintln(os.Stderr, "--- Lexing/Parsing Start ---")
	}
	for i < len(runes) {
		matched := false
		startIdx := i

		if i+5 < len(runes) && runes[i] == runeBeh && runes[i+1] == runeAin && runes[i+2] == runeAin && runes[i+3] == runeAin && runes[i+4] == runeAin && runes[i+5] == runeAin {
			tokens = append(tokens, Token{Type: PTR_LEFT, Value: "بععععع"})
			i += 6
			matched = true
		} else if i+5 < len(runes) && runes[i] == runeBeh && runes[i+1] == runeAin && runes[i+2] == runeBeh && runes[i+3] == runeAin && runes[i+4] == runeBeh && runes[i+5] == runeAin {
			tokens = append(tokens, Token{Type: LOOP_END, Value: "بعبعبع"})
			loopBalance--
			i += 6
			matched = true
		} else if i+4 < len(runes) && runes[i] == runeBeh && runes[i+1] == runeAin && runes[i+2] == runeAin && runes[i+3] == runeAin && runes[i+4] == runeAin {
			tokens = append(tokens, Token{Type: PTR_RIGHT, Value: "بعععع"})
			i += 5
			matched = true
		} else if i+4 < len(runes) && runes[i] == runeBeh && runes[i+1] == runeAin && runes[i+2] == runeBeh && runes[i+3] == runeAin && runes[i+4] == runeAin {
			tokens = append(tokens, Token{Type: LOOP_START, Value: "بعبعع"})
			loopBalance++
			i += 5
			matched = true
		} else if i+3 < len(runes) && runes[i] == runeBeh && runes[i+1] == runeAin && runes[i+2] == runeBeh && runes[i+3] == runeAin {
			tokens = append(tokens, Token{Type: READ, Value: "بعبع"})
			i += 4
			matched = true
		} else if i+3 < len(runes) && runes[i] == runeBeh && runes[i+1] == runeAin && runes[i+2] == runeAin && runes[i+3] == runeAin {
			tokens = append(tokens, Token{Type: DEC, Value: "بععع"})
			i += 4
			matched = true
		} else if i+2 < len(runes) && runes[i] == runeBeh && runes[i+1] == runeAin && runes[i+2] == runeAin {
			tokens = append(tokens, Token{Type: PRINT, Value: "بعع"})
			i += 3
			matched = true
		} else if i+1 < len(runes) && runes[i] == runeBeh && runes[i+1] == runeAin {
			tokens = append(tokens, Token{Type: INC, Value: "بع"})
			i += 2
			matched = true
		}

		if !matched {
			if debugMode {
				fmt.Fprintf(os.Stderr, "DEBUG: Skipping unrecognized character at index %d: '%c' (rune value: %d)\n", startIdx, runes[startIdx], runes[startIdx])
			}
			i++
		} else {
			if debugMode {
				fmt.Fprintf(os.Stderr, "DEBUG: Matched token: %v at index %d\n", tokens[len(tokens)-1], startIdx)
			}
		}
	}
	if debugMode {
		fmt.Fprintln(os.Stderr, "--- Lexing/Parsing End ---")
		fmt.Fprintf(os.Stderr, "DEBUG: Total tokens generated: %d\n", len(tokens))
	}

	if loopBalance != 0 {
		return nil, fmt.Errorf("unmatched loop brackets: balance is %d", loopBalance)
	}

	return tokens, nil
}

func interpret(tokens []Token) error {
	memory := make([]byte, MemorySize)
	dataPtr := 0
	tokenPtr := 0

	loopJumps := make(map[int]int)
	loopStack := []int{}

	for i, token := range tokens {
		if token.Type == LOOP_START {
			loopStack = append(loopStack, i)
		} else if token.Type == LOOP_END {
			if len(loopStack) == 0 {
				return fmt.Errorf("unmatched loop end at token %d", i)
			}
			startIdx := loopStack[len(loopStack)-1]
			loopStack = loopStack[:len(loopStack)-1]
			loopJumps[startIdx] = i
			loopJumps[i] = startIdx
		}
	}

	if len(loopStack) != 0 {
		return fmt.Errorf("unmatched loop start(s) remaining")
	}

	reader := bufio.NewReader(os.Stdin)

	if debugMode {
		fmt.Fprintln(os.Stderr, "--- Interpretation Start ---")
	}
	for tokenPtr < len(tokens) {
		token := tokens[tokenPtr]
		if debugMode {
			fmt.Fprintf(os.Stderr, "DEBUG: Executing token %d: %v, dataPtr: %d, memory[dataPtr]: %d\n", tokenPtr, token, dataPtr, memory[dataPtr])
		}

		switch token.Type {
		case INC:
			memory[dataPtr]++
		case DEC:
			memory[dataPtr]--
		case PTR_RIGHT:
			dataPtr++
			if dataPtr >= MemorySize {
				return fmt.Errorf("data pointer out of bounds (right) at instruction %d", tokenPtr)
			}
		case PTR_LEFT:
			dataPtr--
			if dataPtr < 0 {
				return fmt.Errorf("data pointer out of bounds (left) at instruction %d", tokenPtr)
			}
		case PRINT:
			fmt.Printf("%c", memory[dataPtr])
			if debugMode {
				fmt.Fprintf(os.Stderr, "DEBUG: Printed char: '%c' (ASCII: %d)\n", memory[dataPtr], memory[dataPtr])
			}
		case READ:
			char, _, err := reader.ReadRune()
			if err != nil {
				return fmt.Errorf("error reading input at instruction %d: %w", tokenPtr, err)
			}
			memory[dataPtr] = byte(char)
			if debugMode {
				fmt.Fprintf(os.Stderr, "DEBUG: Read char: '%c' (ASCII: %d)\n", char, byte(char))
			}
		case LOOP_START:
			if memory[dataPtr] == 0 {
				if debugMode {
					fmt.Fprintf(os.Stderr, "DEBUG: Loop start, memory[dataPtr] is 0. Jumping from %d to %d\n", tokenPtr, loopJumps[tokenPtr])
				}
				tokenPtr = loopJumps[tokenPtr]
			}
		case LOOP_END:
			if memory[dataPtr] != 0 {
				if debugMode {
					fmt.Fprintf(os.Stderr, "DEBUG: Loop end, memory[dataPtr] is not 0. Jumping from %d to %d\n", tokenPtr, loopJumps[tokenPtr])
				}
				tokenPtr = loopJumps[tokenPtr]
			}
		case UNKNOWN:
			return fmt.Errorf("unknown token encountered at instruction %d", tokenPtr)
		}
		tokenPtr++
	}
	if debugMode {
		fmt.Fprintln(os.Stderr, "--- Interpretation End ---")
	}
	return nil
}

func main() {
	flag.BoolVar(&debugMode, "debug", false, "Enable debug output")
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("Usage: go run esolang.go [OPTIONS] <filename_or_code>")
		fmt.Println("Example: go run esolang.go -debug myprogram.eso")
		fmt.Println("Example: go run esolang.go \"بع بع بع بع بع بع بع بع بع بعع\"")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var sourceCode string
	filePathOrCode := args[0]

	if _, err := os.Stat(filePathOrCode); err == nil {
		data, err := os.ReadFile(filePathOrCode)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", filePathOrCode, err)
			os.Exit(1)
		}
		sourceCode = string(data)
	} else if strings.HasPrefix(filePathOrCode, "بع") {
		sourceCode = filePathOrCode
	} else {
		fmt.Printf("Error: '%s' is not a file and does not appear to be source code.\n", filePathOrCode)
		os.Exit(1)
	}

	if debugMode {
		fmt.Fprintf(os.Stderr, "DEBUG: Raw source code received: '%s'\n", sourceCode)
	}

	tokens, err := lexAndParse(sourceCode)
	if err != nil {
		fmt.Printf("Lexing/Parsing Error: %v\n", err)
		os.Exit(1)
	}

	err = interpret(tokens)
	if err != nil {
		fmt.Printf("Runtime Error: %v\n", err)
	}
	fmt.Println()
}
