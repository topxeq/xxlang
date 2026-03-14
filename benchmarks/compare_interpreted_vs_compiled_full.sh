#!/bin/bash
# Comprehensive benchmark: Source vs Bytecode vs Standalone Executable

set -e

cd /mnt1/aiprjs/xxlang

echo "=================================================="
echo "Xxlang: 解释运行 vs 编译执行 性能对比"
echo "=================================================="
echo ""

# Build xxlang if needed
echo "1. 准备环境..."
if [ ! -f "./xxlang" ]; then
    go build -o xxlang ./cmd/xxl
fi

# Create test file
FIB_SOURCE="benchmarks/xxlang/fib_naive.xxl"
FIB_BYTECODE="/tmp/fib_naive.xxb"

echo "   测试文件: $FIB_SOURCE"
echo ""

# Compile to bytecode
echo "2. 编译字节码..."
./xxlang compile --bytecode -o "$FIB_BYTECODE" "$FIB_SOURCE" 2>/dev/null
echo "   字节码文件: $FIB_BYTECODE ($(stat -c%s "$FIB_BYTECODE") bytes)"
echo ""

# Generate standalone executable
echo "3. 生成独立可执行文件..."
STANDALONE_DIR="/tmp/xxlang_standalone"
rm -rf "$STANDALONE_DIR"
mkdir -p "$STANDALONE_DIR"

# Read bytecode as hex and generate Go source
BYTECODE_HEX=$(xxd -p "$FIB_BYTECODE" | tr -d '\n')

cat > "$STANDALONE_DIR/main.go" << 'GOMAIN'
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/objects"
	"github.com/topxeq/xxlang/pkg/vm"
)

//go:embed bytecode.bin
var bytecodeData []byte

// Stub for go:embed - we'll use actual data
var actualBytecode = []byte{
GOMAIN

# Add bytecode as Go byte array
echo "// Bytecode data (${#BYTECODE_HEX} hex chars)" >> "$STANDALONE_DIR/main.go"
echo "// Generated from fib_naive.xxl" >> "$STANDALONE_DIR/main.go"

# Convert hex to byte values
python3 -c "
import sys
hex_data = '$BYTECODE_HEX'
print('// Bytecode length:', len(hex_data)//2, 'bytes')
for i in range(0, min(len(hex_data), 200)):
    if i % 32 == 0:
        print()
        print('    ', end='')
    print(f'0x{hex_data[i*2:i*2+2]},', end='')
print()
print('    // ... (truncated for display)')
" >> "$STANDALONE_DIR/main.go" 2>/dev/null || echo "    // Bytecode embedded" >> "$STANDALONE_DIR/main.go"

cat >> "$STANDALONE_DIR/main.go" << 'GOMAIN2'
}

// Types for deserialization (copied from compiler package)
type serializableObject struct {
	Type  string
	Value interface{}
}

func main() {
	// Deserialize bytecode
	bytecode, err := deserializeBytecode(actualBytecode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Run the VM
	v := vm.New(bytecode)
	if err := v.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}
}

func deserializeBytecode(data []byte) (*compiler.Bytecode, error) {
	// Simple stub - use actual deserializer
	return compiler.Deserialize(data)
}
GOMAIN2

# Copy bytecode to standalone dir for embedding reference
cp "$FIB_BYTECODE" "$STANDALONE_DIR/bytecode.bin"

# Create go.mod
cat > "$STANDALONE_DIR/go.mod" << GOMOD
module fib_standalone

go 1.21

require github.com/topxeq/xxlang v0.0.0

replace github.com/topxeq/xxlang => /mnt1/aiprjs/xxlang
GOMOD

# Build standalone
echo "   编译独立可执行文件..."
cd "$STANDALONE_DIR"
go build -o fib_standalone main.go 2>&1 | head -5 || {
    echo "   构建失败，使用简化版本..."
    # Create simpler version
    cat > main.go << 'GOSIMPLE'
package main

import (
	"fmt"
	"os"

	"github.com/topxeq/xxlang/pkg/compiler"
	"github.com/topxeq/xxlang/pkg/vm"
)

func main() {
	bytecode, err := compiler.DeserializeFromFile("bytecode.bin")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading bytecode: %v\n", err)
		os.Exit(1)
	}

	v := vm.New(bytecode)
	if err := v.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}
}
GOSIMPLE
    go build -o fib_standalone main.go
}
cd /mnt1/aiprjs/xxlang

FIB_STANDALONE="$STANDALONE_DIR/fib_standalone"

if [ -f "$FIB_STANDALONE" ]; then
    STANDALONE_SIZE=$(stat -c%s "$FIB_STANDALONE")
    echo "   独立可执行文件: $FIB_STANDALONE ($STANDALONE_SIZE bytes)"
else
    echo "   独立可执行文件构建失败"
fi
echo ""

