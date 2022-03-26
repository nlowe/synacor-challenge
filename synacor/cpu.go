package synacor

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	registerMask      = 0b111
	registerIndicator = 0b1000000000000
)

// IsRegister returns the register index and true iff the provided address is a register. That is,
// the most significant bit is set and the next 11 bits are unset
func IsRegister(addr uint16) (uint8, bool) {
	return uint8(addr & registerMask), (addr >> 3) == registerIndicator
}

// CPU implements tye Synacor Challenge VM spec:
//
// * 15-bit address space, each address points to a 16-bit little-endian value
// * 8 general-purpose registers
// * Reset Vector at address 0
type CPU struct {
	WatchdogTimeout time.Duration
	Debug           io.Writer

	memory    [1 << 15]uint16
	registers [8]uint16
	stack     []uint16

	pc uint16

	in  <-chan rune
	out chan<- rune

	ctx    context.Context
	halt   context.CancelFunc
	halted bool
}

// NewCPUFrom bootstraps a new CPU with initial memory loaded from the provided reader
func NewCPUFrom(ctx context.Context, r io.Reader, in <-chan rune) (*CPU, <-chan rune, error) {
	out := make(chan rune)

	ctx, cancel := context.WithCancel(ctx)

	c := &CPU{
		in:  in,
		out: out,

		ctx:  ctx,
		halt: cancel,

		WatchdogTimeout: 5 * time.Second,
	}

	// We can't just use binary.Read since that would expect to read all of memory
	// Read values one at a time

	var v uint16
	var err error
	for err == nil {
		if err = binary.Read(r, binary.LittleEndian, &v); err == nil {
			c.memory[c.pc] = v
			c.pc++
		}
	}

	// EOF is fine, expected even
	if err == io.EOF {
		err = nil
	}

	// Halt on context cancel in case someone other than the CPU cancels the context
	go func() {
		<-ctx.Done()
		c.shutdown()
	}()

	c.pc = 0
	return c, out, err
}

// registerOrLiteral converts the provided value into the value of the specified register, if it is one. Otherwise, it
// returns the value unmodified.
func (c *CPU) registerOrLiteral(v uint16) uint16 {
	if reg, ok := IsRegister(v); ok {
		return c.registers[reg]
	}

	return v
}

// Write stores the specified value at the provided address (either a register or memory)
func (c *CPU) write(addr, value uint16) {
	if reg, ok := IsRegister(addr); ok {
		c.registers[reg] = value
		return
	}

	c.memory[addr] = value
}

func (c *CPU) popstack() uint16 {
	tail := len(c.stack) - 1
	if tail < 0 {
		panic("stack underflow")
	}

	var v uint16
	v, c.stack = c.stack[tail], c.stack[:tail]

	return v
}

func (c *CPU) shutdown() {
	if !c.halted {
		close(c.out)
	}

	c.halted = true
	c.halt()
}

// Run processes instructions until a HALT is executed or the context of the CPU is cancelled
func (c *CPU) Run() {
	for !c.halted {
		c.Step()
	}
}

// Step executes a single instruction
func (c *CPU) Step() {
	op := c.memory[c.pc]
	impl, ok := microcode[op]
	if !ok {
		panic(fmt.Sprintf("unknown opcode %d", op))
	}

	c.pc += impl(c, c.memory[c.pc+1], c.memory[c.pc+2], c.memory[c.pc+3])
}

// debugStack returns the string representation of the stack slice, left is top-of-stack
func (c *CPU) debugStack() string {
	result := strings.Builder{}
	result.WriteString("[")

	for i := len(c.stack) - 1; i >= 0; i-- {
		result.WriteString(strconv.Itoa(int(c.stack[i])))
		if i != 0 {
			result.WriteString(" ")
		}
	}

	result.WriteString("]")

	return result.String()
}

func (c *CPU) debugState(op string, args ...string) {
	if c.Debug == nil {
		return
	}

	_, _ = fmt.Fprintf(c.Debug, "pc:%5d r0:%5d r1:%5d r2:%5d r3:%5d r4:%5d r5:%5d r6:%5d r7:%5d %s %s %s\n", c.pc, c.registers[0], c.registers[1], c.registers[2], c.registers[3], c.registers[4], c.registers[5], c.registers[6], c.registers[7], op, strings.Join(args, " "), c.debugStack())
}

func debugArg(arg uint16) string {
	if r, ok := IsRegister(arg); ok {
		return fmt.Sprintf("   r%d", r)
	}

	return fmt.Sprintf("%5d", arg)
}

func (c *CPU) debugArgAscii(arg uint16) string {
	if r, ok := IsRegister(arg); ok {
		return fmt.Sprintf("r%d->%s", r, string(rune(c.registers[r])))
	}

	return string(rune(arg))
}
