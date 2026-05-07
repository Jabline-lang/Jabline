package compiler

import (
	"fmt"
	"jabline/pkg/ast"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

func (c *Compiler) compileMeterStatement(node *ast.MeterStatement) error {
	str, ok := node.Name.(*ast.StringLiteral)
	if !ok {
		return fmt.Errorf("meter metric name must be a string literal, got %T", node.Name)
	}

	nameIdx := c.addConstant(&object.String{Value: str.Value})

	if node.Operator == "++" {
		c.emit(code.OpMetricInc, nameIdx)
	} else {
		// We could support -- or other ops if needed
		return fmt.Errorf("unsupported meter operator: %s", node.Operator)
	}

	return nil
}

func (c *Compiler) compileTraceStatement(node *ast.TraceStatement) error {
	str, ok := node.Name.(*ast.StringLiteral)
	if !ok {
		return fmt.Errorf("trace name must be a string literal")
	}

	nameIdx := c.addConstant(&object.String{Value: str.Value})

	// Start Trace
	c.emit(code.OpTraceStart, nameIdx)

	// Compile Body
	if err := c.Compile(node.Body); err != nil {
		return err
	}

	// End Trace
	c.emit(code.OpTraceEnd)

	return nil
}
