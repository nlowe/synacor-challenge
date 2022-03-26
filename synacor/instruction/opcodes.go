package instruction

import "fmt"

const (
	// HALT accepts No arguments. Stop execution and terminate the program.
	HALT uint16 = iota
	// SET accepts Two arguments. Set register <a> to the value of <b>.
	SET
	// PUSH accepts One Argument. Push <a> onto the stack.
	PUSH
	// POP accepts One Argument. Remove the top element from the stack and write it into <a>; empty stack = error.
	POP
	// EQ accepts Three Arguments. Set <a> to 1 if <b> is equal to <c>; set it to 0 otherwise.
	EQ
	// GT accepts Three Arguments. Set <a> to 1 if <b> is greater than <c>; set it to 0 otherwise.
	GT
	// JMP accepts One Argument. Jump to <a>.
	JMP
	// JT accepts Two Arguments. If <a> is nonzero, jump to <b>.
	JT
	// JF accepts Two Arguments. If <a> is zero, jump to <b>.
	JF
	// ADD accepts Three Arguments. Assign into <a> the sum of <b> and <c> (modulo 32768).
	ADD
	// MULT accepts Three Arguments. Store into <a> the product of <b> and <c> (modulo 32768).
	MULT
	// MOD accepts Three Arguments. Store into <a> the remainder of <b> divided by <c>.
	MOD
	// AND accepts Three Arguments. Stores into <a> the bitwise and of <b> and <c>.
	AND
	// OR accepts Three Arguments. Stores into <a> the bitwise or of <b> and <c>.
	OR
	// NOT accepts Two Arguments. Stores 15-bit bitwise inverse of <b> in <a>.
	NOT
	// RMEM accepts Two Arguments. Read memory at address <b> and write it to <a>.
	RMEM
	// WMEM accepts Two Arguments. write the value from <b> into memory at address <a>
	WMEM
	// CALL accepts One Argument. Write the address of the next instruction to the stack and jump to <a>.
	CALL
	// RET accepts No Arguments. Remove the top element from the stack and jump to it; empty stack = halt.
	RET
	// OUT accepts One Argument. Write the character represented by ascii code <a> to the terminal.
	OUT
	// IN accepts One Argument. Read a character from the terminal and write its ascii code to <a>; it can be assumed
	// that once input starts, it will continue until a newline is encountered; this means that you can safely read
	// whole lines from the keyboard and trust that they will be fully read
	IN
	// NOOP accepts No Arguments. No operation.
	NOOP
)

// ArgCount returns the number of arguments that need to be read in order to process the specified instruction
func ArgCount(instr uint16) uint16 {
	v, ok := map[uint16]uint16{
		HALT: 0,
		SET:  2,
		PUSH: 1,
		POP:  1,
		EQ:   3,
		GT:   3,
		JMP:  1,
		JT:   2,
		JF:   2,
		ADD:  3,
		MULT: 3,
		MOD:  3,
		AND:  3,
		OR:   3,
		NOT:  2,
		RMEM: 2,
		WMEM: 2,
		CALL: 1,
		RET:  1,
		OUT:  1,
		IN:   1,
		NOOP: 0,
	}[instr]

	if !ok {
		panic(fmt.Errorf("unknown instruction: %d", instr))
	}

	return v
}