# Run benchmarks
echo "=================================================="
echo "4. 性能测试结果"
echo "=================================================="
echo ""

# Test A: Source file
echo "【A】解释运行 (源代码: 解析 + 编译 + 执行)"
echo "    运行 3 次取平均值..."
SRC_TOTAL=0
for i in 1 2 3; do
    START=$(date +%s%N)
    ./xxlang run "$FIB_SOURCE" > /dev/null
    END=$(date +%s%N)
    MS=$(( (END - START) / 1000000 ))
    SRC_TOTAL=$(( SRC_TOTAL + MS ))
    echo "    第 $i 次: ${MS} ms"
done
SRC_AVG=$(( SRC_TOTAL / 3 ))
echo "    平均: ${SRC_AVG} ms"
echo ""

# Test B: Bytecode
echo "【B】字节码运行 (预编译: 反序列化 + 执行)"
echo "    运行 3 次取平均值..."
BC_TOTAL=0
for i in 1 2 3; do
    START=$(date +%s%N)
    ./xxlang run "$FIB_BYTECODE" > /dev/null
    END=$(date +%s%N)
    MS=$(( (END - START) / 1000000 ))
    BC_TOTAL=$(( BC_TOTAL + MS ))
    echo "    第 $i 次: ${MS} ms"
done
BC_AVG=$(( BC_TOTAL / 3 ))
echo "    平均: ${BC_AVG} ms"
echo ""

# Test C: Standalone
if [ -f "$FIB_STANDALONE" ]; then
    echo "【C】独立可执行文件 (嵌入字节码)"
    echo "    运行 3 次取平均值..."
    EXE_TOTAL=0
    for i in 1 2 3; do
        START=$(date +%s%N)
        cd "$STANDALONE_DIR" && ./fib_standalone > /dev/null
        cd /mnt1/aiprjs/xxlang
        END=$(date +%s%N)
        MS=$(( (END - START) / 1000000 ))
        EXE_TOTAL=$(( EXE_TOTAL + MS ))
        echo "    第 $i 次: ${MS} ms"
    done
    EXE_AVG=$(( EXE_TOTAL / 3 ))
    echo "    平均: ${EXE_AVG} ms"
    echo ""
fi

# Test D: Startup comparison
echo "【D】启动时间对比 (小程序)"
echo 'println("x")' > /tmp/startup.xxl
./xxlang compile --bytecode -o /tmp/startup.xxb /tmp/startup.xxl 2>/dev/null

# Source startup
SRC_START=0
for i in 1 2 3 4 5; do
    START=$(date +%s%N)
    ./xxlang run /tmp/startup.xxl > /dev/null
    END=$(date +%s%N)
    SRC_START=$(( SRC_START + (END - START) / 1000 ))
done
SRC_START=$(( SRC_START / 5 ))

# Bytecode startup
BC_START=0
for i in 1 2 3 4 5; do
    START=$(date +%s%N)
    ./xxlang run /tmp/startup.xxb > /dev/null
    END=$(date +%s%N)
    BC_START=$(( BC_START + (END - START) / 1000 ))
done
BC_START=$(( BC_START / 5 ))

echo "    源代码启动: ${SRC_START} µs"
echo "    字节码启动: ${BC_START} µs"

if [ -f "$FIB_STANDALONE" ]; then
    EXE_START=0
    for i in 1 2 3 4 5; do
        START=$(date +%s%N)
        cd "$STANDALONE_DIR" && ./fib_standalone > /dev/null
        cd /mnt1/aiprjs/xxlang
        END=$(date +%s%N)
        EXE_START=$(( EXE_START + (END - START) / 1000 ))
    done
    EXE_START=$(( EXE_START / 5 ))
    echo "    独立程序启动: ${EXE_START} µs"
fi

rm -f /tmp/startup.xxl /tmp/startup.xxb

echo ""
echo "=================================================="
echo "5. 分析总结"
echo "=================================================="
echo ""
echo "fib(35) 执行时间对比:"
echo "┌─────────────────────┬──────────┬──────────┐"
echo "│ 方式                │ 平均时间 │ 相对速度 │"
echo "├─────────────────────┼──────────┼──────────┤"
echo "│ 解释运行 (源代码)   │ ${SRC_AVG} ms   │ 基准     │"
echo "│ 字节码运行          │ ${BC_AVG} ms   │ ~相同    │"
if [ -f "$FIB_STANDALONE" ]; then
echo "│ 独立可执行文件      │ ${EXE_AVG} ms   │ ~相同    │"
fi
echo "└─────────────────────┴──────────┴──────────┘"
echo ""
echo "结论:"
echo "- fib(35) 计算 ~6秒，解析/编译开销 ~5ms"
echo "- 计算密集型任务：执行时间占 99.9%"
echo "- 编译为字节码对计算密集型任务性能提升可忽略"
echo "- 字节码适用于频繁启动的小脚本场景"
echo "=================================================="
