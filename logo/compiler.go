package logo

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"strings"
)

type CompileCommand func(c *Compiler)

var keywords = map[string]CompileCommand{
	"HOME":    compileHomeCmd,
	"PAPER":   compilePaperCmd,
	"INK":     compileInkCmd,
	"PEN":     compilePenCmd,
	"REPEAT":  compileRepeatCmd,
	"LOOP":    compileLoopCmd,
	"FORWARD": compileForwardCmd,
	"BACK":    compileBackCmd,
	"LEFT":    compileLeftCmd,
	"RIGHT":   compileRightCmd,
}

var colors = map[string]string{
	"BLACK":   "black",
	"WHITE":   "white",
	"RED":     "red",
	"GREEN":   "green",
	"BLUE":    "blue",
	"YELLOW":  "yellow",
	"GRAY":    "gray",
	"MAGENTA": "magenta",
}

type Compiler struct {
	Program []ProgramStep
	writer  *bufio.Writer
	PC      int
	vidx    int
	Trace   bool
}

func compileHomeCmd(c *Compiler) {
	c.trace("HOME")
	c.emit("home();")
}

func compilePaperCmd(c *Compiler) {
	c.trace("PAPER")
	c.emit("paper = '%s';", c.getColor())
}

func compileInkCmd(c *Compiler) {
	c.trace("INK")
	c.emit("ink = '%s';", c.getColor())
}

func compilePenCmd(c *Compiler) {
	c.trace("PEN")
	token := c.getParam(TkIdent)
	value := strings.ToUpper(token.String)
	if value == "UP" || value == "DOWN" {
		c.emit("pendown = %t;", value == "DOWN")
		return
	}

	c.syntaxError(fmt.Sprintf("invalid parameter in line %d", token.Line))
}

func compileForwardCmd(c *Compiler) {
	c.trace("FORWARD")
	c.emit("forward(%d);", c.getParam(TkNumber).Number)
}

func compileBackCmd(c *Compiler) {
	c.trace("BACK")
	c.emit("back(%d);", c.getParam(TkNumber).Number)
}

func compileLeftCmd(c *Compiler) {
	c.trace("LEFT")
	c.emit("left(%d);", c.getParam(TkNumber).Number)
}

func compileRightCmd(c *Compiler) {
	c.trace("RIGHT")
	c.emit("right(%d);", c.getParam(TkNumber).Number)
}

func compileRepeatCmd(c *Compiler) {
	c.trace("REPEAT")
	token := c.getParam(TkNumber)
	if token.Number <= 0 || token.Number >= 65536 {
		c.syntaxError(fmt.Sprintf("the count is too small or too large number in line %d", token.Line))
	}

	variable := c.nextVar()
	c.emit("for(let %s=0;%s<%d;++%s){", variable, variable, token.Number, variable)
}

func compileLoopCmd(c *Compiler) {
	c.trace("LOOP")
	c.emit("}")
}

func (c *Compiler) trace(msg string) {
	if c.Trace {
		log.Printf("TRACE: %s\n", msg)
	}
}

func (c *Compiler) CompilerError(err error) {
	panic(err)
}

func (c *Compiler) syntaxError(err string) {
	c.CompilerError(fmt.Errorf("syntax error: %s", err))
}

func (c *Compiler) getColor() string {
	token := c.getParam(TkIdent)
	color := strings.ToUpper(token.String)
	if value, ok := colors[color]; ok {
		return value
	}
	c.syntaxError(fmt.Sprintf("unrecognized color in line %d", token.Line))
	return "black" // Dummy color
}

func (c *Compiler) isEOP() bool {
	return c.PC == int(len(c.Program))
}

func (c *Compiler) expectToken() {
	if c.isEOP() {
		c.CompilerError(errors.New("unexpected end of program"))
	}
}

func (c *Compiler) next() ProgramStep {
	c.expectToken()
	token := c.Program[c.PC]
	c.PC += 1
	return token
}

func (c *Compiler) getParam(expected Token) ProgramStep {
	param := c.next()
	if param.Token != expected {
		c.syntaxError(fmt.Sprintf("invalid parameter type, expected %d got %d in line %d", expected, param.Token, param.Line))
	}

	return param
}

func NewCompiler(writer *bufio.Writer) *Compiler {
	return &Compiler{
		Program: []ProgramStep{},
		PC:      0,
		vidx:    0,
		Trace:   false,
		writer:  writer,
	}
}

func (c *Compiler) nextVar() string {
	c.vidx += 1
	return fmt.Sprintf("v%d", c.vidx)
}

func (c *Compiler) emit(format string, args ...any) {
	_, err := c.writer.WriteString(fmt.Sprintf(format, args...))
	if err != nil {
		c.CompilerError(err)
	}
}

func (c *Compiler) Compile(program string) error {
	l := NewLexer(program)
	// l.Debug = true
	c.Program = []ProgramStep{}

	// Parsing the source, building the program steps
	for {
		token, err := l.NextToken()
		if err != nil {
			return err
		}

		if token == TkEOF {
			break
		}

		switch token {
		case TkIdent:
			c.Program = append(c.Program, ProgramStep{Token: token, String: l.String, Line: l.Line})
		case TkNumber:
			c.Program = append(c.Program, ProgramStep{Token: token, Number: l.Number, Line: l.Line})
		case TkLiteral: // skipped
		case TkEOL: // skipped
		case TkComment: // skipped
		default:
			return fmt.Errorf("invalid token %d in line %d", token, l.Line)
		}
	}

	// Compile the program
	c.PC = 0 // reset
	for {
		p := c.next()

		if p.Token != TkIdent {
			c.syntaxError(fmt.Sprintf("unexpected token %d in line %d", p.Token, p.Line))
		}

		cmd := strings.ToUpper(p.String)

		if fn, ok := keywords[cmd]; ok {
			fn(c)
		} else {
			c.syntaxError(fmt.Sprintf("unknown keyword in line %d", p.Line))
		}

		if c.isEOP() {
			break
		}
	}

	return nil
}
