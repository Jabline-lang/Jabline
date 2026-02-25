package compiler

import (
	"jabline/pkg/ast"
	"jabline/pkg/code"
	"jabline/pkg/object"
)

func (c *Compiler) compileIntegerLiteral(node *ast.IntegerLiteral) error {
	integer := &object.Integer{Value: node.Value}
	c.emit(code.OpConstant, c.addConstant(integer))
	return nil
}

func (c *Compiler) compileFloatLiteral(node *ast.FloatLiteral) error {
	f := &object.Float{Value: node.Value}
	c.emit(code.OpConstant, c.addConstant(f))
	return nil
}

func (c *Compiler) compileStringLiteral(node *ast.StringLiteral) error {
	str := &object.String{Value: node.Value}
	c.emit(code.OpConstant, c.addConstant(str))
	return nil
}

func (c *Compiler) compileBoolean(node *ast.Boolean) error {
	if node.Value {
		c.emit(code.OpTrue)
	} else {
		c.emit(code.OpFalse)
	}
	return nil
}

func (c *Compiler) compileNull(node *ast.Null) error {

	c.emit(code.OpNull)

	return nil

}



func (c *Compiler) compileArrayLiteral(node *ast.ArrayLiteral) error {

	for _, el := range node.Elements {

		if err := c.Compile(el); err != nil {

			return err

		}

	}

	c.emit(code.OpArray, len(node.Elements))

	return nil

}



func (c *Compiler) compileHashLiteral(node *ast.HashLiteral) error {

	for key, value := range node.Pairs {

		if err := c.Compile(key); err != nil {

			return err

		}

		if err := c.Compile(value); err != nil {

			return err

		}

	}

	c.emit(code.OpHash, len(node.Pairs)*2) // *2 because each pair has key and value

	return nil

}




