package synacor

import (
	"fmt"
	"time"

	"github.com/nlowe/synacor-challenge/synacor/instruction"
)

// The mod base that all arithmetic operations operate in
const base = 32768

type opFunc func(cpu *CPU, a, b, c uint16) uint16

var microcode = map[uint16]opFunc{
	// HALT accepts No arguments. Stop execution and terminate the program.
	instruction.HALT: func(cpu *CPU, _, _, _ uint16) uint16 {
		cpu.debugState("HALT")

		cpu.shutdown()
		return 0
	},
	// SET accepts Two arguments. Set register <a> to the value of <b>.
	instruction.SET: func(cpu *CPU, a, b, _ uint16) uint16 {
		cpu.debugState("SET ", debugArg(a), debugArg(b))

		cpu.write(a, cpu.registerOrLiteral(b))
		return 1 + instruction.ArgCount(instruction.SET)
	},
	// PUSH accepts One Argument. Push <a> onto the stack.
	instruction.PUSH: func(cpu *CPU, a, _, _ uint16) uint16 {
		cpu.debugState("PUSH", debugArg(a))

		cpu.stack = append(cpu.stack, cpu.registerOrLiteral(a))
		return 1 + instruction.ArgCount(instruction.PUSH)
	},
	// POP accepts One Argument. Remove the top element from the stack and write it into <a>; empty stack = error.
	instruction.POP: func(cpu *CPU, a, _, _ uint16) uint16 {
		cpu.debugState("POP ", debugArg(a))

		cpu.write(a, cpu.popstack())

		return 1 + instruction.ArgCount(instruction.POP)
	},
	// EQ accepts Three Arguments. Set <a> to 1 if <b> is equal to <c>; set it to 0 otherwise.
	instruction.EQ: func(cpu *CPU, a, b, c uint16) uint16 {
		cpu.debugState("EQ  ", debugArg(a), debugArg(b), debugArg(c))

		var result uint16
		if cpu.registerOrLiteral(b) == cpu.registerOrLiteral(c) {
			result = 1
		}

		cpu.write(a, result)
		return 1 + instruction.ArgCount(instruction.EQ)
	},
	// GT accepts Three Arguments. Set <a> to 1 if <b> is greater than <c>; set it to 0 otherwise.
	instruction.GT: func(cpu *CPU, a, b, c uint16) uint16 {
		cpu.debugState("GT  ", debugArg(a), debugArg(b), debugArg(c))

		var result uint16
		if cpu.registerOrLiteral(b) > cpu.registerOrLiteral(c) {
			result = 1
		}

		cpu.write(a, result)
		return 1 + instruction.ArgCount(instruction.GT)
	},
	// JMP accepts One Argument. Jump to <a>.
	instruction.JMP: func(cpu *CPU, a, _, _ uint16) uint16 {
		cpu.debugState("JMP ", debugArg(a))

		cpu.pc = cpu.registerOrLiteral(a)

		// Don't increment PC
		return 0
	},
	// JT accepts Two Arguments. If <a> is nonzero, jump to <b>.
	instruction.JT: func(cpu *CPU, a, b, _ uint16) uint16 {
		cpu.debugState("JT  ", debugArg(a), debugArg(b))

		if cpu.registerOrLiteral(a) != 0 {
			cpu.pc = cpu.registerOrLiteral(b)
			return 0
		}

		return 1 + instruction.ArgCount(instruction.JT)
	},
	// JF accepts Two Arguments. If <a> is zero, jump to <b>.
	instruction.JF: func(cpu *CPU, a, b, _ uint16) uint16 {
		cpu.debugState("JF  ", debugArg(a), debugArg(b))

		if cpu.registerOrLiteral(a) == 0 {
			cpu.pc = cpu.registerOrLiteral(b)
			return 0
		}

		return 1 + instruction.ArgCount(instruction.JF)
	},
	// ADD accepts Three Arguments. Assign into <a> the sum of <b> and <c> (modulo 32768).
	instruction.ADD: func(cpu *CPU, a, b, c uint16) uint16 {
		cpu.debugState("ADD ", debugArg(a), debugArg(b), debugArg(c))

		cpu.write(a, (cpu.registerOrLiteral(b)+cpu.registerOrLiteral(c))%base)
		return 1 + instruction.ArgCount(instruction.ADD)
	},
	// MULT accepts Three Arguments. Store into <a> the product of <b> and <c> (modulo 32768).
	instruction.MULT: func(cpu *CPU, a, b, c uint16) uint16 {
		cpu.debugState("MULT", debugArg(a), debugArg(b), debugArg(c))

		cpu.write(a, (cpu.registerOrLiteral(b)*cpu.registerOrLiteral(c))%base)
		return 1 + instruction.ArgCount(instruction.MULT)
	},
	// MOD accepts Three Arguments. Store into <a> the remainder of <b> divided by <c>.
	instruction.MOD: func(cpu *CPU, a, b, c uint16) uint16 {
		cpu.debugState("MOD ", debugArg(a), debugArg(b), debugArg(c))

		cpu.write(a, cpu.registerOrLiteral(b)%cpu.registerOrLiteral(c))
		return 1 + instruction.ArgCount(instruction.MOD)
	},
	// AND accepts Three Arguments. Stores into <a> the bitwise and of <b> and <c>.
	instruction.AND: func(cpu *CPU, a, b, c uint16) uint16 {
		cpu.debugState("AND ", debugArg(a), debugArg(b), debugArg(c))

		cpu.write(a, cpu.registerOrLiteral(b)&cpu.registerOrLiteral(c))
		return 1 + instruction.ArgCount(instruction.AND)
	},
	// OR accepts Three Arguments. Stores into <a> the bitwise or of <b> and <c>.
	instruction.OR: func(cpu *CPU, a, b, c uint16) uint16 {
		cpu.debugState("OR ", debugArg(a), debugArg(b), debugArg(c))

		cpu.write(a, cpu.registerOrLiteral(b)|cpu.registerOrLiteral(c))
		return 1 + instruction.ArgCount(instruction.OR)
	},
	// NOT accepts Two Arguments. Stores 15-bit bitwise inverse of <b> in <a>.
	instruction.NOT: func(cpu *CPU, a, b, _ uint16) uint16 {
		cpu.debugState("NOT ", debugArg(a), debugArg(b))

		cpu.write(a, (^cpu.registerOrLiteral(b))&0x7FFF)
		return 1 + instruction.ArgCount(instruction.NOT)
	},
	// RMEM accepts Two Arguments. Read memory at address <b> and write it to <a>.
	// This is different from SET because if <b> is a register, the value of the
	// address pointed at by <b> is written to <a> instead of the register value.
	instruction.RMEM: func(cpu *CPU, a, b, _ uint16) uint16 {
		cpu.debugState("RMEM", debugArg(a), debugArg(b))

		if reg, ok := IsRegister(b); ok {
			b = cpu.registers[reg]
		}

		cpu.write(a, cpu.memory[b])
		return 1 + instruction.ArgCount(instruction.RMEM)
	},
	// WMEM accepts Two Arguments. write the value from <b> into memory at address <a>
	// This is different from SET because if <a> is a register, <b> is written to the
	// address pointed at by <a> instead of the register.
	instruction.WMEM: func(cpu *CPU, a, b, _ uint16) uint16 {
		cpu.debugState("WMEM", debugArg(a), debugArg(b))

		cpu.memory[cpu.registerOrLiteral(a)] = cpu.registerOrLiteral(b)
		return 1 + instruction.ArgCount(instruction.WMEM)
	},
	// CALL accepts One Argument. Write the address of the next instruction to the stack and jump to <a>.
	instruction.CALL: func(cpu *CPU, a, _, _ uint16) uint16 {
		cpu.debugState("CALL", debugArg(a))

		cpu.stack = append(cpu.stack, cpu.pc+2)
		cpu.pc = cpu.registerOrLiteral(a)

		return 0
	},
	// RET accepts No Arguments. Remove the top element from the stack and jump to it; empty stack = halt.
	instruction.RET: func(cpu *CPU, _, _, _ uint16) uint16 {
		cpu.debugState("RET ")

		if len(cpu.stack) == 0 {
			cpu.shutdown()
			return 0
		}

		cpu.pc = cpu.popstack()
		return 0
	},
	// OUT accepts One Argument. Write the character represented by ascii code <a> to the terminal.
	instruction.OUT: func(cpu *CPU, a, _, _ uint16) uint16 {
		cpu.debugState("OUT ", cpu.debugArgAscii(a))

		select {
		case cpu.out <- rune(cpu.registerOrLiteral(a)):
		case <-cpu.ctx.Done():
			return 0
		case <-time.After(cpu.WatchdogTimeout):
			panic(fmt.Sprintf("blocked on write after %s", cpu.WatchdogTimeout.String()))
		}

		return 1 + instruction.ArgCount(instruction.OUT)
	},
	// IN accepts One Argument. Read a character from the terminal and write its ascii code to <a>; it can be assumed
	// that once input starts, it will continue until a newline is encountered; this means that you can safely read
	// whole lines from the keyboard and trust that they will be fully read
	instruction.IN: func(cpu *CPU, a, _, _ uint16) uint16 {
		cpu.debugState("IN  ", debugArg(a))

		select {
		case v := <-cpu.in:
			cpu.write(a, uint16(v))
		case <-cpu.ctx.Done():
			return 0
		case <-time.After(cpu.WatchdogTimeout):
			panic(fmt.Sprintf("blocked on read after %s", cpu.WatchdogTimeout.String()))
		}

		return 1 + instruction.ArgCount(instruction.IN)
	},
	// NOOP accepts No Arguments. No operation.
	instruction.NOOP: func(cpu *CPU, _, _, _ uint16) uint16 {
		cpu.debugState("NOOP")
		return 1 + instruction.ArgCount(instruction.NOOP)
	},
}
