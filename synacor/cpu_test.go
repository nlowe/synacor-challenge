package synacor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRegister(t *testing.T) {
	t.Run("Memory is not a register", func(t *testing.T) {
		_, r := IsRegister(0)
		assert.False(t, r)

		_, r = IsRegister(32767)
		assert.False(t, r)
	})

	t.Run("Regular Registers", func(t *testing.T) {
		var addr uint16
		for addr = 32768; addr <= 32775; addr++ {
			id, r := IsRegister(addr)

			assert.True(t, r)
			assert.Equal(t, uint8(addr-32768), id)
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		_, r := IsRegister(32776)
		assert.False(t, r)
	})
}

func TestNewCPUFrom(t *testing.T) {
	c, _, err := NewCPUFrom(context.Background(), bytes.NewReader([]byte{0x01, 0x00, 0x02, 0x00, 0x03, 0x00}), nil)

	require.NoError(t, err)
	assert.Equal(t, uint16(1), c.memory[0])
	assert.Equal(t, uint16(2), c.memory[1])
	assert.Equal(t, uint16(3), c.memory[2])
	assert.Equal(t, uint16(0), c.memory[3])
}

func TestCPU_ExampleProgram(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	c, out, err := NewCPUFrom(ctx, bytes.NewReader([]byte{
		0x09, 0x00, // ADD
		0x00, 0x80, //     r0
		0x01, 0x80, //     r1
		0x04, 0x00, //     $4
		0x13, 0x00, // OUT
		0x00, 0x80, //     r0
	}), nil)
	require.NoError(t, err)

	go c.Run()
	require.Equal(t, rune(4), <-out)
}

func TestPOST(t *testing.T) {
	f, err := os.Open("../challenge.bin")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, f.Close())
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, out, err := NewCPUFrom(ctx, f, nil)
	require.NoError(t, err)

	go c.Run()

	buf := strings.Builder{}
	for r := range out {
		if r == '\n' {
			line := buf.String()
			if strings.Contains(line, "challenge website:") {
				// Not sure how secret the flags are, redact them just in case
				before, _, _ := strings.Cut(line, ":")
				fmt.Println(before+":", "***")
			} else {
				fmt.Println(line)
			}

			if strings.Contains(line, "self-test complete, all tests pass") {
				cancel()
				return
			}

			buf.Reset()
			continue
		}

		buf.WriteRune(r)
	}

	t.Error("EOF without successful POST")
}
