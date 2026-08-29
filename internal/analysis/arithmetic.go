package analysis

import (
	"fmt"
	"math"

	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
)

// ArithmeticOperation identifies a checked recursive aggregation operation.
type ArithmeticOperation string

const (
	ArithmeticAdd      ArithmeticOperation = "add"
	ArithmeticMultiply ArithmeticOperation = "multiply"
	ArithmeticDepth    ArithmeticOperation = "depth"
)

// OverflowError reports an invalid or overflowing non-negative int64
// operation without publishing a wrapped or clamped result.
type OverflowError struct {
	Operation ArithmeticOperation
	Left      int64
	Right     int64
}

func (err *OverflowError) Error() string {
	return fmt.Sprintf("recursive arithmetic %s overflow for %d and %d", err.Operation, err.Left, err.Right)
}

// DiagnosticCode exposes the stable arithmetic-overflow diagnostic code.
func (err *OverflowError) DiagnosticCode() diagnostic.Code {
	return diagnostic.CodeArithmeticOverflow
}

func checkedAdd(left, right int64) (int64, error) {
	if left < 0 || right < 0 || left > math.MaxInt64-right {
		return 0, &OverflowError{Operation: ArithmeticAdd, Left: left, Right: right}
	}

	return left + right, nil
}

func checkedMultiply(left, right int64) (int64, error) {
	if left < 0 || right < 0 || (left != 0 && right > math.MaxInt64/left) {
		return 0, &OverflowError{Operation: ArithmeticMultiply, Left: left, Right: right}
	}

	return left * right, nil
}

func checkedDepth(mountDepth, childDepth int64) (int64, error) {
	if mountDepth <= 0 || childDepth <= 0 {
		return 0, &OverflowError{Operation: ArithmeticDepth, Left: mountDepth, Right: childDepth}
	}

	depth, err := checkedAdd(mountDepth, childDepth-1)
	if err != nil {
		return 0, &OverflowError{Operation: ArithmeticDepth, Left: mountDepth, Right: childDepth}
	}

	return depth, nil
}
